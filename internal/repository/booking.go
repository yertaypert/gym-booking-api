package repository

import (
	"context"
	"database/sql"

	"github.com/yertaypert/gym-booking-api/internal/domain"
)

type BookingRepository struct {
	db *sql.DB
}

func NewBookingRepository(db *sql.DB) *BookingRepository {
	return &BookingRepository{db: db}
}

func (r *BookingRepository) Create(tx *sql.Tx, userID, sessionID int) (int, error) {
	var id int
	query := `INSERT INTO bookings (user_id, session_id, status) VALUES ($1, $2, 'pending') RETURNING id`
	err := tx.QueryRow(query, userID, sessionID).Scan(&id)
	return id, err
}

func (r *BookingRepository) UpdateStatus(ctx context.Context, tx *sql.Tx, bookingID int, status string) error {
	query := `UPDATE bookings SET status = $1 WHERE id = $2`
	_, err := tx.ExecContext(ctx, query, status, bookingID)
	return err
}

func (r *BookingRepository) GetByID(ctx context.Context, bookingID int) (*domain.Booking, error) {
	query := `SELECT id, user_id, session_id, status, created_at FROM bookings WHERE id = $1`
	var booking domain.Booking

	err := r.db.QueryRowContext(ctx, query, bookingID).Scan(
		&booking.ID,
		&booking.UserID,
		&booking.SessionID,
		&booking.Status,
		&booking.CreatedAt,
	)
	if err != nil {
		return nil, err
	}
	return &booking, nil
}
func (r *BookingRepository) GetByUserID(ctx context.Context, userID int) ([]domain.Booking, error) {
	query := `SELECT id, user_id, session_id, status, created_at FROM bookings WHERE user_id = $1 ORDER BY created_at DESC`
	rows, err := r.db.QueryContext(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var bookings []domain.Booking
	for rows.Next() {
		var b domain.Booking
		err := rows.Scan(
			&b.ID,
			&b.UserID,
			&b.SessionID,
			&b.Status,
			&b.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		bookings = append(bookings, b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return bookings, nil
}
