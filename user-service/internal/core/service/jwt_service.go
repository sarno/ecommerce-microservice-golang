package service

import (
	"user-service/config"

	"github.com/golang-jwt/jwt/v5"
)

type IJWTService interface {
	GenerateToken(userId int) (string, error)
	ValidateToken(token string) (*jwt.Token, error)
}

type JWTService struct {
	secretKey string
	issuer    string
	expiration int
}

// GenerateToken implements IJWTService.
func (j *JWTService) GenerateToken(userId int) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userId,
		"iss":  j.issuer,
		"exp":     j.expiration,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

// ValidateToken implements IJWTService.
func (j *JWTService) ValidateToken(token string) (*jwt.Token, error) {
	return jwt.Parse(token, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}
		return []byte(j.secretKey), nil
	})
}

func NewJWTService(cfg *config.Config) IJWTService {
	return &JWTService{
		secretKey: cfg.App.JwtSecretKey,
		issuer:    cfg.App.JwtIssuer,
		expiration: cfg.App.JwtExpire,
	}
}
