package domain

import "time"

type TrainerSlot struct {
	ID        int       `json:"id"`
	TrainerID int       `json:"trainerID"`
	StartTime time.Time `json:"startTime"`
	EndTime   time.Time `json:"endTime"`
	Status    string    `json:"status"`
}
