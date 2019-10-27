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

	usersService := &UsersService{
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
		usersService.setUserSpendsCache(user.Username, user.Spends)
		usersService.setUserSpendKindsCache(user.Username, user.SpendKinds)
		usersService.usernames = append(usersService.usernames, user.Username)
	}

	graphite.SimpleSendInt("users.total", len(allUsers))

	log.Debugf("users service [init]: gotten %d users and saved in cache", len(allUsers))

	return usersService
}

func (us *UsersService) GetSpendKind(username string, spendingKindID int) (*models.SpendKind, error) {
	return us.db.GetSpendKind(username, spendingKindID)
}

func (us *UsersService) GetAllDefaultSpendKinds() ([]models.SpendKind, error) {
	return us.db.GetAllDefaultSpendKinds()
}

func (us *UsersService) GetAllUsers() (models.Users, error) {
	var users models.Users
	for _, username := range us.getCachedUsernamesSynced() {
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

	us.setUserSpendsCache(user.Username, user.Spends)
	us.setUserSpendKindsCache(user.Username, user.SpendKinds)

	us.mutex.Lock()
	us.usernames = append(us.usernames, user.Username)
	us.mutex.Unlock()

	return nil
}

func (us *UsersService) GetUser(username string) (*models.User, error) {
	if !us.UserExists(username) {
		return nil, platform.ErrNotFound
	}

	user, err := us.db.GetUser(username, false)
	if err != nil {
		return nil, err
	}

	spends, found := us.getUserSpendsCache(username)
	if spends == nil || !found {
		log.Tracef("users service [get user: %s], spends cache miss. will recreate", username)
		spends, err = us.db.GetSpends(username)
		if err != nil {
			return nil, err
		}
		us.setUserSpendsCache(user.Username, user.Spends)
	}
	user.Spends = spends

	spendKinds, found := us.getUserSpendKindsCache(username)
	if spendKinds == nil || !found {
		log.Tracef("users service [get user: %s], spend kinds cache miss. will recreate", username)
		spendKinds, err = us.db.GetSpendKinds(username)
		if err != nil {
			return nil, err
		}
		us.setUserSpendKindsCache(user.Username, user.SpendKinds)
	}
	user.SpendKinds = spendKinds

	return user, nil
}

func (us *UsersService) UserExists(username string) bool {
	for _, u := range us.getCachedUsernamesSynced() {
		if u == username {
			return true
		}
	}
	return false
}

func (us *UsersService) StoreSpending(user *models.User, spending models.Spending) error {
	id, err := us.db.StoreSpending(user.Username, spending)
	if err != nil {
		return err
	}

	spending.ID = id
	user.Spends = append(user.Spends, spending)
	us.setUserSpendsCache(user.Username, user.Spends)

	return nil
}

func (us *UsersService) DeleteSpending(username, spendID string) error {
	err := us.db.DeleteSpending(username, spendID)
	if err != nil {
		return err
	}

	// delete from cache too
	var spends []models.Spending
	if spendsFromCache, found := us.getUserSpendsCache(username); !found {
		log.Errorf("delete spending from cache error [not found for user: %s]! indicator of bug - db and cache not in sync", username)
		return nil
	} else {
		spends = spendsFromCache
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

	us.setUserSpendsCache(username, spends)

	return nil
}

func (us *UsersService) getUserSpendsCache(username string) ([]models.Spending, bool) {
	us.mutex.Lock()
	defer us.mutex.Unlock()
	if spends, found := us.cache.Get(username); found {
		spendsSlice := spends.([]models.Spending)
		log.Tracef("found %d spends for user %s", len(spendsSlice), username)
		return spendsSlice, found
	}
	return nil, false
}

func (us *UsersService) getUserSpendKindsCache(username string) ([]models.SpendKind, bool) {
	us.mutex.Lock()
	defer us.mutex.Unlock()
	if spendKinds, found := us.cache.Get(username + "|sk"); found {
		spendKindsSlice := spendKinds.([]models.SpendKind)
		log.Tracef("found %d spend kinds for user %s", len(spendKindsSlice), username)
		return spendKindsSlice, found
	}
	return nil, false
}

func (us *UsersService) setUserSpendsCache(username string, spends []models.Spending) bool {
	us.mutex.Lock()
	defer us.mutex.Unlock()
	stored := us.cache.Set(username, spends, 1)
	log.Tracef("user service cache: storing %d spends for user [%s], stored: %t", len(spends), username, stored)
	return stored
}

func (us *UsersService) setUserSpendKindsCache(username string, spendKinds []models.SpendKind) bool {
	us.mutex.Lock()
	defer us.mutex.Unlock()
	stored := us.cache.Set(username+"|sk", spendKinds, 1)
	log.Tracef("user service cache: storing %d spend kinds for user [%s], stored: %t", len(spendKinds), username, stored)
	return stored
}

func (us *UsersService) getCachedUsernamesSynced() []string {
	us.mutex.Lock()
	defer us.mutex.Unlock()
	return us.usernames
}
