package usecase

import (
	"errors"
	"testing"
	"time"

	"github.com/yertaypert/gym-booking-api/internal/domain"
	"github.com/yertaypert/gym-booking-api/internal/repository"
)

func TestCreateGymValidation(t *testing.T) {
	u := &GymUsecase{}

	t.Run("missing owner_id", func(t *testing.T) {
		_, err := u.CreateGym(0, "Gym Name", "addr", "desc")
		if !errors.Is(err, ErrInvalidOwnerID) {
			t.Fatalf("expected ErrInvalidOwnerID, got %v", err)
		}
	})

	t.Run("missing name", func(t *testing.T) {
		_, err := u.CreateGym(1, "   ", "addr", "desc")
		if !errors.Is(err, ErrInvalidGymName) {
			t.Fatalf("expected ErrInvalidGymName, got %v", err)
		}
	})
}

// ─── CreateClass ─────────────────────────────────────────────────────────────

func TestCreateClassValidation(t *testing.T) {
	gym := &domain.Gym{ID: 1, OwnerID: 10}

	tests := []struct {
		name        string
		userID      int
		role        domain.UserRole
		className   string
		maxCapacity int
		wantErr     error
	}{
		{
			name:        "empty class name",
			userID:      10,
			role:        domain.RoleGymOwner,
			className:   "   ",
			maxCapacity: 20,
			wantErr:     ErrInvalidClassName,
		},
		{
			name:        "zero capacity",
			userID:      10,
			role:        domain.RoleGymOwner,
			className:   "Yoga",
			maxCapacity: 0,
			wantErr:     ErrInvalidMaxCapacity,
		},
		{
			name:        "negative capacity",
			userID:      10,
			role:        domain.RoleGymOwner,
			className:   "Yoga",
			maxCapacity: -5,
			wantErr:     ErrInvalidMaxCapacity,
		},
		{
			name:        "not owner",
			userID:      1, // gym принадлежит ownerID=10
			role:        domain.RoleGymOwner,
			className:   "Yoga",
			maxCapacity: 20,
			wantErr:     ErrNotGymOwner,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := newGymUsecase(gym, nil, nil, nil)
			_, err := uc.CreateClass(tt.userID, tt.role, 1, tt.className, tt.maxCapacity)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestCreateClass_GymNotFound(t *testing.T) {
	uc := newGymUsecase(nil, repository.ErrGymNotFound, nil, nil)
	_, err := uc.CreateClass(10, domain.RoleGymOwner, 1, "Yoga", 20)
	if !errors.Is(err, ErrGymNotFound) {
		t.Fatalf("expected ErrGymNotFound, got %v", err)
	}
}

func TestCreateClass_AdminCanCreateForAnyGym(t *testing.T) {
	gym := &domain.Gym{ID: 1, OwnerID: 99}
	created := &domain.Class{ID: 1, GymID: 1, Name: "Yoga", MaxCapacity: 20}
	uc := newGymUsecase(gym, nil, created, nil)

	got, err := uc.CreateClass(1, domain.RoleAdmin, 1, "Yoga", 20)
	if err != nil {
		t.Fatalf("admin should be able to create class, got: %v", err)
	}
	if got.Name != "Yoga" {
		t.Errorf("expected name=Yoga, got %s", got.Name)
	}
}

// ─── CreateSession ────────────────────────────────────────────────────────────

func TestCreateSessionValidation(t *testing.T) {
	gym := &domain.Gym{ID: 1, OwnerID: 10}
	class := &domain.Class{ID: 5, GymID: 1, MaxCapacity: 20}
	now := time.Now()

	tests := []struct {
		name      string
		userID    int
		role      domain.UserRole
		startTime time.Time
		endTime   time.Time
		price     float64
		wantErr   error
	}{
		{
			name:      "end before start",
			userID:    10,
			role:      domain.RoleGymOwner,
			startTime: now.Add(2 * time.Hour),
			endTime:   now,
			price:     500,
			wantErr:   ErrInvalidSessionTime,
		},
		{
			name:      "equal start and end",
			userID:    10,
			role:      domain.RoleGymOwner,
			startTime: now,
			endTime:   now,
			price:     500,
			wantErr:   ErrInvalidSessionTime,
		},
		{
			name:      "zero price",
			userID:    10,
			role:      domain.RoleGymOwner,
			startTime: now,
			endTime:   now.Add(time.Hour),
			price:     0,
			wantErr:   ErrInvalidSessionPrice,
		},
		{
			name:      "negative price",
			userID:    10,
			role:      domain.RoleGymOwner,
			startTime: now,
			endTime:   now.Add(time.Hour),
			price:     -100,
			wantErr:   ErrInvalidSessionPrice,
		},
		{
			name:      "not owner",
			userID:    1, // gym принадлежит ownerID=10
			role:      domain.RoleGymOwner,
			startTime: now,
			endTime:   now.Add(time.Hour),
			price:     500,
			wantErr:   ErrNotGymOwner,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			uc := newGymUsecase(gym, nil, class, nil)
			_, err := uc.CreateSession(tt.userID, tt.role, 1, 5, tt.startTime, tt.endTime, tt.price)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestCreateSession_GymNotFound(t *testing.T) {
	uc := newGymUsecase(nil, repository.ErrGymNotFound, nil, nil)
	now := time.Now()
	_, err := uc.CreateSession(10, domain.RoleGymOwner, 1, 5, now, now.Add(time.Hour), 500)
	if !errors.Is(err, ErrGymNotFound) {
		t.Fatalf("expected ErrGymNotFound, got %v", err)
	}
}

func TestCreateSession_ClassNotFound(t *testing.T) {
	gym := &domain.Gym{ID: 1, OwnerID: 10}
	// gymErr=nil (gym найден), classErr=ErrClassNotFound
	uc := newGymUsecase(gym, nil, nil, repository.ErrClassNotFound)
	now := time.Now()
	_, err := uc.CreateSession(10, domain.RoleGymOwner, 1, 5, now, now.Add(time.Hour), 500)
	if !errors.Is(err, ErrClassNotFound) {
		t.Fatalf("expected ErrClassNotFound, got %v", err)
	}
}

func TestCreateSession_ClassDoesNotBelongToGym(t *testing.T) {
	gym := &domain.Gym{ID: 1, OwnerID: 10}
	class := &domain.Class{ID: 5, GymID: 999, MaxCapacity: 20} // принадлежит другому gym
	uc := newGymUsecase(gym, nil, class, nil)
	now := time.Now()
	_, err := uc.CreateSession(10, domain.RoleGymOwner, 1, 5, now, now.Add(time.Hour), 500)
	if !errors.Is(err, ErrClassDoesNotBelongToGym) {
		t.Fatalf("expected ErrClassDoesNotBelongToGym, got %v", err)
	}
}

func TestCreateSession_Success(t *testing.T) {
	gym := &domain.Gym{ID: 1, OwnerID: 10}
	class := &domain.Class{ID: 5, GymID: 1, MaxCapacity: 20}
	uc := newGymUsecase(gym, nil, class, nil)
	now := time.Now()

	session, err := uc.CreateSession(10, domain.RoleGymOwner, 1, 5, now, now.Add(time.Hour), 500)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if session.AvailableSlots != 20 {
		t.Errorf("expected AvailableSlots=20 (от MaxCapacity класса), got %d", session.AvailableSlots)
	}
	if session.Status != "active" {
		t.Errorf("expected status=active, got %s", session.Status)
	}
	if session.Price != 500 {
		t.Errorf("expected price=500, got %f", session.Price)
	}
}
