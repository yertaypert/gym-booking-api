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
		writeError(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		UserID int     `json:"user_id"`
		Amount float64 `json:"amount"`
		Type   string  `json:"type"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, "invalid data format", http.StatusBadRequest)
		return
	}

	tx, err := h.usecase.CreateTransaction(r.Context(), req.UserID, req.Amount, req.Type)
	if err != nil {
		writeError(w, err.Error(), http.StatusInternalServerError)
		return
	}

	writeJSON(w, http.StatusOK, tx)
}

func (h *TransactionHandler) GetMyTransactions(w http.ResponseWriter, r *http.Request) {
	userIDVal := r.Context().Value("userID")
	if userIDVal == nil {
		userIDVal = r.Context().Value("user_id")
	}

	userID, ok := userIDVal.(int)
	if !ok {
		writeJSON(w, http.StatusUnauthorized, map[string]string{"error": "unauthorized"})
		return
	}

	transactions, err := h.usecase.GetUserTransactions(r.Context(), userID)
	if err != nil {
		writeJSON(w, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	if transactions == nil {
		transactions = []domain.Transaction{}
	}

	writeJSON(w, http.StatusOK, transactions)
}
