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
	return &BookingHandler{
		bookingUsecase: bookingUsecase,
	}
}

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
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	response := CreateBookingResponse{
		BookingID: bookingID,
		Message:   "Booking created successfully",
	}
	_ = json.NewEncoder(w).Encode(response)
}

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
	w.WriteHeader(http.StatusOK)
	response := CancelBookingResponse{
		Message: "Booking cancelled successfully",
	}

	_ = json.NewEncoder(w).Encode(response)
}

func (h *BookingHandler) ListGymBookings(w http.ResponseWriter, r *http.Request) {
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

	gymID, err := parsePathID(r, "id")
	if err != nil {
		http.Error(w, "invalid gym id", http.StatusBadRequest)
		return
	}

	bookings, err := h.bookingUsecase.ListGymBookings(r.Context(), userID, role, gymID)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrBookingForbidden):
			http.Error(w, err.Error(), http.StatusForbidden)
		default:
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(bookings)
}
