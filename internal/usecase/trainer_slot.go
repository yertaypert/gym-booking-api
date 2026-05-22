package usecase

import (
	"context"
	"time"

	"github.com/yertaypert/gym-booking-api/internal/domain"
)

type TrainerSlotUsecase struct {
	trainerSlotRepo TrainerSlotRepository
}

func NewTrainerSlotUsecase(trainerSlotRepo TrainerSlotRepository) *TrainerSlotUsecase {
	return &TrainerSlotUsecase{
		trainerSlotRepo: trainerSlotRepo,
	}
}

func (u *TrainerSlotUsecase) ListAvailableSlots(ctx context.Context) ([]domain.TrainerSlot, error) {
	return u.trainerSlotRepo.ListAvailableSlots(ctx)
}
func (u *TrainerSlotUsecase) CreateTrainerSlot(ctx context.Context, trainerID int, startTime time.Time, endTime time.Time) (*domain.TrainerSlot, error) {
	slot := &domain.TrainerSlot{
		TrainerID: trainerID,
		StartTime: startTime,
		EndTime:   endTime,
		Status:    "available",
	}
	if err := u.trainerSlotRepo.Create(ctx, slot); err != nil {
		return nil, err
	}
	return slot, nil
}
