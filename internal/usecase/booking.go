package usecase

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/yertaypert/gym-booking-api/internal/domain"
)

var (
	ErrBookingNotFound      = errors.New("booking not found")
	ErrBookingForbidden     = errors.New("booking does not belong to user")
	ErrSessionNotActive     = errors.New("session is not active")
	ErrSessionInPast        = errors.New("cannot book a session that has already ended")
	ErrAlreadyAttended      = errors.New("attendance already marked for this booking")
	ErrSessionNotStartedYet = errors.New("session has not started yet — cannot mark attendance")
	ErrBookingNotConfirmed  = errors.New("only confirmed bookings can be marked as attended")
)

type BookingUsecase struct {
	db          *sql.DB
	bookingRepo BookingRepository
	walletRepo  WalletRepository
	userRepo    UserRepository
	sessionRepo SessionRepository
	gymRepo     GymRepository
}

func NewBookingUsecase(
	db *sql.DB,
	bookingRepo BookingRepository,
	walletRepo WalletRepository,
	userRepo UserRepository,
	sessionRepo SessionRepository,
	gymRepo GymRepository,
) *BookingUsecase {
	return &BookingUsecase{
		db:          db,
		bookingRepo: bookingRepo,
		walletRepo:  walletRepo,
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		gymRepo:     gymRepo,
	}
}

// CreateBooking registers a user for a session, deducts balance, and records
// the transaction — all inside a single DB transaction.

func (u *BookingUsecase) ListGymBookings(ctx context.Context, userID int, userRole domain.UserRole, gymID int) ([]domain.Booking, error) {
	gym, err := u.gymRepo.GetGymByID(gymID)
	if err != nil {
		return nil, err
	}

	if userRole != domain.RoleAdmin && gym.OwnerID != userID {
		return nil, ErrBookingForbidden
	}

	return u.bookingRepo.ListByGymID(ctx, gymID)
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

	// Business rules —————————————————————————————————————————————————————————
	if session.Status != "active" {
		return 0, ErrSessionNotActive
	}
	if time.Now().After(session.EndTime) {
		return 0, ErrSessionInPast
	}
	if session.AvailableSlots <= 0 {
		return 0, errors.New("no available slots")
	}
	if user.Balance < session.Price {
		return 0, errors.New("insufficient balance")
	}

	// Explicit duplicate check before hitting the DB constraint so we can
	// return a friendlier error code in the handler.
	duplicate, err := u.bookingRepo.ExistsByUserAndSession(ctx, userID, sessionID)
	if err != nil {
		return 0, err
	}
	if duplicate {
		return 0, errors.New("you are already booked for this session")
	}
	// —————————————————————————————————————————————————————————————————————————

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

	if err = u.walletRepo.CreateTransaction(tx, userID, &bookingID, -session.Price, string(domain.TransactionTypePayment)); err != nil {
		return 0, err
	}

	if err = tx.Commit(); err != nil {
		return 0, err
	}

	return bookingID, nil
}

// CancelBooking cancels a booking and refunds the user's balance.
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
	if booking.Status == "attended" {
		return errors.New("attended bookings cannot be cancelled")
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

	if err = u.walletRepo.CreateTransaction(tx, booking.UserID, &bookingID, session.Price, string(domain.TransactionTypeRefund)); err != nil {
		return err
	}

	if err = u.sessionRepo.IncreaseAvailableSlots(ctx, tx, booking.SessionID); err != nil {
		return err
	}

	return tx.Commit()
}
func (u *BookingUsecase) GetUserBookings(ctx context.Context, userID int) ([]domain.Booking, error) {
	return u.bookingRepo.GetByUserID(ctx, userID)
}

// MarkAttended marks a booking as attended.  Only admins (or trainers) should
// call this endpoint — the handler enforces the role.
// Rules:
//   - Booking must be in "confirmed" state.
//   - The session must have already started (start_time <= now).
func (u *BookingUsecase) MarkAttended(ctx context.Context, bookingID int) error {
	booking, err := u.bookingRepo.GetByID(ctx, bookingID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return ErrBookingNotFound
		}
		return err
	}

	if booking.Status == "attended" {
		return ErrAlreadyAttended
	}
	if booking.Status != "confirmed" {
		return ErrBookingNotConfirmed
	}

	session, err := u.sessionRepo.GetByID(ctx, booking.SessionID)
	if err != nil {
		return err
	}
	if time.Now().Before(session.StartTime) {
		return ErrSessionNotStartedYet
	}

	tx, err := u.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	if err = u.bookingRepo.MarkAttended(ctx, tx, bookingID); err != nil {
		return err
	}

	return tx.Commit()
}

// GetMyBookings returns all bookings (with session detail) for the calling user.
func (u *BookingUsecase) GetMyBookings(ctx context.Context, userID int) ([]domain.BookingDetail, error) {
	return u.bookingRepo.GetDetailsByUserID(ctx, userID)
}

// GetSessionAttendees returns all bookings for a session.  Intended for admin/trainer use.
func (u *BookingUsecase) GetSessionAttendees(ctx context.Context, sessionID int) ([]domain.BookingDetail, error) {
	// Verify session exists first.
	if _, err := u.sessionRepo.GetByID(ctx, sessionID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, errors.New("session not found")
		}
		return nil, err
	}
	return u.bookingRepo.GetBySessionID(ctx, sessionID)
}
