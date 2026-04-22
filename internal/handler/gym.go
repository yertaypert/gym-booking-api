package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/yertaypert/gym-booking-api/internal/domain"
	"github.com/yertaypert/gym-booking-api/internal/middleware"
	"github.com/yertaypert/gym-booking-api/internal/usecase"
)

type GymHandler struct {
	usecase *usecase.GymUsecase
}

type CreateGymRequest struct {
	OwnerID     int    `json:"owner_id"`
	Name        string `json:"name"`
	Address     string `json:"address"`
	Description string `json:"description"`
}

type CreateClassRequest struct {
	Name        string `json:"name"`
	MaxCapacity int    `json:"max_capacity"`
}

type CreateSessionRequest struct {
	StartTime string  `json:"start_time"`
	EndTime   string  `json:"end_time"`
	Price     float64 `json:"price"`
}

func NewGymHandler(u *usecase.GymUsecase) *GymHandler {
	return &GymHandler{usecase: u}
}

func (h *GymHandler) ListGyms(w http.ResponseWriter, r *http.Request) {
	gyms, err := h.usecase.ListGyms()
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, gyms)
}

func (h *GymHandler) ListMyGyms(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "invalid user id", http.StatusUnauthorized)
		return
	}

	gyms, err := h.usecase.ListGymsByOwner(userID)
	if err != nil {
		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, gyms)
}

func (h *GymHandler) GetGym(w http.ResponseWriter, r *http.Request) {
	gymID, err := parsePathID(r, "id")
	if err != nil {
		http.Error(w, "invalid gym id", http.StatusBadRequest)
		return
	}

	gym, err := h.usecase.GetGym(gymID)
	if err != nil {
		if errors.Is(err, usecase.ErrGymNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, gym)
}

func (h *GymHandler) ListGymClasses(w http.ResponseWriter, r *http.Request) {
	gymID, err := parsePathID(r, "id")
	if err != nil {
		http.Error(w, "invalid gym id", http.StatusBadRequest)
		return
	}

	classes, err := h.usecase.ListGymClasses(gymID)
	if err != nil {
		if errors.Is(err, usecase.ErrGymNotFound) {
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}

		http.Error(w, "internal server error", http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, classes)
}

func (h *GymHandler) ListClassSessions(w http.ResponseWriter, r *http.Request) {
	gymID, err := parsePathID(r, "gymId")
	if err != nil {
		http.Error(w, "invalid gym id", http.StatusBadRequest)
		return
	}

	classID, err := parsePathID(r, "classId")
	if err != nil {
		http.Error(w, "invalid class id", http.StatusBadRequest)
		return
	}

	sessions, err := h.usecase.ListClassSessions(gymID, classID)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrGymNotFound), errors.Is(err, usecase.ErrClassNotFound), errors.Is(err, usecase.ErrClassDoesNotBelongToGym):
			http.Error(w, err.Error(), http.StatusNotFound)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, http.StatusOK, sessions)
}

func (h *GymHandler) CreateGym(w http.ResponseWriter, r *http.Request) {
	var req CreateGymRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	gym, err := h.usecase.CreateGym(req.OwnerID, req.Name, req.Address, req.Description)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidGymName), errors.Is(err, usecase.ErrInvalidOwnerID):
			http.Error(w, err.Error(), http.StatusBadRequest)
		case errors.Is(err, usecase.ErrGymAlreadyExists):
			http.Error(w, err.Error(), http.StatusConflict)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, http.StatusCreated, gym)
}

func (h *GymHandler) CreateClass(w http.ResponseWriter, r *http.Request) {
	gymID, err := parsePathID(r, "id")
	if err != nil {
		http.Error(w, "invalid gym id", http.StatusBadRequest)
		return
	}

	userID, _ := r.Context().Value(middleware.UserIDKey).(int)
	userRole, _ := r.Context().Value(middleware.UserRoleKey).(domain.UserRole)

	var req CreateClassRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	class, err := h.usecase.CreateClass(userID, userRole, gymID, req.Name, req.MaxCapacity)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrGymNotFound):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, usecase.ErrNotGymOwner):
			http.Error(w, err.Error(), http.StatusForbidden)
		case errors.Is(err, usecase.ErrInvalidClassName), errors.Is(err, usecase.ErrInvalidMaxCapacity):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, http.StatusCreated, class)
}

func (h *GymHandler) CreateSession(w http.ResponseWriter, r *http.Request) {
	gymID, err := parsePathID(r, "gymId")
	if err != nil {
		http.Error(w, "invalid gym id", http.StatusBadRequest)
		return
	}

	classID, err := parsePathID(r, "classId")
	if err != nil {
		http.Error(w, "invalid class id", http.StatusBadRequest)
		return
	}

	userID, _ := r.Context().Value(middleware.UserIDKey).(int)
	userRole, _ := r.Context().Value(middleware.UserRoleKey).(domain.UserRole)

	var req CreateSessionRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid request", http.StatusBadRequest)
		return
	}

	startTime, err := time.Parse(time.RFC3339, req.StartTime)
	if err != nil {
		http.Error(w, "start_time must be RFC3339", http.StatusBadRequest)
		return
	}

	endTime, err := time.Parse(time.RFC3339, req.EndTime)
	if err != nil {
		http.Error(w, "end_time must be RFC3339", http.StatusBadRequest)
		return
	}

	session, err := h.usecase.CreateSession(userID, userRole, gymID, classID, startTime, endTime, req.Price)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrGymNotFound), errors.Is(err, usecase.ErrClassNotFound), errors.Is(err, usecase.ErrClassDoesNotBelongToGym):
			http.Error(w, err.Error(), http.StatusNotFound)
		case errors.Is(err, usecase.ErrNotGymOwner):
			http.Error(w, err.Error(), http.StatusForbidden)
		case errors.Is(err, usecase.ErrInvalidSessionTime), errors.Is(err, usecase.ErrInvalidSessionPrice):
			http.Error(w, err.Error(), http.StatusBadRequest)
		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	writeJSON(w, http.StatusCreated, session)
}

func parsePathID(r *http.Request, name string) (int, error) {
	return strconv.Atoi(r.PathValue(name))
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(v)
}
