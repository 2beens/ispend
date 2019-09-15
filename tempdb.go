package ispend

import "errors"

type TempDB struct {
	Users []User
}

func NewTempDB() *TempDB {
	return &TempDB{
		Users: []User{},
	}
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
	return nil, errors.New("cannot find user with username: " + username)
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
