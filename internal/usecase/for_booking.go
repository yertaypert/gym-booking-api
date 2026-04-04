package usecase

import (
	"context"
	"database/sql"
	"errors"

	"github.com/yertaypert/gym-booking-api/internal/domain"
	"github.com/yertaypert/gym-booking-api/internal/repository"
)

var ErrBookingNotFound = errors.New("booking not found")
var ErrBookingForbidden = errors.New("booking does not belong to user")

type BookingUsecase struct {
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
) *BookingUsecase {
	return &BookingUsecase{
		db:          db,
		bookingRepo: bookingRepo,
		walletRepo:  walletRepo,
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
}

func (u *BookingUsecase) CreateBooking(ctx context.Context, userID, sessionID int) (int, error) {
	user, err := u.userRepo.GetByID(userID)
	if err != nil {
		return 0, err
	}
	session, err := u.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return 0, err
	}

	if session.AvailableSlots <= 0 {
		return 0, errors.New("no available slots")
	}
	if user.Balance < session.Price {
		return 0, errors.New("insufficient balance")
	}

	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()

	bookingID, err := u.bookingRepo.Create(tx, userID, sessionID)
	if err != nil {
		return 0, err
	}

	if err = u.walletRepo.UpdateBalance(tx, userID, -session.Price); err != nil {
		return 0, err
	}

	if err = u.sessionRepo.DecreaseAvailableSlots(ctx, tx, sessionID); err != nil {
		return 0, err
	}

	if err = u.bookingRepo.UpdateStatus(ctx, tx, bookingID, "confirmed"); err != nil {
		return 0, err
	}

	if err = u.walletRepo.CreateTransaction(tx, userID, bookingID, -session.Price, string(domain.TransactionTypePayment)); err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return bookingID, nil
}

func (u *BookingUsecase) CancelBooking(ctx context.Context, requesterID, bookingID int, isAdmin bool) error {
	booking, err := u.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrBookingNotFound
		}
		return err
	}
	if booking.UserID != requesterID && !isAdmin {
		return ErrBookingForbidden
	}
	if booking.Status == "cancelled" {
		return errors.New("booking already cancelled")
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

	if err = u.bookingRepo.UpdateStatus(ctx, tx, bookingID, "cancelled"); err != nil {
		return err
	}

	if err = u.walletRepo.UpdateBalance(tx, booking.UserID, session.Price); err != nil {
		return err
	}

	if err = u.walletRepo.CreateTransaction(tx, booking.UserID, bookingID, session.Price, string(domain.TransactionTypeRefund)); err != nil {
		return err
	}

	if err = u.sessionRepo.IncreaseAvailableSlots(ctx, tx, booking.SessionID); err != nil {
		return err
	}

	return tx.Commit()
}
