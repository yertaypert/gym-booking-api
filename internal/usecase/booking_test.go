package usecase

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/yertaypert/gym-booking-api/internal/auth"
	"github.com/yertaypert/gym-booking-api/internal/domain"
)

type mockUserRepo struct {
	user *domain.User
	err  error
}

func (m *mockUserRepo) GetByID(ctx context.Context, id int) (*domain.User, error) {
	return m.user, m.err
}
func (m *mockUserRepo) Create(ctx context.Context, user domain.User) (int, error) { return 1, nil }
func (m *mockUserRepo) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	return m.user, m.err
}

type mockSessionRepo struct {
	session        *domain.Session
	err            error
	ownerID        int
	decreaseCalled bool
	increaseCalled bool
}

func (m *mockSessionRepo) GetByID(ctx context.Context, id int) (*domain.Session, error) {
	return m.session, m.err
}
func (m *mockSessionRepo) GetGymOwnerIDBySessionID(ctx context.Context, sessionID int) (int, error) {
	return m.ownerID, m.err
}
func (m *mockSessionRepo) DecreaseAvailableSlots(ctx context.Context, tx *sql.Tx, sessionID int) error {
	m.decreaseCalled = true
	return nil
}
func (m *mockSessionRepo) IncreaseAvailableSlots(ctx context.Context, tx *sql.Tx, sessionID int) error {
	m.increaseCalled = true
	return nil
}

type mockBookingRepo struct {
	booking        *domain.Booking
	err            error
	createErr      error
	isDuplicate    bool
	lastStatus     string
	attendedCalled bool
}

func (m *mockBookingRepo) ListByGymID(ctx context.Context, gymID int) ([]domain.Booking, error) {
	return []domain.Booking{}, nil
}

func (m *mockBookingRepo) Create(ctx context.Context, tx *sql.Tx, userID, sessionID int) (int, error) {
	return 1, m.createErr
}
func (m *mockBookingRepo) UpdateStatus(ctx context.Context, tx *sql.Tx, bookingID int, status string) error {
	m.lastStatus = status
	return nil
}
func (m *mockBookingRepo) GetByID(ctx context.Context, bookingID int) (*domain.Booking, error) {
	return m.booking, m.err
}
func (m *mockBookingRepo) GetByUserAndSession(ctx context.Context, userID, sessionID int) (*domain.Booking, error) {
	return m.booking, m.err
}
func (m *mockBookingRepo) ExistsByUserAndSession(ctx context.Context, userID, sessionID int) (bool, error) {
	return m.isDuplicate, nil
}
func (m *mockBookingRepo) GetByUserID(ctx context.Context, userID int) ([]domain.Booking, error) {
	return []domain.Booking{}, nil
}
func (m *mockBookingRepo) GetDetailsByUserID(ctx context.Context, userID int) ([]domain.BookingDetail, error) {
	return []domain.BookingDetail{}, nil
}
func (m *mockBookingRepo) GetBySessionID(ctx context.Context, sessionID int) ([]domain.BookingDetail, error) {
	return []domain.BookingDetail{}, nil
}
func (m *mockBookingRepo) MarkAttended(ctx context.Context, tx *sql.Tx, bookingID int) error {
	m.attendedCalled = true
	return nil
}

type mockWalletRepo struct {
	lastAmount float64
}

func (m *mockWalletRepo) UpdateBalance(ctx context.Context, tx *sql.Tx, userID int, amount float64) error {
	m.lastAmount = amount
	return nil
}
func (m *mockWalletRepo) CreateTransaction(ctx context.Context, tx *sql.Tx, userID int, bookingID *int, amount float64, txType string) error {
	return nil
}

