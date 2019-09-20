package ispend

import "log"

type TempDB struct {
	DefaultSpendKinds []SpendKind
	Users             []User
}

func NewTempDB() *TempDB {
	tempDB := &TempDB{
		DefaultSpendKinds: []SpendKind{},
		Users:             []User{},
	}

	tempDB.prepareDebuggingData()

	return tempDB
}

func (db *TempDB) StoreDefaultSpendKind(kind SpendKind) error {
	db.DefaultSpendKinds = append(db.DefaultSpendKinds, kind)
	return nil
}

func (db *TempDB) GetDefaultSpendKind(name string) (*SpendKind, error) {
	for _, k := range db.DefaultSpendKinds {
		if k.Name == name {
			return &k, nil
		}
	}
	return nil, ErrNotFound
}

func (db *TempDB) GetAllDefaultSpendKinds() ([]SpendKind, error) {
	return db.DefaultSpendKinds, nil
}

func (db *TempDB) GetSpendKinds(username string) ([]SpendKind, error) {
	return nil, nil
}

func (db *TempDB) StoreUser(user User) error {
	db.Users = append(db.Users, user)
	return nil
}

func (db *TempDB) GetUser(username string) (*User, error) {
	for _, u := range db.Users {
		if u.Username == username {
			return &u, nil
		}
	}
	return nil, ErrNotFound
}

func (db *TempDB) GetAllUsers() []User {
	return db.Users
}

func (db *TempDB) StoreSpending(username string, spending Spending) error {
	user, err := db.GetUser(username)
	if err != nil {
		return err
	}

	user.Spendings = append(user.Spendings, spending)
	return nil
}

func (db *TempDB) GetSpendings(username string) ([]Spending, error) {
	user, err := db.GetUser(username)
	if err != nil {
		return nil, err
	}
	return user.Spendings, nil
}

func (db *TempDB) prepareDebuggingData() *TempDB {
	skNightlife := SpendKind{"nightlife"}
	skTravel := SpendKind{"travel"}
	skFood := SpendKind{"food"}
	skRent := SpendKind{"rent"}
	defSpendKinds := []SpendKind{skNightlife, skTravel, skFood, skRent}

	adminUser := NewUser("admin", defSpendKinds)
	adminUser.Spendings = append(adminUser.Spendings, Spending{
		Amount:   100,
		Currency: "RSD",
		Kind:     skNightlife,
	})
	adminUser.Spendings = append(adminUser.Spendings, Spending{
		Amount:   2300,
		Currency: "RSD",
		Kind:     skTravel,
	})
	lazarUser := NewUser("lazar", defSpendKinds)
	lazarUser.Spendings = append(lazarUser.Spendings, Spending{
		Amount:   89.99,
		Currency: "USD",
		Kind:     skTravel,
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
