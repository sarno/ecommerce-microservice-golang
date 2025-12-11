package service

import (
	"time"
	"user-service/config"

	"github.com/golang-jwt/jwt/v5"
)

type IJWTService interface {
	GenerateToken(userId int) (string, error)
	ValidateToken(token string) (*jwt.Token, error)
}

type jwtService struct {
	secretKey  string
	issuer     string
	expiration int
}

// GenerateToken implements IJWTService.
func (j *jwtService) GenerateToken(userId int) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userId,
		"iss":     j.issuer,
		"exp":     time.Now().Add(time.Second * time.Duration(j.expiration)).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(j.secretKey))
}

// ValidateToken implements IJWTService.
func (j *jwtService) ValidateToken(encodetoken string) (*jwt.Token, error) {
	return jwt.Parse(encodetoken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrSignatureInvalid
		}

		return []byte(j.secretKey), nil
	})
}

func NewJWTService(cfg *config.Config) IJWTService {
	return &jwtService{
		secretKey:  cfg.App.JwtSecretKey,
		issuer:     cfg.App.JwtIssuer,
		expiration: cfg.App.JwtExpire,
	}
}
