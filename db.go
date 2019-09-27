package ispend

type SpenderDB interface {
	Open() error
	Close() error

	StoreDefaultSpendKind(kind SpendKind) (int, error)
	GetDefaultSpendKind(name string) (*SpendKind, error)
	GetAllDefaultSpendKinds() ([]SpendKind, error)
	GetSpendKind(username string, spendingKindID int) (*SpendKind, error)
	GetSpendKinds(username string) ([]SpendKind, error)
	StoreSpendKind(username string, kind *SpendKind) (int, error)
	StoreUser(user *User) (int, error)
	GetUser(username string) (*User, error)
	GetAllUsers() (Users, error)
	StoreSpending(username string, spending Spending) error
	GetSpends(username string) ([]Spending, error)
}
