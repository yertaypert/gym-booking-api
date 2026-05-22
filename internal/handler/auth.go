package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/yertaypert/gym-booking-api/internal/middleware"
	"github.com/yertaypert/gym-booking-api/internal/usecase"
)

type AuthHandler struct {
	usecase *usecase.AuthUsecase
}

func NewAuthHandler(u *usecase.AuthUsecase) *AuthHandler {
	return &AuthHandler{usecase: u}
}

type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	FullName string `json:"full_name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type UpdateRoleRequest struct {
	Role string `json:"role"`
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req RegisterRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, "invalid request", http.StatusBadRequest)
		return
	}

	err = h.usecase.Register(r.Context(), req.Email, req.Password, req.FullName)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidEmail) || errors.Is(err, usecase.ErrInvalidFullName) || errors.Is(err, usecase.ErrWeakPassword) {
			writeError(w, err.Error(), http.StatusBadRequest)
			return
		}

		if errors.Is(err, usecase.ErrEmailAlreadyExists) {
			writeError(w, err.Error(), http.StatusConflict)
			return
		}

		writeError(w, "internal server error", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("user created"))
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest

	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, "invalid request", http.StatusBadRequest)
		return
	}

	token, err := h.usecase.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		writeError(w, err.Error(), http.StatusUnauthorized)
		return
	}

	writeJSON(w, http.StatusOK, map[string]string{"token": token})
}

func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		writeError(w, "invalid user context", http.StatusUnauthorized)
		return
	}

	user, err := h.usecase.Me(r.Context(), userID)
	if err != nil {
		writeError(w, "user not found", http.StatusNotFound)
		return
	}

	writeJSON(w, http.StatusOK, user)
}

func (h *AuthHandler) UpdateUserRole(w http.ResponseWriter, r *http.Request) {
	targetUserIDStr := r.PathValue("id")
	if targetUserIDStr == "" {
		writeError(w, "user id is required", http.StatusBadRequest)
		return
	}

	targetUserID, err := strconv.Atoi(targetUserIDStr)
	if err != nil {
		writeError(w, "invalid user id", http.StatusBadRequest)
		return
	}

	var req UpdateRoleRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		writeError(w, "invalid request body", http.StatusBadRequest)
		return
	}

	err = h.usecase.UpdateUserRole(r.Context(), targetUserID, req.Role)
	if err != nil {
		if err.Error() == "invalid role" {
			writeError(w, err.Error(), http.StatusBadRequest)
			return
		}
		writeError(w, "failed to update role", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte("role updated"))
}
