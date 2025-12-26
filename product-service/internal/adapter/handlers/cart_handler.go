package handlers

import (
	"encoding/json"
	"net/http"
	"product-service/config"
	"product-service/internal/adapter"
	"product-service/internal/adapter/handlers/request"
	"product-service/internal/adapter/handlers/response"
	"product-service/internal/core/domain/entities"
	"product-service/internal/core/service"
	"product-service/utils/conv"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

type ICartHandler interface {
	AddToCart(c echo.Context) error
	GetCart(c echo.Context) error
	RemoveFromCart(c echo.Context) error
	RemoveAllCart(c echo.Context) error
}

type CartHandler struct {
	cartService    service.ICartService
	ProductService service.IProductService
}

// RemoveAllCart implements [ICartHandler].
func (ch *CartHandler) RemoveAllCart(c echo.Context) error {
	var (
		resp        = response.DefaultResponse{}
		ctx         = c.Request().Context()
		jwtUserData = entities.JwtUserData{}
	)

	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[CartHandler-1] RemoveFromCart: %s", "data token not found")
		resp.Message = "data token not found"
		resp.Data = nil
		return c.JSON(http.StatusNotFound, resp)
	}

	err := json.Unmarshal([]byte(user), &jwtUserData)
	if err != nil {
		log.Errorf("[CartHandler-2] RemoveFromCart: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	userID := jwtUserData.UserID
	err = ch.cartService.RemoveAllCart(ctx, userID)
	if err != nil {
		log.Errorf("[CartHandler-1] RemoveAllCart: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	resp.Message = "success"
	resp.Data = nil
	return c.JSON(http.StatusOK, resp)
}

// RemoveFromCart implements [ICartHandler].
func (ch *CartHandler) RemoveFromCart(c echo.Context) error {
	var (
		resp        = response.DefaultResponse{}
		ctx         = c.Request().Context()
		jwtUserData = entities.JwtUserData{}
	)

	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[CartHandler-1] RemoveFromCart: %s", "data token not found")
		resp.Message = "data token not found"
		resp.Data = nil
		return c.JSON(http.StatusNotFound, resp)
	}

	err := json.Unmarshal([]byte(user), &jwtUserData)
	if err != nil {
		log.Errorf("[CartHandler-2] RemoveFromCart: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	userID := jwtUserData.UserID
	productID := c.QueryParam("product_id")
	if productID == "" {
		log.Errorf("[CartHandler-3] RemoveFromCart: %s", "product_id is required")
		resp.Message = "product_id is required"
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	prodID, err := conv.StringToInt64(productID)
	err = ch.cartService.RemoveFromCart(ctx, userID, prodID)
	if err != nil {
		log.Errorf("[CartHandler-4] RemoveFromCart: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	resp.Message = "success"
	resp.Data = nil
	return c.JSON(http.StatusOK, resp)
}

// GetCart implements [ICartHandler].
func (ch *CartHandler) GetCart(c echo.Context) error {
	var (
		resp        = response.DefaultResponse{}
		ctx         = c.Request().Context()
		respList    = []response.CartResponse{}
		jwtUserData = entities.JwtUserData{}
	)

	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[CartHandler-1] GetCart: %s", "data token not found")
		resp.Message = "data token not found"
		resp.Data = nil
		return c.JSON(http.StatusNotFound, resp)
	}

	err := json.Unmarshal([]byte(user), &jwtUserData)
	if err != nil {
		log.Errorf("[CartHandler-2] GetCart: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	userID := jwtUserData.UserID
	items, err := ch.cartService.GetCartByUserID(ctx, userID)
	if err != nil {
		log.Errorf("[CartHandler-3] GetCart: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	for _, item := range items {
		product, err := ch.ProductService.GetByID(ctx, item.ProductID)
		if err != nil {
			log.Errorf("[CartHandler-4] GetCart: %v", err)
			resp.Message = err.Error()
			resp.Data = nil
			return c.JSON(http.StatusInternalServerError, resp)
		}

		respList = append(respList, response.CartResponse{
			ID:            item.ProductID,
			ProductName:   product.Name,
			ProductImage:  product.Image,
			ProductStatus: product.Status,
			SalePrice:     int64(product.SalePrice),
			Quantity:      item.Quantity,
			Unit:          product.Unit,
			Weight:        int64(product.Weight),
		})
	}

	resp.Message = "success"
	resp.Data = respList

	return c.JSON(http.StatusOK, resp)
}

// AddToCart implements [ICartHandler].
func (ch *CartHandler) AddToCart(c echo.Context) error {
	var (
		resp        = response.DefaultResponse{}
		ctx         = c.Request().Context()
		request     = request.CartRequest{}
		jwtUserData = entities.JwtUserData{}
	)

	if err := c.Bind(&request); err != nil {
		log.Errorf("[CartHandler-1] AddToCart: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	if err := c.Validate(request); err != nil {
		log.Errorf("[CartHandler-2] AddToCart: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[CartHandler-3] AddToCart: %s", "data token not found")
		resp.Message = "data token not found"
		resp.Data = nil
		return c.JSON(http.StatusNotFound, resp)
	}

	err := json.Unmarshal([]byte(user), &jwtUserData)
	if err != nil {
		log.Errorf("[CartHandler-4] AddToCart: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	userID := jwtUserData.UserID
	reqEntity := entities.CartItem{
		ProductID: request.ProductID,
		Quantity:  request.Quantity,
	}

	err = ch.cartService.AddToCart(ctx, userID, reqEntity)
	if err != nil {
		log.Errorf("[CartHandler-5] AddToCart: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	resp.Message = "success"
	resp.Data = nil
	return c.JSON(http.StatusCreated, resp)
}

func NewCartHandler(e *echo.Echo, cfg *config.Config, cartService service.ICartService, productService service.IProductService) ICartHandler {
	cartHandler := &CartHandler{
		cartService:    cartService,
		ProductService: productService,
	}

	e.Use(middleware.Recover())
	mid := adapter.NewMiddlewareAdapter(cfg)
	authGroup := e.Group("/auth", mid.CheckToken())
	authGroup.POST("/cart", cartHandler.AddToCart)
	authGroup.GET("/cart", cartHandler.GetCart)
	authGroup.DELETE("/cart", cartHandler.RemoveFromCart)
	authGroup.DELETE("/cart/all", cartHandler.RemoveAllCart)

	return cartHandler
}
