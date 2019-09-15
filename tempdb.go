package ispend

type TempDB struct {
	SpendKinds []SpendKind
	Users      []User
}

func NewTempDB() *TempDB {
	return &TempDB{
		SpendKinds: []SpendKind{},
		Users:      []User{},
	}
}

func (db *TempDB) StoreSpendKind(kind SpendKind) error {
	db.SpendKinds = append(db.SpendKinds, kind)
	return nil
}

func (db *TempDB) GetSpendKind(name string) (*SpendKind, error) {
	for _, k := range db.SpendKinds {
		if k.Name == name {
			return &k, nil
		}
	}
	return nil, ErrNotFound
}

func (db *TempDB) GetAllSpendKinds() ([]SpendKind, error) {
	return db.SpendKinds, nil
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
