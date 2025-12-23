package message

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"order-service/config"
	"order-service/internal/core/domain/entity"

	"github.com/labstack/gommon/log"
)

func StartOrderConsumer() {
	conn, err := config.NewConfig().NewRabbitMQ()
	if err != nil {
		log.Errorf("[StartOrderConsumer-1] Failed to connect to RabbitMQ: %v", err)
		return
	}

	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Errorf("[StartOrderConsumer-2] Failed to open a channel: %v", err)
		return
	}

	defer ch.Close()

	q, err := ch.QueueDeclare(
		config.NewConfig().PublisherName.OrderPublish,
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Fatalf("[StartOrderConsumer-3] Failed to declare queue: %v", err)
		return
	}

	msgs, err := ch.Consume(
		q.Name,
		"",
		true,
		false,
		false,
		false,
		nil,
	)

	if err != nil {
		log.Fatalf("[StartOrderConsumer-4] Failed to register consumer: %v", err)
		return
	}
	log.Info("RabbitMQ Consumer order started...")

	esClient, err := config.NewConfig().InitElasticsearch()
	if err != nil {
		log.Errorf("[StartOrderConsumer-5] Failed initialize Elasticsearch client: %v", err)
		return
	}

	forever := make(chan bool)
	go func() {
		for d := range msgs {
			var order entity.OrderEntity
			err := json.Unmarshal(d.Body, &order)
			
			if err != nil {
				log.Errorf("[StartOrderConsumer-6] Error decoding message: %v", err)
				continue
			}

			// Convert order struct ke JSON
			orderJSON, err := json.Marshal(order)
			if err != nil {
				log.Errorf("[StartOrderConsumer-7] Error encoding order to JSON: %v", err)
				continue
			}

			// Indexing ke Elasticsearch
			res, err := esClient.Index(
				"orders",                   // Nama index di Elasticsearch
				bytes.NewReader(orderJSON), // Data JSON
				esClient.Index.WithDocumentID(fmt.Sprintf("%d", order.ID)), // ID dokumen
				esClient.Index.WithContext(context.Background()),
				esClient.Index.WithRefresh("true"),
			)

			if err != nil {
				log.Errorf("[StartOrderConsumer-8] Error indexing to Elasticsearch: %v", err)
				continue
			}
			defer res.Body.Close()

			body, _ := io.ReadAll(res.Body)
			log.Infof("[StartOrderConsumer-9] Order %d berhasil diindex ke Elasticsearch %v", order.ID, string(body))
		}
	}()

	log.Infof("[StartOrderConsumer-10] Waiting for messages. To exit press CTRL+C")
	<-forever
}

