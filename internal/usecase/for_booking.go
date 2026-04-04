package usecase

import (
	"context"
	"database/sql"
	"errors"

	"github.com/yertaypert/gym-booking-api/internal/domain"
	"github.com/yertaypert/gym-booking-api/internal/repository"
)

type ForBooking struct {
	db          *sql.DB
	bookingRepo repository.BookingRepository
	walletRepo  repository.WalletRepository
	userRepo    repository.UserRepository
	sessionRepo repository.SessionRepository
}

func NewBookingUsecase(
	db *sql.DB,
	bookingRepo repository.BookingRepository,
	walletRepo repository.WalletRepository,
	userRepo repository.UserRepository,
	sessionRepo repository.SessionRepository,
) *ForBooking {
	return &ForBooking{
		db:          db,
		bookingRepo: bookingRepo,
		walletRepo:  walletRepo,
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}
func (u *ForBooking) CreateBooking(ctx context.Context, userID, sessionID int) (int, error) {
	user, err := u.userRepo.GetByID(ctx, userID)
	session, err := u.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return 0, err
	}
	if session.AvailableSlots <= 0 {
		return 0, errors.New("No available slots")
	}
	if user.Balance < session.Price {
		return 0, errors.New("Payment failed")
	}
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	bookingID, err := u.bookingRepo.Create(ctx, tx, userID, sessionID)
	if err != nil {
		return 0, err
	}
	err = u.walletRepo.UpdateBalance(ctx, tx, userID, -session.Price)
	if err != nil {
		return 0, err
	}
	err = u.sessionRepo.DecreaseAvailableSlots(ctx, tx, sessionID)
	if err != nil {
		return 0, err
	}
	err = u.bookingRepo.UpdateStatus(ctx, tx, bookingID, "confirmed") // если оплата удалась, меняем pending на confirmed
	if err != nil {
		return 0, err
	}
	err = u.walletRepo.CreateTransaction(
		ctx,
		tx,
		userID,
		&bookingID,
		-session.Price,
		domain.TransactionTypeBooking,
	)
	if err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return bookingID, nil
}
func (u *ForBooking) CancelBooking(ctx context.Context, bookingID int) error {
	booking, err := u.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		return err
	}
	if booking.Status == "cancelled" {
		return errors.New("Booking already cancelled")
	}
	session, err := u.sessionRepo.GetByID(ctx, booking.SessionID)
	if err != nil {
		return err
	}
	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	err = u.bookingRepo.UpdateStatus(ctx, tx, bookingID, "cancelled")
	if err != nil {
		return err
	}

	err = u.walletRepo.UpdateBalance(ctx, tx, booking.UserID, session.Price)
	if err != nil {
		return err
	}

	err = u.walletRepo.CreateTransaction(
		ctx,
		tx,
		booking.UserID,
		&bookingID,
		session.Price,
		domain.TransactionTypeRefund,
	)
	if err != nil {
		return err
	}

	err = u.sessionRepo.IncreaseAvailableSlots(ctx, tx, booking.SessionID)
	if err != nil {
		return err
	}
	return tx.Commit()
}
