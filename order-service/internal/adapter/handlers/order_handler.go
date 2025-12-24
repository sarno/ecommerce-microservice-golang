package handlers

import (
	"encoding/json"
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
	GetByIDAdmin(c echo.Context) error
	CreateOrder(c echo.Context) error
	GetDetailCustomer(c echo.Context) error

	GetAllCustomer(c echo.Context) error
	GetOrderByOrderCode(c echo.Context) error
	UpdateStatus(c echo.Context) error
}

type orderHandler struct {
	orderService service.IOrderService
}

// UpdateStatus implements [IOrderHandler].
func (o *orderHandler) UpdateStatus(c echo.Context) error {
	var (
		ctx = c.Request().Context()
		req = request.OrderUpdateStatusRequest{}
	)
	
	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[OrderHandler-1] UpdateStatus: %s", "data token not found")
		return c.JSON(http.StatusUnauthorized, response.ResponseError("data token not found"))
	}

	if err := c.Bind(&req); err != nil {
		log.Errorf("[OrderHandler-2] UpdateStatus: %v", err)
		return c.JSON(http.StatusBadRequest, response.ResponseError(err.Error()))
	}

	if err := c.Validate(&req); err != nil {
		log.Errorf("[OrderHandler-3] UpdateStatus: %v", err)
		return c.JSON(http.StatusUnprocessableEntity, response.ResponseError(err.Error()))
	}

	idParams := c.Param("orderID")
	if idParams == "" {
		log.Errorf("[OrderHandler-4] UpdateStatus: %s", "orderID not found")
		return c.JSON(http.StatusNotFound, response.ResponseError("orderID not found"))
	}

	orderID, err := conv.StringToInt64(idParams)
	if err != nil {
		log.Errorf("[OrderHandler-5] UpdateStatus: %v", err)
		return c.JSON(http.StatusInternalServerError, response.ResponseError(err.Error()))
	}

	reqEntity := entity.OrderEntity{
		Remarks: req.Remarks,
		Status:  req.Status,
		ID:      orderID,
	}

	err = o.orderService.UpdateStatus(ctx, reqEntity, user)
	if err != nil {
		log.Errorf("[OrderHandler-6] UpdateStatus: %v", err)
		if err.Error() == "404" {
			return c.JSON(http.StatusNotFound, response.ResponseError("data not found"))
		}

		if err.Error() == "400" {
			return c.JSON(http.StatusBadRequest, response.ResponseError("Invalid status transition"))
		}
		return c.JSON(http.StatusInternalServerError, response.ResponseError(err.Error()))
	
	}

	return c.JSON(http.StatusOK, response.ResponseSuccess("success", nil))

}

// GetOrderByOrderCode implements [IOrderHandler].
func (o *orderHandler) GetOrderByOrderCode(c echo.Context) error {
	var (
		ctx       = c.Request().Context()
		respOrder = response.OrderAdminDetail{}
	)

	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[OrderHandler-1] GetOrderByOrderCode: %s", "data token not found")
		return c.JSON(http.StatusNotFound, response.ResponseError("data token not found"))
	}

	orderCode := c.Param("orderCode")
	if orderCode == "" {
		log.Errorf("[OrderHandler-2] GetOrderByOrderCode: %s", "orderCode not found")
		return c.JSON(http.StatusNotFound, response.ResponseError("orderCode not found"))
	}
	order, err := o.orderService.GetOrderByOrderCode(ctx, orderCode, user)

	if err != nil {
		log.Errorf("[OrderHandler-3] GetOrderByOrderCode: %v", err)
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
	respOrder.Remarks = order.Remarks
	respOrder.PaymentMethod = order.PaymentMethod
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

// GetByIDAdmin implements [IOrderHandler].
func (o *orderHandler) GetByIDAdmin(c echo.Context) error {
	var (
		ctx       = c.Request().Context()
		respOrder = response.OrderAdminDetail{}
	)

	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[OrderHandler-1] GetByIDAdmin: %s", "data token not found")
		return c.JSON(http.StatusNotFound, response.ResponseError("data token not found"))
	}

	orderIDStr := c.Param("orderID")
	if orderIDStr == "" {
		log.Errorf("[OrderHandler-2] GetByIDAdmin: %s", "orderID not found")
		return c.JSON(http.StatusNotFound, response.ResponseError("orderID not found"))
	}

	orderID, err := conv.StringToInt64(orderIDStr)
	if err != nil {
		log.Errorf("[OrderHandler-3] GetByIDAdmin: %v", err)
		return c.JSON(http.StatusInternalServerError, response.ResponseError(err.Error()))
	}

	order, err := o.orderService.GetByID(ctx, orderID, user)
	if err != nil {
		log.Errorf("[OrderHandler-4] GetByIDAdmin: %v", err)
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

// GetAllCustomer implements [IOrderHandler].
func (o *orderHandler) GetAllCustomer(c echo.Context) error {
	var (
		ctx         = c.Request().Context()
		respOrders  = []response.OrderCustomerList{} // Ini akan diisi dengan data yang diformat
		jwtUserData = entity.JwtUserData{}
	)

	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[OrderHandler-1] GetAllCustomer: %s", "data token not found")
		return c.JSON(http.StatusUnauthorized, response.ResponseError("data token not found"))
	}

	err := json.Unmarshal([]byte(user), &jwtUserData)
	if err != nil {
		log.Errorf("[OrderHandler-2] GetAllCustomer: %v", err)
		return c.JSON(http.StatusBadRequest, response.ResponseError(err.Error()))
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

	reqEntity := entity.QueryStringEntity{
		Search:  search,
		Status:  status,
		Page:    page,
		Limit:   perPage,
		BuyerID: userID,
	}

	results, totalData, totalPage, err := o.orderService.GetAllCustomer(ctx, reqEntity, user)
	if err != nil {
		log.Errorf("[OrderHandler-3] GetAllCustomer: %v", err) // Mengubah log level error
		if err.Error() == "404" {
			return c.JSON(http.StatusNotFound, response.ResponseError("data not found"))
		}
		return c.JSON(http.StatusInternalServerError, response.ResponseError(err.Error()))
	}

	// Mempopulasi respOrders dari results
	for _, result := range results {
		respOrders = append(respOrders, response.OrderCustomerList{
			ID:            result.ID,
			OrderCode:     result.OrderCode,
			OrderDateTime: result.OrderDate,
			Status:        result.Status,
			TotalAmount:   result.TotalAmount,
		})
	}

	return c.JSON(http.StatusOK, response.ResponseSuccessWithPagination("success", respOrders, page, totalData, totalPage, perPage))
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
	authGroup.GET("/orders", ordHandler.GetAllCustomer)
	authGroup.GET("/orders/:orderCode/code", ordHandler.GetOrderByOrderCode)

	adminGroup := e.Group("/admin", mid.CheckToken())
	adminGroup.GET("/orders", ordHandler.GetAllAdmin)
	adminGroup.GET("/orders/:orderID", ordHandler.GetByIDAdmin)
	adminGroup.PUT("/orders/:orderID/status", ordHandler.UpdateStatus)

	return ordHandler
}
