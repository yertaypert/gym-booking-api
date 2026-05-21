package usecase

import (
	"errors"
	"testing"

	"golang.org/x/crypto/bcrypt"

	"github.com/yertaypert/gym-booking-api/internal/domain"
)

// ─── Хелпер ──────────────────────────────────────────────────────────────────

func newAuthUsecase(user *domain.User, repoErr error) *AuthUsecase {
	return NewAuthUsecase(&mockUserRepo{user: user, err: repoErr})
}

// ─── Register ────────────────────────────────────────────────────────────────

func TestRegister_InvalidEmail(t *testing.T) {
	uc := newAuthUsecase(nil, nil)
	err := uc.Register("not-an-email", "StrongPass1", "John Doe")
	if !errors.Is(err, ErrInvalidEmail) {
		t.Fatalf("expected ErrInvalidEmail, got %v", err)
	}
}

func TestRegister_WeakPassword(t *testing.T) {
	uc := newAuthUsecase(nil, nil)
	err := uc.Register("user@example.com", "weak", "John Doe")
	if !errors.Is(err, ErrWeakPassword) {
		t.Fatalf("expected ErrWeakPassword, got %v", err)
	}
}

func TestRegister_EmptyFullName(t *testing.T) {
	uc := newAuthUsecase(nil, nil)
	err := uc.Register("user@example.com", "StrongPass1", "   ")
	if !errors.Is(err, ErrInvalidFullName) {
		t.Fatalf("expected ErrInvalidFullName, got %v", err)
	}
}

func TestRegister_Success(t *testing.T) {
	// mockUserRepo.Create возвращает (1, nil) — регистрация проходит
	uc := newAuthUsecase(nil, nil)
	err := uc.Register("user@example.com", "StrongPass1", "John Doe")
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
}

func TestRegister_NormalizesEmail(t *testing.T) {
	// Email с пробелами и заглавными — должен нормализоваться и пройти валидацию
	uc := newAuthUsecase(nil, nil)
	err := uc.Register("  USER@Example.COM  ", "StrongPass1", "John Doe")
	if err != nil {
		t.Fatalf("expected success after normalization, got %v", err)
	}
}

// ─── Login ────────────────────────────────────────────────────────────────────

func TestLogin_UserNotFound(t *testing.T) {
	// mockUserRepo.GetByEmail вернёт ошибку — пользователь не найден
	uc := newAuthUsecase(nil, errors.New("not found"))
	_, err := uc.Login("user@example.com", "StrongPass1")
	if err == nil || err.Error() != "user not found" {
		t.Fatalf("expected 'user not found', got %v", err)
	}
}

func TestLogin_WrongPassword(t *testing.T) {
	// Создаём пользователя с реальным bcrypt-хешем правильного пароля
	hash, _ := bcrypt.GenerateFromPassword([]byte("CorrectPass1"), bcrypt.MinCost)
	user := &domain.User{
		ID:           1,
		Email:        "user@example.com",
		PasswordHash: string(hash),
		Role:         domain.RoleUser,
	}

	uc := newAuthUsecase(user, nil)
	_, err := uc.Login("user@example.com", "WrongPass1")
	if err == nil || err.Error() != "invalid password" {
		t.Fatalf("expected 'invalid password', got %v", err)
	}
}

func TestLogin_Success(t *testing.T) {
	hash, _ := bcrypt.GenerateFromPassword([]byte("CorrectPass1"), bcrypt.MinCost)
	user := &domain.User{
		ID:           1,
		Email:        "user@example.com",
		PasswordHash: string(hash),
		Role:         domain.RoleUser,
	}

	uc := newAuthUsecase(user, nil)
	token, err := uc.Login("user@example.com", "CorrectPass1")
	if err != nil {
		t.Fatalf("expected success, got %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}

func TestLogin_NormalizesEmail(t *testing.T) {
	// Логин с заглавными буквами в email — должен нормализоваться и найти пользователя
	hash, _ := bcrypt.GenerateFromPassword([]byte("CorrectPass1"), bcrypt.MinCost)
	user := &domain.User{
		ID:           1,
		Email:        "user@example.com",
		PasswordHash: string(hash),
		Role:         domain.RoleUser,
	}

	uc := newAuthUsecase(user, nil)
	token, err := uc.Login("  USER@Example.COM  ", "CorrectPass1")
	if err != nil {
		t.Fatalf("expected success after email normalization, got %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}
}
