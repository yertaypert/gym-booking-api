package repository

import (
	"database/sql"
	"github.com/yertaypert/gym-booking-api/internal/models"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(u models.User) (int, error) {
	var id int
	query := `INSERT INTO users (email, password_hash, full_name, role, balance) 
              VALUES ($1, $2, $3, $4, $5) RETURNING id`

	err := r.db.QueryRow(query, u.Email, u.PasswordHash, u.FullName, u.Role, u.Balance).Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func (r *UserRepository) GetByEmail(email string) (*models.User, error) {
	u := &models.User{}
	query := `SELECT id, email, password_hash, full_name, role, balance FROM users WHERE email = $1`

	err := r.db.QueryRow(query, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.Role, &u.Balance)
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepository) GetByID(id int) (*models.User, error) {
	u := &models.User{}
	query := `SELECT id, email, full_name, role, balance FROM users WHERE id = $1`

	err := r.db.QueryRow(query, id).Scan(&u.ID, &u.Email, &u.FullName, &u.Role, &u.Balance)
	if err != nil {
		return nil, err
	}
	return u, nil
}
