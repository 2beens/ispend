package ispend

type SpenderDB interface {
	StoreDefaultSpendKind(kind SpendKind) (int, error)
	GetDefaultSpendKind(name string) (*SpendKind, error)
	GetAllDefaultSpendKinds() ([]SpendKind, error)
	GetSpendKind(username string, spendingKindID int) (*SpendKind, error)
	GetSpendKinds(username string) ([]SpendKind, error)
	StoreUser(user *User) error
	GetUser(username string) (*User, error)
	GetAllUsers() (Users, error)
	StoreSpending(username string, spending Spending) error
	GetSpends(username string) ([]Spending, error)
}
