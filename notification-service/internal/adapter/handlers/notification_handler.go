package handlers

import (
	"encoding/json"
	"net/http"
	"notification-service/config"
	"notification-service/internal/adapter"
	"notification-service/internal/adapter/handlers/response"
	"notification-service/internal/core/domain/entities"
	"notification-service/internal/core/service"
	"notification-service/utils/conv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

// interface

type INotificationHandler interface {
	GetAll(c echo.Context) error
	GetById(c echo.Context) error
	MarkAsRead(c echo.Context) error
}

// struct

type NotificationHandler struct {
	service service.INotifService
}

// GetAll implements [INotificationHandler].
func (n *NotificationHandler) GetAll(c echo.Context) error {
	var (
		ctx = c.Request().Context()
		respNotifs = []response.ListResponse{}
		jwtUserData = entities.JwtUserData{}
	)

	user := c.Get("user").(string)
	if err := c.Bind(&jwtUserData); err != nil {
		log.Errorf("[NotificationHandler-1] GetAll: %s", "data token not found")
		return err
	}

	err := json.Unmarshal([]byte(user), &jwtUserData)
	if err != nil {
		log.Errorf("[NotificationHandler-2] GetAll: %v", err)
		return err
	}

	userID := jwtUserData.UserID
	search := c.QueryParam("search")
	var page int64 = 1
	if pageStr := c.QueryParam("page"); pageStr != "" {
		page, _ = conv.StringToInt64(pageStr)
		if page <= 0 {
			page = 1
		}
	}

	var perPage int64 = 10
	if perPageStr := c.QueryParam("perPage"); perPageStr != "" {
		perPage, _ = conv.StringToInt64(perPageStr)
		if perPage <= 0 {
			perPage = 10
		}
	}

	status := ""
	if statusStr := c.QueryParam("status"); statusStr != "" {
		status = statusStr
	}

	orderBy := "created_at"
	if orderByStr := c.QueryParam("orderBy"); orderByStr != "" {
		orderBy = orderByStr
	}

	orderType := "desc"
	if orderTypeStr := c.QueryParam("orderType"); orderTypeStr != "" {
		orderType = orderTypeStr
	}

	isRead := false
	if isReadStr := c.QueryParam("isRead"); isReadStr != "" {
		if isReadStr == "true" {
			isRead = true
		}
	}

	reqEntity := entities.NotifyQueryString{
		Search:    search,
		Status:    status,
		Page:      page,
		Limit:     perPage,
		UserID:    uint(userID),
		OrderBy:   orderBy,
		OrderType: orderType,
		IsRead:    isRead,
	}

	results, totalData, totalPage, err := n.service.GetAll(ctx, reqEntity)
	if err != nil {
		log.Errorf("[NotificationHandler-3] GetAll: %v", err)
		if err.Error() == "404" {
			return c.JSON(http.StatusNotFound, response.Response("data not found", nil))
		}
		return c.JSON(http.StatusInternalServerError, response.Response(err.Error(), nil))
	}

	for _, result := range results {
		respNotifs = append(respNotifs, response.ListResponse{
			ID:      result.ID,
			Subject: *result.Subject,
			Status:  result.Status,
			SentAt:  result.SentAt.Format("2006-01-02 15:04:05"),
		})
	}

	return c.JSON(http.StatusOK, response.ResponseWithPagination("success", respNotifs, page, totalData, totalPage, perPage))
}

// GetById implements [INotificationHandler].
func (n *NotificationHandler) GetById(c echo.Context) error {
	var (
		ctx        = c.Request().Context()
		respDetail = response.DetailResponse{}
	)

	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[NotificationHandler-1] GetByID: %s", "data token not found")
		return c.JSON(http.StatusNotFound, response.Response("data token not found", nil))
	}

	idStr := c.Param("id")
	id, err := conv.StringToInt64(idStr)
	if err != nil {
		log.Errorf("[NotificationHandler-2] GetByID: %v", err)
		return c.JSON(http.StatusBadRequest, response.Response(err.Error(), nil))
	}

	result, err := n.service.GetByID(ctx, uint(id))
	if err != nil {
		log.Errorf("[NotificationHandler-3] GetByID: %v", err)
		if err.Error() == "404" {
			return c.JSON(http.StatusNotFound, response.Response("data not found", nil))
		}
		return c.JSON(http.StatusInternalServerError, response.Response(err.Error(), nil))
	}

	respDetail.ID = result.ID
	respDetail.Subject = *result.Subject
	respDetail.Message = result.Message
	respDetail.Status = result.Status
	respDetail.SentAt = result.SentAt.Format("2006-01-02 15:04:05")
	respDetail.ReadAt = result.ReadAt.Format("2006-01-02 15:04:05")
	respDetail.NotificationType = result.NotificationType

	return c.JSON(http.StatusOK, response.Response("success", respDetail))
}

// MarkAsRead implements [INotificationHandler].
func (n *NotificationHandler) MarkAsRead(c echo.Context) error {
	var (
		ctx = c.Request().Context()
	)

	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[NotificationHandler-1] MarkAsRead: %s", "data token not found")
		return c.JSON(http.StatusNotFound, response.Response("data token not found", nil))
	}

	idStr := c.Param("id")
	id, err := conv.StringToInt64(idStr)
	if err != nil {
		log.Errorf("[NotificationHandler-2] MarkAsRead: %v", err)
		return c.JSON(http.StatusBadRequest, response.Response(err.Error(), nil))
	}

	err = n.service.MarkAsRead(ctx, uint(id))
	if err != nil {
		log.Errorf("[NotificationHandler-3] MarkAsRead: %v", err)
		if err.Error() == "404" {
			return c.JSON(http.StatusNotFound, response.Response("data not found", nil))
		}
		return c.JSON(http.StatusInternalServerError, response.Response(err.Error(), nil))
	}
	return c.JSON(http.StatusOK, response.Response("success", nil))
}

func NewNotificationHandler(service service.INotifService, e *echo.Echo, cfg *config.Config) INotificationHandler {
	notifHandler := &NotificationHandler{
		service: service,
	}

	e.Use(middleware.Recover())
	mid := adapter.NewMiddlewareAdapter(cfg)
	authGroup := e.Group("auth", mid.AuthMiddleware())
	authGroup.GET("/notifications", notifHandler.GetAll)
	authGroup.GET("/notifications/:id", notifHandler.GetById)
	authGroup.PUT("/notifications/:id", notifHandler.MarkAsRead)

	return notifHandler
}
