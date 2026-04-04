package repository

import (
	"database/sql"
	"time"
)

type GymRepository struct {
	db *sql.DB
}

func NewGymRepository(db *sql.DB) *GymRepository {
	return &GymRepository{db: db}
}

type Session struct {
	ID             int
	GymName        string
	ClassName      string
	StartTime      time.Time
	AvailableSlots int
	Price          float64
}

func (r *GymRepository) GetAvailableSessions() ([]Session, error) {
	query := `
		SELECT s.id, g.name, c.name, s.start_time, s.available_slots, s.price
		FROM class_sessions s
		JOIN classes c ON s.class_id = c.id
		JOIN gyms g ON c.gym_id = g.id
		WHERE s.available_slots > 0 AND s.start_time > NOW()
	`
	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []Session
	for rows.Next() {
		var s Session
		err := rows.Scan(&s.ID, &s.GymName, &s.ClassName, &s.StartTime, &s.AvailableSlots, &s.Price)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}
