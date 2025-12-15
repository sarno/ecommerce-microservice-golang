package repositories

import (
	"context"
	"errors"
	"notification-service/internal/core/domain/entities"
	"notification-service/internal/core/domain/models"
	"time"

	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
)

// buat interface

type INotifRepository interface {
	// buat method
	CreateNotification(ctx context.Context, notification *entities.NotificationEntity) error
	MarkAsSent(notifID uint) error
}

type NotifRepository struct {
	db *gorm.DB
}

// MarkAsSent implements [INotifRepository].
func (n *NotifRepository) MarkAsSent(notifID uint) error {
	modelNotif := models.Notification{}

	if err := n.db.First(&modelNotif, notifID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("404")
			log.Errorf("[NotifRepository-1] MarkAsSent: Notification not found")
			return err
		}
		log.Errorf("[NotifRepository-2] MarkAsSent: %v", err)
		return err
	}

	modelNotif.Status = entities.NotificationStatusSent
	if err := n.db.UpdateColumns(&modelNotif).Error; err != nil {
		log.Errorf("[MarkAsSent-3] Failed to save notification: %v", err)
		return err
	}

	return nil
}

// CreateNotification implements [INotifRepository].
func (n *NotifRepository) CreateNotification(ctx context.Context, notification *entities.NotificationEntity) error {
	now := time.Now()
	notifMdl := models.Notification{
		NotificationType: notification.NotificationType,
		ReceiverID:       notification.ReceiverID,
		ReceiverEmail:    notification.ReceiverEmail,
		Subject:          notification.Subject,
		Message:          notification.Message,
		Status:           notification.Status,
		SentAt:           &now,
		ReadAt:           notification.ReadAt,
	}

	err := n.db.WithContext(ctx).Create(&notifMdl).Error
	if err != nil {
		return err
	}

	return nil
}

func NewNotifRepository(db *gorm.DB) INotifRepository {
	return &NotifRepository{
		db: db,
	}
}
