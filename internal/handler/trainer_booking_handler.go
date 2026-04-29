package handler

import (
	"encoding/json"
	"errors"
	"github.com/yertaypert/gym-booking-api/internal/middleware"
	"github.com/yertaypert/gym-booking-api/internal/usecase"
	"net/http"
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
