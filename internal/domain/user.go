package domain

import (
	"time"
)

type UserRole string

const (
	RoleAdmin    UserRole = "admin"
	RoleUser     UserRole = "user"
	RoleGymOwner UserRole = "gym_owner"
	RoleTrainer  UserRole = "trainer"
)

type User struct {
	ID           int       `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	FullName     string    `json:"full_name"`
	Role         UserRole  `json:"role"`
	Balance      float64   `json:"balance"`
	CreatedAt    time.Time `json:"created_at"`
}
