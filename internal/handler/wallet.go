package handler

import (
	"encoding/json"
	"net/http"

	"github.com/yertaypert/gym-booking-api/internal/middleware"
	"github.com/yertaypert/gym-booking-api/internal/usecase"
)

type WalletHandler struct {
	walletUsecase *usecase.WalletUsecase
}

func NewWalletHandler(walletUsecase *usecase.WalletUsecase) *WalletHandler {
	return &WalletHandler{walletUsecase: walletUsecase}
}

type TopUpRequest struct {
	UserID int     `json:"user_id"`
	Amount float64 `json:"amount"`
}

func (h *WalletHandler) TopUp(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDKey).(int)
	if !ok {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
		return
	}

	w.Header().Set("Content-Type", "application/json")

	var req TopUpRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`{"error": "invalid request body"}`))
		return
	}

	err := h.walletUsecase.TopUpBalance(r.Context(), userID, req.Amount)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "` + err.Error() + `"}`))
		return
	}

	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "balance topped up successfully"}`))
}
