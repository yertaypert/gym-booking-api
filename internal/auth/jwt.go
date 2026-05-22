package auth

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

var jwtKey = []byte("your_secret_key")

type Claims struct {
	UserID int    `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type AttendanceClaims struct {
	SessionID int    `json:"session_id"`
	Type      string `json:"type"`
	jwt.RegisteredClaims
}

func SetJWTSecret(secret string) {
	if secret == "" {
		return
	}
	jwtKey = []byte(secret)
}

func GenerateJWT(userID int, role string) (string, error) {
	expirationTime := time.Now().Add(24 * time.Hour)
	claims := &Claims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expirationTime),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtKey)
}

func ParseJWT(tokenString string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})

	if err != nil {
		return nil, err
	}

	if claims, ok := token.Claims.(*Claims); ok && token.Valid {
		return claims, nil
	}

	return nil, errors.New("invalid token")
}

func GenerateAttendanceQRToken(sessionID int, ttl time.Duration) (string, time.Time, error) {
	if ttl <= 0 {
		ttl = 15 * time.Minute
	}

	expiresAt := time.Now().Add(ttl)
	jti, err := randomTokenID()
	if err != nil {
		return "", time.Time{}, err
	}

	claims := &AttendanceClaims{
		SessionID: sessionID,
		Type:      "attendance_qr",
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(expiresAt),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ID:        jti,
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(jwtKey)
	if err != nil {
		return "", time.Time{}, err
	}

	return signed, expiresAt, nil
}

func ParseAttendanceQRToken(tokenString string) (*AttendanceClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AttendanceClaims{}, func(token *jwt.Token) (interface{}, error) {
		return jwtKey, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*AttendanceClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid attendance token")
	}
	if claims.Type != "attendance_qr" {
		return nil, errors.New("invalid attendance token")
	}

	return claims, nil
}

func randomTokenID() (string, error) {
	buf := make([]byte, 16)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	return hex.EncodeToString(buf), nil
}
