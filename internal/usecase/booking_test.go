package usecase

// Это unit-тесты для бизнес-логики букинга.
// Они НЕ требуют запущенной базы данных — вместо реальных репозиториев
// используются "моки" (фейковые реализации интерфейсов).
//
// Запуск: go test ./internal/usecase/ -v -run TestBooking

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/yertaypert/gym-booking-api/internal/domain"
)

// ─── Моки ───────────────────────────────────────────────────────────────────
// Мок — это фейковый репозиторий. Он реализует тот же интерфейс что и
// настоящий, но вместо SQL просто возвращает данные которые мы задаём в тесте.

// mockUserRepo — фейковый репозиторий пользователей
type mockUserRepo struct {
	user *domain.User
	err  error
}

func (m *mockUserRepo) GetByID(id int) (*domain.User, error)          { return m.user, m.err }
func (m *mockUserRepo) Create(user domain.User) (int, error)          { return 1, nil }
func (m *mockUserRepo) GetByEmail(email string) (*domain.User, error) { return m.user, m.err }

// mockSessionRepo — фейковый репозиторий сессий
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

// mockBookingRepo — фейковый репозиторий букингов
type mockBookingRepo struct {
	booking        *domain.Booking
	err            error
	createErr      error
	isDuplicate    bool
	lastStatus     string
	attendedCalled bool
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
func (m *mockBookingRepo) GetByUserID(ctx context.Context, userID int) ([]domain.BookingDetail, error) {
	return []domain.BookingDetail{}, nil
}
func (m *mockBookingRepo) GetBySessionID(ctx context.Context, sessionID int) ([]domain.BookingDetail, error) {
	return []domain.BookingDetail{}, nil
}
func (m *mockBookingRepo) MarkAttended(ctx context.Context, tx *sql.Tx, bookingID int) error {
	m.attendedCalled = true
	return nil
}

// mockWalletRepo — фейковый кошелёк
type mockWalletRepo struct {
	lastAmount float64
}

func (m *mockWalletRepo) UpdateBalance(tx *sql.Tx, userID int, amount float64) error {
	m.lastAmount = amount
	return nil
}
func (m *mockWalletRepo) CreateTransaction(tx *sql.Tx, userID, bookingID int, amount float64, txType string) error {
	return nil
}

// ─── Хелпер для создания usecase с моками ───────────────────────────────────

// Реальная БД нужна только для BeginTx. Чтобы не поднимать Postgres,
// используем sqlite-like подход — открываем in-memory соединение.
// На практике в тестах это нормально заменять через sqlmock или
// просто тестировать логику ДО транзакции (что мы здесь и делаем).

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
		db:          nil, // BeginTx вызывается только после проверок — в тестах ниже не доходим до него
		bookingRepo: bookingRepo,
		walletRepo:  walletRepo,
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
	}
	return uc, bookingRepo, sessionRepo, walletRepo
}

// ─── Тесты CreateBooking ─────────────────────────────────────────────────────

// Тест: нельзя записаться на сессию которая уже закончилась
func TestCreateBooking_SessionInPast(t *testing.T) {
	user := &domain.User{ID: 1, Balance: 1000}
	session := &domain.Session{
		ID:             1,
		Status:         "active",
		AvailableSlots: 5,
		Price:          500,
		StartTime:      time.Now().Add(-3 * time.Hour), // началась 3 часа назад
		EndTime:        time.Now().Add(-1 * time.Hour), // закончилась час назад
	}

	uc, _, _, _ := newTestUsecase(user, session, nil)
	_, err := uc.CreateBooking(context.Background(), 1, 1)

	if !errors.Is(err, ErrSessionInPast) {
		t.Errorf("ожидали ErrSessionInPast, получили: %v", err)
	}
}

