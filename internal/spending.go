package internal

import "time"

type Spending struct {
	ID        string     `json:"id"`
	Currency  string     `json:"currency"`
	Amount    float32    `json:"amount"`
	Kind      *SpendKind `json:"kind"`
	Timestamp time.Time  `json:"timestamp"`
}
