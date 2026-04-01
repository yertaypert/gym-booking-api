package repository

import (
	"context"
	"database/sql"

	"github.com/yertaypert/gym-booking-api/internal/domain"
)

type walletRepository struct {
	db *sql.DB
}

func NewWalletRepository(db *sql.DB) WalletRepository {
	return &walletRepository{db: db}
}

func (r *walletRepository) UpdateBalance(ctx context.Context, tx *sql.Tx, userID int, amount float64) error {
	query := `UPDATE users SET balance = balance + $1 WHERE id = $2`
	_, err := tx.ExecContext(ctx, query, amount, userID)
	return err
}

func (r *walletRepository) CreateTransaction(ctx context.Context, tx *sql.Tx, userID int, bookingID *int, amount float64, txType domain.TransactionType) error {
	query := `INSERT INTO transactions (user_id, booking_id, amount, type) 
              VALUES ($1, $2, $3, $4)`
	_, err := tx.ExecContext(ctx, query, userID, bookingID, amount, txType)
	return err
}
