package app

import (
	"context"
	"log"
	"os"
	"os/signal"
	"payment-service/config"
	"payment-service/internal/adapter/handler"
	httpclient "payment-service/internal/adapter/http_client"
	"payment-service/internal/adapter/message"
	"payment-service/internal/adapter/repository"
	"payment-service/internal/core/service"
	"payment-service/utils/validator"
	"syscall"
	"time"

	"github.com/go-playground/validator/v10/translations/en"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)


func RunServer() {
	cfg := config.NewAppConfig()
	db, err := cfg.ConnectionPostgres()
	if err != nil {
		log.Fatalf("[RunServer-1] %v", err)
		return
	}

	httpClient := httpclient.NewHttpClient(cfg)
	serviceCall := httpclient.NewServiceCall(cfg, httpClient)
	midtrans := httpclient.NewMidtransClient(cfg)
	publisherRabbitMQ := message.NewPublishRabbitMQ(cfg)
	
	paymentRepo := repository.NewPaymentRepository(db.DB)
	paymentService := service.NewPaymentService(paymentRepo, cfg, serviceCall, midtrans, publisherRabbitMQ)

	e := echo.New()
	e.Use(middleware.CORS())

	customValidator := validator.NewValidator()
	en.RegisterDefaultTranslations(customValidator.Validator, customValidator.Translator)
	e.Validator = customValidator

	e.GET("/api/check", func(c echo.Context) error {
		return c.String(200, "OK")
	})

	handler.NewPaymentHandler(paymentService, e, cfg)

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