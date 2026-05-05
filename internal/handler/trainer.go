package handler

import (
	"encoding/json"
	"net/http"

	"github.com/yertaypert/gym-booking-api/internal/usecase"
)

type TrainerHandler struct {
	trainerUsecase *usecase.TrainerUsecase
}

func NewTrainerHandler(tu *usecase.TrainerUsecase) *TrainerHandler {
	return &TrainerHandler{trainerUsecase: tu}
}

func (h *TrainerHandler) PromoteToTrainer(w http.ResponseWriter, r *http.Request) {
	userID, err := parsePathID(r, "id")
	if err != nil {
		http.Error(w, "invalid user id", http.StatusBadRequest)
		return
	}

	var req struct {
		Specialization string  `json:"specialization"`
		ExtraFee       float64 `json:"extra_fee"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err = h.trainerUsecase.PromoteToTrainer(r.Context(), userID, req.Specialization, req.ExtraFee)
	if err != nil {
		HandleError(w, err)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{"message": "user promoted to trainer successfully"})
}
