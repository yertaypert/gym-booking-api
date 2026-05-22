package usecase

import (
	"context"
	"errors"

	"github.com/yertaypert/gym-booking-api/internal/domain"
)

var (
	ErrTrainerSlotNotFound     = errors.New("trainer slot not found")
	ErrTrainerSlotNotAvailable = errors.New("trainer slot not available")
	ErrTrainerBookingNotFound  = errors.New("trainer booking not found")
	ErrTrainerBookingForbidden = errors.New("trainer booking does not belong to this user")
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
func (u *TrainerBookingUsecase) GetMyTrainerBookings(ctx context.Context, userID int) ([]domain.TrainerBooking, error) {
	return u.trainerBookingRepo.GetByUserID(ctx, userID)
}
func (u *TrainerBookingUsecase) CancelTrainerBooking(ctx context.Context, userID int, bookingID int) error {
	booking, err := u.trainerBookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return err
	}
	if booking == nil {
		return ErrTrainerBookingNotFound
	}
	if booking.UserID != userID {
		return ErrTrainerBookingForbidden
	}
	if err := u.trainerSlotRepo.UpdateStatus(ctx, bookingID, "cancelled"); err != nil {
		return err
	}
	return u.trainerSlotRepo.UpdateStatus(ctx, booking.TrainerSlotID, "available")
}
