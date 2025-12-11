package models

import "time"

type Role struct {
	ID   int  `gorm:"primaryKey"`  
	Name string `gorm:"type:varchar(255);unique;not null"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt *time.Time 
	DeletedAt time.Time 
	Users []User `gorm:"many2many:user_roles"`
}

// table name
func (Role) TableName() string {
	return "roles"
}