package repository

import (
	"context"
	"database/sql"

	"github.com/yertaypert/gym-booking-api/internal/domain"
)

type BookingRepository interface {
	Create(ctx context.Context, tx *sql.Tx, userID, sessionID int) (int, error)
}

type WalletRepository interface { // or rename to TransactionRepository later
	UpdateBalance(ctx context.Context, tx *sql.Tx, userID int, amount float64) error
	CreateTransaction(ctx context.Context, tx *sql.Tx, userID int, bookingID *int, amount float64, txType domain.TransactionType) error
}

type UserRepository interface {
	GetByID(ctx context.Context, id int) (*domain.User, error)
	// ... other methods later
}
