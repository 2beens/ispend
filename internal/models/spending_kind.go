package models

import "fmt"

type SpendKind struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

func (sk *SpendKind) String() string {
	return fmt.Sprintf("%s [%d]", sk.Name, sk.ID)
}
