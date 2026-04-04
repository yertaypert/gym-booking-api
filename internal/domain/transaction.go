package domain

type TransactionType string

const (
	TransactionTypeDeposit TransactionType = "deposit"
	TransactionTypeBooking TransactionType = "booking"
	TransactionTypeRefund  TransactionType = "refund"
)
