package repository

import (
	"database/sql"
	"errors"
	"time"

	"github.com/lib/pq"
	"github.com/yertaypert/gym-booking-api/internal/domain"
)

type GymRepository struct {
	db *sql.DB
}

var ErrGymNotFound = errors.New("gym not found")
var ErrGymAlreadyExists = errors.New("gym already exists")
var ErrClassNotFound = errors.New("class not found")
var ErrClassDoesNotBelongToGym = errors.New("class does not belong to gym")

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

func (r *GymRepository) ListGyms() ([]domain.Gym, error) {
	rows, err := r.db.Query(`SELECT id, owner_id, name, address, description FROM gyms ORDER BY id`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var gyms []domain.Gym
	for rows.Next() {
		var gym domain.Gym
		var ownerID sql.NullInt64
		if err := rows.Scan(&gym.ID, &ownerID, &gym.Name, &gym.Address, &gym.Description); err != nil {
			return nil, err
		}
		if ownerID.Valid {
			gym.OwnerID = int(ownerID.Int64)
		}
		gyms = append(gyms, gym)
	}

	return gyms, rows.Err()
}

func (r *GymRepository) CreateGym(gym domain.Gym) (*domain.Gym, error) {
	created := &domain.Gym{}
	var ownerID sql.NullInt64
	err := r.db.QueryRow(
		`INSERT INTO gyms (owner_id, name, address, description) VALUES ($1, $2, $3, $4) RETURNING id, owner_id, name, address, description`,
		gym.OwnerID,
		gym.Name,
		gym.Address,
		gym.Description,
	).Scan(&created.ID, &ownerID, &created.Name, &created.Address, &created.Description)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok {
			if pgErr.Code == "23505" {
				return nil, ErrGymAlreadyExists
			}
		}
		return nil, err
	}
	if ownerID.Valid {
		created.OwnerID = int(ownerID.Int64)
	}

	return created, nil
}

func (r *GymRepository) GetGymByID(id int) (*domain.Gym, error) {
	var gym domain.Gym
	var ownerID sql.NullInt64

	err := r.db.QueryRow(
		`SELECT id, owner_id, name, address, description FROM gyms WHERE id = $1`,
		id,
	).Scan(&gym.ID, &ownerID, &gym.Name, &gym.Address, &gym.Description)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrGymNotFound
		}
		return nil, err
	}
	if ownerID.Valid {
		gym.OwnerID = int(ownerID.Int64)
	}

	return &gym, nil
}

