package ispend

type Spending struct {
	Currency string    `json:"currency"`
	Amount   float32   `json:"amount"`
	Kind     SpendKind `json:"kind"`
}
