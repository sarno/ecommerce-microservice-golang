package config

import (
	"fmt"

	"github.com/labstack/gommon/log"
	"github.com/streadway/amqp"
)

func (cfg Config) NewRabbitMQ() (*amqp.Connection, error) {
	connect := fmt.Sprintf("amqp://%s:%s@%s:%s/",
		cfg.RabbitMQ.User,
		cfg.RabbitMQ.Password,
		cfg.RabbitMQ.Host,
		cfg.RabbitMQ.Port,
	)

	connection, err := amqp.Dial(connect)

	if err != nil {
		log.Errorf("[NewRabbitMQ-1] NewRabbitMQ: %v", err)
		return nil, err
	}

	return connection, nil
}
