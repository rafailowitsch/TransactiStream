package domain

import "time"

type Transaction struct {
	ID        string
	UserID    string
	Amount    float64
	Currency  string
	Timestamp time.Time
}
