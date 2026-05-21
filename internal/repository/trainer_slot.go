package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/yertaypert/gym-booking-api/internal/domain"
)

type TrainerSlotRepository struct {
	db *sql.DB
}

func NewTrainerSlotRepository(db *sql.DB) *TrainerSlotRepository {
	return &TrainerSlotRepository{db: db}
}
func (r *TrainerSlotRepository) GetByID(ctx context.Context, slotID int) (*domain.TrainerSlot, error) {
	query := `SELECT id, trainer_id, start_time, end_time, status FROM trainer_slot WHERE id = $1`

	var slot domain.TrainerSlot
	err := r.db.QueryRowContext(ctx, query, slotID).Scan(
		&slot.ID,
		&slot.TrainerID,
		&slot.StartTime,
		&slot.EndTime,
		&slot.Status,
	)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &slot, nil
}
func (r *TrainerSlotRepository) UpdateStatus(ctx context.Context, slotID int, status string) error {
	query := `UPDATE trainer_slot SET status = $1 WHERE id = $2`
	_, err := r.db.ExecContext(ctx, query, status, slotID)
	return err
}
