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

func (r *TransactionRepository) GetByUserID(userID int) ([]domain.Transaction, error) {
	query := `SELECT id, user_id, booking_id, amount, type, created_at 
	          FROM transactions WHERE user_id = $1 ORDER BY created_at DESC`

	rows, err := r.db.Query(query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var transactions []domain.Transaction
	for rows.Next() {
		var t domain.Transaction
		var bookingID sql.NullInt64

		err := rows.Scan(&t.ID, &t.UserID, &bookingID, &t.Amount, &t.Type, &t.CreatedAt)
		if err != nil {
			return nil, err
		}

		if bookingID.Valid {
			id := int(bookingID.Int64)
			t.BookingID = &id
		}

		transactions = append(transactions, t)
	}
	return transactions, rows.Err()
}
