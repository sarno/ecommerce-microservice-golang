package handlers

import (
	"net/http"
	"order-service/config"
	"order-service/internal/adapter"
	"order-service/internal/adapter/handlers/request"
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
	CreateOrder(c echo.Context) error
	GetDetailCustomer(c echo.Context) error
}

type orderHandler struct {
	orderService service.IOrderService
}

// GetDetailCustomer implements [IOrderHandler].
func (o *orderHandler) GetDetailCustomer(c echo.Context) error {
	var (
		ctx       = c.Request().Context()
		respOrder = response.OrderAdminDetail{}
	)

	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[OrderHandler-1] GetDetailCustomer: %s", "data token not found")
		return c.JSON(http.StatusNotFound, response.ResponseError("data token not found"))
	}

	orderIDStr := c.Param("orderID")
	if orderIDStr == "" {
		log.Errorf("[OrderHandler-2] GetDetailCustomer: %s", "orderID not found")
		return c.JSON(http.StatusNotFound, response.ResponseError("orderID not found"))
	}

	orderID, err := conv.StringToInt64(orderIDStr)
	if err != nil {
		log.Errorf("[OrderHandler-3] GetDetailCustomer: %v", err)
		return c.JSON(http.StatusInternalServerError, response.ResponseError(err.Error()))
	}

	order, err := o.orderService.GetDetailCustomer(ctx, orderID, user)
	if err != nil {
		log.Errorf("[OrderHandler-4] GetDetailCustomer: %v", err)
		if err.Error() == "404" {
			return c.JSON(http.StatusNotFound, response.ResponseError("data not found"))
		}
		return c.JSON(http.StatusInternalServerError, response.ResponseError(err.Error()))
	}

	respOrder.ID = order.ID
	respOrder.OrderCode = order.OrderCode
	respOrder.Status = order.Status
	respOrder.TotalAmount = order.TotalAmount
	respOrder.OrderDatetime = order.OrderDate
	respOrder.ShippingFee = order.ShippingFee
	respOrder.ShippingType = order.ShippingType
	respOrder.Remarks = order.Remarks
	respOrder.Customer = response.CustomerOrder{
		CustomerName:    order.BuyerName,
		CustomerPhone:   order.BuyerPhone,
		CustomerAddress: order.BuyerAddress,
		CustomerEmail:   order.BuyerEmail,
		CustomerID:      order.BuyerId,
	}


	for _, item := range order.OrderItems {
		respOrder.OrderDetail = append(respOrder.OrderDetail, response.OrderDetail{
			ProductName:  item.ProductName,
			ProductImage: item.ProductImage,
			ProductPrice: item.Price,
			Quantity:     item.Quantity,
		})
	}

	
	return c.JSON(http.StatusOK, response.ResponseSuccess("success", respOrder))
}

// CreateOrder implements [IOrderHandler].
func (o *orderHandler) CreateOrder(c echo.Context) error {
	var (
		ctx = c.Request().Context()
		req = request.CreateOrderRequest{}
	)

	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[OrderHandler-1] CreateOrder: %s", "data token not found")
		return c.JSON(http.StatusNotFound, response.ResponseError("data token not found"))
	}

	if err := c.Bind(&req); err != nil {
		log.Errorf("[OrderHandler-2] CreateOrder: %v", err)
		return c.JSON(http.StatusBadRequest, response.ResponseError(err.Error()))
	}

	if err := c.Validate(&req); err != nil {
		log.Errorf("[OrderHandler-3] CreateOrder: %v", err)
		return c.JSON(http.StatusUnprocessableEntity, response.ResponseError(err.Error()))
	}

	reqEntity := entity.OrderEntity{
		BuyerId:      req.BuyerID,
		OrderDate:    req.OrderDate,
		TotalAmount:  req.TotalAmount,
		ShippingType: req.ShippingType,
		Remarks:      req.Remarks,
		OrderTime:    req.OrderTime,
	}

	orderDetails := []entity.OrderItemEntity{}
	for _, val := range req.OrderDetails {
		orderDetails = append(orderDetails, entity.OrderItemEntity{
			ProductID: val.ProductID,
			Quantity:  val.Quantity,
		})
	}

	reqEntity.OrderItems = orderDetails
	orderID, err := o.orderService.CreateOrder(ctx, reqEntity, user)
	if err != nil {
		log.Errorf("[OrderHandler-4] CreateOrder: %v", err)
		return c.JSON(http.StatusInternalServerError, response.ResponseError(err.Error()))
	}

	return c.JSON(
		http.StatusCreated,
		response.ResponseSuccess("success", map[string]interface{}{
			"order_id": orderID,
		}))
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
	authGroup := e.Group("auth", mid.CheckToken())
	authGroup.POST("/orders", ordHandler.CreateOrder, mid.DistanceCheck())
	authGroup.GET("/orders/:orderID", ordHandler.GetDetailCustomer)

	adminGroup := e.Group("/admin", mid.CheckToken())
	adminGroup.GET("/orders", ordHandler.GetAllAdmin)

	return ordHandler
}
