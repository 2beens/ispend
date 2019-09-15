package ispend

type User struct {
	Username  string     `json:"username"`
	Spendings []Spending `json:"spendings"`
}

func NewUser(username string) User {
	return User{
		Username:  username,
		Spendings: []Spending{},
	}
}
