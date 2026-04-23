package usecase

import (
	"context"
	"database/sql"
	"time"

	"github.com/yertaypert/gym-booking-api/internal/domain"
	"github.com/yertaypert/gym-booking-api/internal/repository"
)

type UserRepository interface {
	Create(user domain.User) (int, error)
	GetByEmail(email string) (*domain.User, error)
	GetByID(id int) (*domain.User, error)
}

type GymRepository interface {
	ListGyms() ([]domain.Gym, error)
	GetGymByID(id int) (*domain.Gym, error)
	ListClassesByGymID(gymID int) ([]domain.Class, error)
	CreateGym(gym domain.Gym) (*domain.Gym, error)
	CreateClass(class domain.Class) (*domain.Class, error)
	ListSessionsByGymAndClassID(gymID, classID int) ([]domain.Session, error)
	GetClassByID(classID int) (*domain.Class, error)
	CreateSession(gymID int, session domain.Session) (*domain.Session, error)
	ListGymsByOwnerID(ownerID int) ([]domain.Gym, error)
	AssignTrainer(gymID int, trainerID int) error
}

type BookingRepository interface {
	GetByID(ctx context.Context, bookingID int) (*domain.Booking, error)
	Create(tx *sql.Tx, userID, sessionID int) (int, error)
	UpdateStatus(ctx context.Context, tx *sql.Tx, bookingID int, status string) error
	GetByUserID(ctx context.Context, userID int) ([]domain.Booking, error)
	GetDetailsByUserID(ctx context.Context, userID int) ([]domain.BookingDetail, error)
	GetBySessionID(ctx context.Context, sessionID int) ([]domain.BookingDetail, error)
	ExistsByUserAndSession(ctx context.Context, userID, sessionID int) (bool, error)
	MarkAttended(ctx context.Context, tx *sql.Tx, bookingID int) error
	ListByGymID(ctx context.Context, gymID int) ([]domain.Booking, error)
}

type WalletRepository interface {
	UpdateBalance(tx *sql.Tx, userID int, amount float64) error
	CreateTransaction(tx *sql.Tx, userID int, bookingID *int, amount float64, txType string) error
}

type SessionRepository interface {
	GetByID(ctx context.Context, id int) (*domain.Session, error)
	DecreaseAvailableSlots(ctx context.Context, tx *sql.Tx, sessionID int) error
	IncreaseAvailableSlots(ctx context.Context, tx *sql.Tx, sessionID int) error
}

type TransactionRepository interface {
	Create(transaction *domain.Transaction) error
	GetByUserID(userID int) ([]domain.Transaction, error)
}

type ClassRepository interface {
	ListDistinctClasses(ctx context.Context) ([]string, error)
	ListGymsByClassName(ctx context.Context, name string) ([]repository.GymWithClass, error)
	SearchSessionsByClassName(ctx context.Context, name string, startTime, endTime *time.Time) ([]repository.SessionWithGym, error)
	GetSessionWithDetails(ctx context.Context, sessionID int) (*repository.SessionWithGym, error)
}
