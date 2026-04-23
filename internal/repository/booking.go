package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

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
	if err != nil {
		// Detect the unique constraint violation and return a friendly error.
		if strings.Contains(err.Error(), "unique_user_session") ||
			strings.Contains(err.Error(), "unique constraint") {
			return 0, fmt.Errorf("you are already booked for this session")
		}
		return 0, err
	}
	return id, nil
}

func (r *BookingRepository) UpdateStatus(ctx context.Context, tx *sql.Tx, bookingID int, status string) error {
	query := `UPDATE bookings SET status = $1 WHERE id = $2`
	_, err := tx.ExecContext(ctx, query, status, bookingID)
	return err
}

func (r *BookingRepository) GetByID(ctx context.Context, bookingID int) (*domain.Booking, error) {
	query := `SELECT id, user_id, session_id, status, created_at, attended_at FROM bookings WHERE id = $1`
	var booking domain.Booking
	var attendedAt sql.NullTime

	err := r.db.QueryRowContext(ctx, query, bookingID).Scan(
		&booking.ID,
		&booking.UserID,
		&booking.SessionID,
		&booking.Status,
		&booking.CreatedAt,
		&attendedAt,
	)
	if err != nil {
		return nil, err
	}
	if attendedAt.Valid {
		booking.AttendedAt = &attendedAt.Time
	}
	return &booking, nil
}

// ExistsByUserAndSession returns true when an active (non-cancelled) booking
// already exists for the given user/session pair.
func (r *BookingRepository) ExistsByUserAndSession(ctx context.Context, userID, sessionID int) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(
		SELECT 1 FROM bookings
		WHERE user_id = $1 AND session_id = $2 AND status != 'cancelled'
	)`
	err := r.db.QueryRowContext(ctx, query, userID, sessionID).Scan(&exists)
	return exists, err
}

// GetByUserID returns all bookings for a user with session/class/gym detail.
func (r *BookingRepository) GetByUserID(ctx context.Context, userID int) ([]domain.BookingDetail, error) {
	query := `
		SELECT
			b.id,
			b.session_id,
			cl.name          AS class_name,
			g.name           AS gym_name,
			cs.start_time,
			cs.end_time,
			cs.price,
			b.status,
			b.user_id,
			u.email,
			u.full_name,
			b.created_at,
			b.attended_at
		FROM bookings b
		JOIN class_sessions cs ON cs.id = b.session_id
		JOIN classes cl        ON cl.id = cs.class_id
		JOIN gyms g            ON g.id  = cl.gym_id
		JOIN users u           ON u.id  = b.user_id
		WHERE b.user_id = $1
		ORDER BY cs.start_time DESC
	`
	return r.scanBookingDetails(ctx, query, userID)
}

// GetBySessionID returns all bookings for a session (used by admin/trainer).
func (r *BookingRepository) GetBySessionID(ctx context.Context, sessionID int) ([]domain.BookingDetail, error) {
	query := `
		SELECT
			b.id,
			b.session_id,
			cl.name          AS class_name,
			g.name           AS gym_name,
			cs.start_time,
			cs.end_time,
			cs.price,
			b.status,
			b.user_id,
			u.email,
			u.full_name,
			b.created_at,
			b.attended_at
		FROM bookings b
		JOIN class_sessions cs ON cs.id = b.session_id
		JOIN classes cl        ON cl.id = cs.class_id
		JOIN gyms g            ON g.id  = cl.gym_id
		JOIN users u           ON u.id  = b.user_id
		WHERE b.session_id = $1
		ORDER BY b.created_at ASC
	`
	return r.scanBookingDetails(ctx, query, sessionID)
}

func (r *BookingRepository) scanBookingDetails(ctx context.Context, query string, arg int) ([]domain.BookingDetail, error) {
	rows, err := r.db.QueryContext(ctx, query, arg)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []domain.BookingDetail
	for rows.Next() {
		var d domain.BookingDetail
		var fullName sql.NullString
		var attendedAt sql.NullTime
		if err := rows.Scan(
			&d.BookingID,
			&d.SessionID,
			&d.ClassName,
			&d.GymName,
			&d.StartTime,
			&d.EndTime,
			&d.Price,
			&d.Status,
			&d.UserID,
			&d.UserEmail,
			&fullName,
			&d.BookedAt,
			&attendedAt,
		); err != nil {
			return nil, err
		}
		if fullName.Valid {
			d.UserName = fullName.String
		}
		if attendedAt.Valid {
			d.AttendedAt = &attendedAt.Time
		}
		result = append(result, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if result == nil {
		return []domain.BookingDetail{}, nil
	}
	return result, nil
}

// MarkAttended sets status = 'attended' and stamps attended_at = NOW().
// Must be called inside an open transaction.
func (r *BookingRepository) MarkAttended(ctx context.Context, tx *sql.Tx, bookingID int) error {
	query := `UPDATE bookings SET status = 'attended', attended_at = NOW() WHERE id = $1`
	_, err := tx.ExecContext(ctx, query, bookingID)
	return err
}

// ErrNoRows is a sentinel so callers don't import database/sql directly.
var ErrNoRows = errors.New("record not found")
