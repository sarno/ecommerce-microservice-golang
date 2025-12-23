package app

import (
	"context"
	"log"
	"order-service/config"
	"order-service/internal/adapter/handlers"
	httpclient "order-service/internal/adapter/http_client"
	"order-service/internal/adapter/message"
	"order-service/internal/adapter/repository"
	"order-service/internal/core/service"
	"order-service/utils/validator"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10/translations/en"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func RunServer() {
	cfg := config.NewConfig()
	db, err := cfg.ConnectionPostgres()

	if err != nil {
		log.Fatalf("[RunServer-1] %v", err)
		return
	}

	elasticInit, err := cfg.InitElasticsearch()
	if err != nil {
		log.Fatalf("[RunServer-2] %v", err)
		return
	}

	publisher := message.NewPublisherRabbitMQ(cfg)
	orderRepo := repository.NewOrderRepository(db.DB)
	elasticRepo := repository.NewElasticRepository(elasticInit)

	httpClient := httpclient.NewHttpClient(cfg)
	orderService := service.NewOrderService(orderRepo, cfg, httpClient, publisher, elasticRepo)

	e := echo.New()
	e.Use(middleware.CORS())

	customValidator := validator.NewValidator()
	en.RegisterDefaultTranslations(customValidator.Validator, customValidator.Translator)

	e.Validator = customValidator
	e.GET("/api/check", func(c echo.Context) error {
		return c.String(200, "OK")
	})

	handlers.NewOrderHandler(orderService, e, cfg)

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