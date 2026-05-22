package domain

import "time"

type BookingStatus string

const (
	BookingStatusPending   BookingStatus = "pending"
	BookingStatusConfirmed BookingStatus = "confirmed"
	BookingStatusCancelled BookingStatus = "cancelled"
	BookingStatusAttended  BookingStatus = "attended"
)

type Booking struct {
	ID         int           `json:"id"`
	UserID     int           `json:"user_id"`
	SessionID  int           `json:"session_id"`
	Status     BookingStatus `json:"status"`
	CreatedAt  time.Time     `json:"created_at"`
	AttendedAt *time.Time    `json:"attended_at"`
}

// BookingDetail is a richer view used in list responses —
// it joins booking with session and class info so callers
// don't have to make extra round-trips.
type BookingDetail struct {
	BookingID  int           `json:"booking_id"`
	SessionID  int           `json:"session_id"`
	ClassName  string        `json:"class_name"`
	GymName    string        `json:"gym_name"`
	StartTime  time.Time     `json:"start_time"`
	EndTime    time.Time     `json:"end_time"`
	Price      float64       `json:"price"`
	Status     BookingStatus `json:"status"`
	UserID     int           `json:"user_id,omitempty"` // populated in session-attendance list
	UserEmail  string        `json:"user_email,omitempty"`
	UserName   string        `json:"user_name,omitempty"`
	BookedAt   time.Time     `json:"booked_at"`
	AttendedAt *time.Time    `json:"attended_at,omitempty"`
}
