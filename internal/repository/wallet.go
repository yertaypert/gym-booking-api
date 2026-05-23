package repository

import (
	"context"
	"database/sql"
)

type WalletRepository struct {
	db *sql.DB
}

func NewWalletRepository(db *sql.DB) *WalletRepository {
	return &WalletRepository{db: db}
}

func (r *WalletRepository) UpdateBalance(ctx context.Context, tx *sql.Tx, userID int, amount float64) error {
	query := `UPDATE users SET balance = balance + $1 WHERE id = $2`
	result, err := tx.ExecContext(ctx, query, amount, userID)
	if err != nil {
		return err
	}
	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *WalletRepository) CreateTransaction(ctx context.Context, tx *sql.Tx, userID int, bookingID *int, amount float64, txType string) error {
	query := `INSERT INTO transactions (user_id, booking_id, amount, type) VALUES ($1, $2, $3, $4)`
	_, err := tx.ExecContext(ctx, query, userID, bookingID, amount, txType)
	return err
}
