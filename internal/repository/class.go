package repository

import (
	"context"
	"database/sql"
	"fmt"
	"time"

	"github.com/yertaypert/gym-booking-api/internal/domain"
)

type ClassRepository struct {
	db *sql.DB
}

func NewClassRepository(db *sql.DB) *ClassRepository {
	return &ClassRepository{db: db}
}

func (r *ClassRepository) ListDistinctClasses(ctx context.Context) ([]string, error) {
	query := `SELECT DISTINCT name FROM classes ORDER BY name`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var classes []string
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		classes = append(classes, name)
	}
	return classes, nil
}

type SessionWithGym struct {
	domain.Session
	GymName    string `json:"gym_name"`
	GymAddress string `json:"gym_address"`
	ClassName  string `json:"class_name"`
}

func (r *ClassRepository) SearchSessionsByClassName(
	ctx context.Context,
	name string,
	startTime, endTime *time.Time,
) ([]SessionWithGym, error) {
	query := `
		SELECT s.id, s.class_id, s.start_time, s.end_time, s.available_slots, s.price, s.status,
		       g.name as gym_name, g.address as gym_address, c.name as class_name
		FROM class_sessions s
		JOIN classes c ON s.class_id = c.id
		JOIN gyms g ON c.gym_id = g.id
		WHERE c.name ILIKE $1
	`
	args := []any{"%" + name + "%"}
	argCount := 2

	if startTime != nil {
		query += fmt.Sprintf(" AND s.start_time >= $%d", argCount)
		args = append(args, *startTime)
		argCount++
	}
	if endTime != nil {
		query += fmt.Sprintf(" AND s.end_time <= $%d", argCount)
		args = append(args, *endTime)
		argCount++
	}

	query += ` ORDER BY s.start_time`

	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []SessionWithGym
	for rows.Next() {
		var s SessionWithGym
		err := rows.Scan(
			&s.ID, &s.ClassID, &s.StartTime, &s.EndTime, &s.AvailableSlots, &s.Price, &s.Status,
			&s.GymName, &s.GymAddress, &s.ClassName,
		)
		if err != nil {
			return nil, err
		}
		sessions = append(sessions, s)
	}
	return sessions, nil
}

func (r *ClassRepository) GetSessionWithDetails(ctx context.Context, sessionID int) (*SessionWithGym, error) {
	query := `
		SELECT s.id, s.class_id, s.start_time, s.end_time, s.available_slots, s.price, s.status,
		       g.name as gym_name, g.address as gym_address, c.name as class_name
		FROM class_sessions s
		JOIN classes c ON s.class_id = c.id
		JOIN gyms g ON c.gym_id = g.id
		WHERE s.id = $1
	`
	var s SessionWithGym
	err := r.db.QueryRowContext(ctx, query, sessionID).Scan(
		&s.ID, &s.ClassID, &s.StartTime, &s.EndTime, &s.AvailableSlots, &s.Price, &s.Status,
		&s.GymName, &s.GymAddress, &s.ClassName,
	)
	if err != nil {
		return nil, err
	}
	return &s, nil
}

type GymWithClass struct {
	GymID      int    `json:"gym_id"`
	GymName    string `json:"gym_name"`
	GymAddress string `json:"gym_address"`
	ClassID    int    `json:"class_id"`
}

func (r *ClassRepository) ListGymsByClassName(ctx context.Context, name string) ([]GymWithClass, error) {
	query := `
		SELECT g.id, g.name, g.address, c.id
		FROM gyms g
		JOIN classes c ON g.id = c.gym_id
		WHERE c.name = $1
	`
	rows, err := r.db.QueryContext(ctx, query, name)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var gyms []GymWithClass
	for rows.Next() {
		var g GymWithClass
		if err := rows.Scan(&g.GymID, &g.GymName, &g.GymAddress, &g.ClassID); err != nil {
			return nil, err
		}
		gyms = append(gyms, g)
	}
	return gyms, nil
}
