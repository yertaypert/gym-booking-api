package domain

import "time"

//type User struct {
//	ID           int       `json:"id"`
//	Email        string    `json:"email"`
//	PasswordHash string    `json:"-"`
//	FullName     string    `json:"full_name"`
//	Role         string    `json:"role"`
//	Balance      float64   `json:"balance"`
//	CreatedAt    time.Time `json:"created_at"`
//}

type Booking struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	SessionID int       `json:"session_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

type Transaction struct {
	ID        int       `json:"id"`
	UserID    int       `json:"user_id"`
	BookingID *int      `json:"booking_id,omitempty"`
	Amount    float64   `json:"amount"`
	Type      string    `json:"type"`
	CreatedAt time.Time `json:"created_at"`
}

type Session struct {
	ID             int       `json:"id"`
	ClassID        int       `json:"class_id"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	AvailableSlots int       `json:"available_slots"`
	Price          float64   `json:"price"`
	Status         string    `json:"status"`
}

type Gym struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Address     string `json:"address"`
	Description string `json:"description"`
}

type Class struct {
	ID          int    `json:"id"`
	GymID       int    `json:"gym_id"`
	Name        string `json:"name"`
	MaxCapacity int    `json:"max_capacity"`
}
