package models

import (
	"time"

	"gorm.io/gorm"
)

type VerificationToken struct {
	ID        int `gorm:"primaryKey"`
	UserID    int `gorm:"not null,index:idx_verification_token_user_id"`
	Token     string `gorm:"type:varchar(255);not null"`
	TokenType string `gorm:"type:varchar(50);not null"`
	ExpiresAt time.Time `gorm:"type:timestamp;not null"`
	CreatedAt time.Time `gorm:"type:timestamp;default:CURRENT_TIMESTAMP"`
	UpdatedAt *time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`
	
	User      User `gorm:"foreignKey:UserID;constraints:OnUpdate:CASCADE,OnDelete:CASCADE;"`
}