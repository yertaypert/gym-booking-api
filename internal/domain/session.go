package domain

import "time"

type Session struct {
	ID             int       `json:"id"`
	ClassID        int       `json:"class_id"`
	StartTime      time.Time `json:"start_time"`
	EndTime        time.Time `json:"end_time"`
	AvailableSlots int       `json:"available_slots"`
	Price          float64   `json:"price"`
	Status         string    `json:"status"`
}
