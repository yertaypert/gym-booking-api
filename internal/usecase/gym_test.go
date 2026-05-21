package usecase

import (
	"context"
	"errors"
	"testing"
)

func TestCreateGymValidation(t *testing.T) {
	u := &GymUsecase{}

	t.Run("missing owner_id", func(t *testing.T) {
		_, err := u.CreateGym(context.Background(), 0, "Gym Name", "addr", "desc")
		if !errors.Is(err, ErrInvalidOwnerID) {
			t.Fatalf("expected ErrInvalidOwnerID, got %v", err)
		}
	})

	t.Run("missing name", func(t *testing.T) {
		_, err := u.CreateGym(context.Background(), 1, "   ", "addr", "desc")
		if !errors.Is(err, ErrInvalidGymName) {
			t.Fatalf("expected ErrInvalidGymName, got %v", err)
		}
	})
}

func TestCreateClassValidation(t *testing.T) {
	// These require a mock repo because they call GetGym first
}

func TestCreateSessionValidation(t *testing.T) {
	// These require a mock repo because they call GetGym first
}
