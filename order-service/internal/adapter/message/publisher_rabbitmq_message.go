package message

import (
	"encoding/json"
	"fmt"
	"order-service/config"
	"order-service/internal/core/domain/entity"
	"order-service/utils"

	"github.com/labstack/gommon/log"
	"github.com/streadway/amqp"
)

type IPublisherRabbitMQ interface {
	PublishOrderToQueue(order entity.OrderEntity) error
	PublishUpdateStock(productID int64, quantity int64)
	PublishSendEmailUpdateStatus(email, message, queuename string, userID int64) error

	PublishSendPushNotifUpdateStatus(message, queuename string, userID int64) error
	PublishUpdateStatus(queuename string, orderID int64, status string) error
	PublishDeleteOrderFromQueue(orderID int64) error
}

type PublisherRabbitMQ struct {
	cfg *config.Config
}

// PublishDeleteOrderFromQueue implements [IPublisherRabbitMQ].
func (p *PublisherRabbitMQ) PublishDeleteOrderFromQueue(orderID int64) error {
	conn, err := p.cfg.NewRabbitMQ()
	if err != nil {
		log.Errorf("[PublishDeleteOrderFromQueue-1] Failed to connect to RabbitMQ: %v", err)
		return err
	}

	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("[PublishDeleteOrderFromQueue-2] Failed to open a channel: %v", err)
		return err
	}

	defer ch.Close()

	queue, err := ch.QueueDeclare(
		p.cfg.PublisherName.PublisherDeleteOrder,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Errorf("[PublishDeleteOrderFromQueue-3] Failed to declare a queue: %v", err)
		return err
	}

	order := map[string]string{
		"orderID": fmt.Sprintf("%d", orderID),
	}

	body, err := json.Marshal(order)
	if err != nil {
		log.Errorf("[PublishDeleteOrderFromQueue-4] Failed to marshal JSON: %v", err)
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

// PublishUpdateStatus implements [IPublisherRabbitMQ].
func (p *PublisherRabbitMQ) PublishUpdateStatus(queuename string, orderID int64, status string) error {
	conn, err := p.cfg.NewRabbitMQ()
	if err != nil {
		log.Errorf("[PublishUpdateStatus-1] Failed to connect to RabbitMQ: %v", err)
		return err
	}

	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("[PublishUpdateStatus-2] Failed to open a channel: %v", err)
		return err
	}

	defer ch.Close()

	queue, err := ch.QueueDeclare(
		queuename,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Errorf("[PublishUpdateStatus-3] Failed to declare a queue: %v", err)
		return err
	}

	orderStatus := map[string]string{
		"orderID": fmt.Sprintf("%d", orderID),
		"status":  status,
	}

	body, err := json.Marshal(orderStatus)
	if err != nil {
		log.Errorf("[PublishUpdateStatus-4] Failed to marshal JSON: %v", err)
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

// PublishSendPushNotifUpdateStatus implements [IPublisherRabbitMQ].
func (p *PublisherRabbitMQ) PublishSendPushNotifUpdateStatus(message string, queuename string, userID int64) error {
	conn, err := p.cfg.NewRabbitMQ()
	if err != nil {
		log.Errorf("[PublishSendEmailUpdateStatus-1] Failed to connect to RabbitMQ: %v", err)
		return err
	}

	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("[PublishSendEmailUpdateStatus-2] Failed to open a channel: %v", err)
		return err
	}

	defer ch.Close()

	queue, err := ch.QueueDeclare(
		queuename,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Errorf("[PublishSendEmailUpdateStatus-3] Failed to declare a queue: %v", err)
		return err
	}

	notifType := "EMAIL"
	if queuename == utils.PUSH_NOTIF {
		notifType = "PUSH"
	}

	notification := map[string]interface{}{
		"receiver_email":    "",
		"message":           message,
		"subject":           "Update Status Order",
		"type":              "UPDATE_STATUS",
		"receiver_id":       userID,
		"notification_type": notifType,
	}

	body, err := json.Marshal(notification)
	if err != nil {
		log.Errorf("[PublishSendEmailUpdateStatus-4] Failed to marshal JSON: %v", err)
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

// PublishSendEmailUpdateStatus implements [IPublisherRabbitMQ].
func (p *PublisherRabbitMQ) PublishSendEmailUpdateStatus(email string, message string, queuename string, userID int64) error {
	conn, err := p.cfg.NewRabbitMQ()
	if err != nil {
		log.Errorf("[PublishSendEmailUpdateStatus-1] Failed to connect to RabbitMQ: %v", err)
		return err
	}

	defer conn.Close()
	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("[PublishSendEmailUpdateStatus-2] Failed to open a channel: %v", err)
		return err
	}

	defer ch.Close()
	queue, err := ch.QueueDeclare(
		queuename,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Errorf("[PublishSendEmailUpdateStatus-3] Failed to declare a queue: %v", err)
		return err
	}

	notifType := "EMAIL"
	if queuename == utils.PUSH_NOTIF {
		notifType = "PUSH"
	}

	notification := map[string]interface{}{
		"receiver_email":    email,
		"message":           message,
		"subject":           "Update Status Order",
		"type":              "UPDATE_STATUS",
		"receiver_id":       userID,
		"notification_type": notifType,
	}

	body, err := json.Marshal(notification)
	if err != nil {
		log.Errorf("[PublishSendEmailUpdateStatus-4] Failed to marshal JSON: %v", err)
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

// PublishUpdateStock implements [IPublisherRabbitMQ].
func (p *PublisherRabbitMQ) PublishUpdateStock(productID int64, quantity int64) {
	conn, err := p.cfg.NewRabbitMQ()
	if err != nil {
		log.Errorf("[PublishUpdateStock-1] Failed to connect to RabbitMQ: %v", err)
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("[PublishUpdateStock-2] Failed to open a channel: %v", err)
		return
	}

	defer ch.Close()

	q, err := ch.QueueDeclare(p.cfg.PublisherName.ProductUpdateStock, true, false, false, false, nil)
	if err != nil {
		log.Errorf("[PublishUpdateStock-3] Failed to declare a queue: %v", err)
		return
	}

	order := entity.PublishOrderItemEntity{
		ProductID: productID,
		Quantity:  quantity,
	}

	data, err := json.Marshal(order)
	if err != nil {
		log.Errorf("[PublishUpdateStock-4] Failed to marshal order: %v", err)
		return
	}

	err = ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        data,
		},
	)

	if err != nil {
		log.Errorf("[PublishUpdateStock-5] Failed to publish message: %v", err)
		return
	}
	log.Info("Pesan order dikirim ke RabbitMQ")
}

// PublishOrderToQueue implements [IPublisherRabbitMQ].
func (p *PublisherRabbitMQ) PublishOrderToQueue(order entity.OrderEntity) error {
	conn, err := p.cfg.NewRabbitMQ()

	if err != nil {
		log.Errorf("[PublishOrderToQueue-1] Failed to connect to RabbitMQ: %v", err)
		return err
	}

	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("[PublishOrderToQueue-2] Failed to open a channel: %v", err)
		return err
	}

	defer ch.Close()

	q, err := ch.QueueDeclare(
		p.cfg.PublisherName.OrderPublish,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Errorf("[PublishOrderToQueue-3] Failed to declare queue: %v", err)
		return err
	}

	data, _ := json.Marshal(order)
	err = ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        data,
		},
	)
	if err != nil {
		log.Errorf("[PublishOrderToQueue-4] Failed to publish message: %v", err)
		return err
	}

	return nil
}

func NewPublisherRabbitMQ(cfg *config.Config) IPublisherRabbitMQ {
	return &PublisherRabbitMQ{
		cfg: cfg,
	}
}
