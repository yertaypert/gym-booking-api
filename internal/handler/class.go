package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/yertaypert/gym-booking-api/internal/usecase"
)

type ClassHandler struct {
	usecase *usecase.ClassUsecase
}

func NewClassHandler(u *usecase.ClassUsecase) *ClassHandler {
	return &ClassHandler{usecase: u}
}

func (h *ClassHandler) ListClasses(w http.ResponseWriter, r *http.Request) {
	classes, err := h.usecase.ListDistinctClasses(r.Context())
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(classes)
}

func (h *ClassHandler) SearchSessions(w http.ResponseWriter, r *http.Request) {
	name := r.PathValue("name")

	var startTime, endTime *time.Time

	if st := r.URL.Query().Get("start_time"); st != "" {
		t, err := time.Parse(time.RFC3339, st)
		if err != nil {
			http.Error(w, "invalid start_time format (RFC3339 required)", http.StatusBadRequest)
			return
		}
		startTime = &t
	}

	if et := r.URL.Query().Get("end_time"); et != "" {
		t, err := time.Parse(time.RFC3339, et)
		if err != nil {
			http.Error(w, "invalid end_time format (RFC3339 required)", http.StatusBadRequest)
			return
		}
		endTime = &t
	}

	// Handle "date" filter if provided
	if dateStr := r.URL.Query().Get("date"); dateStr != "" {
		d, err := time.Parse("2006-01-02", dateStr)
		if err != nil {
			http.Error(w, "invalid date format (YYYY-MM-DD required)", http.StatusBadRequest)
			return
		}

		// If date is provided, it overrides or sets the range to that specific day
		st := d
		et := d.Add(24 * time.Hour).Add(-time.Second)
		startTime = &st
		endTime = &et
	}

	sessions, err := h.usecase.SearchSessions(r.Context(), name, startTime, endTime)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(sessions)
}
