package handlers

import (
	"net/http"
	"order-service/config"
	"order-service/internal/adapter"
	"order-service/internal/adapter/handlers/response"
	"order-service/internal/core/domain/entity"
	"order-service/internal/core/service"
	"order-service/utils/conv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

type IOrderHandler interface {
	GetAllAdmin(c echo.Context) error
}

type orderHandler struct {
	orderService service.IOrderService
}

// GetAllAdmin implements [IOrderHandler].
func (o *orderHandler) GetAllAdmin(c echo.Context) error {
	var (
		ctx        = c.Request().Context()
		respOrders = []response.OrderAdminList{}
	)

	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[OrderHandler-1] GetAllAdmin: %s", "data token not found")
		return c.JSON(http.StatusNotFound, response.ResponseError("data token not found"))
	}
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

	reqEntity := entity.QueryStringEntity{
		Search: search,
		Status: status,
		Page:   page,
		Limit:  perPage,
	}

	results, totalData, totalPage, err := o.orderService.GetAll(ctx, reqEntity, user)
	if err != nil {
		log.Errorf("[OrderHandler-1] GetAllAdmin: %v", err)
		if err.Error() == "404" {
			return c.JSON(http.StatusNotFound, response.ResponseError("data not found"))
		}
		return c.JSON(http.StatusInternalServerError, response.ResponseError(err.Error()))
	}

	for _, result := range results {
		respOrders = append(respOrders, response.NewOrderAdminList(result))
	}
	
	return c.JSON(http.StatusOK, response.ResponseSuccessWithPagination("success", respOrders, page, totalData, totalPage, perPage))
}

func NewOrderHandler(orderService service.IOrderService, e *echo.Echo, cfg *config.Config) IOrderHandler {
	ordHandler := &orderHandler{
		orderService: orderService,
	}

	e.Use(middleware.Recover())
	mid := adapter.NewMiddlewareAdapter(cfg)
	adminGroup := e.Group("/admin", mid.CheckToken())
	adminGroup.GET("/orders", ordHandler.GetAllAdmin)

	return ordHandler
}
