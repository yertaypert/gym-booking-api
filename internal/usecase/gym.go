package usecase

import (
	"errors"
	"strings"
	"time"

	"github.com/yertaypert/gym-booking-api/internal/domain"
	"github.com/yertaypert/gym-booking-api/internal/repository"
)

var ErrGymNotFound = errors.New("gym not found")
var ErrClassNotFound = errors.New("class not found")
var ErrClassDoesNotBelongToGym = errors.New("class does not belong to gym")
var ErrInvalidGymName = errors.New("gym name is required")
var ErrInvalidClassName = errors.New("class name is required")
var ErrInvalidMaxCapacity = errors.New("max_capacity must be greater than 0")
var ErrInvalidSessionTime = errors.New("end_time must be after start_time")
var ErrInvalidSessionPrice = errors.New("price must be greater than 0")

type GymUsecase struct {
	gymRepo *repository.GymRepository
}

func NewGymUsecase(repo *repository.GymRepository) *GymUsecase {
	return &GymUsecase{gymRepo: repo}
}

func (u *GymUsecase) ListGyms() ([]domain.Gym, error) {
	return u.gymRepo.ListGyms()
}

func (u *GymUsecase) GetGym(id int) (*domain.Gym, error) {
	gym, err := u.gymRepo.GetGymByID(id)
	if err != nil {
		if errors.Is(err, repository.ErrGymNotFound) {
			return nil, ErrGymNotFound
		}
		return nil, err
	}

	return gym, nil
}

func (u *GymUsecase) ListGymClasses(gymID int) ([]domain.Class, error) {
	classes, err := u.gymRepo.ListClassesByGymID(gymID)
	if err != nil {
		if errors.Is(err, repository.ErrGymNotFound) {
			return nil, ErrGymNotFound
		}
		return nil, err
	}

	return classes, nil
}

func (u *GymUsecase) ListClassSessions(gymID, classID int) ([]domain.Session, error) {
	sessions, err := u.gymRepo.ListSessionsByGymAndClassID(gymID, classID)
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

func (u *GymUsecase) CreateGym(name, address, description string) (*domain.Gym, error) {
	gym := domain.Gym{
		Name:        strings.TrimSpace(name),
		Address:     strings.TrimSpace(address),
		Description: strings.TrimSpace(description),
	}

	if gym.Name == "" {
		return nil, ErrInvalidGymName
	}

	return u.gymRepo.CreateGym(gym)
}

func (u *GymUsecase) CreateClass(gymID int, name string, maxCapacity int) (*domain.Class, error) {
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

	created, err := u.gymRepo.CreateClass(class)
	if err != nil {
		if errors.Is(err, repository.ErrGymNotFound) {
			return nil, ErrGymNotFound
		}
		return nil, err
	}

	return created, nil
}

func (u *GymUsecase) CreateSession(gymID, classID int, startTime, endTime time.Time, price float64) (*domain.Session, error) {
	if !endTime.After(startTime) {
		return nil, ErrInvalidSessionTime
	}
	if price <= 0 {
		return nil, ErrInvalidSessionPrice
	}

	class, err := u.gymRepo.GetClassByID(classID)
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

	created, err := u.gymRepo.CreateSession(gymID, session)
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
