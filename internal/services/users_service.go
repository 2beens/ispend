package services

import (
	"errors"
	"sync"

	"github.com/2beens/ispend/internal/db"
	"github.com/2beens/ispend/internal/metrics"
	"github.com/2beens/ispend/internal/models"
	"github.com/2beens/ispend/internal/platform"
	"github.com/dgraph-io/ristretto"
	log "github.com/sirupsen/logrus"
)

type UsersService struct {
	db        db.SpenderDB
	mutex     *sync.Mutex
	cache     *ristretto.Cache
	graphite  *metrics.GraphiteClient
	usernames []string
}

func NewUsersService(db db.SpenderDB, graphite *metrics.GraphiteClient) *UsersService {
	config := &ristretto.Config{
		NumCounters: 1e7,     // number of keys to track frequency of (10M).
		MaxCost:     1 << 30, // maximum cost of cache (1GB).
		BufferItems: 64,      // number of keys per Get buffer.
	}
	cache, err := ristretto.NewCache(config)
	if err != nil {
		log.Fatalf("cannot initialize UsersService: %s", err.Error())
	}

	cacheService := &UsersService{
		db:        db,
		mutex:     &sync.Mutex{},
		cache:     cache,
		graphite:  graphite,
		usernames: []string{},
	}

	allUsers, err := db.GetAllUsers(true)
	if err != nil {
		log.Fatalf("cannot initialize UsersService: %s", err.Error())
	}
	for _, user := range allUsers {
		cacheService.setUserSpends(user.Username, user.Spends)
		cacheService.setUserSpendKinds(user.Username, user.SpendKinds)
		cacheService.usernames = append(cacheService.usernames, user.Username)
	}

	graphite.SimpleSendInt("users.total", len(allUsers))

	log.Debugf("users service [init]: gotten %d users and saved in cache", len(allUsers))

	return cacheService
}

func (us *UsersService) GetSpendKind(username string, spendingKindID int) (*models.SpendKind, error) {
	return us.db.GetSpendKind(username, spendingKindID)
}

func (us *UsersService) GetAllDefaultSpendKinds() ([]models.SpendKind, error) {
	return us.db.GetAllDefaultSpendKinds()
}

func (us *UsersService) GetAllUsers() (models.Users, error) {
	//us.mutex.Lock()
	//defer us.mutex.Unlock()
	var users models.Users
	for _, username := range us.usernames {
		user, err := us.GetUser(username)
		if err != nil {
			log.Errorf("user service error [get all users]: %s", err.Error())
			continue
		}
		users = append(users, user)
	}
	return users, nil
}

func (us *UsersService) AddUser(user *models.User) error {
	if user == nil {
		return errors.New("user is nil, cannot add")
	}
	_, err := us.db.StoreUser(user)
	if err != nil {
		return err
	}

	for i := range user.SpendKinds {
		spendKindID, err := us.db.StoreSpendKind(user.Username, &user.SpendKinds[i])
		if err != nil {
			log.Errorf("users service add user - add spend kind error: %s", err.Error())
			continue
		}
		user.SpendKinds[i].ID = spendKindID
	}

	// TODO: solve multithreaded issues
	//us.mutex.Lock()
	//defer us.mutex.Unlock()
	us.setUserSpends(user.Username, user.Spends)
	us.setUserSpendKinds(user.Username, user.SpendKinds)
	us.usernames = append(us.usernames, user.Username)

	return nil
}

func (us *UsersService) GetUser(username string) (*models.User, error) {
	// TODO:
	//us.mutex.Lock()
	//defer us.mutex.Unlock()

	if !us.UserExists(username) {
		return nil, platform.ErrNotFound
	}

	user, err := us.db.GetUser(username, false)
	if err != nil {
		return nil, err
	}

	spends, found := us.getUserSpends(username)
	if spends == nil || !found {
		log.Tracef("users service [get user: %s], spends cache miss. will recreate", username)
		spends, err = us.db.GetSpends(username)
		if err != nil {
			return nil, err
		}
		us.setUserSpends(user.Username, spends)
	}
	user.Spends = spends

	spendKinds, found := us.getUserSpendKinds(username)
	if spendKinds == nil || !found {
		log.Tracef("users service [get user: %s], spend kinds cache miss. will recreate", username)
		spendKinds, err = us.db.GetSpendKinds(username)
		if err != nil {
			return nil, err
		}
		us.setUserSpendKinds(user.Username, spendKinds)
	}
	user.SpendKinds = spendKinds

	return user, nil
}

func (us *UsersService) UserExists(username string) bool {
	// TODO:
	//us.mutex.Lock()
	//defer us.mutex.Unlock()
	for _, u := range us.usernames {
		if u == username {
			return true
		}
	}
	return false
}

func (us *UsersService) StoreSpending(user *models.User, spending models.Spending) error {
	//us.mutex.Lock()
	//defer us.mutex.Unlock()
	id, err := us.db.StoreSpending(user.Username, spending)
	if err != nil {
		return err
	}

	spending.ID = id
	user.Spends = append(user.Spends, spending)

	//spends, _ := us.getUserSpends(user.Username)
	//us.cache.Del(user.Username)
	//us.setUserSpends(user.Username, append(spends, spending))

	us.cache.Del(user.Username)
	us.setUserSpends(user.Username, user.Spends)

	return nil
}

func (us *UsersService) DeleteSpending(username, spendID string) error {
	//us.mutex.Lock()
	//defer us.mutex.Unlock()
	err := us.db.DeleteSpending(username, spendID)
	if err != nil {
		return err
	}

	spends, found := us.getUserSpends(username)
	if !found {
		// TODO: create or something

	}

	indexToRemove := -1
	for i := range spends {
		if spends[i].ID == spendID {
			indexToRemove = i
			break
		}
	}

	if indexToRemove < 0 {
		return platform.ErrNotFound
	}

	// remove spending by its index
	spends = append(spends[:indexToRemove], spends[indexToRemove+1:]...)

	us.cache.Del(username)
	us.setUserSpends(username, spends)

	return nil
}

func (us *UsersService) getUserSpends(username string) ([]models.Spending, bool) {
	//us.mutex.Lock()
	//defer us.mutex.Unlock()
	if spends, found := us.cache.Get(username); found {
		spendsSlice := spends.([]models.Spending)
		log.Tracef("found %d spends for user %s", len(spendsSlice), username)
		return spendsSlice, found
	}
	return nil, false
}

func (us *UsersService) getUserSpendKinds(username string) ([]models.SpendKind, bool) {
	//us.mutex.Lock()
	//defer us.mutex.Unlock()
	if spendKinds, found := us.cache.Get(username + "|sk"); found {
		spendKindsSlice := spendKinds.([]models.SpendKind)
		log.Tracef("found %d spend kinds for user %s", len(spendKindsSlice), username)
		return spendKindsSlice, found
	}
	return nil, false
}

func (us *UsersService) setUserSpends(username string, spends []models.Spending) bool {
	//us.mutex.Lock()
	//defer us.mutex.Unlock()
	log.Tracef("user service cache: storing %d spends for user %s", len(spends), username)
	return us.cache.Set(username, spends, 1)
}

func (us *UsersService) setUserSpendKinds(username string, spendKinds []models.SpendKind) bool {
	//us.mutex.Lock()
	//defer us.mutex.Unlock()
	log.Tracef("user service cache: storing %d spend kinds for user %s", len(spendKinds), username)
	return us.cache.Set(username+"|sk", spendKinds, 1)
}