func (r *GymRepository) ListGymsByOwnerID(ownerID int) ([]domain.Gym, error) {
	rows, err := r.db.Query(`SELECT id, owner_id, name, address, description FROM gyms WHERE owner_id = $1 ORDER BY id`, ownerID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var gyms []domain.Gym
	for rows.Next() {
		var gym domain.Gym
		var ownerID sql.NullInt64
		if err := rows.Scan(&gym.ID, &ownerID, &gym.Name, &gym.Address, &gym.Description); err != nil {
			return nil, err
		}
		if ownerID.Valid {
			gym.OwnerID = int(ownerID.Int64)
		}
		gyms = append(gyms, gym)
	}

	return gyms, rows.Err()
}

func (r *GymRepository) ListClassesByGymID(gymID int) ([]domain.Class, error) {
	if err := r.ensureGymExists(gymID); err != nil {
		return nil, err
	}

	rows, err := r.db.Query(
		`SELECT id, gym_id, name, max_capacity FROM classes WHERE gym_id = $1 ORDER BY id`,
		gymID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var classes []domain.Class
	for rows.Next() {
		var class domain.Class
		if err := rows.Scan(&class.ID, &class.GymID, &class.Name, &class.MaxCapacity); err != nil {
			return nil, err
		}
		classes = append(classes, class)
	}

	return classes, rows.Err()
}

func (r *GymRepository) CreateClass(class domain.Class) (*domain.Class, error) {
	if err := r.ensureGymExists(class.GymID); err != nil {
		return nil, err
	}

	created := &domain.Class{}
	err := r.db.QueryRow(
		`INSERT INTO classes (gym_id, name, max_capacity) VALUES ($1, $2, $3) RETURNING id, gym_id, name, max_capacity`,
		class.GymID,
		class.Name,
		class.MaxCapacity,
	).Scan(&created.ID, &created.GymID, &created.Name, &created.MaxCapacity)
	if err != nil {
		return nil, err
	}

	return created, nil
}

func (r *GymRepository) ListSessionsByGymAndClassID(gymID, classID int) ([]domain.Session, error) {
	if err := r.ensureClassBelongsToGym(gymID, classID); err != nil {
		return nil, err
	}

	rows, err := r.db.Query(
		`SELECT id, class_id, start_time, end_time, available_slots, price, status
		FROM class_sessions
		WHERE class_id = $1
		ORDER BY start_time`,
		classID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var sessions []domain.Session
	for rows.Next() {
		var session domain.Session
		if err := rows.Scan(
			&session.ID,
			&session.ClassID,
			&session.StartTime,
			&session.EndTime,
			&session.AvailableSlots,
			&session.Price,
			&session.Status,
		); err != nil {
			return nil, err
		}
		sessions = append(sessions, session)
	}

	return sessions, rows.Err()
}

func (r *GymRepository) CreateSession(gymID int, session domain.Session) (*domain.Session, error) {
	if err := r.ensureClassBelongsToGym(gymID, session.ClassID); err != nil {
		return nil, err
	}

	created := &domain.Session{}
	err := r.db.QueryRow(
		`INSERT INTO class_sessions (class_id, start_time, end_time, available_slots, price, status)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id, class_id, start_time, end_time, available_slots, price, status`,
		session.ClassID,
		session.StartTime,
		session.EndTime,
		session.AvailableSlots,
		session.Price,
		session.Status,
	).Scan(
		&created.ID,
		&created.ClassID,
		&created.StartTime,
		&created.EndTime,
		&created.AvailableSlots,
		&created.Price,
		&created.Status,
	)
	if err != nil {
		return nil, err
	}

	return created, nil
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

func (r *GymRepository) ensureGymExists(gymID int) error {
	var exists bool
	err := r.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM gyms WHERE id = $1)`, gymID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return ErrGymNotFound
	}
	return nil
}

func (r *GymRepository) ensureClassExists(classID int) error {
	var exists bool
	err := r.db.QueryRow(`SELECT EXISTS(SELECT 1 FROM classes WHERE id = $1)`, classID).Scan(&exists)
	if err != nil {
		return err
	}
	if !exists {
		return ErrClassNotFound
	}
	return nil
}

func (r *GymRepository) GetClassByID(classID int) (*domain.Class, error) {
	class := &domain.Class{}
	err := r.db.QueryRow(
		`SELECT id, gym_id, name, max_capacity FROM classes WHERE id = $1`,
		classID,
	).Scan(&class.ID, &class.GymID, &class.Name, &class.MaxCapacity)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrClassNotFound
		}
		return nil, err
	}

	return class, nil
}

func (r *GymRepository) ensureClassBelongsToGym(gymID, classID int) error {
	if err := r.ensureGymExists(gymID); err != nil {
		return err
	}

	class, err := r.GetClassByID(classID)
	if err != nil {
		return err
	}

	if class.GymID != gymID {
		return ErrClassDoesNotBelongToGym
	}

	return nil
}

func (r *GymRepository) AssignTrainer(gymID int, trainerID int) error {
	if err := r.ensureGymExists(gymID); err != nil {
		return err
	}

	_, err := r.db.Exec(`INSERT INTO gym_trainers (gym_id, user_id) VALUES ($1, $2)`, gymID, trainerID)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			return errors.New("trainer is already assigned to this gym")
		}
		return err
	}

	return nil
}
