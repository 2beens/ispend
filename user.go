package ispend

type User struct {
	Username   string      `json:"username"`
	Password   string      `json:"password"`
	Spendings  []Spending  `json:"spendings"`
	SpendKinds []SpendKind `json:"spending_kinds"`
}

func NewUser(username string, password string, spendKinds []SpendKind) User {
	return User{
		Username:   username,
		Password:   password,
		Spendings:  []Spending{},
		SpendKinds: spendKinds,
	}
}
