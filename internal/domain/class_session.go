package domain

import "time"

type ClassSession struct {
	ID        int       `json:"id"`
	GymID     int       `json:"gym_id"`
	Title     string    `json:"title"`
	StartTime time.Time `json:"start_time"`
	EndTime   time.Time `json:"end_time"`
	Capacity  int       `json:"capacity"`
	Booked    int       `json:"booked"`
}
