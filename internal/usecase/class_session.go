package usecase

import (
	"errors"
	"time"

	"github.com/yertaypert/gym-booking-api/internal/domain"
	"github.com/yertaypert/gym-booking-api/internal/repository"
)

type ClassSessionUsecase struct {
	repo *repository.ClassSessionRepository
}

func NewClassSessionUsecase(repo *repository.ClassSessionRepository) *ClassSessionUsecase {
	return &ClassSessionUsecase{repo: repo}
}

func (u *ClassSessionUsecase) CreateSession(gymID int, title string, startTime, endTime time.Time, capacity int) (*domain.ClassSession, error) {
	if capacity <= 0 {
		return nil, errors.New("capacity must be greater than zero")
	}
	if startTime.After(endTime) || startTime.Equal(endTime) {
		return nil, errors.New("start time must be before end time")
	}

	session := &domain.ClassSession{
		GymID:     gymID,
		Title:     title,
		StartTime: startTime,
		EndTime:   endTime,
		Capacity:  capacity,
		Booked:    0,
	}

	err := u.repo.Create(session)
	if err != nil {
		return nil, err
	}

	return session, nil
}
