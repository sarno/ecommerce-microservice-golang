package seeds

import (
	"user-service/internal/core/domain/models"

	"gorm.io/gorm"
)

func SeedUserRole(db *gorm.DB) {
	userRoles := []models.UserRole{
		{
			UserID: 1,
			RoleID: 1,
		},
	}

	for _, userRole := range userRoles {
		if err := db.FirstOrCreate(&userRole).Error; err != nil {
			return
		}
	}
}
