package usecase

import (
	"database/sql"

	"github.com/yertaypert/gym-booking-api/internal/domain"
)

type WalletUsecase struct {
	db         *sql.DB
	walletRepo WalletRepository
}

func NewWalletUsecase(db *sql.DB, walletRepo WalletRepository) *WalletUsecase {
	return &WalletUsecase{
		db:         db,
		walletRepo: walletRepo,
	}
}

func (u *WalletUsecase) TopUpBalance(userID int, amount float64) error {
	if amount <= 0 {
		return domain.ErrInvalidAmount
	}

	tx, err := u.db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = u.walletRepo.UpdateBalance(tx, userID, amount)
	if err != nil {
		return err
	}

	err = u.walletRepo.CreateTransaction(tx, userID, nil, amount, "top_up")
	if err != nil {
		return err
	}

	return tx.Commit()
}
