package ispend

type SpenderDB interface {
	Open() error
	Close() error

	StoreDefaultSpendKind(kind SpendKind) (int, error)
	GetAllDefaultSpendKinds() ([]SpendKind, error)
	GetSpendKind(username string, spendingKindID int) (*SpendKind, error)
	GetSpendKinds(username string) ([]SpendKind, error)
	StoreSpendKind(username string, kind *SpendKind) (int, error)

	StoreUser(user *User) (int, error)
	GetUser(username string, loadAllData bool) (*User, error)
	GetAllUsers(loadAllUserData bool) (Users, error)

	StoreSpending(username string, spending Spending) (string, error)
	GetSpends(username string) ([]Spending, error)
}
