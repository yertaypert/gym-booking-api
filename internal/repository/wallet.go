package repository

import (
	"database/sql"
)

type WalletRepository struct {
	db *sql.DB
}

func NewWalletRepository(db *sql.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

// UpdateBalance изменяет баланс пользователя внутри транзакции
func (r *WalletRepository) UpdateBalance(tx *sql.Tx, userID int, amount float64) error {
	query := `UPDATE users SET balance = balance + $1 WHERE id = $2`
	_, err := tx.Exec(query, amount, userID)
	return err
}

// CreateTransaction записывает историю операции
func (r *WalletRepository) CreateTransaction(tx *sql.Tx, userID int, bookingID int, amount float64, txType string) error {
	query := `INSERT INTO transactions (user_id, booking_id, amount, type) VALUES ($1, $2, $3, $4)`
	_, err := tx.Exec(query, userID, bookingID, amount, txType)
	return err
}
