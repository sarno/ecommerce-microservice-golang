package seeds

import (
	"user-service/internal/core/domain/models"
	"user-service/utils/conv"

	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
)

func SeedUser(db *gorm.DB) {
	passHash, err := conv.HashPassword("54321")

	if err != nil {
		return
	}

	var adminRole models.Role
	if err := db.Where("name = ?", "admin").First(&adminRole).Error; err != nil {
		adminRole = models.Role{Name: "admin"}
		db.FirstOrCreate(&adminRole)
	}


	user := models.User{
		Name: "admin",
		Email: "admin@example.com",
		Password: passHash,
		Phone: "08123456789",
		Photo: "admin",
		Address: "jl. xyz",
		Lat: "0.0",
		Lng: "0.0",
		IsVerified: true,
	}

	user.Roles = append(user.Roles, adminRole)

	if err := db.FirstOrCreate(&user).Error; err != nil {
		log.Fatal(err)
		return
	} 
}  