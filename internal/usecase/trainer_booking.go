package usecase

import (
	"context"
	"errors"
	"github.com/yertaypert/gym-booking-api/internal/domain"
)

var (
	ErrTrainerSlotNotFound     = errors.New("trainer slot not found")
	ErrTrainerSlotNotAvailable = errors.New("trainer slot not available")
)

type TrainerBookingUsecase struct {
	trainerSlotRepo    TrainerSlotRepository
	trainerBookingRepo TrainerBookingRepository
}

func NewTrainerBookingUsecase(
	trainerSlotRepo TrainerSlotRepository,
	trainerBookingRepo TrainerBookingRepository,
) *TrainerBookingUsecase {
	return &TrainerBookingUsecase{
		trainerSlotRepo:    trainerSlotRepo,
		trainerBookingRepo: trainerBookingRepo,
	}
}
func (u *TrainerBookingUsecase) BookTrainerSlot(ctx context.Context, userID int, slotID int) error {
	slot, err := u.trainerSlotRepo.GetByID(ctx, slotID)
	if err != nil {
		return err
	}
	if slot == nil {
		return ErrTrainerSlotNotFound
	}
	if slot.Status != "available" {
		return ErrTrainerSlotNotAvailable
	}
	booking := &domain.TrainerBooking{
		UserID:        userID,
		TrainerSlotID: slot.ID,
		Status:        "active",
	}
	if err := u.trainerBookingRepo.Create(ctx, booking); err != nil {
		return err
	}
	if err := u.trainerSlotRepo.UpdateStatus(ctx, slotID, "booked"); err != nil {
		return err
	}
	return nil
}
