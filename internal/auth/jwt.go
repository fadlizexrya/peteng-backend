package auth

import (
	"errors"
	"os"
	"strconv"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	secret     []byte
	expireTime time.Duration
}

func NewJWTManager() *JWTManager {
	secret := os.Getenv("JWT_SECRET")

	if secret == "" {
		panic("JWT_SECRET belum diset di environment")
	}

	expireHours := 24

	if value := os.Getenv("JWT_EXPIRE_HOURS"); value != "" {
		if hours, err := strconv.Atoi(value); err == nil && hours > 0 {
			expireHours = hours
		}
	}

	return &JWTManager{
		secret:     []byte(secret),
		expireTime: time.Duration(expireHours) * time.Hour,
	}
}

type Claims struct {
	UserID int    `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`

	jwt.RegisteredClaims
}

func (m *JWTManager) GenerateToken(
	userID int,
	email string,
	role string,
) (string, error) {

	now := time.Now()

	claims := Claims{
		UserID: userID,
		Email:  email,
		Role:   role,

		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(userID),
			Issuer:    "peteng-backend",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(m.expireTime)),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString(m.secret)
}

func (m *JWTManager) ValidateToken(
	tokenString string,
) (*Claims, error) {

	token, err := jwt.ParseWithClaims(
		tokenString,
		&Claims{},
		func(token *jwt.Token) (interface{}, error) {

			if token.Method.Alg() != jwt.SigningMethodHS256.Alg() {
				return nil, errors.New("algoritma JWT tidak valid")
			}

			return m.secret, nil
		},
		jwt.WithIssuer("peteng-backend"),
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*Claims)

	if !ok || !token.Valid {
		return nil, errors.New("token tidak valid")
	}

	return claims, nil
}
