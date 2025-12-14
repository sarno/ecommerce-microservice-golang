package models

import "time"

type User struct {
	ID         int `gorm:"primaryKey"`
	Name       string
	Email      string
	Password   string
	Phone      string
	Photo      string
	Address    string
	Lat        string
	Lng        string
	IsVerified bool
	CreatedAt  time.Time
	UpdatedAt  time.Time
	DeletedAt  time.Time
	Roles      []Role `gorm:"many2many:user_roles"`
}

// table name
func (User) TableName() string {
	return "users"
}
