package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/yertaypert/gym-booking-api/internal/domain"
)

type sqlSessionRepository struct {
	db *sql.DB
}

func NewSessionRepository(db *sql.DB) SessionRepository {
	return &sqlSessionRepository{db: db}
}

func (r *sqlSessionRepository) GetByID(ctx context.Context, sessionID int) (*domain.Session, error) {
	query := `SELECT id, class_id, start_time, end_time, available_slots, price, status
              FROM class_sessions
              WHERE id = $1`
	var session domain.Session
	err := r.db.QueryRowContext(ctx, query, sessionID).Scan(
		&session.ID,
		&session.ClassID,
		&session.StartTime,
		&session.EndTime,
		&session.AvailableSlots,
		&session.Price,
		&session.Status,
	)
	if err != nil {
		return nil, err
	}
	return &session, nil
}

func (r *sqlSessionRepository) DecreaseAvailableSlots(ctx context.Context, tx *sql.Tx, sessionID int) error {
	query := `UPDATE class_sessions SET available_slots = available_slots - 1
              WHERE id = $1 AND available_slots > 0`
	result, err := tx.ExecContext(ctx, query, sessionID)
	if err != nil {
		return err
	}
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rowsAffected == 0 {
		return errors.New("session does not have enough available slots")
	}
	return nil
}

func (r *sqlSessionRepository) IncreaseAvailableSlots(ctx context.Context, tx *sql.Tx, sessionID int) error {
	query := `UPDATE class_sessions SET available_slots = available_slots + 1
              WHERE id = $1`
	_, err := tx.ExecContext(ctx, query, sessionID)
	return err
}
