package usecase

import (
	"errors"

	"github.com/lib/pq"
	"github.com/yertaypert/gym-booking-api/internal/auth"
	"github.com/yertaypert/gym-booking-api/internal/domain"
	"github.com/yertaypert/gym-booking-api/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

var ErrEmailAlreadyExists = errors.New("email already exists")

type AuthUsecase struct {
	userRepo *repository.UserRepository
}

func NewAuthUsecase(repo *repository.UserRepository) *AuthUsecase {
	return &AuthUsecase{userRepo: repo}
}

func (u *AuthUsecase) Register(email, password, fullName string) error {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	newUser := domain.User{
		Email:        email,
		PasswordHash: string(hashedPassword),
		FullName:     fullName,
		Role:         (domain.RoleUser),
		Balance:      0,
	}

	_, err = u.userRepo.Create(newUser)
	if err != nil {
		var pqErr *pq.Error
		if errors.As(err, &pqErr) && pqErr.Code == "23505" {
			return ErrEmailAlreadyExists
		}
	}

	return err
}

func (u *AuthUsecase) Login(email, password string) (string, error) {
	user, err := u.userRepo.GetByEmail(email)
	if err != nil {
		return "", errors.New("user not found")
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	)
	if err != nil {
		return "", errors.New("invalid password")
	}

	token, err := auth.GenerateJWT(user.ID, string(user.Role))
	if err != nil {
		return "", err
	}

	return token, nil
}

func (u *AuthUsecase) Me(userID int) (*domain.User, error) {
	user, err := u.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}
	return user, nil
}