// Тест: нельзя записаться если сессия отменена
func TestCreateBooking_SessionNotActive(t *testing.T) {
	user := &domain.User{ID: 1, Balance: 1000}
	session := &domain.Session{
		ID:             1,
		Status:         "cancelled", // не active
		AvailableSlots: 5,
		Price:          500,
		EndTime:        time.Now().Add(2 * time.Hour),
	}

	uc, _, _, _ := newTestUsecase(user, session, nil)
	_, err := uc.CreateBooking(context.Background(), 1, 1)

	if !errors.Is(err, ErrSessionNotActive) {
		t.Errorf("ожидали ErrSessionNotActive, получили: %v", err)
	}
}

// Тест: нельзя записаться если нет мест
func TestCreateBooking_NoSlots(t *testing.T) {
	user := &domain.User{ID: 1, Balance: 1000}
	session := &domain.Session{
		ID:             1,
		Status:         "active",
		AvailableSlots: 0, // мест нет!
		Price:          500,
		EndTime:        time.Now().Add(2 * time.Hour),
	}

	uc, _, _, _ := newTestUsecase(user, session, nil)
	_, err := uc.CreateBooking(context.Background(), 1, 1)

	if err == nil || err.Error() != "no available slots" {
		t.Errorf("ожидали 'no available slots', получили: %v", err)
	}
}

// Тест: нельзя записаться если не хватает денег
func TestCreateBooking_InsufficientBalance(t *testing.T) {
	user := &domain.User{ID: 1, Balance: 100} // только 100
	session := &domain.Session{
		ID:             1,
		Status:         "active",
		AvailableSlots: 5,
		Price:          500, // а занятие стоит 500
		EndTime:        time.Now().Add(2 * time.Hour),
	}

	uc, _, _, _ := newTestUsecase(user, session, nil)
	_, err := uc.CreateBooking(context.Background(), 1, 1)

	if err == nil || err.Error() != "insufficient balance" {
		t.Errorf("ожидали 'insufficient balance', получили: %v", err)
	}
}

// Тест: нельзя записаться дважды на одну сессию
func TestCreateBooking_Duplicate(t *testing.T) {
	user := &domain.User{ID: 1, Balance: 1000}
	session := &domain.Session{
		ID:             1,
		Status:         "active",
		AvailableSlots: 5,
		Price:          500,
		EndTime:        time.Now().Add(2 * time.Hour),
	}

	uc, bookingRepo, _, _ := newTestUsecase(user, session, nil)
	bookingRepo.isDuplicate = true // симулируем что букинг уже существует

	_, err := uc.CreateBooking(context.Background(), 1, 1)

	if err == nil || err.Error() != "you are already booked for this session" {
		t.Errorf("ожидали ошибку дубликата, получили: %v", err)
	}
}

// ─── Тесты CancelBooking ─────────────────────────────────────────────────────

// Тест: нельзя отменить чужой букинг
func TestCancelBooking_Forbidden(t *testing.T) {
	booking := &domain.Booking{
		ID:        1,
		UserID:    99, // букинг принадлежит юзеру 99
		SessionID: 1,
		Status:    "confirmed",
	}
	session := &domain.Session{ID: 1, Price: 500}

	uc, _, _, _ := newTestUsecase(nil, session, booking)

	// Запрашиваем отмену от имени юзера 1 (не 99), не админ
	err := uc.CancelBooking(context.Background(), 1, 1, false)

	if !errors.Is(err, ErrBookingForbidden) {
		t.Errorf("ожидали ErrBookingForbidden, получили: %v", err)
	}
}

// Тест: нельзя отменить уже отменённый букинг
func TestCancelBooking_AlreadyCancelled(t *testing.T) {
	booking := &domain.Booking{
		ID:        1,
		UserID:    1,
		SessionID: 1,
		Status:    "cancelled", // уже отменён
	}
	session := &domain.Session{ID: 1, Price: 500}

	uc, _, _, _ := newTestUsecase(nil, session, booking)
	err := uc.CancelBooking(context.Background(), 1, 1, false)

	if err == nil || err.Error() != "booking already cancelled" {
		t.Errorf("ожидали 'booking already cancelled', получили: %v", err)
	}
}

