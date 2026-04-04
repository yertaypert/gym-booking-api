package domain

type TransactionType string

const (
	TransactionTypeTopUp   TransactionType = "top_up"
	TransactionTypeFreeze  TransactionType = "freeze"
	TransactionTypePayment TransactionType = "payment"
	TransactionTypeRefund  TransactionType = "refund"
)
