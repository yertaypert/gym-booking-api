package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/yertaypert/gym-booking-api/internal/usecase"
)

type ClassSessionHandler struct {
	usecase *usecase.ClassSessionUsecase
}

func NewClassSessionHandler(u *usecase.ClassSessionUsecase) *ClassSessionHandler {
	return &ClassSessionHandler{usecase: u}
}

func (h *ClassSessionHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		GymID     int       `json:"gym_id"`
		Title     string    `json:"title"`
		StartTime time.Time `json:"start_time"`
		EndTime   time.Time `json:"end_time"`
		Capacity  int       `json:"capacity"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid data format", http.StatusBadRequest)
		return
	}

	session, err := h.usecase.CreateSession(req.GymID, req.Title, req.StartTime, req.EndTime, req.Capacity)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(session)
}
