package ispend

type TempDB struct {
	DefaultSpendKinds []SpendKind
	Users             []User
}

func NewTempDB() *TempDB {
	return &TempDB{
		DefaultSpendKinds: []SpendKind{},
		Users:             []User{},
	}
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
