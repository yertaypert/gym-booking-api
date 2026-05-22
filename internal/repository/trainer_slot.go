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

func (r *TrainerSlotRepository) ListAvailableSlots(ctx context.Context) ([]domain.TrainerSlot, error) {
	query := `SELECT id, trainer_id, start_time, end_time, status FROM trainer_slot WHERE status = 'available' ORDER BY start_time ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var slots []domain.TrainerSlot
	for rows.Next() {
		var slot domain.TrainerSlot
		err := rows.Scan(
			&slot.ID,
			&slot.TrainerID,
			&slot.StartTime,
			&slot.EndTime,
			&slot.Status)
		if err != nil {
			return nil, err
		}
		slots = append(slots, slot)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return slots, nil
}
func (r *TrainerSlotRepository) Create(ctx context.Context, slot *domain.TrainerSlot) error {
	query := `INSERT INTO trainer_slot (trainer_id, start_time, end_time, status) VALUES ($1, $2, $3, $4) returning id`

	return r.db.QueryRowContext(
		ctx,
		query,
		slot.TrainerID,
		slot.StartTime,
		slot.EndTime,
		slot.Status).Scan(&slot.ID)
}