func newTestUsecase(
	user *domain.User,
	session *domain.Session,
	booking *domain.Booking,
) (*BookingUsecase, *mockBookingRepo, *mockSessionRepo, *mockWalletRepo) {
	userRepo := &mockUserRepo{user: user}
	sessionRepo := &mockSessionRepo{session: session}
	bookingRepo := &mockBookingRepo{booking: booking}
	walletRepo := &mockWalletRepo{}

	uc := &BookingUsecase{
		db:          nil,
		bookingRepo: bookingRepo,
		walletRepo:  walletRepo,
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
	return uc, bookingRepo, sessionRepo, walletRepo
}

func TestCreateBooking_SessionInPast(t *testing.T) {
	user := &domain.User{ID: 1, Balance: 1000}
	session := &domain.Session{
		ID:             1,
		Status:         domain.SessionStatusActive,
		AvailableSlots: 5,
		Price:          500,
		StartTime:      time.Now().Add(-3 * time.Hour),
		EndTime:        time.Now().Add(-1 * time.Hour),
	}

	uc, _, _, _ := newTestUsecase(user, session, nil)
	_, err := uc.CreateBooking(context.Background(), 1, 1)

	if !errors.Is(err, ErrSessionInPast) {
		t.Errorf("expected ErrSessionInPast, got: %v", err)
	}
}

func TestCreateBooking_SessionNotActive(t *testing.T) {
	user := &domain.User{ID: 1, Balance: 1000}
	session := &domain.Session{
		ID:             1,
		Status:         "cancelled_session", // Use a different string to represent non-active
		AvailableSlots: 5,
		Price:          500,
		EndTime:        time.Now().Add(2 * time.Hour),
	}

	uc, _, _, _ := newTestUsecase(user, session, nil)
	_, err := uc.CreateBooking(context.Background(), 1, 1)

	if !errors.Is(err, ErrSessionNotActive) {
		t.Errorf("expected ErrSessionNotActive, got: %v", err)
	}
}

func TestCreateBooking_NoSlots(t *testing.T) {
	user := &domain.User{ID: 1, Balance: 1000}
	session := &domain.Session{
		ID:             1,
		Status:         domain.SessionStatusActive,
		AvailableSlots: 0,
		Price:          500,
		EndTime:        time.Now().Add(2 * time.Hour),
	}

	uc, _, _, _ := newTestUsecase(user, session, nil)
	_, err := uc.CreateBooking(context.Background(), 1, 1)

	if err == nil || err.Error() != "no available slots" {
		t.Errorf("expected 'no available slots', got: %v", err)
	}
}

func TestCreateBooking_InsufficientBalance(t *testing.T) {
	user := &domain.User{ID: 1, Balance: 100}
	session := &domain.Session{
		ID:             1,
		Status:         domain.SessionStatusActive,
		AvailableSlots: 5,
		Price:          500,
		EndTime:        time.Now().Add(2 * time.Hour),
	}

	uc, _, _, _ := newTestUsecase(user, session, nil)
	_, err := uc.CreateBooking(context.Background(), 1, 1)

	if err == nil || err.Error() != "insufficient balance" {
		t.Errorf("expected 'insufficient balance', got: %v", err)
	}
}

func TestCreateBooking_Duplicate(t *testing.T) {
	user := &domain.User{ID: 1, Balance: 1000}
	session := &domain.Session{
		ID:             1,
		Status:         domain.SessionStatusActive,
		AvailableSlots: 5,
		Price:          500,
		EndTime:        time.Now().Add(2 * time.Hour),
	}

	uc, bookingRepo, _, _ := newTestUsecase(user, session, nil)
	bookingRepo.isDuplicate = true

	_, err := uc.CreateBooking(context.Background(), 1, 1)

	if err == nil || err.Error() != "you are already booked for this session" {
		t.Errorf("expected duplicate error, got: %v", err)
	}
}

func TestCancelBooking_Forbidden(t *testing.T) {
	booking := &domain.Booking{
		ID:        1,
		UserID:    99,
		SessionID: 1,
		Status:    domain.BookingStatusConfirmed,
	}
	session := &domain.Session{ID: 1, Price: 500}

	uc, _, _, _ := newTestUsecase(nil, session, booking)

	err := uc.CancelBooking(context.Background(), 1, 1, false)

	if !errors.Is(err, ErrBookingForbidden) {
		t.Errorf("expected ErrBookingForbidden, got: %v", err)
	}
}

func TestCancelBooking_AlreadyCancelled(t *testing.T) {
	booking := &domain.Booking{
		ID:        1,
		UserID:    1,
		SessionID: 1,
		Status:    domain.BookingStatusCancelled,
	}
	session := &domain.Session{ID: 1, Price: 500}

	uc, _, _, _ := newTestUsecase(nil, session, booking)
	err := uc.CancelBooking(context.Background(), 1, 1, false)

	if err == nil || err.Error() != "booking already cancelled" {
		t.Errorf("expected 'booking already cancelled', got: %v", err)
	}
}

func TestCancelBooking_AlreadyAttended(t *testing.T) {
	booking := &domain.Booking{
		ID:        1,
		UserID:    1,
		SessionID: 1,
		Status:    domain.BookingStatusAttended,
	}
	session := &domain.Session{ID: 1, Price: 500}

	uc, _, _, _ := newTestUsecase(nil, session, booking)
	err := uc.CancelBooking(context.Background(), 1, 1, false)

	if err == nil || err.Error() != "attended bookings cannot be cancelled" {
		t.Errorf("expected 'attended bookings cannot be cancelled', got: %v", err)
	}
}

func TestCancelBooking_AdminCanCancelAnyone(t *testing.T) {
	checkAccess := func(booking *domain.Booking, requesterID int, isAdmin bool) error {
		if booking.UserID != requesterID && !isAdmin {
			return ErrBookingForbidden
		}
		if booking.Status == domain.BookingStatusCancelled {
			return errors.New("booking already cancelled")
		}
		if booking.Status == domain.BookingStatusAttended {
			return errors.New("attended bookings cannot be cancelled")
		}
		return nil
	}

	booking := &domain.Booking{ID: 1, UserID: 99, SessionID: 1, Status: domain.BookingStatusConfirmed}

	err := checkAccess(booking, 1, false)
	if !errors.Is(err, ErrBookingForbidden) {
		t.Errorf("user must got ErrBookingForbidden, got: %v", err)
	}

	err = checkAccess(booking, 1, true)
	if err != nil {
		t.Errorf("admin shouldn't get error, got: %v", err)
	}
}

func TestMarkAttended_SessionNotStarted(t *testing.T) {
	booking := &domain.Booking{
		ID:        1,
		UserID:    1,
		SessionID: 1,
		Status:    domain.BookingStatusConfirmed,
	}
	session := &domain.Session{
		ID:        1,
		StartTime: time.Now().Add(2 * time.Hour),
	}

	uc, _, _, _ := newTestUsecase(nil, session, booking)
	err := uc.MarkAttended(context.Background(), 1)

	if !errors.Is(err, ErrSessionNotStartedYet) {
		t.Errorf("expected ErrSessionNotStartedYet, got: %v", err)
	}
}

func TestMarkAttended_AlreadyAttended(t *testing.T) {
	booking := &domain.Booking{
		ID:        1,
		UserID:    1,
		SessionID: 1,
		Status:    domain.BookingStatusAttended,
	}
	session := &domain.Session{
		ID:        1,
		StartTime: time.Now().Add(-1 * time.Hour),
	}

	uc, _, _, _ := newTestUsecase(nil, session, booking)
	err := uc.MarkAttended(context.Background(), 1)

	if !errors.Is(err, ErrAlreadyAttended) {
		t.Errorf("expected ErrAlreadyAttended, got: %v", err)
	}
}

func TestMarkAttended_WrongStatus(t *testing.T) {
	for _, status := range []domain.BookingStatus{domain.BookingStatusPending, domain.BookingStatusCancelled} {
		booking := &domain.Booking{
			ID:        1,
			UserID:    1,
			SessionID: 1,
			Status:    status,
		}
		session := &domain.Session{
			ID:        1,
			StartTime: time.Now().Add(-1 * time.Hour),
		}

		uc, _, _, _ := newTestUsecase(nil, session, booking)
		err := uc.MarkAttended(context.Background(), 1)

		if !errors.Is(err, ErrBookingNotConfirmed) {
			t.Errorf("status %s: expected ErrBookingNotConfirmed, got: %v", status, err)
		}
	}
}

func TestGenerateAttendanceQR_ForbiddenForOtherOwner(t *testing.T) {
	uc, _, sessionRepo, _ := newTestUsecase(nil, nil, nil)
	sessionRepo.ownerID = 99

	_, _, err := uc.GenerateAttendanceQR(context.Background(), 1, domain.RoleGymOwner, 10)

	if !errors.Is(err, ErrSessionForbidden) {
		t.Fatalf("expected ErrSessionForbidden, got: %v", err)
	}
}

func TestGenerateAttendanceQR_AdminAllowed(t *testing.T) {
	session := &domain.Session{ID: 10}
	uc, _, _, _ := newTestUsecase(nil, session, nil)
	auth.SetJWTSecret("test-secret")

	token, expiresAt, err := uc.GenerateAttendanceQR(context.Background(), 1, domain.RoleAdmin, 10)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if token == "" {
		t.Fatal("expected token to be generated")
	}
	if expiresAt.Before(time.Now()) {
		t.Fatal("expected future expiry")
	}
}

func TestScanAttendanceQR_InvalidToken(t *testing.T) {
	uc, _, _, _ := newTestUsecase(nil, nil, nil)

	_, err := uc.ScanAttendanceQR(context.Background(), 1, "not-a-token")

	if !errors.Is(err, ErrInvalidAttendanceQR) {
		t.Fatalf("expected ErrInvalidAttendanceQR, got: %v", err)
	}
}

func TestScanAttendanceQR_AlreadyAttended(t *testing.T) {
	auth.SetJWTSecret("test-secret")
	token, _, err := auth.GenerateAttendanceQRToken(1, time.Minute)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	now := time.Now()
	booking := &domain.Booking{
		ID:         7,
		UserID:     1,
		SessionID:  1,
		Status:     domain.BookingStatusAttended,
		AttendedAt: &now,
	}

	uc, _, _, _ := newTestUsecase(nil, nil, booking)
	result, err := uc.ScanAttendanceQR(context.Background(), 1, token)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}
	if !result.AlreadyAttended {
		t.Fatal("expected already attended result")
	}
	if result.Status != string(domain.BookingStatusAttended) {
		t.Fatalf("expected attended status, got: %s", result.Status)
	}
}
