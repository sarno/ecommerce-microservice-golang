package app

import (
	"context"
	"log"
	"notification-service/config"
	"notification-service/internal/adapter/handlers"
	"notification-service/internal/adapter/rabbitmq"
	"notification-service/internal/adapter/repositories"
	"notification-service/internal/core/service"
	"notification-service/utils"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func RunServer() {
	cfg :=config.NewConfig()

	db, err := cfg.ConnectionPostgres()
	if err != nil {
		log.Fatal(err)
	}

	notifRepository := repositories.NewNotifRepository(db.DB)
	notifService := service.NewNotifService(notifRepository)

	emailService := service.NewEmailService(cfg)
	rabbitMQAdapter := rabbitmq.NewConsumeRabbitMQ(notifService, emailService )

	e := echo.New()
	e.Use(middleware.CORS())

	go func() {
		err = rabbitMQAdapter.ConsumeMessage(utils.NOTIF_EMAIL_VERIFICATION)
		if err != nil {
			e.Logger.Errorf("Failed to consume RabbitMQ for %s: %v", utils.NOTIF_EMAIL_VERIFICATION, err)
		}
	}()

	go func() {
		err = rabbitMQAdapter.ConsumeMessage(utils.NOTIF_EMAIL_FORGOT_PASSWORD)
		if err != nil {
			e.Logger.Errorf("Failed to consume RabbitMQ for %s: %v", utils.NOTIF_EMAIL_FORGOT_PASSWORD, err)
		}
	}()

	go func() {
		err = rabbitMQAdapter.ConsumeMessage(utils.NOTIF_EMAIL_CREATE_CUSTOMER)
		if err != nil {
			e.Logger.Errorf("Failed to consume RabbitMQ for %s: %v", utils.NOTIF_EMAIL_CREATE_CUSTOMER, err)
		}
	}()

	go func() {
		err = rabbitMQAdapter.ConsumeMessage(utils.NOTIF_EMAIL_UPDATE_STATUS_ORDER)
		if err != nil {
			e.Logger.Errorf("Failed to consume RabbitMQ for %s: %v", utils.NOTIF_EMAIL_UPDATE_STATUS_ORDER, err)
		}
	}()

	go func() {
		err = rabbitMQAdapter.ConsumeMessage(utils.PUSH_NOTIF)
		if err != nil {
			e.Logger.Errorf("Failed to consume RabbitMQ for %s: %v", utils.PUSH_NOTIF, err)
		}
	}()

	handlers.NewNotificationHandler(notifService, e, cfg)

	go func() {
		if cfg.App.AppPort == "" {
			cfg.App.AppPort = os.Getenv("APP_PORT")
		}

		err = e.Start(":" + cfg.App.AppPort)
		if err != nil {
			log.Fatalf("[RunServer-2] %v", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	signal.Notify(quit, syscall.SIGTERM)
	<-quit

	log.Print("[RunServer-3] Shutting down server of 5 second...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	e.Shutdown(ctx)
}