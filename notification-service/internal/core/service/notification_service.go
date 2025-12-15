package service

import (
	"context"
	"notification-service/internal/adapter/repositories"
	"notification-service/internal/core/domain/entities"
	"notification-service/utils"

	"github.com/labstack/gommon/log"
)

// buat interface

type INotifService interface {
	CreateNotification(ctx context.Context, notification *entities.NotificationEntity) error
	SendPushNotification(ctx context.Context, notification entities.NotificationEntity)
}

// buat struct
type NotifService struct {
	notifRepository repositories.INotifRepository
}

// SendPushNotification implements [INotifService].
func (n *NotifService) SendPushNotification(ctx context.Context, notification entities.NotificationEntity) {
	if notification.ReceiverID == nil {
		return
	}

	conn := utils.GetWebSocketClientConn(*notification.ReceiverID)
	if conn == nil {
		log.Errorf("[SendPushNotification-1] WebSocket connection not found for user %d", *notification.ReceiverID)
		return
	}

	msg := map[string]interface{}{
		"type":    notification.NotificationType,
		"subject": notification.Subject,
		"message": notification.Message,
		"sent_at": notification.SentAt,
	}

	if err := conn.WriteJSON(msg); err != nil {
		log.Errorf("[SendPushNotification-2] Error sending push notification: %v", err)
	}

	if err := n.notifRepository.MarkAsSent(notification.ID); err != nil {
		log.Errorf("[SendPushNotification-3] Failed to mark notification as sent: %v", err)
	}
	
}

// CreateNotification implements [INotifService].
func (n *NotifService) CreateNotification(ctx context.Context, notification *entities.NotificationEntity) error {
	return n.notifRepository.CreateNotification(ctx, notification)
}

// buat method
func NewNotifService(notifRepository repositories.INotifRepository) INotifService {
	return &NotifService{
		notifRepository: notifRepository,
	}
}
