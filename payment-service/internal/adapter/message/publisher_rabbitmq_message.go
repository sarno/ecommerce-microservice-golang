package message

import (
	"encoding/json"
	"fmt"
	"payment-service/config"
	"payment-service/internal/core/domain/entity"

	"github.com/labstack/gommon/log"
	"github.com/streadway/amqp"
)

type IPublishRabbitMQ interface {
	PublishPaymentSuccess(payment entity.PaymentEntity) error
}

type PublishRabbitMQ struct {
	cfg *config.Config
}

// PublishPaymentSuccess implements [IPublishRabbitMQ].
func (p *PublishRabbitMQ) PublishPaymentSuccess(payment entity.PaymentEntity) error {
	conn, err := p.cfg.NewRabbitMQ()
	if err != nil {
		log.Errorf("[PublishPaymentSuccess-1] Failed to connect to RabbitMQ: %v", err)
		return err
	}

	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("[PublishPaymentSuccess-2] Failed to open a channel: %v", err)
		return err
	}

	defer ch.Close()

	q, err := ch.QueueDeclare(
		p.cfg.PublisherName.PaymentSuccess,
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Errorf("[PublishPaymentSuccess-3] Failed to declare queue: %v", err)
		return err
	}

	paymentOrder := map[string]string{
		"orderID":       fmt.Sprintf("%d", payment.OrderID),
		"paymentMethod": payment.PaymentMethod,
	}

	data, _ := json.Marshal(paymentOrder)
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
		log.Errorf("[PublishPaymentSuccess-4] Failed to publish message: %v", err)
		return err
	}

	return nil
}

func NewPublishRabbitMQ(cfg *config.Config) IPublishRabbitMQ {
	return &PublishRabbitMQ{
		cfg: cfg,
	}
}
