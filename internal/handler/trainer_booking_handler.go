package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/yertaypert/gym-booking-api/internal/middleware"
	"github.com/yertaypert/gym-booking-api/internal/usecase"
)

type TrainerBookingHandler struct {
	trainerBookingUsecase *usecase.TrainerBookingUsecase
}

func NewTrainerBookingHandler(trainerBookingUsecase *usecase.TrainerBookingUsecase) *TrainerBookingHandler {
	return &TrainerBookingHandler{
		trainerBookingUsecase: trainerBookingUsecase,
	}
}
func (h *TrainerBookingHandler) BookTrainerSlot(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "invalid user context", http.StatusUnauthorized)
		return
	}
	slotID, err := parsePathID(r, "id")
	if err != nil {
		http.Error(w, "invalid trainer slot id", http.StatusBadRequest)
		return
	}
	err = h.trainerBookingUsecase.BookTrainerSlot(r.Context(), userID, slotID)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrTrainerSlotNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, usecase.ErrTrainerSlotNotAvailable):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "failed to book trainer slot", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "trainer slot is booked successfully",
	})
}
func (h *TrainerBookingHandler) GetMyTrainerBookings(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "invalid user context", http.StatusUnauthorized)
		return
	}
	bookings, err := h.trainerBookingUsecase.GetMyTrainerBookings(r.Context(), userID)
	if err != nil {
		http.Error(w, "failed to get trainer bookings", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(bookings)
}
func (h *TrainerBookingHandler) CancelTrainerBooking(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "invalid user context", http.StatusUnauthorized)
		return
	}
	bookingID, err := parsePathID(r, "id")
	if err != nil {
		http.Error(w, "invalid trainer booking id", http.StatusBadRequest)
		return
	}
	err = h.trainerBookingUsecase.CancelTrainerBooking(r.Context(), userID, bookingID)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrTrainerBookingNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, usecase.ErrBookingForbidden):
			http.Error(w, err.Error(), http.StatusForbidden)
		default:
			http.Error(w, "failed to cancel trainer booking", http.StatusInternalServerError)
		}
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{
		"message": "trainer booking is canceled successfully",
	})
}
