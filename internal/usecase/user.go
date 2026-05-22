package usecase

import (
	"context"
	"github.com/yertaypert/gym-booking-api/internal/domain"
)

type UserUsecase struct {
	userRepo UserRepository
}

func NewUserUsecase(repo UserRepository) *UserUsecase {
	return &UserUsecase{userRepo: repo}
}

func (u *UserUsecase) ListAllUsers(ctx context.Context) ([]domain.User, error) {
	return u.userRepo.ListAll(ctx)
}

func (u *UserUsecase) GetUserByID(ctx context.Context, id int) (*domain.User, error) {
	return u.userRepo.GetByID(id)
}

func (u *UserUsecase) UpdateUser(ctx context.Context, user *domain.User) error {
	// Add validations if needed
	return u.userRepo.Update(ctx, user)
}

func (u *UserUsecase) DeleteUser(ctx context.Context, id int) error {
	return u.userRepo.Delete(ctx, id)
}
