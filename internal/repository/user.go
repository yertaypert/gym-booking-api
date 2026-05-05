package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/lib/pq"
	"github.com/yertaypert/gym-booking-api/internal/domain"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(u domain.User) (int, error) {
	var id int
	query := `INSERT INTO users (email, password_hash, full_name, role, balance) 
              VALUES ($1, $2, $3, $4, $5) RETURNING id`

	err := r.db.QueryRow(query, u.Email, u.PasswordHash, u.FullName, u.Role, u.Balance).Scan(&id)
	if err != nil {
		if pgErr, ok := err.(*pq.Error); ok && pgErr.Code == "23505" {
			return 0, domain.ErrEmailAlreadyExists
		}
		return 0, err
	}
	return id, nil
}

func (r *UserRepository) GetByEmail(email string) (*domain.User, error) {
	u := &domain.User{}
	query := `SELECT id, email, password_hash, full_name, role, balance FROM users WHERE email = $1`

	err := r.db.QueryRow(query, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.Role, &u.Balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return u, nil
}

func (r *UserRepository) GetByID(id int) (*domain.User, error) {
	u := &domain.User{}
	query := `SELECT id, email, full_name, role, balance, created_at FROM users WHERE id = $1`

	err := r.db.QueryRow(query, id).Scan(&u.ID, &u.Email, &u.FullName, &u.Role, &u.Balance, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	return u, nil
}

func (r *UserRepository) UpdateRole(ctx context.Context, tx *sql.Tx, userID int, role domain.UserRole) error {
	query := `UPDATE users SET role = $1 WHERE id = $2`
	var err error
	if tx != nil {
		_, err = tx.ExecContext(ctx, query, role, userID)
	} else {
		_, err = r.db.ExecContext(ctx, query, role, userID)
	}
	return err
}
