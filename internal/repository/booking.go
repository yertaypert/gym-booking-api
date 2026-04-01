package repository

import (
	"context"
	"database/sql"
)

type sqlBookingRepository struct {
	db *sql.DB
}

func NewBookingRepository(db *sql.DB) BookingRepository {
	return &sqlBookingRepository{db: db}
}

func (r *sqlBookingRepository) Create(ctx context.Context, tx *sql.Tx, userID, sessionID int) (int, error) {
	var id int
	query := `INSERT INTO bookings (user_id, session_id, status) VALUES ($1, $2, 'pending') RETURNING id`

	err := tx.QueryRowContext(ctx, query, userID, sessionID).Scan(&id)
	return id, err
}
