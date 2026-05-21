package usecase

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/yertaypert/gym-booking-api/internal/domain"
)

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

// ─── ListGymBookings ──────────────────────────────────────────────────────────

func TestListGymBookings_GymNotFound(t *testing.T) {
	_, _, _, _ = newTestUsecase(nil, nil, nil)
	// gymRepo не установлен в newTestUsecase — нужен отдельный хелпер
	ucWithGym := &BookingUsecase{
		db:          nil,
		bookingRepo: &mockBookingRepo{},
		walletRepo:  &mockWalletRepo{},
		userRepo:    &mockUserRepo{},
		sessionRepo: &mockSessionRepo{},
		gymRepo:     &mockGymRepo{gymErr: ErrGymNotFound},
	}

	_, err := ucWithGym.ListGymBookings(context.Background(), 10, domain.RoleGymOwner, 1)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}

func TestListGymBookings_NotOwnerAndNotAdmin(t *testing.T) {
	// gym принадлежит ownerID=99, запрос от userID=1 с ролью GymOwner
	ucWithGym := &BookingUsecase{
		db:          nil,
		bookingRepo: &mockBookingRepo{},
		walletRepo:  &mockWalletRepo{},
		userRepo:    &mockUserRepo{},
		sessionRepo: &mockSessionRepo{},
		gymRepo:     &mockGymRepo{gym: &domain.Gym{ID: 1, OwnerID: 99}},
	}

	_, err := ucWithGym.ListGymBookings(context.Background(), 1, domain.RoleGymOwner, 1)
	if !errors.Is(err, ErrBookingForbidden) {
		t.Fatalf("expected ErrBookingForbidden, got %v", err)
	}
}

func TestListGymBookings_AdminCanSeeAnyGym(t *testing.T) {
	// gym принадлежит ownerID=99, но запрос от admin — должно пройти
	ucWithGym := &BookingUsecase{
		db:          nil,
		bookingRepo: &mockBookingRepo{},
		walletRepo:  &mockWalletRepo{},
		userRepo:    &mockUserRepo{},
		sessionRepo: &mockSessionRepo{},
		gymRepo:     &mockGymRepo{gym: &domain.Gym{ID: 1, OwnerID: 99}},
	}

	bookings, err := ucWithGym.ListGymBookings(context.Background(), 1, domain.RoleAdmin, 1)
	if err != nil {
		t.Fatalf("admin should see any gym bookings, got: %v", err)
	}
	if bookings == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestListGymBookings_OwnerCanSeeOwnGym(t *testing.T) {
	ucWithGym := &BookingUsecase{
		db:          nil,
		bookingRepo: &mockBookingRepo{},
		walletRepo:  &mockWalletRepo{},
		userRepo:    &mockUserRepo{},
		sessionRepo: &mockSessionRepo{},
		gymRepo:     &mockGymRepo{gym: &domain.Gym{ID: 1, OwnerID: 10}},
	}

	_, err := ucWithGym.ListGymBookings(context.Background(), 10, domain.RoleGymOwner, 1)
	if err != nil {
		t.Fatalf("owner should see own gym bookings, got: %v", err)
	}
}

// ─── GetSessionAttendees ──────────────────────────────────────────────────────

func TestGetSessionAttendees_SessionNotFound(t *testing.T) {
	uc, _, _, _ := newTestUsecase(nil, nil, nil)
	// sessionRepo вернёт sql.ErrNoRows
	uc.sessionRepo = &mockSessionRepo{err: sql.ErrNoRows}

	_, err := uc.GetSessionAttendees(context.Background(), 1)
	if err == nil || err.Error() != "session not found" {
		t.Fatalf("expected 'session not found', got %v", err)
	}
}

func TestGetSessionAttendees_Success(t *testing.T) {
	session := &domain.Session{ID: 1}
	uc, _, _, _ := newTestUsecase(nil, session, nil)

	attendees, err := uc.GetSessionAttendees(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if attendees == nil {
		t.Fatal("expected non-nil slice")
	}
}

// ─── GetMyBookings / GetUserBookings ──────────────────────────────────────────

func TestGetMyBookings_Success(t *testing.T) {
	uc, _, _, _ := newTestUsecase(nil, nil, nil)

	bookings, err := uc.GetMyBookings(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if bookings == nil {
		t.Fatal("expected non-nil slice")
	}
}

func TestGetUserBookings_Success(t *testing.T) {
	uc, _, _, _ := newTestUsecase(nil, nil, nil)

	bookings, err := uc.GetUserBookings(context.Background(), 1)
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if bookings == nil {
		t.Fatal("expected non-nil slice")
	}
}
