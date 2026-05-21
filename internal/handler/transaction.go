package handler

import (
	"encoding/json"
	"net/http"

	"github.com/yertaypert/gym-booking-api/internal/domain"
	"github.com/yertaypert/gym-booking-api/internal/usecase"
)

type TransactionHandler struct {
	usecase *usecase.TransactionUsecase
}

func NewTransactionHandler(u *usecase.TransactionUsecase) *TransactionHandler {
	return &TransactionHandler{usecase: u}
}

func (h *TransactionHandler) CreateTransaction(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID int     `json:"user_id"`
		Amount float64 `json:"amount"`
		Type   string  `json:"type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "invalid data format", http.StatusBadRequest)
		return
	}

	tx, err := h.usecase.CreateTransaction(r.Context(), req.UserID, req.Amount, req.Type)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(tx)
}

func (h *TransactionHandler) GetMyTransactions(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	userIDVal := r.Context().Value("userID")
	if userIDVal == nil {
		userIDVal = r.Context().Value("user_id")
	}

	userID, ok := userIDVal.(int)
	if !ok {
		w.WriteHeader(http.StatusUnauthorized)
		w.Write([]byte(`{"error": "unauthorized"}`))
		return
	}

	transactions, err := h.usecase.GetUserTransactions(r.Context(), userID)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "` + err.Error() + `"}`))
		return
	}

	if transactions == nil {
		transactions = []domain.Transaction{}
	}

	json.NewEncoder(w).Encode(transactions)
}
