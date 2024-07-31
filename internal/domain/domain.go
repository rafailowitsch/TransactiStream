package domain

import "time"

type Transaction struct {
	ID        string    `json:"id"`
	UserID    string    `json:"user_id"`
	Amount    float64   `json:"amount"`
	Currency  string    `json:"currency"`
	Done      bool      `json:"done"`
	Timestamp time.Time `json:"timestamp"`
}

type Statistics struct {
	TotalTransactions     int      `json:"total_transactions"`
	FailedTransactions    int      `json:"failed_transactions"`
	TotalUsers            int      `json:"total_users"`
	AverageProcessingTime float64  `json:"average_processing_time"`
	Currencies            []string `json:"currencies"`
}
