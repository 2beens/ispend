package ispend

type User struct {
	Username   string      `json:"username"`
	Spendings  []Spending  `json:"spendings"`
	SpendKinds []SpendKind `json:"spending_kinds"`
}

func NewUser(username string, spendKinds []SpendKind) User {
	return User{
		Username:   username,
		Spendings:  []Spending{},
		SpendKinds: spendKinds,
	}
}
