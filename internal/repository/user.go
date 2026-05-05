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

func (r *UserRepository) ListAll(ctx context.Context) ([]domain.User, error) {
	query := `SELECT id, email, full_name, role, balance, created_at FROM users ORDER BY id ASC`
	rows, err := r.db.QueryContext(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []domain.User
	for rows.Next() {
		var u domain.User
		err := rows.Scan(&u.ID, &u.Email, &u.FullName, &u.Role, &u.Balance, &u.CreatedAt)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

func (r *UserRepository) Update(ctx context.Context, u *domain.User) error {
	query := `UPDATE users SET email = $1, full_name = $2, role = $3, balance = $4 WHERE id = $5`
	_, err := r.db.ExecContext(ctx, query, u.Email, u.FullName, u.Role, u.Balance, u.ID)
	return err
}

func (r *UserRepository) Delete(ctx context.Context, id int) error {
	query := `DELETE FROM users WHERE id = $1`
	_, err := r.db.ExecContext(ctx, query, id)
	return err
}
