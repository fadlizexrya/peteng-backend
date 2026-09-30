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
	IDWarga int    `json:"id_warga"`
	Email   string `json:"email"`

	jwt.RegisteredClaims
}

func (m *JWTManager) GenerateToken(
	idWarga int,
	email string,
) (string, error) {

	now := time.Now()

	claims := Claims{
		IDWarga: idWarga,
		Email:   email,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.Itoa(idWarga),
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

			if token.Method != jwt.SigningMethodHS256 {
				return nil, errors.New("algoritma JWT tidak valid")
			}

			return m.secret, nil
		},
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
