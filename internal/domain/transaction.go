package domain

import "time"

type TransactionType string

const (
	TransactionTypeTopUp   TransactionType = "top_up"
	TransactionTypeFreeze  TransactionType = "freeze"
	TransactionTypePayment TransactionType = "payment"
	TransactionTypeRefund  TransactionType = "refund"
)

type Transaction struct {
	ID        int             `json:"id"`
	UserID    int             `json:"user_id"`
	BookingID *int            `json:"booking_id,omitempty"`
	Amount    float64         `json:"amount"`
	Type      TransactionType `json:"type"`
	CreatedAt time.Time       `json:"created_at"`
}
