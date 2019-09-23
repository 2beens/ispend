package ispend

import "time"

type UserDTO struct {
	Email      string         `json:"email"`
	Username   string         `json:"username"`
	Spendings  []SpendingDTO  `json:"spendings"`
	SpendKinds []SpendKindDTO `json:"spending_kinds"`
}

type SpendKindDTO struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

type SpendingDTO struct {
	ID        string       `json:"id"`
	Currency  string       `json:"currency"`
	Amount    float32      `json:"amount"`
	Kind      SpendKindDTO `json:"kind"`
	Timestamp time.Time    `json:"timestamp"`
}

func NewUserDTO(user *User) UserDTO {
	var spendKinds []SpendKindDTO
	for _, sk := range user.SpendKinds {
		spendKinds = append(spendKinds, NewSpendKindDTO(sk))
	}
	var spendings []SpendingDTO
	for _, s := range user.Spendings {
		spendings = append(spendings, NewSpendingDTO(s))
	}
	return UserDTO{
		Email:      user.Email,
		Username:   user.Username,
		SpendKinds: spendKinds,
		Spendings:  spendings,
	}
}

func NewSpendKindDTO(spendKind SpendKind) SpendKindDTO {
	return SpendKindDTO{
		ID:   spendKind.ID,
		Name: spendKind.Name,
	}
}

func NewSpendingDTO(spending Spending) SpendingDTO {
	return SpendingDTO{
		ID:        spending.ID,
		Currency:  spending.Currency,
		Amount:    spending.Amount,
		Kind:      NewSpendKindDTO(spending.Kind),
		Timestamp: spending.Timestamp,
	}
}
