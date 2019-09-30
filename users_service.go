package ispend

import (
	"errors"
	"sync"

	"github.com/dgraph-io/ristretto"
	log "github.com/sirupsen/logrus"
)

type UsersService struct {
	db        SpenderDB
	mutex     *sync.Mutex
	cache     *ristretto.Cache
	usernames []string
}

func NewUsersService(db SpenderDB) *UsersService {
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

	log.Debugf("users service [init]: gotten %d users and saved in cache", len(allUsers))

	return cacheService
}

func (us *UsersService) GetSpendKind(username string, spendingKindID int) (*SpendKind, error) {
	return us.db.GetSpendKind(username, spendingKindID)
}

func (us *UsersService) GetAllDefaultSpendKinds() ([]SpendKind, error) {
	return us.db.GetAllDefaultSpendKinds()
}

func (us *UsersService) GetAllUsers() (Users, error) {
	var users Users
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

func (us *UsersService) AddUser(user *User) error {
	if user == nil {
		return errors.New("user is nil, cannot add")
	}
	_, err := us.db.StoreUser(user)
	if err != nil {
		return err
	}

	for i, _ := range user.SpendKinds {
		spendKindID, err := us.db.StoreSpendKind(user.Username, &user.SpendKinds[i])
		if err != nil {
			log.Errorf("users service add user - add spend kind error: %s", err.Error())
			continue
		}
		user.SpendKinds[i].ID = spendKindID
	}

	us.setUserSpends(user.Username, user.Spends)
	us.setUserSpendKinds(user.Username, user.SpendKinds)
	us.usernames = append(us.usernames, user.Username)

	return nil
}

func (us *UsersService) GetUser(username string) (*User, error) {
	if !us.UserExists(username) {
		return nil, ErrNotFound
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
	us.mutex.Lock()
	defer us.mutex.Unlock()
	for _, u := range us.usernames {
		if u == username {
			return true
		}
	}
	return false
}

func (us *UsersService) StoreSpending(username string, spending Spending) error {
	var spends []Spending
	spends, _ = us.getUserSpends(username)
	us.cache.Del(username)
	us.setUserSpends(username, append(spends, spending))
	return us.db.StoreSpending(username, spending)
}

func (us *UsersService) getUserSpends(username string) ([]Spending, bool) {
	us.mutex.Lock()
	defer us.mutex.Unlock()
	if spends, found := us.cache.Get(username); found {
		spendsSlice := spends.([]Spending)
		log.Tracef("found %d spends for user %s", len(spendsSlice), username)
		return spendsSlice, found
	}
	return nil, false
}

func (us *UsersService) getUserSpendKinds(username string) ([]SpendKind, bool) {
	us.mutex.Lock()
	defer us.mutex.Unlock()
	if spendKinds, found := us.cache.Get(username + "|sk"); found {
		spendKindsSlice := spendKinds.([]SpendKind)
		log.Tracef("found %d spend kinds for user %s", len(spendKindsSlice), username)
		return spendKindsSlice, found
	}
	return nil, false
}

func (us *UsersService) setUserSpends(username string, spends []Spending) bool {
	us.mutex.Lock()
	defer us.mutex.Unlock()
	log.Tracef("user service cache: storing %d spends for user %s", len(spends), username)
	return us.cache.Set(username, spends, 1)
}

func (us *UsersService) setUserSpendKinds(username string, spendKinds []SpendKind) bool {
	us.mutex.Lock()
	defer us.mutex.Unlock()
	log.Tracef("user service cache: storing %d spend kinds for user %s", len(spendKinds), username)
	return us.cache.Set(username+"|sk", spendKinds, 1)
}
