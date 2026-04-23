package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/yertaypert/gym-booking-api/internal/domain"
	"github.com/yertaypert/gym-booking-api/internal/middleware"
	"github.com/yertaypert/gym-booking-api/internal/usecase"
)

type BookingHandler struct {
	bookingUsecase *usecase.BookingUsecase
}

func NewBookingHandler(bookingUsecase *usecase.BookingUsecase) *BookingHandler {
	return &BookingHandler{bookingUsecase: bookingUsecase}
}

// ─── POST /sessions/{sessionId}/bookings ────────────────────────────────────

type CreateBookingResponse struct {
	BookingID int    `json:"booking_id"`
	Message   string `json:"message"`
}

func (h *BookingHandler) CreateBooking(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "invalid user context", http.StatusUnauthorized)
		return
	}
	sessionID, err := parsePathID(r, "sessionId")
	if err != nil {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}
	bookingID, err := h.bookingUsecase.CreateBooking(r.Context(), userID, sessionID)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrSessionNotActive), errors.Is(err, usecase.ErrSessionInPast):
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(CreateBookingResponse{BookingID: bookingID, Message: "Booking created successfully"})
}

// ─── POST /bookings/{bookingId}/cancel ──────────────────────────────────────

type CancelBookingResponse struct {
	Message string `json:"message"`
}

func (h *BookingHandler) CancelBooking(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "invalid user context", http.StatusUnauthorized)
		return
	}
	role, ok := r.Context().Value(middleware.UserRoleKey).(domain.UserRole)
	if !ok {
		http.Error(w, "invalid user role", http.StatusUnauthorized)
		return
	}
	bookingID, err := parsePathID(r, "bookingId")
	if err != nil {
		http.Error(w, "invalid booking id", http.StatusBadRequest)
		return
	}
	err = h.bookingUsecase.CancelBooking(r.Context(), userID, bookingID, role == domain.RoleAdmin)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrBookingNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, usecase.ErrBookingForbidden):
			http.Error(w, err.Error(), http.StatusForbidden)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(CancelBookingResponse{Message: "Booking cancelled successfully"})
}

// ─── POST /bookings/{bookingId}/attend  (admin only) ────────────────────────

func (h *BookingHandler) MarkAttended(w http.ResponseWriter, r *http.Request) {
	bookingID, err := parsePathID(r, "bookingId")
	if err != nil {
		http.Error(w, "invalid booking id", http.StatusBadRequest)
		return
	}
	err = h.bookingUsecase.MarkAttended(r.Context(), bookingID)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrBookingNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, usecase.ErrAlreadyAttended),
			errors.Is(err, usecase.ErrBookingNotConfirmed),
			errors.Is(err, usecase.ErrSessionNotStartedYet):
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, err.Error(), http.StatusBadRequest)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"message": "Attendance marked successfully"})
}

// ─── GET /users/me/bookings ──────────────────────────────────────────────────

func (h *BookingHandler) GetMyBookings(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "invalid user context", http.StatusUnauthorized)
		return
	}
	bookings, err := h.bookingUsecase.GetMyBookings(r.Context(), userID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(bookings)
}

// ─── GET /sessions/{sessionId}/bookings  (admin only) ───────────────────────

func (h *BookingHandler) GetSessionAttendees(w http.ResponseWriter, r *http.Request) {
	sessionID, err := parsePathID(r, "sessionId")
	if err != nil {
		http.Error(w, "invalid session id", http.StatusBadRequest)
		return
	}
	attendees, err := h.bookingUsecase.GetSessionAttendees(r.Context(), sessionID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(attendees)
}
