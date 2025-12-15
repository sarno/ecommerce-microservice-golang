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
		log.Errorf("[ConsumeMessage-1] Failed to connect to RabbitMQ: %v", err)
		return err
	}

	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("[ConsumeMessage-2] Failed to open a channel: %v", err) // More specific error message
		return err
	}

	defer ch.Close()

	// Declare the queue to ensure it exists
	_, err = ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)
	
	if err != nil {
		log.Errorf("[ConsumeMessage-Declare] Failed to declare queue '%s': %v", queueName, err)
		return err
	}

	msgs, err := ch.Consume(queueName, "", true, false, false, false, nil)
	if err != nil {
		log.Errorf("[ConsumeMessage-3] Failed to consume messages from queue '%s': %v", queueName, err) // More specific error message
		return err
	}

	for msg := range msgs {
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
	return &consumeRabbitMQ{
		notificationService: notificationService,
		emailService:        emailService,
	}
}
