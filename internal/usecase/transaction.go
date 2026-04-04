package usecase

import (
	"errors"
	"time"

	"github.com/yertaypert/gym-booking-api/internal/domain"
	"github.com/yertaypert/gym-booking-api/internal/repository"
)

type TransactionUsecase struct {
	repo *repository.TransactionRepository
}

func NewTransactionUsecase(repo *repository.TransactionRepository) *TransactionUsecase {
	return &TransactionUsecase{repo: repo}
}

func (u *TransactionUsecase) CreateTransaction(userID int, amount float64, txType string) (*domain.Transaction, error) {
	if amount <= 0 {
		return nil, errors.New("amount must be greater than zero")
	}

	tx := &domain.Transaction{
		UserID:    userID,
		Amount:    amount,
		Type:      txType,
		CreatedAt: time.Now(),
	}

	err := u.repo.Create(tx)
	if err != nil {
		return nil, err
	}

	return tx, nil
}