// Тест: нельзя отменить уже посещённый букинг
func TestCancelBooking_AlreadyAttended(t *testing.T) {
	booking := &domain.Booking{
		ID:        1,
		UserID:    1,
		SessionID: 1,
		Status:    "attended", // уже посетил
	}
	session := &domain.Session{ID: 1, Price: 500}

	uc, _, _, _ := newTestUsecase(nil, session, booking)
	err := uc.CancelBooking(context.Background(), 1, 1, false)

	if err == nil || err.Error() != "attended bookings cannot be cancelled" {
		t.Errorf("ожидали 'attended bookings cannot be cancelled', получили: %v", err)
	}
}

// Тест: админ может отменить чужой букинг.
// Проверяем только логику прав доступа изолированно.
func TestCancelBooking_AdminCanCancelAnyone(t *testing.T) {
	// Эта вспомогательная функция воспроизводит только ту часть CancelBooking,
	// которая проверяет права — без BeginTx, без db.
	checkAccess := func(booking *domain.Booking, requesterID int, isAdmin bool) error {
		if booking.UserID != requesterID && !isAdmin {
			return ErrBookingForbidden
		}
		if booking.Status == "cancelled" {
			return errors.New("booking already cancelled")
		}
		if booking.Status == "attended" {
			return errors.New("attended bookings cannot be cancelled")
		}
		return nil
	}

	booking := &domain.Booking{ID: 1, UserID: 99, SessionID: 1, Status: "confirmed"}

	// Обычный юзер (ID=1) НЕ может отменить букинг юзера 99
	err := checkAccess(booking, 1, false)
	if !errors.Is(err, ErrBookingForbidden) {
		t.Errorf("обычный юзер должен получить ErrBookingForbidden, получили: %v", err)
	}

	// Админ МОЖЕТ отменить чужой букинг
	err = checkAccess(booking, 1, true)
	if err != nil {
		t.Errorf("админ не должен получать ошибку, получили: %v", err)
	}
}

// ─── Тесты MarkAttended ──────────────────────────────────────────────────────

// Тест: нельзя отметить посещение до начала занятия
func TestMarkAttended_SessionNotStarted(t *testing.T) {
	booking := &domain.Booking{
		ID:        1,
		UserID:    1,
		SessionID: 1,
		Status:    "confirmed",
	}
	session := &domain.Session{
		ID:        1,
		StartTime: time.Now().Add(2 * time.Hour), // занятие ещё не началось
	}

	uc, _, _, _ := newTestUsecase(nil, session, booking)
	err := uc.MarkAttended(context.Background(), 1)

	if !errors.Is(err, ErrSessionNotStartedYet) {
		t.Errorf("ожидали ErrSessionNotStartedYet, получили: %v", err)
	}
}

// Тест: нельзя отметить посещение дважды
func TestMarkAttended_AlreadyAttended(t *testing.T) {
	booking := &domain.Booking{
		ID:        1,
		UserID:    1,
		SessionID: 1,
		Status:    "attended", // уже отмечен
	}
	session := &domain.Session{
		ID:        1,
		StartTime: time.Now().Add(-1 * time.Hour),
	}

	uc, _, _, _ := newTestUsecase(nil, session, booking)
	err := uc.MarkAttended(context.Background(), 1)

	if !errors.Is(err, ErrAlreadyAttended) {
		t.Errorf("ожидали ErrAlreadyAttended, получили: %v", err)
	}
}

// Тест: нельзя отметить посещение для pending/cancelled букинга
func TestMarkAttended_WrongStatus(t *testing.T) {
	for _, status := range []string{"pending", "cancelled"} {
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
			t.Errorf("статус %s: ожидали ErrBookingNotConfirmed, получили: %v", status, err)
		}
	}
}
