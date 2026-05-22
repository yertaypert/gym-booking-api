package handler

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/yertaypert/gym-booking-api/internal/domain"
	"github.com/yertaypert/gym-booking-api/internal/middleware"
	"github.com/yertaypert/gym-booking-api/internal/usecase"
)

func TestParsePathID(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/gyms/12", nil)
	req.SetPathValue("id", "12")

	id, err := parsePathID(req, "id")
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if id != 12 {
		t.Fatalf("expected id 12, got %d", id)
	}
}

func TestParsePathIDInvalid(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/gyms/abc", nil)
	req.SetPathValue("id", "abc")

	_, err := parsePathID(req, "id")
	if err == nil {
		t.Fatal("expected error for invalid id")
	}
}

// mock repo for handler tests

type mockGymRepoForHandler struct {
	gym   *domain.Gym
	class *domain.Class
	err   error
}

func (m *mockGymRepoForHandler) GetGymByID(id int) (*domain.Gym, error)        { return m.gym, m.err }
func (m *mockGymRepoForHandler) GetClassByID(id int) (*domain.Class, error)    { return m.class, m.err }
func (m *mockGymRepoForHandler) CreateGym(gym domain.Gym) (*domain.Gym, error) { return m.gym, m.err }
func (m *mockGymRepoForHandler) CreateClass(c domain.Class) (*domain.Class, error) {
	return m.class, m.err
}
func (m *mockGymRepoForHandler) CreateSession(gymID int, s domain.Session) (*domain.Session, error) {
	return &s, m.err
}
func (m *mockGymRepoForHandler) ListGyms() ([]domain.Gym, error)                { return nil, m.err }
func (m *mockGymRepoForHandler) ListGymsByOwnerID(id int) ([]domain.Gym, error) { return nil, m.err }
func (m *mockGymRepoForHandler) ListClassesByGymID(id int) ([]domain.Class, error) {
	return nil, m.err
}
func (m *mockGymRepoForHandler) ListSessionsByGymAndClassID(gymID, classID int) ([]domain.Session, error) {
	return nil, nil
}
func (m *mockGymRepoForHandler) AssignTrainer(gymID, trainerID int) error { return nil }

// Helper

func newGymHandler(gym *domain.Gym, class *domain.Class, repoErr error) *GymHandler {
	repo := &mockGymRepoForHandler{gym: gym, class: class, err: repoErr}
	uc := usecase.NewGymUsecase(repo)
	return NewGymHandler(uc)
}

// GET /gyms/{id} 

