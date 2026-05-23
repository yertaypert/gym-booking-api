package usecase

import (
	"context"
	"database/sql"
	"errors"
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

func (u *WalletUsecase) TopUpBalance(ctx context.Context, userID int, amount float64) error {
	if amount <= 0 {
		return errors.New("amount must be greater than zero")
	}

	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = u.walletRepo.UpdateBalance(ctx, tx, userID, amount)
	if err != nil {
		return err
	}

	err = u.walletRepo.CreateTransaction(ctx, tx, userID, nil, amount, "top_up")
	if err != nil {
		return err
	}

	return tx.Commit()
}
