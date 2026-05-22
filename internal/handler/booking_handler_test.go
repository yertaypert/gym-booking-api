package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// CreateBooking нет UserID в контексте - 401
func TestCreateBooking_NoUserContext(t *testing.T) {
	h := &BookingHandler{bookingUsecase: nil}

	req := httptest.NewRequest(http.MethodPost, "/sessions/1/bookings", nil)
	rr := httptest.NewRecorder()

	h.CreateBooking(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

// CancelBooking невалидный bookingId - 400
func TestCancelBooking_InvalidID(t *testing.T) {
	h := &BookingHandler{bookingUsecase: nil}

	req := httptest.NewRequest(http.MethodPost, "/bookings/abc/cancel", nil)
	req.SetPathValue("bookingId", "abc")
	ctx := withUserContext(req.Context(), 1, "user")
	req = req.WithContext(ctx)
	rr := httptest.NewRecorder()

	h.CancelBooking(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

// MarkAttended невалидный bookingId - 400
func TestMarkAttended_InvalidID(t *testing.T) {
	h := &BookingHandler{bookingUsecase: nil}

	req := httptest.NewRequest(http.MethodPost, "/bookings/xyz/attend", nil)
	req.SetPathValue("bookingId", "xyz")
	rr := httptest.NewRecorder()

	h.MarkAttended(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}
