package db

import (
	"log"

	"github.com/2beens/ispend/internal/models"
	"github.com/2beens/ispend/internal/platform"
)

type InMemoryDB struct {
	DefaultSpendKinds []models.SpendKind
	Users             models.Users
}

func NewInMemoryDB() *InMemoryDB {
	inMemDB := &InMemoryDB{
		DefaultSpendKinds: []models.SpendKind{},
		Users:             models.Users{},
	}

	inMemDB.prepareDebuggingData()

	return inMemDB
}

func (db *InMemoryDB) Open() error {
	return nil
}

func (db *InMemoryDB) Close() error {
	return nil
}

func (db *InMemoryDB) StoreDefaultSpendKind(kind models.SpendKind) (int, error) {
	db.DefaultSpendKinds = append(db.DefaultSpendKinds, kind)
	return kind.ID, nil
}

func (db *InMemoryDB) GetAllDefaultSpendKinds() ([]models.SpendKind, error) {
	return db.DefaultSpendKinds, nil
}

func (db *InMemoryDB) GetSpendKind(username string, spendingKindID int) (*models.SpendKind, error) {
	user, err := db.GetUser(username, true)
	if err != nil {
		return nil, err
	}

	for _, sk := range user.SpendKinds {
		if sk.ID == spendingKindID {
			return &sk, nil
		}
	}

	return nil, platform.ErrNotFound
}

func (db *InMemoryDB) GetSpendKinds(username string) ([]models.SpendKind, error) {
	user, err := db.GetUser(username, true)
	if err != nil {
		return nil, err
	}
	return user.SpendKinds, nil
}

func (db *InMemoryDB) StoreSpendKind(username string, kind *models.SpendKind) (int, error) {
	return -1, nil
}

func (db *InMemoryDB) StoreUser(user *models.User) (int, error) {
	db.Users = append(db.Users, user)
	return 0, nil
}

func (db *InMemoryDB) GetUser(username string, loadAllData bool) (*models.User, error) {
	for i := range db.Users {
		if db.Users[i].Username == username {
			return db.Users[i], nil
		}
	}
	return nil, platform.ErrNotFound
}

func (db *InMemoryDB) GetAllUsers(loadAllUserData bool) (models.Users, error) {
	return db.Users, nil
}

func (db *InMemoryDB) StoreSpending(username string, spending models.Spending) (string, error) {
	user, err := db.GetUser(username, true)
	if err != nil {
		return "", err
	}

	user.Spends = append(user.Spends, spending)
	return platform.GenerateRandomString(10), nil
}

func (db *InMemoryDB) GetSpends(username string) ([]models.Spending, error) {
	user, err := db.GetUser(username, true)
	if err != nil {
		return nil, err
	}
	return user.Spends, nil
}

func (db *InMemoryDB) DeleteSpending(username, spendID string) error {
	user, err := db.GetUser(username, true)
	if err != nil {
		return err
	}

	indexToRemove := -1
	for i := range user.Spends {
		if user.Spends[i].ID == spendID {
			indexToRemove = i
			break
		}
	}

	if indexToRemove < 0 {
		return platform.ErrNotFound
	}

	// remove spending by its index
	user.Spends = append(user.Spends[:indexToRemove], user.Spends[indexToRemove+1:]...)

	return nil
}

func (db *InMemoryDB) prepareDebuggingData() *InMemoryDB {
	skNightlife := models.SpendKind{ID: 1, Name: "nightlife"}
	skTravel := models.SpendKind{ID: 2, Name: "travel"}
	skFood := models.SpendKind{ID: 3, Name: "food"}
	skRent := models.SpendKind{ID: 4, Name: "rent"}
	defSpendKinds := []models.SpendKind{skNightlife, skTravel, skFood, skRent}

	adminUser := models.NewUser("admin@serjspends.de", "admin", "admin1", defSpendKinds)
	adminUser.Spends = append(adminUser.Spends, models.Spending{
		ID:       "sp1",
		Amount:   100,
		Currency: "RSD",
		Kind:     &skNightlife,
	})
	adminUser.Spends = append(adminUser.Spends, models.Spending{
		ID:       "sp2",
		Amount:   2300,
		Currency: "RSD",
		Kind:     &skTravel,
	})
	lazarUser := models.NewUser("lazar@serjspends.de", "lazar", "lazar1", defSpendKinds)
	lazarUser.Spends = append(lazarUser.Spends, models.Spending{
		ID:       "sp3",
		Amount:   89.99,
		Currency: "USD",
		Kind:     &skTravel,
	})

	_, err := db.StoreUser(adminUser)
	if err != nil {
		log.Panic(err.Error())
	}
	_, err = db.StoreUser(lazarUser)
	if err != nil {
		log.Panic(err.Error())
	}

	_, err = db.StoreDefaultSpendKind(skNightlife)
	if err != nil {
		log.Panic(err.Error())
	}
	_, err = db.StoreDefaultSpendKind(skFood)
	if err != nil {
		log.Panic(err.Error())
	}
	_, err = db.StoreDefaultSpendKind(skRent)
	if err != nil {
		log.Panic(err.Error())
	}
	_, err = db.StoreDefaultSpendKind(skTravel)
	if err != nil {
		log.Panic(err.Error())
	}

	return db
}
