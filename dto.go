package ispend

type UserDTO struct {
	Email      string         `json:"email"`
	Username   string         `json:"username"`
	SpendKinds []SpendKindDTO `json:"spending_kinds"`
}

type SpendKindDTO struct {
	Name string `json:"name"`
}

func NewUserDTO(user *User) UserDTO {
	var spendKinds []SpendKindDTO
	for _, sk := range user.SpendKinds {
		spendKinds = append(spendKinds, NewSpendKindDTO(sk))
	}
	return UserDTO{
		Email:      user.Email,
		Username:   user.Username,
		SpendKinds: spendKinds,
	}
}

func NewSpendKindDTO(spendKind SpendKind) SpendKindDTO {
	return SpendKindDTO{Name: spendKind.Name}
}
