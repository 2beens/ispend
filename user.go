package ispend

type User struct {
	Email      string      `json:"email"`
	Username   string      `json:"username"`
	Password   string      `json:"password"`
	Spendings  []Spending  `json:"spendings"`
	SpendKinds []SpendKind `json:"spending_kinds"`
}

func NewUser(email string, username string, password string, spendKinds []SpendKind) User {
	return User{
		Email:      email,
		Username:   username,
		Password:   password,
		Spendings:  []Spending{},
		SpendKinds: spendKinds,
	}
}
