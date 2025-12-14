package seeds

import (
	"user-service/internal/core/domain/models"

	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
)

func SeedRole(db *gorm.DB) {
	roles := []models.Role{
		{
			Name: "admin",
		},
		{
			Name: "user",
		},
	}
	for _, role := range roles {
		if err := db.FirstOrCreate(&models.Role{}, &role).Error; err != nil {
			log.Fatalf("cannot seed roles table: %v", err)
		}
	}
}
