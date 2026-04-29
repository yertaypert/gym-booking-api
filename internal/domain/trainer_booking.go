package domain

import "time"

type TrainerBooking struct {
	ID            int       `json:"id"`
	UserID        int       `json:"user_id"`
	TrainerSlotID int       `json:"trainer_slot_id"`
	Status        string    `json:"status"`
	CreatedAt     time.Time `json:"created_at"`
}
