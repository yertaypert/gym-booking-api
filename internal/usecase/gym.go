package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/yertaypert/gym-booking-api/internal/domain"
	"github.com/yertaypert/gym-booking-api/internal/repository"
)

var ErrGymNotFound = errors.New("gym not found")
var ErrGymAlreadyExists = errors.New("gym with this name already exists")
var ErrClassNotFound = errors.New("class not found")
var ErrClassDoesNotBelongToGym = errors.New("class does not belong to gym")
var ErrInvalidGymName = errors.New("gym name is required")
var ErrInvalidOwnerID = errors.New("owner_id is required")
var ErrInvalidClassName = errors.New("class name is required")
var ErrInvalidMaxCapacity = errors.New("max_capacity must be greater than 0")
var ErrInvalidSessionTime = errors.New("end_time must be after start_time")
var ErrInvalidSessionPrice = errors.New("price must be greater than 0")
var ErrNotGymOwner = errors.New("user is not the owner of this gym")

type GymUsecase struct {
	gymRepo GymRepository
}

func NewGymUsecase(repo GymRepository) *GymUsecase {
	return &GymUsecase{gymRepo: repo}
}

func (u *GymUsecase) ListGyms(ctx context.Context) ([]domain.Gym, error) {
	return u.gymRepo.ListGyms(ctx)
}

func (u *GymUsecase) ListGymsByOwner(ctx context.Context, ownerID int) ([]domain.Gym, error) {
	return u.gymRepo.ListGymsByOwnerID(ctx, ownerID)
}

func (u *GymUsecase) GetGym(ctx context.Context, id int) (*domain.Gym, error) {
	gym, err := u.gymRepo.GetGymByID(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrGymNotFound) {
			return nil, ErrGymNotFound
		}
		return nil, err
	}

	return gym, nil
}

func (u *GymUsecase) ListGymClasses(ctx context.Context, gymID int) ([]domain.Class, error) {
	classes, err := u.gymRepo.ListClassesByGymID(ctx, gymID)
	if err != nil {
		if errors.Is(err, repository.ErrGymNotFound) {
			return nil, ErrGymNotFound
		}
		return nil, err
	}

	return classes, nil
}

func (u *GymUsecase) ListClassSessions(ctx context.Context, gymID, classID int) ([]domain.Session, error) {
	sessions, err := u.gymRepo.ListSessionsByGymAndClassID(ctx, gymID, classID)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrGymNotFound):
			return nil, ErrGymNotFound
		case errors.Is(err, repository.ErrClassNotFound):
			return nil, ErrClassNotFound
		case errors.Is(err, repository.ErrClassDoesNotBelongToGym):
			return nil, ErrClassDoesNotBelongToGym
		}
		return nil, err
	}

	return sessions, nil
}

func (u *GymUsecase) CreateGym(ctx context.Context, ownerID int, name, address, description string) (*domain.Gym, error) {
	if ownerID <= 0 {
		return nil, ErrInvalidOwnerID
	}

	gym := domain.Gym{
		OwnerID:     ownerID,
		Name:        strings.TrimSpace(name),
		Address:     strings.TrimSpace(address),
		Description: strings.TrimSpace(description),
	}

	if gym.Name == "" {
		return nil, ErrInvalidGymName
	}

	created, err := u.gymRepo.CreateGym(ctx, gym)
	if err != nil {
		if errors.Is(err, repository.ErrGymAlreadyExists) {
			return nil, ErrGymAlreadyExists
		}
		return nil, err
	}

	return created, nil
}

func (u *GymUsecase) CreateClass(ctx context.Context, userID int, userRole domain.UserRole, gymID int, name string, maxCapacity int) (*domain.Class, error) {
	gym, err := u.GetGym(ctx, gymID)
	if err != nil {
		return nil, err
	}

	if userRole != domain.RoleAdmin && gym.OwnerID != userID {
		return nil, ErrNotGymOwner
	}

	class := domain.Class{
		GymID:       gymID,
		Name:        strings.TrimSpace(name),
		MaxCapacity: maxCapacity,
	}

	if class.Name == "" {
		return nil, ErrInvalidClassName
	}
	if class.MaxCapacity <= 0 {
		return nil, ErrInvalidMaxCapacity
	}

	created, err := u.gymRepo.CreateClass(ctx, class)
	if err != nil {
		if errors.Is(err, repository.ErrGymNotFound) {
			return nil, ErrGymNotFound
		}
		return nil, err
	}

	return created, nil
}

func (u *GymUsecase) CreateSession(ctx context.Context, userID int, userRole domain.UserRole, gymID, classID int, startTime, endTime time.Time, price float64) (*domain.Session, error) {
	gym, err := u.GetGym(ctx, gymID)
	if err != nil {
		return nil, err
	}

	if userRole != domain.RoleAdmin && gym.OwnerID != userID {
		return nil, ErrNotGymOwner
	}

	if !endTime.After(startTime) {
		return nil, ErrInvalidSessionTime
	}
	if price <= 0 {
		return nil, ErrInvalidSessionPrice
	}

	class, err := u.gymRepo.GetClassByID(ctx, classID)
	if err != nil {
		if errors.Is(err, repository.ErrClassNotFound) {
			return nil, ErrClassNotFound
		}
		return nil, err
	}
	if class.GymID != gymID {
		return nil, ErrClassDoesNotBelongToGym
	}

	session := domain.Session{
		ClassID:        classID,
		StartTime:      startTime,
		EndTime:        endTime,
		AvailableSlots: class.MaxCapacity,
		Price:          price,
		Status:         "active",
	}

	created, err := u.gymRepo.CreateSession(ctx, gymID, session)
	if err != nil {
		switch {
		case errors.Is(err, repository.ErrGymNotFound):
			return nil, ErrGymNotFound
		case errors.Is(err, repository.ErrClassNotFound):
			return nil, ErrClassNotFound
		case errors.Is(err, repository.ErrClassDoesNotBelongToGym):
			return nil, ErrClassDoesNotBelongToGym
		}
		return nil, err
	}

	return created, nil
}

func (u *GymUsecase) AssignTrainer(ctx context.Context, gymID int, trainerID int) error {
	return u.gymRepo.AssignTrainer(ctx, gymID, trainerID)
}
