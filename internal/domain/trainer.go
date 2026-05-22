package domain

import "time"

type Trainer struct {
	ID             int     `json:"id"`
	UserID         int     `json:"user_id"`
	Specialization string  `json:"specialization"`
	ExtraFee       float64 `json:"extraFee"`
}

type TrainerInfo struct {
	UserID         int     `json:"user_id"`
	FullName       string  `json:"full_name"`
	Specialization string  `json:"specialization"`
	ExtraFee       float64 `json:"extra_fee"`
}
type TrainerSlot struct {
	ID        int       `json:"id"`
	TrainerID int       `json:"trainer_id"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Status    string    `json:"status"`
}
type TrainerBooking struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	TrainerSlotID int       `json:"trainer_slot_id"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}
