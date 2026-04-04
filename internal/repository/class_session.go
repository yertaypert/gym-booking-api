package repository

import (
	"database/sql"

	"github.com/yertaypert/gym-booking-api/internal/domain"
)

type ClassSessionRepository struct {
	db *sql.DB
}

func NewClassSessionRepository(db *sql.DB) *ClassSessionRepository {
	return &ClassSessionRepository{db: db}
}

func (r *ClassSessionRepository) Create(session *domain.ClassSession) error {
	query := `INSERT INTO class_sessions (gym_id, title, start_time, end_time, capacity, booked) 
	          VALUES ($1, $2, $3, $4, $5, $6) RETURNING id`

	err := r.db.QueryRow(query, session.GymID, session.Title, session.StartTime, session.EndTime, session.Capacity, session.Booked).Scan(&session.ID)
	return err
}
