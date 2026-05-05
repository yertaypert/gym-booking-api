package handler

import (
	"errors"
	"net/http"

	"github.com/yertaypert/gym-booking-api/internal/domain"
)

func HandleError(w http.ResponseWriter, err error) {
	if err == nil {
		return
	}

	status := http.StatusInternalServerError

	switch {
	// Not Found
	case errors.Is(err, domain.ErrUserNotFound),
		errors.Is(err, domain.ErrGymNotFound),
		errors.Is(err, domain.ErrClassNotFound),
		errors.Is(err, domain.ErrSessionNotFound),
		errors.Is(err, domain.ErrBookingNotFound),
		errors.Is(err, domain.ErrClassDoesNotBelongToGym):
		status = http.StatusNotFound

	// Conflict / Already Exists
	case errors.Is(err, domain.ErrEmailAlreadyExists),
		errors.Is(err, domain.ErrGymAlreadyExists),
		errors.Is(err, domain.ErrAlreadyBooked),
		errors.Is(err, domain.ErrAlreadyAttended),
		errors.Is(err, domain.ErrSessionNotActive),
		errors.Is(err, domain.ErrSessionInPast),
		errors.Is(err, domain.ErrBookingNotConfirmed),
		errors.Is(err, domain.ErrBookingCancelled),
		errors.Is(err, domain.ErrBookingAttended),
		errors.Is(err, domain.ErrTrainerAlreadyAssigned),
		errors.Is(err, domain.ErrTrainerNotAssignedToGym):
		status = http.StatusConflict

	// Bad Request / Validation
	case errors.Is(err, domain.ErrInvalidEmail),
		errors.Is(err, domain.ErrInvalidFullName),
		errors.Is(err, domain.ErrWeakPassword),
		errors.Is(err, domain.ErrInvalidPassword),
		errors.Is(err, domain.ErrInvalidGymName),
		errors.Is(err, domain.ErrInvalidOwnerID),
		errors.Is(err, domain.ErrInvalidClassName),
		errors.Is(err, domain.ErrInvalidMaxCapacity),
		errors.Is(err, domain.ErrInvalidSessionTime),
		errors.Is(err, domain.ErrInvalidSessionPrice),
		errors.Is(err, domain.ErrNoAvailableSlots),
		errors.Is(err, domain.ErrInsufficientBalance),
		errors.Is(err, domain.ErrInvalidAmount),
		errors.Is(err, domain.ErrSessionNotStarted),
		errors.Is(err, domain.ErrUserIsNotTrainer):
		status = http.StatusBadRequest

	// Forbidden
	case errors.Is(err, domain.ErrPermissionDenied),
		errors.Is(err, domain.ErrNotGymOwner),
		errors.Is(err, domain.ErrBookingForbidden):
		status = http.StatusForbidden

	// Unauthorized
	case errors.Is(err, domain.ErrUnauthorized):
		status = http.StatusUnauthorized
	}

	http.Error(w, err.Error(), status)
}
