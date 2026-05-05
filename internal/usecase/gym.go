package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/yertaypert/gym-booking-api/internal/domain"
)

var ErrGymNotFound = domain.ErrGymNotFound
var ErrGymAlreadyExists = domain.ErrGymAlreadyExists
var ErrClassNotFound = domain.ErrClassNotFound
var ErrClassDoesNotBelongToGym = domain.ErrClassDoesNotBelongToGym
var ErrInvalidGymName = domain.ErrInvalidGymName
var ErrInvalidOwnerID = domain.ErrInvalidOwnerID
var ErrInvalidClassName = domain.ErrInvalidClassName
var ErrInvalidMaxCapacity = domain.ErrInvalidMaxCapacity
var ErrInvalidSessionTime = domain.ErrInvalidSessionTime
var ErrInvalidSessionPrice = domain.ErrInvalidSessionPrice
var ErrNotGymOwner = domain.ErrNotGymOwner
var ErrUserIsNotTrainer = domain.ErrUserIsNotTrainer
var ErrTrainerNotAssignedToGym = domain.ErrTrainerNotAssignedToGym

type GymUsecase struct {
	gymRepo     GymRepository
	sessionRepo SessionRepository
	userRepo    UserRepository
}

func NewGymUsecase(repo GymRepository, sessionRepo SessionRepository, userRepo UserRepository) *GymUsecase {
	return &GymUsecase{
		gymRepo:     repo,
		sessionRepo: sessionRepo,
		userRepo:    userRepo,
	}
}

func (u *GymUsecase) ListGyms() ([]domain.Gym, error) {
	return u.gymRepo.ListGyms()
}

func (u *GymUsecase) ListGymsByOwner(ownerID int) ([]domain.Gym, error) {
	return u.gymRepo.ListGymsByOwnerID(ownerID)
}

func (u *GymUsecase) GetGym(id int) (*domain.Gym, error) {
	return u.gymRepo.GetGymByID(id)
}

func (u *GymUsecase) ListGymClasses(gymID int) ([]domain.Class, error) {
	return u.gymRepo.ListClassesByGymID(gymID)
}

func (u *GymUsecase) ListClassSessions(gymID, classID int) ([]domain.Session, error) {
	return u.gymRepo.ListSessionsByGymAndClassID(gymID, classID)
}

func (u *GymUsecase) CreateGym(ownerID int, name, address, description string) (*domain.Gym, error) {
	if ownerID <= 0 {
		return nil, ErrInvalidOwnerID
	}

	gym := domain.Gym{
		OwnerID:     ownerID,
		Name:        strings.TrimSpace(name),
		Address:     strings.TrimSpace(address),
		Description: strings.TrimSpace(description),
	}

	if gym.Name == "" {
		return nil, ErrInvalidGymName
	}

	return u.gymRepo.CreateGym(gym)
}

func (u *GymUsecase) CreateClass(userID int, userRole domain.UserRole, gymID int, name string, maxCapacity int) (*domain.Class, error) {
	gym, err := u.GetGym(gymID)
	if err != nil {
		return nil, err
	}

	if userRole != domain.RoleAdmin && gym.OwnerID != userID {
		return nil, ErrNotGymOwner
	}

	class := domain.Class{
		GymID:       gymID,
		Name:        strings.TrimSpace(name),
		MaxCapacity: maxCapacity,
	}

	if class.Name == "" {
		return nil, ErrInvalidClassName
	}
	if class.MaxCapacity <= 0 {
		return nil, ErrInvalidMaxCapacity
	}

	return u.gymRepo.CreateClass(class)
}

func (u *GymUsecase) CreateSession(userID int, userRole domain.UserRole, gymID, classID int, startTime, endTime time.Time, price float64) (*domain.Session, error) {
	gym, err := u.GetGym(gymID)
	if err != nil {
		return nil, err
	}

	if userRole != domain.RoleAdmin && gym.OwnerID != userID {
		return nil, ErrNotGymOwner
	}

	if !endTime.After(startTime) {
		return nil, ErrInvalidSessionTime
	}
	if price <= 0 {
		return nil, ErrInvalidSessionPrice
	}

	class, err := u.gymRepo.GetClassByID(classID)
	if err != nil {
		return nil, err
	}
	if class.GymID != gymID {
		return nil, ErrClassDoesNotBelongToGym
	}

	session := domain.Session{
		ClassID:        classID,
		StartTime:      startTime,
		EndTime:        endTime,
		AvailableSlots: class.MaxCapacity,
		Price:          price,
		Status:         "active",
	}

	return u.gymRepo.CreateSession(gymID, session)
}

func (u *GymUsecase) AssignTrainer(ctx context.Context, userID int, userRole domain.UserRole, gymID int, trainerID int) error {
	gym, err := u.gymRepo.GetGymByID(gymID)
	if err != nil {
		return err
	}

	if userRole != domain.RoleAdmin && gym.OwnerID != userID {
		return ErrNotGymOwner
	}

	trainer, err := u.userRepo.GetByID(trainerID)
	if err != nil {
		return err
	}

	if trainer.Role != domain.RoleTrainer {
		return ErrUserIsNotTrainer
	}

	return u.gymRepo.AssignTrainer(gymID, trainerID)
}

func (u *GymUsecase) AssignTrainerToSession(ctx context.Context, userID int, userRole domain.UserRole, sessionID, trainerID int) error {
	session, err := u.sessionRepo.GetByID(ctx, sessionID)
	if err != nil {
		return err
	}

	class, err := u.gymRepo.GetClassByID(session.ClassID)
	if err != nil {
		return err
	}

	gym, err := u.gymRepo.GetGymByID(class.GymID)
	if err != nil {
		return err
	}

	if userRole != domain.RoleAdmin && gym.OwnerID != userID {
		return ErrNotGymOwner
	}

	trainer, err := u.userRepo.GetByID(trainerID)
	if err != nil {
		return err
	}

	if trainer.Role != domain.RoleTrainer {
		return ErrUserIsNotTrainer
	}

	isAssigned, err := u.gymRepo.IsTrainerInGym(gym.ID, trainerID)
	if err != nil {
		return err
	}
	if !isAssigned {
		return ErrTrainerNotAssignedToGym
	}

	return u.sessionRepo.AssignTrainer(ctx, sessionID, trainerID)
}

func (u *GymUsecase) ListGymTrainers(userID int, userRole domain.UserRole, gymID int) ([]domain.TrainerInfo, error) {
	gym, err := u.gymRepo.GetGymByID(gymID)
	if err != nil {
		return nil, err
	}

	if userRole != domain.RoleAdmin && gym.OwnerID != userID {
		return nil, ErrNotGymOwner
	}

	return u.gymRepo.ListTrainersByGymID(gymID)
}

func (u *GymUsecase) ListAllMyGymTrainers(ownerID int) ([]domain.GymWithTrainers, error) {
	gyms, err := u.gymRepo.ListGymsByOwnerID(ownerID)
	if err != nil {
		return nil, err
	}

	var result []domain.GymWithTrainers
	for _, g := range gyms {
		trainers, err := u.gymRepo.ListTrainersByGymID(g.ID)
		if err != nil {
			return nil, err
		}
		result = append(result, domain.GymWithTrainers{
			GymID:    g.ID,
			GymName:  g.Name,
			Trainers: trainers,
		})
	}

	return result, nil
}
