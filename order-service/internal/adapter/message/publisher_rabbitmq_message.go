package message

import (
	"encoding/json"
	"order-service/config"
	"order-service/internal/core/domain/entity"

	"github.com/labstack/gommon/log"
	"github.com/streadway/amqp"
)

type IPublisherRabbitMQ interface {
	PublishOrderToQueue(order entity.OrderEntity) error
	PublishUpdateStock(productID int64, quantity int64)
}

type PublisherRabbitMQ struct {
	cfg *config.Config
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
