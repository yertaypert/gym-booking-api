package repository

import (
	"context"
	"database/sql"

	"github.com/yertaypert/gym-booking-api/internal/domain"
)

type TrainerBookingRepository struct {
	db *sql.DB
}

func NewTrainerBookingRepository(db *sql.DB) *TrainerBookingRepository {
	return &TrainerBookingRepository{db: db}
}

func (r *TrainerBookingRepository) Create(ctx context.Context, tx *sql.Tx, booking *domain.TrainerBooking) error {
	query := `INSERT INTO trainer_bookings (user_id, trainer_slot_id, status) VALUES ($1, $2, $3) RETURNING id, created_at`

	var err error
	if tx != nil {
		err = tx.QueryRowContext(
			ctx,
			query,
			booking.UserID,
			booking.TrainerSlotID,
			booking.Status).Scan(&booking.ID, &booking.CreatedAt)
	} else {
		err = r.db.QueryRowContext(
			ctx,
			query,
			booking.UserID,
			booking.TrainerSlotID,
			booking.Status).Scan(&booking.ID, &booking.CreatedAt)
	}
	return err
}
