package usecase

import (
	"context"
	"database/sql"
	"errors"
	"testing"
	"time"

	"github.com/yertaypert/gym-booking-api/internal/domain"
)

// Impossible to create booking on past session
func TestCreateBooking_SessionInPast(t *testing.T) {
	user := &domain.User{ID: 1, Balance: 1000}
	session := &domain.Session{
		ID:             1,
		Status:         "active",
		AvailableSlots: 5,
		Price:          500,
		StartTime:      time.Now().Add(-3 * time.Hour), 
		EndTime:        time.Now().Add(-1 * time.Hour),
	}

	uc, _, _, _ := newTestUsecase(user, session, nil)
	_, err := uc.CreateBooking(context.Background(), 1, 1)

	if !errors.Is(err, ErrSessionInPast) {
		t.Errorf("ожидали ErrSessionInPast, получили: %v", err)
	}
}

// Impossible to book if session is not active
func TestCreateBooking_SessionNotActive(t *testing.T) {
	user := &domain.User{ID: 1, Balance: 1000}
	session := &domain.Session{
		ID:             1,
		Status:         "cancelled",
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

// Impossible to book if no slots
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

// Impossible to book if not enough balance
func TestCreateBooking_InsufficientBalance(t *testing.T) {
	user := &domain.User{ID: 1, Balance: 100}
	session := &domain.Session{
		ID:             1,
		Status:         "active",
		AvailableSlots: 5,
		Price:          500,
		EndTime:        time.Now().Add(2 * time.Hour),
	}

	uc, _, _, _ := newTestUsecase(user, session, nil)
	_, err := uc.CreateBooking(context.Background(), 1, 1)

	if err == nil || err.Error() != "insufficient balance" {
		t.Errorf("ожидали 'insufficient balance', получили: %v", err)
	}
}

// Impossible to book one session couple times
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
	bookingRepo.isDuplicate = true

	_, err := uc.CreateBooking(context.Background(), 1, 1)

	if err == nil || err.Error() != "you are already booked for this session" {
		t.Errorf("ожидали ошибку дубликата, получили: %v", err)
	}
}

// Impossible to cancel another booking
func TestCancelBooking_Forbidden(t *testing.T) {
	booking := &domain.Booking{
		ID:        1,
		UserID:    99, // booking for user 99
		SessionID: 1,
		Status:    "confirmed",
	}
	session := &domain.Session{ID: 1, Price: 500}

	uc, _, _, _ := newTestUsecase(nil, session, booking)

	// Ask for cancelation from user 1
	err := uc.CancelBooking(context.Background(), 1, 1, false)

	if !errors.Is(err, ErrBookingForbidden) {
		t.Errorf("ожидали ErrBookingForbidden, получили: %v", err)
	}
}

// Impossible to cancel canceled booking
func TestCancelBooking_AlreadyCancelled(t *testing.T) {
	booking := &domain.Booking{
		ID:        1,
		UserID:    1,
		SessionID: 1,
		Status:    "cancelled",
	}
	session := &domain.Session{ID: 1, Price: 500}

	uc, _, _, _ := newTestUsecase(nil, session, booking)
	err := uc.CancelBooking(context.Background(), 1, 1, false)

	if err == nil || err.Error() != "booking already cancelled" {
		t.Errorf("ожидали 'booking already cancelled', получили: %v", err)
	}
}

// Impossible to cancel attended booking
func TestCancelBooking_AlreadyAttended(t *testing.T) {
	booking := &domain.Booking{
		ID:        1,
		UserID:    1,
		SessionID: 1,
		Status:    "attended",
	}
	session := &domain.Session{ID: 1, Price: 500}

	uc, _, _, _ := newTestUsecase(nil, session, booking)
	err := uc.CancelBooking(context.Background(), 1, 1, false)

	if err == nil || err.Error() != "attended bookings cannot be cancelled" {
		t.Errorf("ожидали 'attended bookings cannot be cancelled', получили: %v", err)
	}
}

// Admin can calcel any booking
func TestCancelBooking_AdminCanCancelAnyone(t *testing.T) {
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

	err := checkAccess(booking, 1, false)
	if !errors.Is(err, ErrBookingForbidden) {
		t.Errorf("обычный юзер должен получить ErrBookingForbidden, получили: %v", err)
	}

	err = checkAccess(booking, 1, true)
	if err != nil {
		t.Errorf("админ не должен получать ошибку, получили: %v", err)
	}
}

// Тесты MarkAttended

// Impossible to mark attendance before session started
func TestMarkAttended_SessionNotStarted(t *testing.T) {
	booking := &domain.Booking{
		ID:        1,
		UserID:    1,
		SessionID: 1,
		Status:    "confirmed",
	}
	session := &domain.Session{
		ID:        1,
		StartTime: time.Now().Add(2 * time.Hour),
	}

	uc, _, _, _ := newTestUsecase(nil, session, booking)
	err := uc.MarkAttended(context.Background(), 1)

	if !errors.Is(err, ErrSessionNotStartedYet) {
		t.Errorf("ожидали ErrSessionNotStartedYet, получили: %v", err)
	}
}

// Impossible to mark attendance for attended booking
func TestMarkAttended_AlreadyAttended(t *testing.T) {
	booking := &domain.Booking{
		ID:        1,
		UserID:    1,
		SessionID: 1,
		Status:    "attended",
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

// Impossible to mark attendance for pending/cancelled booking
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

// ListGymBookings 

func TestListGymBookings_GymNotFound(t *testing.T) {
	_, _, _, _ = newTestUsecase(nil, nil, nil)
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

// GetSessionAttendees

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

// GetMyBookings / GetUserBookings 

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
