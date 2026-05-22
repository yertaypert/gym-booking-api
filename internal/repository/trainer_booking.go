package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/yertaypert/gym-booking-api/internal/domain"
)

type TrainerBookingRepository struct {
	db *sql.DB
}

func NewTrainerBookingRepository(db *sql.DB) *TrainerBookingRepository {
	return &TrainerBookingRepository{db: db}
}

func (r *TrainerBookingRepository) Create(ctx context.Context, booking *domain.TrainerBooking) error {
	query := `INSERT INTO trainer_bookings (user_id, trainer_slot_id, status) VALUES ($1, $2, $3) returning id, created_at`

	return r.db.QueryRowContext(
		ctx,
		query,
		booking.UserID,
		booking.TrainerSlotID,
		booking.Status).Scan(&booking.ID, &booking.CreatedAt)
}
func (r *TrainerBookingRepository) GetByUserID(ctx context.Context, userID int) ([]domain.TrainerBooking, error) {
	query := `SELECT id, used_id, trainer_slot_id, status, created_at FROM trainer_bookings WHERE user_id = $1 order by created_at desc`

	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []domain.TrainerBooking
	for rows.Next() {
		var booking domain.TrainerBooking
		err := rows.Scan(
			&booking.ID,
			&booking.UserID,
			&booking.TrainerSlotID,
			&booking.Status,
			&booking.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, booking)
	}
	return bookings, rows.Err()
}
func (r *TrainerBookingRepository) GetByID(ctx context.Context, bookingID int) (*domain.TrainerBooking, error) {
	query := `SELECT id, user_id, trainer_slot_id, status, created_at FROM trainer_bookings WHERE id = $1`

	var booking domain.TrainerBooking
	err := r.db.QueryRowContext(ctx, query, bookingID).Scan(
		&booking.ID,
		&booking.UserID,
		&booking.TrainerSlotID,
		&booking.Status,
		&booking.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &booking, nil
}
func (r *TrainerBookingRepository) UpdateStatus(ctx context.Context, bookingID int, status string) error {
	query := `UPDATE trainer_bookings SET status = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, bookingID)
	return err
}
