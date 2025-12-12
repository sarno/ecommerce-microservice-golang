package entity

import "time"

type VerificationTokenEntity struct {
	ID        int
	UserID    int
	Token     string
	TokenType string
	ExpiresAt time.Time
	User      UserEntity
}
