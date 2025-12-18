package message

import (
	"encoding/json"
	"fmt"
	"product-service/config"
	"product-service/internal/core/domain/entities"

	"github.com/labstack/gommon/log"
	"github.com/streadway/amqp"
)

type IPublishRabbitMQ interface {
	PublishProductToQueue(product entities.ProductEntity) error
	DeleteProductFromQueue(productID int64) error
}

type PublishRabbitMQ struct {
	cfg *config.Config
}

// DeleteProductFromQueue implements [IPublishRabbitMQ].
func (p *PublishRabbitMQ) DeleteProductFromQueue(productID int64) error {
	conn, err := p.cfg.NewRabbitMQ()
	if err != nil {
		log.Errorf("[DeleteProductFromQueue-1] Failed to connect to RabbitMQ: %v", err)
		return err
	}	
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("[DeleteProductFromQueue-2] Failed to open a channel: %v", err)
		return err
	}

	defer ch.Close()
	q, err := ch.QueueDeclare(
		p.cfg.PublisherName.ProductDelete,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Errorf("[DeleteProductFromQueue-3] Failed to declare queue: %v", err)
		return err
	}

	data, _ := json.Marshal(map[string]string{"ProductID": fmt.Sprintf("%d", productID)})
	err = ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        data,
			DeliveryMode: amqp.Persistent,
		},
	)

	if err != nil {
		log.Errorf("[DeleteProductFromQueue-4] Failed to publish message: %v", err)
		return err
	}

	return nil
}

// PublishProductToQueue implements [IPublishRabbitMQ].
func (p *PublishRabbitMQ) PublishProductToQueue(product entities.ProductEntity) error {
	conn, err := p.cfg.NewRabbitMQ()
	if err != nil {
		log.Errorf("[PublishProductToQueue-1] Failed to connect to RabbitMQ: %v", err)
		return err
	}

	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("[PublishProductToQueue-2] Failed to open a channel: %v", err)
		return err
	}

	defer ch.Close()
	q, err := ch.QueueDeclare(
		p.cfg.PublisherName.ProductPublish,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Errorf("[PublishProductToQueue-3] Failed to declare queue: %v", err)
		return err
	}

	data, _ := json.Marshal(product)
	err = ch.Publish(
		"",
		q.Name,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         data,
			DeliveryMode: amqp.Persistent,
		},
	)
	if err != nil {
		log.Errorf("[PublishProductToQueue-4] Failed to publish message: %v", err)
		return err
	}

	return nil
}

func NewPublishRabbitMQ(cfg *config.Config) IPublishRabbitMQ {
	return &PublishRabbitMQ{
		cfg: cfg,
	}
}
