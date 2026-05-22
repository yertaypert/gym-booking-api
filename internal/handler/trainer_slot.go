package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/yertaypert/gym-booking-api/internal/usecase"
)

type TrainerSlotHandler struct {
	trainerSlotUsecase *usecase.TrainerSlotUsecase
}

type CreateTrainerSlotRequest struct {
	TrainerID int       `json:"trainer_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
}

func NewTrainerSlotHandler(trainerSlotUsecase *usecase.TrainerSlotUsecase) *TrainerSlotHandler {
	return &TrainerSlotHandler{
		trainerSlotUsecase: trainerSlotUsecase,
	}
}

func (h *TrainerSlotHandler) ListAvailableSlots(w http.ResponseWriter, r *http.Request) {
	slots, err := h.trainerSlotUsecase.ListAvailableSlots(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(slots)
}
func (h *TrainerSlotHandler) CreateTrainerSlot(w http.ResponseWriter, r *http.Request) {
	var req CreateTrainerSlotRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}
	slot, err := h.trainerSlotUsecase.CreateTrainerSlot(r.Context(), req.TrainerID, req.StartTime, req.EndTime)
	if err != nil {
		http.Error(w, "failed to create trainer slot", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(slot)
}
