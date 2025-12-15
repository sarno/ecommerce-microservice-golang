package response

import "notification-service/internal/core/domain/entities"

type ListResponse struct {
	ID      uint   `json:"id"`
	Subject string `json:"subject"`
	Status  entities.NotificationStatus `json:"status"`
	SentAt  string `json:"sent_at"`
}

type DetailResponse struct {
	ID               uint   `json:"id"`
	Subject          string `json:"subject"`
	Message          string `json:"message"`
	Status           entities.NotificationStatus `json:"status"`
	SentAt           string `json:"sent_at"`
	ReadAt           string `json:"read_at"`
	NotificationType entities.NotificationType `json:"notification_type"`
}