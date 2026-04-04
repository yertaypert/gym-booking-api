package usecase

import (
	"errors"
	"testing"
	"time"
)

func TestCreateSessionValidation(t *testing.T) {
	u := &GymUsecase{}
	start := time.Date(2026, 4, 5, 10, 0, 0, 0, time.UTC)

	_, err := u.CreateSession(1, 1, start, start, 10)
	if !errors.Is(err, ErrInvalidSessionTime) {
		t.Fatalf("expected ErrInvalidSessionTime, got %v", err)
	}

	_, err = u.CreateSession(1, 1, start, start.Add(time.Hour), 0)
	if !errors.Is(err, ErrInvalidSessionPrice) {
		t.Fatalf("expected ErrInvalidSessionPrice, got %v", err)
	}
}

func TestCreateGymValidation(t *testing.T) {
	u := &GymUsecase{}

	_, err := u.CreateGym("   ", "addr", "desc")
	if !errors.Is(err, ErrInvalidGymName) {
		t.Fatalf("expected ErrInvalidGymName, got %v", err)
	}
}

func TestCreateClassValidation(t *testing.T) {
	u := &GymUsecase{}

	_, err := u.CreateClass(1, "   ", 10)
	if !errors.Is(err, ErrInvalidClassName) {
		t.Fatalf("expected ErrInvalidClassName, got %v", err)
	}

	_, err = u.CreateClass(1, "Yoga", 0)
	if !errors.Is(err, ErrInvalidMaxCapacity) {
		t.Fatalf("expected ErrInvalidMaxCapacity, got %v", err)
	}
}
