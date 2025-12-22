package adapter

import (
	"encoding/json"
	"net/http"
	"order-service/config"
	"order-service/internal/adapter/handlers/response"
	"order-service/internal/core/domain/entity"
	"strings"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type IMiddlewareAdapter interface {
	CheckToken() echo.MiddlewareFunc
}

type middlewareAdapter struct {
	cfg *config.Config
}

// CheckToken implements [IMiddlewareAdapter].
func (m *middlewareAdapter) CheckToken() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			redisConn := config.NewConfig().NewRedisClient()
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				log.Errorf("[MiddlewareAdapter-1] CheckToken: %s", "missing or invalid token")
				return c.JSON(http.StatusUnauthorized, response.ResponseError("missing or invalid token"))
			}
			tokenString := strings.TrimPrefix(authHeader, "Bearer ")
			_, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}

				return []byte(m.cfg.App.JwtSecretKey), nil
			})
			if err != nil {
				log.Errorf("[MiddlewareAdapter-2] CheckToken: %s", err.Error())
				return c.JSON(http.StatusUnauthorized, response.ResponseError(err.Error()))
			}

			getSession, err := redisConn.Get(c.Request().Context(), tokenString).Result()
			if err != nil || len(getSession) == 0 {
				log.Errorf("[MiddlewareAdapter-3] CheckToken: %s", err.Error())
				return c.JSON(http.StatusUnauthorized, response.ResponseError(err.Error()))
			}
			
			jwtUserData := entity.JwtUserData{}
			err = json.Unmarshal([]byte(getSession), &jwtUserData)
			if err != nil {
				log.Errorf("[MiddlewareAdapter-4] CheckToken: %v", err)
				return c.JSON(http.StatusInternalServerError, response.ResponseError(err.Error()))
			}

			path := c.Request().URL.Path
			segments := strings.Split(strings.Trim(path, "/"), "/")
			if jwtUserData.RoleName == "user" && segments[0] == "admin" {
				log.Infof("[MiddlewareAdapter-5] CheckToken: %s", "customer cannot access admin routes")
				return c.JSON(http.StatusForbidden, response.ResponseError("customer cannot access admin routes"))
			}

			c.Set("user", getSession)
			return next(c)
		}
	}
}

func NewMiddlewareAdapter(cfg *config.Config) IMiddlewareAdapter {
	return &middlewareAdapter{
		cfg: cfg,
	}
}
