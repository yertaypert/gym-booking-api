package repository

import (
	"context"
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/yertaypert/gym-booking-api/internal/domain"
)

type UserRepository struct {
	db *sql.DB
}

func NewUserRepository(db *sql.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) Create(ctx context.Context, u domain.User) (int, error) {
	var id int
	query := `INSERT INTO users (email, password_hash, full_name, role, balance) 
              VALUES ($1, $2, $3, $4, $5) RETURNING id`

	err := r.db.QueryRowContext(ctx, query, u.Email, u.PasswordHash, u.FullName, u.Role, u.Balance).Scan(&id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return 0, ErrAlreadyExists
		}
		return 0, err
	}
	return id, nil
}

func (r *UserRepository) GetByEmail(ctx context.Context, email string) (*domain.User, error) {
	u := &domain.User{}
	query := `SELECT id, email, password_hash, full_name, role, balance FROM users WHERE email = $1`

	err := r.db.QueryRowContext(ctx, query, email).Scan(&u.ID, &u.Email, &u.PasswordHash, &u.FullName, &u.Role, &u.Balance)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return u, nil
}

func (r *UserRepository) GetByID(ctx context.Context, id int) (*domain.User, error) {
	u := &domain.User{}
	query := `SELECT id, email, full_name, role, balance, created_at FROM users WHERE id = $1`

	err := r.db.QueryRowContext(ctx, query, id).Scan(&u.ID, &u.Email, &u.FullName, &u.Role, &u.Balance, &u.CreatedAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, err
	}
	return u, nil
}

func (r *UserRepository) UpdateRole(ctx context.Context, userID int, role domain.UserRole) error {
	query := `UPDATE users SET role = $1 WHERE id = $2`
	result, err := r.db.ExecContext(ctx, query, role, userID)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}
