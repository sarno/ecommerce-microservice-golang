package rabbitmq

import (
	"context"
	"encoding/json"
	"notification-service/config"
	"notification-service/internal/core/domain/entities" // Added import
	"notification-service/internal/core/service"

	"github.com/labstack/gommon/log"
)

type IConsumeRabbitMQ interface {
	ConsumeMessage(queueName string) error
}

type consumeRabbitMQ struct {
	notificationService service.INotifService
	emailService        service.IEmailService
}

// ConsumeMessage implements [IConsumeRabbitMQ].
func (c *consumeRabbitMQ) ConsumeMessage(queueName string) error {
	conn, err := config.NewConfig().NewRabbitMQ()

	if err != nil {
		log.Errorf("[ConsumeMessage-1] ConsumeMessage: %v", err)
		return err
	}

	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("[ConsumeMessage-2] ConsumeMessage: %v", err)
		return err
	}

	defer ch.Close()

	msqs, err := ch.Consume(
		queueName, // queue
		"",        // consumer
		true,      // auto-ack
		false,     // exclusive
		false,     // no-local
		false,     // no-wait
		nil,       // args
	)

	if err != nil {
		log.Errorf("[ConsumeMessage-3] ConsumeMessage: %v", err)
		return err
	}

	for msg := range msqs {
		var notificationEntity entities.NotificationEntity
		log.Infof("Received a message: %s", msg.Body)
		if err = json.Unmarshal(msg.Body, &notificationEntity); err != nil {
			log.Errorf("Failed to unmarshal JSON: %v", err)
			continue
		}

		notificationEntity.Status = entities.NotificationStatusPending // Changed to enum
		if notificationEntity.NotificationType == entities.NotificationTypeEmail {
			notificationEntity.Status = entities.NotificationStatusSent // Changed to enum
		}

		err = c.notificationService.CreateNotification(context.Background(), &notificationEntity)
		if err != nil {
			log.Errorf("Failed to create notification: %v", err)
			continue
		}

		go c.SendNotification(notificationEntity)
	}

	return nil
}

func (c *consumeRabbitMQ) SendNotification(notificationEntity entities.NotificationEntity) {
	switch notificationEntity.NotificationType {
	case entities.NotificationTypeEmail:
		err := c.emailService.SendEmail(*notificationEntity.ReceiverEmail, *notificationEntity.Subject, notificationEntity.Message)
		if err != nil {
			log.Errorf("Failed to send email: %v", err)
		}
	case entities.NotificationTypePush:
		c.notificationService.SendPushNotification(context.Background(), notificationEntity)
	}
}

func NewConsumeRabbitMQ(notificationService service.INotifService, emailService service.IEmailService) IConsumeRabbitMQ {
	return &consumeRabbitMQ{}
}
