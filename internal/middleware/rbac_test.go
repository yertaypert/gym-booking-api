package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/yertaypert/gym-booking-api/internal/domain"
)

func TestRequireRolesAllowsMatchingRole(t *testing.T) {
	nextCalled := false
	handler := RequireRoles(domain.RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		nextCalled = true
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin/me", nil)
	req = req.WithContext(context.WithValue(req.Context(), UserRoleKey, domain.RoleAdmin))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rr.Code)
	}
	if !nextCalled {
		t.Fatal("expected next handler to be called")
	}
}

func TestRequireRolesRejectsNonMatchingRole(t *testing.T) {
	handler := RequireRoles(domain.RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/admin/me", nil)
	req = req.WithContext(context.WithValue(req.Context(), UserRoleKey, domain.RoleUser))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, rr.Code)
	}
}

// TestRequireRoles_NoRoleInContext — нет ключа в контексте вообще
func TestRequireRoles_NoRoleInContext(t *testing.T) {
	handler := RequireRoles(domain.RoleAdmin)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	// контекст пустой — UserRoleKey не установлен
	req := httptest.NewRequest(http.MethodGet, "/admin/me", nil)
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", rr.Code)
	}
}

// TestRequireRoles_MultipleRoles_FirstMatches — роль есть в списке (первая)
func TestRequireRoles_MultipleRoles_FirstMatches(t *testing.T) {
	handler := RequireRoles(domain.RoleAdmin, domain.RoleGymOwner)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), UserRoleKey, domain.RoleAdmin))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

// TestRequireRoles_MultipleRoles_SecondMatches — роль есть в списке (вторая)
func TestRequireRoles_MultipleRoles_SecondMatches(t *testing.T) {
	handler := RequireRoles(domain.RoleAdmin, domain.RoleGymOwner)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req = req.WithContext(context.WithValue(req.Context(), UserRoleKey, domain.RoleGymOwner))
	rr := httptest.NewRecorder()

	handler.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}
