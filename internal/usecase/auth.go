package usecase

import (
	"errors"
	"net/mail"
	"strings"
	"unicode"

	"github.com/yertaypert/gym-booking-api/internal/auth"
	"github.com/yertaypert/gym-booking-api/internal/domain"
	"golang.org/x/crypto/bcrypt"
)

var ErrEmailAlreadyExists = domain.ErrEmailAlreadyExists
var ErrInvalidEmail = domain.ErrInvalidEmail
var ErrInvalidFullName = domain.ErrInvalidFullName
var ErrWeakPassword = domain.ErrWeakPassword

type AuthUsecase struct {
	userRepo UserRepository
}

func NewAuthUsecase(repo UserRepository) *AuthUsecase {
	return &AuthUsecase{userRepo: repo}
}

func (u *AuthUsecase) Register(email, password, fullName string) error {
	email = normalizeEmail(email)
	fullName = strings.TrimSpace(fullName)

	if err := validateRegistrationInput(email, password, fullName); err != nil {
		return err
	}

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
	return err
}

func (u *AuthUsecase) Login(email, password string) (string, error) {
	email = normalizeEmail(email)

	user, err := u.userRepo.GetByEmail(email)
	if err != nil {
		if errors.Is(err, domain.ErrUserNotFound) {
			return "", domain.ErrUserNotFound
		}
		return "", err
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(user.PasswordHash),
		[]byte(password),
	)
	if err != nil {
		return "", domain.ErrInvalidPassword
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

func normalizeEmail(email string) string {
	return strings.ToLower(strings.TrimSpace(email))
}

func validateRegistrationInput(email, password, fullName string) error {
	if _, err := mail.ParseAddress(email); err != nil {
		return ErrInvalidEmail
	}

	if strings.TrimSpace(fullName) == "" {
		return ErrInvalidFullName
	}

	if !isStrongPassword(password) {
		return ErrWeakPassword
	}

	return nil
}

func isStrongPassword(password string) bool {
	if len(password) < 8 {
		return false
	}

	var hasUpper bool
	var hasLower bool
	var hasDigit bool

	for _, r := range password {
		switch {
		case unicode.IsUpper(r):
			hasUpper = true
		case unicode.IsLower(r):
			hasLower = true
		case unicode.IsDigit(r):
			hasDigit = true
		}
	}

	return hasUpper && hasLower && hasDigit
}
