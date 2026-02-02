package auth

import (
	"errors"
	"fmt"
	"log/slog"
	"time"

	"github.com/golang-jwt/jwt/v5"
)

type JWTManager struct {
	secretKey     []byte
	tokenDuration time.Duration
}

type Claims struct {
	UserID int64  `json:"user_id"`
	Email  string `json:"email"`
	Role   string `json:"role"`
	Status string `json:"status"`
	jwt.RegisteredClaims
}

func NewJWTManager(secret string, duration time.Duration) *JWTManager {
	return &JWTManager{
		secretKey:     []byte(secret), //  convert to []byte here
		tokenDuration: duration,
	}
}

// Generate a JWT token for a user
func (j *JWTManager) GenerateToken(userID int64, email, role string) (string, error) {
	claims := &Claims{
		UserID: userID,
		Email:  email,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(j.tokenDuration)),
			IssuedAt:  jwt.NewNumericDate(time.Now()), // good for auditing
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, err := token.SignedString(j.secretKey)
	if err != nil {
		// this usually only happens if the secret key is bad
		return "", fmt.Errorf("failed to sign token: %w", err)
	}
	return signed, nil
}

// Verify a JWT token and return claims
func (j *JWTManager) VerifyToken(tokenStr string) (*Claims, error) {
	token, err := jwt.ParseWithClaims(tokenStr, &Claims{}, func(token *jwt.Token) (interface{}, error) {
		// check : attacker could change the header of the JWT to alg: none,
		//  and some older libraries might accept it as "verified" without a signature.
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return j.secretKey, nil // ✅ must be []byte
	})
	if err != nil {
		// log the reason for failure (internal logic, not the token itself)
		slog.Debug("token verification failed", "error", err)
		return nil, fmt.Errorf("invalid token: %w", err)
	}

	claims, ok := token.Claims.(*Claims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid token claims")
	}

	return claims, nil
}
