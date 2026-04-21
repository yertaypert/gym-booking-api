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

func (r *BookingRepository) ListByGymID(ctx context.Context, gymID int) ([]domain.Booking, error) {
	query := `
		SELECT b.id, b.user_id, b.session_id, b.status, b.created_at
		FROM bookings b
		JOIN class_sessions s ON b.session_id = s.id
		JOIN classes c ON s.class_id = c.id
		WHERE c.gym_id = $1
		ORDER BY b.created_at DESC
	`
	rows, err := r.db.QueryContext(ctx, query, gymID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var bookings []domain.Booking
	for rows.Next() {
		var b domain.Booking
		if err := rows.Scan(&b.ID, &b.UserID, &b.SessionID, &b.Status, &b.CreatedAt); err != nil {
			return nil, err
		}
		bookings = append(bookings, b)
	}
	return bookings, rows.Err()
}
