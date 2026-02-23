// internal/infrastructure/security/jwt_service.go
package security

import (
	"errors"
	"fmt"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"go.uber.org/zap"
)

type JWTService interface {
	ValidateAccessToken(token string) (userID, role, jti string, err error)
	GetAccessTokenExpiration(token string) (time.Time, error)
}

type AccessTokenClaims struct {
	UserID string `json:"sub"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

type JWTServiceImpl struct {
	secret string
	logger *zap.Logger
}

func NewJWTService(secret string, logger *zap.Logger) JWTService {
	return &JWTServiceImpl{secret: secret, logger: logger}
}

func (s *JWTServiceImpl) ValidateAccessToken(tokenString string) (string, string, string, error) {
	claims := &AccessTokenClaims{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secret), nil
	})
	if err != nil {
		return "", "", "", fmt.Errorf("error parsing token: %w", err)
	}
	if !token.Valid {
		return "", "", "", errors.New("invalid token")
	}
	return claims.UserID, claims.Role, claims.ID, nil
}

func (s *JWTServiceImpl) GetAccessTokenExpiration(tokenString string) (time.Time, error) {
	claims := &AccessTokenClaims{}
	_, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(s.secret), nil
	})
	if err != nil {
		return time.Time{}, fmt.Errorf("error parsing token expiration: %w", err)
	}
	if claims.ExpiresAt == nil {
		return time.Time{}, errors.New("token sin fecha de expiración")
	}
	return claims.ExpiresAt.Time, nil
}

var _ JWTService = (*JWTServiceImpl)(nil)
