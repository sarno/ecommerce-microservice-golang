package repositories

import (
	"context"
	"errors"
	"fmt"
	"math"
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
	GetAll(ctx context.Context, query entities.NotifyQueryString) ([]entities.NotificationEntity, int64, int64, error)
	GetByID(ctx context.Context, notifID uint) (*entities.NotificationEntity, error)
	MarkAsRead(ctx context.Context, notifID uint) error
}

type NotifRepository struct {
	db *gorm.DB
}

// GetByID implements [INotifRepository].
func (n *NotifRepository) GetByID(ctx context.Context, notifID uint) (*entities.NotificationEntity, error) {
	modelNotif := models.Notification{}
	if err := n.db.Select("id", "subject", "status", "sent_at", "read_at", "message", "notification_type").First(&modelNotif, notifID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Errorf("[GetByID-1] Record not found for notification ID %d", notifID)
			err = errors.New("404")
			return nil, err
		}
		log.Errorf("[GetByID-2] Failed to find notification by ID: %v", err)
		return nil, err
	}

	return &entities.NotificationEntity{
		ID:               modelNotif.ID,
		Subject:          modelNotif.Subject,
		Status:           modelNotif.Status,
		SentAt:           modelNotif.SentAt,
		ReadAt:           modelNotif.ReadAt,
		Message:          modelNotif.Message,
		NotificationType: modelNotif.NotificationType,
	}, nil
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

// GetAll implements [INotifRepository].
func (n *NotifRepository) GetAll(ctx context.Context, query entities.NotifyQueryString) ([]entities.NotificationEntity, int64, int64, error) {
	modelNotifes := []models.Notification{}

	var countData int64
	offset := (query.Page - 1) * query.Limit

	sqlMain := n.db.
		Select("id", "subject", "status", "sent_at").
		Where("subject ILIKE ? OR message ILIKE ? OR status ILIKE ?", "%"+query.Search+"%", "%"+query.Search+"%", "%"+query.Status+"%")

	if query.UserID != 0 {
		sqlMain = sqlMain.Where("reciever_id = ?", query.UserID)
	}

	if query.IsRead {
		sqlMain = sqlMain.Where("read_at IS NOT NULL")
	}

	if err := sqlMain.Model(&modelNotifes).Count(&countData).Error; err != nil {
		log.Errorf("[NotificationRepository-1] GetAll: %v", err)
		return nil, 0, 0, err
	}

	order := fmt.Sprintf("%s %s", query.OrderBy, query.OrderType)
	totalPage := int(math.Ceil(float64(countData) / float64(query.Limit)))

	if err := sqlMain.Order(order).Limit(int(query.Limit)).Offset(int(offset)).Find(&modelNotifes).Error; err != nil {
		log.Errorf("[NotificationRepository-2] GetAll: %v", err)
		return nil, 0, 0, err
	}

	if len(modelNotifes) == 0 {
		err := errors.New("404")
		log.Infof("[NotificationRepository-3] GetAll: No notification found")
		return nil, 0, 0, err
	}

	notifEntities := []entities.NotificationEntity{}
	for _, val := range modelNotifes {
		notifEntities = append(notifEntities, entities.NotificationEntity{
			ID:      val.ID,
			Subject: val.Subject,
			Status:  val.Status,
			SentAt:  val.SentAt,
		})
	}

	return notifEntities, countData, int64(totalPage), nil
}

// MarkAsRead implements [INotifRepository].
func (n *NotifRepository) MarkAsRead(ctx context.Context, notifID uint) error {
	modelNotif := models.Notification{}

	if err := n.db.First(&modelNotif, notifID).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			log.Errorf("[MarkAsRead-1] Record not found for notification ID %d", notifID)
			err = errors.New("404")
			return err
		}
		log.Errorf("[MarkAsRead-2] Failed to find notification by ID: %v", err)
		return err
	}
	now := time.Now()
	modelNotif.ReadAt = &now
	if err := n.db.UpdateColumns(&modelNotif).Error; err != nil {
		log.Errorf("[MarkAsRead-3] Failed to save notification: %v", err)
		return err
	}
	return nil
}

func NewNotifRepository(db *gorm.DB) INotifRepository {
	return &NotifRepository{
		db: db,
	}
}
