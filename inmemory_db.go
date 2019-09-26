package ispend

import "log"

type InMemoryDB struct {
	DefaultSpendKinds []SpendKind
	Users             Users
}

func NewInMemoryDB() *InMemoryDB {
	inMemDB := &InMemoryDB{
		DefaultSpendKinds: []SpendKind{},
		Users:             Users{},
	}

	inMemDB.prepareDebuggingData()

	return inMemDB
}

func (db *InMemoryDB) StoreDefaultSpendKind(kind SpendKind) error {
	db.DefaultSpendKinds = append(db.DefaultSpendKinds, kind)
	return nil
}

func (db *InMemoryDB) GetDefaultSpendKind(name string) (*SpendKind, error) {
	for _, k := range db.DefaultSpendKinds {
		if k.Name == name {
			return &k, nil
		}
	}
	return nil, ErrNotFound
}

func (db *InMemoryDB) GetAllDefaultSpendKinds() ([]SpendKind, error) {
	return db.DefaultSpendKinds, nil
}

func (db *InMemoryDB) GetSpendKind(username string, spendingKindID int) (*SpendKind, error) {
	user, err := db.GetUser(username)
	if err != nil {
		return nil, err
	}

	for _, sk := range user.SpendKinds {
		if sk.ID == spendingKindID {
			return &sk, nil
		}
	}

	return nil, ErrNotFound
}

func (db *InMemoryDB) GetSpendKinds(username string) ([]SpendKind, error) {
	user, err := db.GetUser(username)
	if err != nil {
		return nil, err
	}
	return user.SpendKinds, nil
}

func (db *InMemoryDB) StoreUser(user *User) error {
	db.Users = append(db.Users, user)
	return nil
}

func (db *InMemoryDB) GetUser(username string) (*User, error) {
	for i, _ := range db.Users {
		if db.Users[i].Username == username {
			return db.Users[i], nil
		}
	}
	return nil, ErrNotFound
}

func (db *InMemoryDB) GetAllUsers() (Users, error) {
	return db.Users, nil
}

func (db *InMemoryDB) StoreSpending(username string, spending Spending) error {
	user, err := db.GetUser(username)
	if err != nil {
		return err
	}

	user.Spends = append(user.Spends, spending)
	return nil
}

func (db *InMemoryDB) GetSpendings(username string) ([]Spending, error) {
	user, err := db.GetUser(username)
	if err != nil {
		return nil, err
	}
	return user.Spends, nil
}

func (db *InMemoryDB) prepareDebuggingData() *InMemoryDB {
	skNightlife := SpendKind{1, "nightlife"}
	skTravel := SpendKind{2, "travel"}
	skFood := SpendKind{3, "food"}
	skRent := SpendKind{4, "rent"}
	defSpendKinds := []SpendKind{skNightlife, skTravel, skFood, skRent}

	adminUser := NewUser("admin@serjspends.de", "admin", "admin1", defSpendKinds)
	adminUser.Spends = append(adminUser.Spends, Spending{
		ID:       "sp1",
		Amount:   100,
		Currency: "RSD",
		Kind:     &skNightlife,
	})
	adminUser.Spends = append(adminUser.Spends, Spending{
		ID:       "sp2",
		Amount:   2300,
		Currency: "RSD",
		Kind:     &skTravel,
	})
	lazarUser := NewUser("lazar@serjspends.de", "lazar", "lazar1", defSpendKinds)
	lazarUser.Spends = append(lazarUser.Spends, Spending{
		ID:       "sp3",
		Amount:   89.99,
		Currency: "USD",
		Kind:     &skTravel,
	})

	err := db.StoreUser(adminUser)
	if err != nil {
		log.Panic(err.Error())
	}
	err = db.StoreUser(lazarUser)
	if err != nil {
		log.Panic(err.Error())
	}

	err = db.StoreDefaultSpendKind(skNightlife)
	if err != nil {
		log.Panic(err.Error())
	}
	err = db.StoreDefaultSpendKind(skFood)
	if err != nil {
		log.Panic(err.Error())
	}
	err = db.StoreDefaultSpendKind(skRent)
	if err != nil {
		log.Panic(err.Error())
	}
	err = db.StoreDefaultSpendKind(skTravel)
	if err != nil {
		log.Panic(err.Error())
	}

	return db
}
