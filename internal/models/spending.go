package models

import (
	"fmt"
	"time"
)

type Spending struct {
	ID        string     `json:"id"`
	Currency  string     `json:"currency"`
	Amount    float32    `json:"amount"`
	Kind      *SpendKind `json:"kind"`
	Timestamp time.Time  `json:"timestamp"`
}

func (s *Spending) String() string {
	return fmt.Sprintf("Spend ID[%s] %f[%s] %s %v", s.ID, s.Amount, s.Currency, s.Kind.Name, s.Timestamp)
}
