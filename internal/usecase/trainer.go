package usecase

import (
	"context"
	"database/sql"
	"github.com/yertaypert/gym-booking-api/internal/domain"
)

type TrainerUsecase struct {
	db          *sql.DB
	userRepo    UserRepository
	trainerRepo TrainerRepository
}

func NewTrainerUsecase(db *sql.DB, userRepo UserRepository, trainerRepo TrainerRepository) *TrainerUsecase {
	return &TrainerUsecase{
		db:          db,
		userRepo:    userRepo,
		trainerRepo: trainerRepo,
	}
}

func (u *TrainerUsecase) PromoteToTrainer(ctx context.Context, userID int, specialization string, extraFee float64) error {
	// Check if user exists
	if _, err := u.userRepo.GetByID(userID); err != nil {
		return err
	}

	// Check if already a trainer profile exists
	existing, err := u.trainerRepo.GetByUserID(ctx, userID)
	if err != nil {
		return err
	}
	if existing != nil {
		return domain.ErrTrainerAlreadyAssigned // Or a more specific error like ErrAlreadyATrainer
	}

	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	// Update user role
	if err := u.userRepo.UpdateRole(ctx, tx, userID, domain.RoleTrainer); err != nil {
		return err
	}

	// Create trainer profile
	trainer := &domain.Trainer{
		UserID:         userID,
		Specialization: specialization,
		ExtraFee:       extraFee,
	}
	if err := u.trainerRepo.Create(ctx, tx, trainer); err != nil {
		return err
	}

	return tx.Commit()
}
