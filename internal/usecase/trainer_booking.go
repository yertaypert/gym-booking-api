package usecase

import (
	"context"
	"database/sql"

	"github.com/yertaypert/gym-booking-api/internal/domain"
)

var (
	ErrTrainerSlotNotFound     = domain.ErrTrainerSlotNotFound
	ErrTrainerSlotNotAvailable = domain.ErrTrainerSlotNotAvailable
)

type TrainerBookingUsecase struct {
	db                 *sql.DB
	trainerSlotRepo    TrainerSlotRepository
	trainerBookingRepo TrainerBookingRepository
}

func NewTrainerBookingUsecase(
	db *sql.DB,
	trainerSlotRepo TrainerSlotRepository,
	trainerBookingRepo TrainerBookingRepository,
) *TrainerBookingUsecase {
	return &TrainerBookingUsecase{
		db:                 db,
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

	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	booking := &domain.TrainerBooking{
		UserID:        userID,
		TrainerSlotID: slot.ID,
		Status:        "active",
	}

	if err := u.trainerBookingRepo.Create(ctx, tx, booking); err != nil {
		return err
	}

	if err := u.trainerSlotRepo.UpdateStatus(ctx, tx, slotID, "booked"); err != nil {
		return err
	}

	return tx.Commit()
}