func TestGetGym_NotFound(t *testing.T) {
	h := newGymHandler(nil, nil, usecase.ErrGymNotFound)

	req := httptest.NewRequest(http.MethodGet, "/gyms/1", nil)
	req.SetPathValue("id", "1")
	rr := httptest.NewRecorder()

	h.GetGym(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestGetGym_InvalidID(t *testing.T) {
	h := newGymHandler(nil, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/gyms/abc", nil)
	req.SetPathValue("id", "abc")
	rr := httptest.NewRecorder()

	h.GetGym(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestGetGym_Success(t *testing.T) {
	gym := &domain.Gym{ID: 1, Name: "FitGym", OwnerID: 10}
	h := newGymHandler(gym, nil, nil)

	req := httptest.NewRequest(http.MethodGet, "/gyms/1", nil)
	req.SetPathValue("id", "1")
	rr := httptest.NewRecorder()

	h.GetGym(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	if !strings.Contains(rr.Body.String(), "FitGym") {
		t.Errorf("expected body to contain FitGym, got: %s", rr.Body.String())
	}
}

// POST /gyms

func TestCreateGym_InvalidJSON(t *testing.T) {
	h := newGymHandler(nil, nil, nil)

	req := httptest.NewRequest(http.MethodPost, "/gyms", strings.NewReader("not json"))
	rr := httptest.NewRecorder()

	h.CreateGym(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestCreateGym_ValidationError(t *testing.T) {
	h := newGymHandler(nil, nil, nil)

	body := `{"owner_id": 0, "name": "Gym", "address": "addr"}`
	req := httptest.NewRequest(http.MethodPost, "/gyms", strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.CreateGym(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestCreateGym_Success(t *testing.T) {
	gym := &domain.Gym{ID: 1, Name: "FitGym", OwnerID: 5}
	h := newGymHandler(gym, nil, nil)

	body := `{"owner_id": 5, "name": "FitGym", "address": "Almaty"}`
	req := httptest.NewRequest(http.MethodPost, "/gyms", strings.NewReader(body))
	rr := httptest.NewRecorder()

	h.CreateGym(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}
}

// POST /gyms/{id}/classes

func TestCreateClass_Forbidden(t *testing.T) {
	gym := &domain.Gym{ID: 1, OwnerID: 99}
	h := newGymHandler(gym, nil, nil)

	body := `{"name": "Yoga", "max_capacity": 20}`
	req := httptest.NewRequest(http.MethodPost, "/gyms/1/classes", strings.NewReader(body))
	req.SetPathValue("id", "1")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, 1)
	ctx = context.WithValue(ctx, middleware.UserRoleKey, domain.RoleGymOwner)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	h.CreateClass(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

func TestCreateClass_Success(t *testing.T) {
	gym := &domain.Gym{ID: 1, OwnerID: 10}
	class := &domain.Class{ID: 1, GymID: 1, Name: "Yoga", MaxCapacity: 20}
	h := newGymHandler(gym, class, nil)

	body := `{"name": "Yoga", "max_capacity": 20}`
	req := httptest.NewRequest(http.MethodPost, "/gyms/1/classes", strings.NewReader(body))
	req.SetPathValue("id", "1")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, 10)
	ctx = context.WithValue(ctx, middleware.UserRoleKey, domain.RoleGymOwner)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	h.CreateClass(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}
}

// POST /gyms/{gymId}/classes/{classId}/sessions

func TestCreateSession_InvalidTimeFormat(t *testing.T) {
	gym := &domain.Gym{ID: 1, OwnerID: 10}
	h := newGymHandler(gym, nil, nil)

	body := `{"start_time": "not-a-time", "end_time": "also-not", "price": 500}`
	req := httptest.NewRequest(http.MethodPost, "/gyms/1/classes/5/sessions", strings.NewReader(body))
	req.SetPathValue("gymId", "1")
	req.SetPathValue("classId", "5")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, 10)
	ctx = context.WithValue(ctx, middleware.UserRoleKey, domain.RoleGymOwner)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	h.CreateSession(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestCreateSession_Success(t *testing.T) {
	gym := &domain.Gym{ID: 1, OwnerID: 10}
	class := &domain.Class{ID: 5, GymID: 1, MaxCapacity: 20}
	h := newGymHandler(gym, class, nil)

	start := time.Now().Add(time.Hour).UTC().Format(time.RFC3339)
	end := time.Now().Add(2 * time.Hour).UTC().Format(time.RFC3339)
	body := `{"start_time": "` + start + `", "end_time": "` + end + `", "price": 500}`

	req := httptest.NewRequest(http.MethodPost, "/gyms/1/classes/5/sessions", strings.NewReader(body))
	req.SetPathValue("gymId", "1")
	req.SetPathValue("classId", "5")
	ctx := context.WithValue(req.Context(), middleware.UserIDKey, 10)
	ctx = context.WithValue(ctx, middleware.UserRoleKey, domain.RoleGymOwner)
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	h.CreateSession(rr, req)

	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d", rr.Code)
	}
}

// adds userID and role in context request
func withUserContext(ctx context.Context, userID int, role domain.UserRole) context.Context {
	ctx = context.WithValue(ctx, middleware.UserIDKey, userID)
	ctx = context.WithValue(ctx, middleware.UserRoleKey, role)
	return ctx
}
