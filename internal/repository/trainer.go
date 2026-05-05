package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/yertaypert/gym-booking-api/internal/domain"
)

type TrainerRepository struct {
	db *sql.DB
}

func NewTrainerRepository(db *sql.DB) *TrainerRepository {
	return &TrainerRepository{db: db}
}

func (r *TrainerRepository) Create(ctx context.Context, tx *sql.Tx, t *domain.Trainer) error {
	query := `INSERT INTO trainers (user_id, specialization, extra_fee) VALUES ($1, $2, $3) RETURNING id`

	var err error
	if tx != nil {
		err = tx.QueryRowContext(ctx, query, t.UserID, t.Specialization, t.ExtraFee).Scan(&t.ID)
	} else {
		err = r.db.QueryRowContext(ctx, query, t.UserID, t.Specialization, t.ExtraFee).Scan(&t.ID)
	}

	return err
}

func (r *TrainerRepository) GetByUserID(ctx context.Context, userID int) (*domain.Trainer, error) {
	query := `SELECT id, user_id, specialization, extra_fee FROM trainers WHERE user_id = $1`
	t := &domain.Trainer{}
	err := r.db.QueryRowContext(ctx, query, userID).Scan(&t.ID, &t.UserID, &t.Specialization, &t.ExtraFee)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return t, nil
}
