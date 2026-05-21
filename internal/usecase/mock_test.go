package usecase

import (
	"context"
	"database/sql"

	"github.com/yertaypert/gym-booking-api/internal/domain"
)

// ─── mockUserRepo ─────────────────────────────────────────────────────────────

type mockUserRepo struct {
	user *domain.User
	err  error
}

func (m *mockUserRepo) GetByID(id int) (*domain.User, error)          { return m.user, m.err }
func (m *mockUserRepo) Create(user domain.User) (int, error)          { return 1, nil }
func (m *mockUserRepo) GetByEmail(email string) (*domain.User, error) { return m.user, m.err }

// ─── mockSessionRepo ──────────────────────────────────────────────────────────

type mockSessionRepo struct {
	session        *domain.Session
	err            error
	decreaseCalled bool
	increaseCalled bool
}

func (m *mockSessionRepo) GetByID(ctx context.Context, id int) (*domain.Session, error) {
	return m.session, m.err
}
func (m *mockSessionRepo) DecreaseAvailableSlots(ctx context.Context, tx *sql.Tx, sessionID int) error {
	m.decreaseCalled = true
	return nil
}
func (m *mockSessionRepo) IncreaseAvailableSlots(ctx context.Context, tx *sql.Tx, sessionID int) error {
	m.increaseCalled = true
	return nil
}

// ─── mockBookingRepo ──────────────────────────────────────────────────────────

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
func (m *mockBookingRepo) Create(tx *sql.Tx, userID, sessionID int) (int, error) {
	return 1, m.createErr
}
func (m *mockBookingRepo) UpdateStatus(ctx context.Context, tx *sql.Tx, bookingID int, status string) error {
	m.lastStatus = status
	return nil
}
func (m *mockBookingRepo) GetByID(ctx context.Context, bookingID int) (*domain.Booking, error) {
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

// ─── mockWalletRepo ───────────────────────────────────────────────────────────

type mockWalletRepo struct {
	lastAmount float64
}

func (m *mockWalletRepo) UpdateBalance(tx *sql.Tx, userID int, amount float64) error {
	m.lastAmount = amount
	return nil
}
func (m *mockWalletRepo) CreateTransaction(tx *sql.Tx, userID int, bookingID *int, amount float64, txType string) error {
	return nil
}

// ─── mockGymRepo ──────────────────────────────────────────────────────────────

type mockGymRepo struct {
	gym      *domain.Gym
	gymErr   error
	class    *domain.Class
	classErr error
}

func (m *mockGymRepo) GetGymByID(id int) (*domain.Gym, error)        { return m.gym, m.gymErr }
func (m *mockGymRepo) GetClassByID(id int) (*domain.Class, error)    { return m.class, m.classErr }
func (m *mockGymRepo) CreateGym(gym domain.Gym) (*domain.Gym, error) { return m.gym, m.gymErr }
func (m *mockGymRepo) CreateClass(c domain.Class) (*domain.Class, error) {
	return m.class, m.classErr
}
func (m *mockGymRepo) CreateSession(gymID int, s domain.Session) (*domain.Session, error) {
	return &s, m.classErr
}
func (m *mockGymRepo) ListGyms() ([]domain.Gym, error)                   { return nil, nil }
func (m *mockGymRepo) ListGymsByOwnerID(id int) ([]domain.Gym, error)    { return nil, nil }
func (m *mockGymRepo) ListClassesByGymID(id int) ([]domain.Class, error) { return nil, nil }
func (m *mockGymRepo) ListSessionsByGymAndClassID(gymID, classID int) ([]domain.Session, error) {
	return nil, nil
}
func (m *mockGymRepo) AssignTrainer(gymID, trainerID int) error { return nil }

// ─── Хелперы ──────────────────────────────────────────────────────────────────

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
		gymRepo:     &mockGymRepo{},
	}
	return uc, bookingRepo, sessionRepo, walletRepo
}

func newGymUsecase(gym *domain.Gym, gymErr error, class *domain.Class, classErr error) *GymUsecase {
	return NewGymUsecase(&mockGymRepo{
		gym:      gym,
		gymErr:   gymErr,
		class:    class,
		classErr: classErr,
	})
}
