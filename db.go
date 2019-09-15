package ispend

type SpenderDB interface {
	StoreSpendKind(kind SpendKind) error
	GetSpendKind(name string) (*SpendKind, error)
	GetAllSpendKinds() ([]SpendKind, error)
	StoreUser(user User) error
	GetUser(username string) (*User, error)
	GetAllUsers() []User
	StoreSpending(username string, spending Spending) error
	GetSpendings(username string) ([]Spending, error)
}
