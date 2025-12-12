package message

import (
	"encoding/json"
	"user-service/config"
	"user-service/utils"

	"github.com/labstack/gommon/log"
	"github.com/streadway/amqp"
)


func PublishMessage(userId int, email, message, queueName, subject string) error  {
	conn, err := config.NewConfig().NewRabbitMQ()
	if err != nil {
		log.Errorf("[PublishMessage-1] PublishMessage: %v", err)
		return err
	}

	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("[PublishMessage-2] PublishMessage: %v", err)
		return err
	}

	defer ch.Close()

	queue, err := ch.QueueDeclare(
		queueName, // name
		false,     // durable
		false,     // delete when unused
		false,     // exclusive
		false,     // no-wait
		nil,       // arguments
	)

	if err != nil {
		log.Errorf("[PublishMessage-3] PublishMessage: %v", err)
		return err
	}

	notifType := "EMAIL"
	
	if queueName == utils.PUSH_NOTIF {
		notifType = "PUSH"
	}

	notification := map[string]interface{}{
		"receiver_email": email,
		"message": message,
		"receiver_id": userId,
		"subject": subject,
		"notification_type": notifType,
	}

	body, err := json.Marshal(notification)
	if err != nil {
		log.Errorf("[PublishMessage-4] PublishMessage: %v", err)
		return err
	}

	return ch.Publish(
		"",
		queue.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
}