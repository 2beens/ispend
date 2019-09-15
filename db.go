package ispend

type SpenderDB interface {
	StoreUser(user User) error
	GetUser(username string) (*User, error)
	GetAllUsers() []User
	StoreSpending(username string, spending Spending) error
	GetSpendings(username string) ([]Spending, error)
}
