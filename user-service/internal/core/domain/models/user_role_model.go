package models

import "time"

type UserRole struct {
	ID   int   `gorm:"primaryKey"`
	UserID int `gorm:"index"`
	RoleID int `gorm:"index"`
	CreatedAt time.Time 
	UpdatedAt time.Time 
	DeletedAt *time.Time 
}

//change table name
func (UserRole) TableName() string {
	return "user_roles"
}