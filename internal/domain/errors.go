package domain

import "errors"

var (
	ErrNotFound         = errors.New("resource not found")
	ErrAlreadyExists    = errors.New("resource already exists")
	ErrUnauthorized     = errors.New("unauthorized")
	ErrInternal         = errors.New("internal server error")
	ErrInvalidInput     = errors.New("invalid input")
	ErrConflict         = errors.New("resource conflict")
	ErrPermissionDenied = errors.New("permission denied")
)

// Specific domain errors
var (
	// User / Auth
	ErrUserNotFound       = errors.New("user not found")
	ErrEmailAlreadyExists = errors.New("email already exists")
	ErrInvalidEmail       = errors.New("email must be a valid address")
	ErrInvalidFullName    = errors.New("full_name is required")
	ErrWeakPassword       = errors.New("password must be at least 8 characters and include uppercase, lowercase, and a digit")
	ErrInvalidPassword    = errors.New("invalid password")

	// Gym
	ErrGymNotFound             = errors.New("gym not found")
	ErrGymAlreadyExists       = errors.New("gym already exists")
	ErrNotGymOwner             = errors.New("user is not the owner of this gym")
	ErrTrainerAlreadyAssigned  = errors.New("trainer is already assigned to this gym")
	ErrInvalidGymName          = errors.New("gym name is required")
	ErrInvalidOwnerID          = errors.New("owner_id is required")

	// Class
	ErrClassNotFound           = errors.New("class not found")
	ErrClassDoesNotBelongToGym = errors.New("class does not belong to gym")
	ErrInvalidClassName        = errors.New("class name is required")
	ErrInvalidMaxCapacity      = errors.New("max_capacity must be greater than 0")

	// Session
	ErrSessionNotFound     = errors.New("session not found")
	ErrSessionNotActive    = errors.New("session is not active")
	ErrSessionInPast       = errors.New("cannot book a session that has already ended")
	ErrSessionNotStarted   = errors.New("session has not started yet")
	ErrInvalidSessionTime  = errors.New("end_time must be after start_time")
	ErrInvalidSessionPrice = errors.New("price must be greater than 0")

	// Booking
	ErrBookingNotFound     = errors.New("booking not found")
	ErrBookingForbidden    = errors.New("booking does not belong to user")
	ErrAlreadyBooked       = errors.New("you are already booked for this session")
	ErrAlreadyAttended     = errors.New("attendance already marked for this booking")
	ErrBookingNotConfirmed = errors.New("only confirmed bookings can be marked as attended")
	ErrNoAvailableSlots    = errors.New("no available slots")
	ErrBookingCancelled    = errors.New("booking already cancelled")
	ErrBookingAttended     = errors.New("attended bookings cannot be cancelled")

	// Wallet / Transaction
	ErrInsufficientBalance = errors.New("insufficient balance")
	ErrInvalidAmount       = errors.New("amount must be greater than zero")
)
