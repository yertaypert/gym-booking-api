package repository

import (
	"database/sql"

	"github.com/yertaypert/gym-booking-api/internal/domain"
)

type TransactionRepository struct {
	db *sql.DB
}

func NewTransactionRepository(db *sql.DB) *TransactionRepository {
	return &TransactionRepository{db: db}
}

func (r *TransactionRepository) Create(t *domain.Transaction) error {
	query := `INSERT INTO transactions (user_id, amount, type, created_at) 
	          VALUES ($1, $2, $3, $4) RETURNING id`

	err := r.db.QueryRow(query, t.UserID, t.Amount, t.Type, t.CreatedAt).Scan(&t.ID)
	return err
}
