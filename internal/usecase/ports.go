package usecase

import (
	"context"
	"database/sql"
	"time"

	"github.com/yertaypert/gym-booking-api/internal/domain"
	"github.com/yertaypert/gym-booking-api/internal/repository"
)

type UserRepository interface {
	Create(ctx context.Context, user domain.User) (int, error)
	GetByEmail(ctx context.Context, email string) (*domain.User, error)
	GetByID(ctx context.Context, id int) (*domain.User, error)
	UpdateRole(ctx context.Context, userID int, role domain.UserRole) error
}

type GymRepository interface {
	ListGyms(ctx context.Context) ([]domain.Gym, error)
	GetGymByID(ctx context.Context, id int) (*domain.Gym, error)
	ListClassesByGymID(ctx context.Context, gymID int) ([]domain.Class, error)
	CreateGym(ctx context.Context, gym domain.Gym) (*domain.Gym, error)
	CreateClass(ctx context.Context, class domain.Class) (*domain.Class, error)
	ListSessionsByGymAndClassID(ctx context.Context, gymID, classID int) ([]domain.Session, error)
	GetClassByID(ctx context.Context, classID int) (*domain.Class, error)
	CreateSession(ctx context.Context, gymID int, session domain.Session) (*domain.Session, error)
	ListGymsByOwnerID(ctx context.Context, ownerID int) ([]domain.Gym, error)
	AssignTrainer(ctx context.Context, gymID int, trainerID int) error
}

type BookingRepository interface {
	GetByID(ctx context.Context, bookingID int) (*domain.Booking, error)
	GetByUserAndSession(ctx context.Context, userID, sessionID int) (*domain.Booking, error)
	Create(ctx context.Context, tx *sql.Tx, userID, sessionID int) (int, error)
	UpdateStatus(ctx context.Context, tx *sql.Tx, bookingID int, status string) error
	GetByUserID(ctx context.Context, userID int) ([]domain.Booking, error)
	GetDetailsByUserID(ctx context.Context, userID int) ([]domain.BookingDetail, error)
	GetBySessionID(ctx context.Context, sessionID int) ([]domain.BookingDetail, error)
	ExistsByUserAndSession(ctx context.Context, userID, sessionID int) (bool, error)
	MarkAttended(ctx context.Context, tx *sql.Tx, bookingID int) error
	ListByGymID(ctx context.Context, gymID int) ([]domain.Booking, error)
}

type WalletRepository interface {
	UpdateBalance(ctx context.Context, tx *sql.Tx, userID int, amount float64) error
	CreateTransaction(ctx context.Context, tx *sql.Tx, userID int, bookingID *int, amount float64, txType string) error
}

type SessionRepository interface {
	GetByID(ctx context.Context, id int) (*domain.Session, error)
	GetGymOwnerIDBySessionID(ctx context.Context, sessionID int) (int, error)
	DecreaseAvailableSlots(ctx context.Context, tx *sql.Tx, sessionID int) error
	IncreaseAvailableSlots(ctx context.Context, tx *sql.Tx, sessionID int) error
	UpdateExpiredSessions(ctx context.Context) (int64, error)
}

type TransactionRepository interface {
	Create(ctx context.Context, transaction *domain.Transaction) error
	GetByUserID(ctx context.Context, userID int) ([]domain.Transaction, error)
}

type ClassRepository interface {
	ListDistinctClasses(ctx context.Context) ([]string, error)
	ListGymsByClassName(ctx context.Context, name string) ([]repository.GymWithClass, error)
	SearchSessionsByClassName(ctx context.Context, name string, startTime, endTime *time.Time) ([]repository.SessionWithGym, error)
	GetSessionWithDetails(ctx context.Context, sessionID int) (*repository.SessionWithGym, error)
}
