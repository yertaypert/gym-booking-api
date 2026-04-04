package repository

import (
	"context"
	"database/sql"

	"github.com/yertaypert/gym-booking-api/internal/domain"
)

type BookingRepository interface {
	GetByID(ctx context.Context, bookingID int) (*domain.Booking, error)
	Create(ctx context.Context, tx *sql.Tx, userID, sessionID int) (int, error)
	UpdateStatus(ctx context.Context, tx *sql.Tx, bookingID int, status string) error
}

type WalletRepository interface { // or rename to TransactionRepository later
	UpdateBalance(ctx context.Context, tx *sql.Tx, userID int, amount float64) error
	CreateTransaction(ctx context.Context, tx *sql.Tx, userID int, bookingID *int, amount float64, txType domain.TransactionType) error
}

type UserRepository interface {
	GetByID(ctx context.Context, id int) (*domain.User, error)
	// ... other methods later
}

type SessionRepository interface {
	GetByID(ctx context.Context, id int) (*domain.Session, error)
	DecreaseAvailableSlots(ctx context.Context, tx *sql.Tx, sessionID int) error
	IncreaseAvailableSlots(ctx context.Context, tx *sql.Tx, sessionID int) error
}
