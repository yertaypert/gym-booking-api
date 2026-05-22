package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Register невалидный JSON - 400
func TestRegister_InvalidJSON(t *testing.T) {
	h := &AuthHandler{usecase: nil}

	req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader("not json"))
	rr := httptest.NewRecorder()

	h.Register(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

// Login невалидный JSON - 400
func TestLogin_InvalidJSON(t *testing.T) {
	h := &AuthHandler{usecase: nil}

	req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("{bad json"))
	rr := httptest.NewRecorder()

	h.Login(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

// Me нет UserID в контексте - 401
func TestMe_NoUserContext(t *testing.T) {
	h := &AuthHandler{usecase: nil}

	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	rr := httptest.NewRecorder()

	h.Me(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}
