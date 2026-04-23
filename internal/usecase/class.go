package usecase

import (
	"context"
	"time"

	"github.com/yertaypert/gym-booking-api/internal/repository"
)

type ClassUsecase struct {
	classRepo ClassRepository
}

func NewClassUsecase(classRepo ClassRepository) *ClassUsecase {
	return &ClassUsecase{classRepo: classRepo}
}

func (u *ClassUsecase) ListDistinctClasses(ctx context.Context) ([]string, error) {
	return u.classRepo.ListDistinctClasses(ctx)
}

func (u *ClassUsecase) ListGymsByClassName(ctx context.Context, name string) ([]repository.GymWithClass, error) {
	return u.classRepo.ListGymsByClassName(ctx, name)
}

func (u *ClassUsecase) SearchSessions(
	ctx context.Context,
	name string,
	startTime, endTime *time.Time,
) ([]repository.SessionWithGym, error) {
	return u.classRepo.SearchSessionsByClassName(ctx, name, startTime, endTime)
}

func (u *ClassUsecase) GetSession(ctx context.Context, sessionID int) (*repository.SessionWithGym, error) {
	return u.classRepo.GetSessionWithDetails(ctx, sessionID)
}
