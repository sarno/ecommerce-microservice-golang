package app

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"
	"user-service/config"
	"user-service/internal/adapter/handler"
	"user-service/internal/adapter/repository"
	"user-service/internal/core/service"
	"user-service/utils/validator"

	"github.com/go-playground/validator/v10/translations/en"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

func RunServer() {
	cfg := config.NewConfig()

	db, err := cfg.ConnectionPostgres()
	if err != nil {
		log.Fatalf(
			"[RunServer-1] Failed to connect to database: %v",
			err,
		)
	}
	defer db.Close()

	userRepo := repository.NewUserRepository(db.DB)
	tokenRepo := repository.NewVerificationTokenRepository(db.DB)

	jwtService := service.NewJWTService(cfg)
	userService := service.NewUserService(userRepo, cfg, jwtService, tokenRepo)


	e := echo.New()
	e.Use(middleware.CORS())
	

	customValidator := validator.NewValidator()
	en.RegisterDefaultTranslations(customValidator.Validator, customValidator.Translator)
	e.Validator = customValidator

	e.GET("/api/check", func(c echo.Context) error {
		return c.String(200, "OK")
	})

	handler.NewUserHandler(e, userService, cfg, jwtService)

	go func() {
		if cfg.App.AppPort=="" {
			cfg.App.AppPort = os.Getenv("APP_PORT")
		}

		err = e.Start(":"+ cfg.App.AppPort)

		if err != nil {
			log.Fatalf(
				"[RunServer-2] Failed to start server: %v",
				err,
			)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt)
	signal.Notify(quit, syscall.SIGTERM)

	<-quit

	log.Print("Shutting down server of 5 seconds...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}

	log.Print("Server shut down")
	
	e.Shutdown(ctx)
}