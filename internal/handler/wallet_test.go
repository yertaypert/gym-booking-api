package handler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// TopUp нет UserID в контексте - 401
func TestTopUp_NoUserContext(t *testing.T) {
	h := &WalletHandler{walletUsecase: nil}

	req := httptest.NewRequest(http.MethodPost, "/wallet/topup", nil)
	rr := httptest.NewRecorder()

	h.TopUp(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

// TopUp невалидный JSON - 400
func TestTopUp_InvalidJSON(t *testing.T) {
	h := &WalletHandler{walletUsecase: nil}

	req := httptest.NewRequest(http.MethodPost, "/wallet/topup", strings.NewReader("bad json"))
	ctx := withUserContext(req.Context(), 1, "user")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	h.TopUp(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
