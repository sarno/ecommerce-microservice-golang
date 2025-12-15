package handlers

import (
	"net/http"
	"notification-service/config"
	"notification-service/utils"
	"strconv"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// interface
type IWebSocketHandler interface {
	HandleWebSocket(c echo.Context) error
}

type WebSocketHandler struct {
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// HandleWebSocket implements [IWebSocketHandler].
func (w *WebSocketHandler) HandleWebSocket(c echo.Context) error {
	userIdStr := c.QueryParam("user_id")
	userId, err := strconv.Atoi(userIdStr)
	if err != nil {
		
		return c.String(http.StatusBadRequest, "Invalid user_id")
	}

	conn, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}

	utils.AddWebSocketClientConn(userId, conn)
	defer utils.RemoveWebSocketClientConn(userId)
	defer conn.Close()

	for {
		if _,_, err := conn.NextReader(); err != nil {
			break
		}
	}
	
	return nil

}


func NewWebSocketHandler(e *echo.Echo, cfg *config.Config) IWebSocketHandler {

	wsHandler := &WebSocketHandler{}
	e.Use(middleware.Recover())
	e.GET("/ws", wsHandler.HandleWebSocket)

	return wsHandler
}
