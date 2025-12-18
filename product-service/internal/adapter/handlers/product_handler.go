package handlers

import (
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

// interface
type IProductHandler interface {
	GetAllAdmin(c echo.Context) error
	CreateAdmin(c echo.Context) error
	GetByIDAdmin(c echo.Context) error
	EditAdmin(c echo.Context) error
	DeleteAdmin(c echo.Context) error
}

// struct
type productHandler struct {
	productService service.IProductService
}

// DeleteAdmin implements [IProductHandler].
func (p *productHandler) DeleteAdmin(c echo.Context) error {
	var (
		resp = response.DefaultResponse{}
		ctx  = c.Request().Context()
	)

	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[ProductHandler-1] DeleteAdmin: %s", "data token not found")
		resp.Message = "data token not found"
		resp.Data = nil
		return c.JSON(http.StatusNotFound, resp)
	}

	idStr := c.Param("id")
	if idStr == "" {
		log.Errorf("[ProductHandler-2] DeleteAdmin: %v", "Invalid id")
		resp.Message = "ID is required"
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	id, err := conv.StringToInt64(idStr)
	if err != nil {
		log.Errorf("[ProductHandler-3] DeleteAdmin: %v", err.Error())
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	err = p.productService.Delete(ctx, id)
	if err != nil {
		log.Errorf("[ProductHandler-4] DeleteAdmin: %v", err)
		if err.Error() == "404" {
			resp.Message = "Data not found"
			resp.Data = nil
			return c.JSON(http.StatusNotFound, resp)
		}
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	resp.Message = "success"
	resp.Data = nil
	return c.JSON(http.StatusOK, resp)
}

// EditAdmin implements [IProductHandler].
func (p *productHandler) EditAdmin(c echo.Context) error {
	var (
		resp = response.DefaultResponse{}
		ctx  = c.Request().Context()
		req  = request.ProductRequest{}
	)

	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[ProductHandler-1] EditAdmin: %s", "data token not found")
		resp.Message = "data token not found"
		resp.Data = nil
		return c.JSON(http.StatusNotFound, resp)
	}

	idStr := c.Param("id")
	if idStr == "" {
		log.Errorf("[ProductHandler-2] EditAdmin: %v", "Invalid id")
		resp.Message = "ID is required"
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	id, err := conv.StringToInt64(idStr)
	if err != nil {
		log.Errorf("[ProductHandler-3] EditAdmin: %v", err.Error())
		if err.Error() == "404" {
			resp.Message = "Product not found"
			resp.Data = nil
			return c.JSON(http.StatusNotFound, resp)
		}
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	if err := c.Bind(&req); err != nil {
		log.Errorf("[ProductHandler-4] EditAdmin: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	if err := c.Validate(req); err != nil {
		log.Errorf("[ProductHandler-3] CreateAdmin: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	reqEntity := entities.ProductEntity{
		ID:           id,
		CategorySlug: req.CategorySlug,
		ParentID:     nil,
		Name:         req.ProductName,
		Image:        req.VariantDetail[0].ProductImage,
		Description:  req.ProductDescription,
		RegulerPrice: float64(req.VariantDetail[0].RegulerPrice),
		SalePrice:    float64(req.VariantDetail[0].SalePrice),
		Unit:         req.Unit,
		Weight:       req.VariantDetail[0].Weight,
		Stock:        req.VariantDetail[0].Stock,
		Variant:      req.Variant,
		Status:       req.Status,
	}

	productChilds := []entities.ProductEntity{}
	if len(req.VariantDetail) > 1 {
		for i := 1; i < len(req.VariantDetail); i++ {
			productChilds = append(productChilds, entities.ProductEntity{
				Image:        req.VariantDetail[i].ProductImage,
				RegulerPrice: float64(req.VariantDetail[i].RegulerPrice),
				SalePrice:    float64(req.VariantDetail[i].SalePrice),
				Weight:       req.VariantDetail[i].Weight,
				Stock:        req.VariantDetail[i].Stock,
			})
		}

		reqEntity.Child = productChilds
	}

	err = p.productService.Update(ctx, reqEntity)
	if err != nil {
		log.Errorf("[ProductHandler-4] EditAdmin: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	resp.Message = "success"
	resp.Data = nil
	return c.JSON(http.StatusOK, resp)
}

// GetByIDAdmin implements [IProductHandler].
func (p *productHandler) GetByIDAdmin(c echo.Context) error {
	var (
		resp        = response.DefaultResponse{}
		ctx         = c.Request().Context()
		respProduct = response.ProductDetailResponse{}
	)

	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[ProductHandler-1] GetByIDAdmin: %s", "data token not found")
		resp.Message = "data token not found"
		resp.Data = nil
		return c.JSON(http.StatusNotFound, resp)
	}

	idStr := c.Param("id")
	if idStr == "" {
		log.Errorf("[ProductHandler-2] GetByIDAdmin: %v", "Invalid id")
		resp.Message = "ID is required"
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	id, err := conv.StringToInt64(idStr)
	if err != nil {
		log.Errorf("[ProductHandler-3] GetByIDAdmin: %v", err.Error())
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	result, err := p.productService.GetByID(ctx, id)
	if err != nil {
		log.Errorf("[ProductHandler-4] GetByIDAdmin: %v", err)
		if err.Error() == "404" {
			resp.Message = "Data not found"
			resp.Data = nil
			return c.JSON(http.StatusNotFound, resp)
		}
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	responseChilds := []response.ProductChildResponse{}
	if len(result.Child) > 0 {
		for _, child := range result.Child {
			responseChilds = append(responseChilds, response.ProductChildResponse{
				ID:           child.ID,
				SalePrice:    int64(child.SalePrice),
				RegulerPrice: int64(child.RegulerPrice),
				Weight:       child.Weight,
				Stock:        child.Stock,
			})
		}
	}

	respProduct = response.ProductDetailResponse{
		ID:                 result.ID,
		ProductName:        result.Name,
		ParentID:           conv.Int64PointerToInt64(result.ParentID),
		ProductImage:       result.Image,
		CategorySlug:       result.CategorySlug,
		CategoryName:       result.CategoryName,
		ProductStatus:      result.Status,
		ProductDescription: result.Description,
		SalePrice:          int64(result.SalePrice),
		RegulerPrice:       int64(result.RegulerPrice),
		Unit:               result.Unit,
		Weight:             result.Weight,
		Stock:              result.Stock,
		CreatedAt:          result.CreatedAt,
		Child:              responseChilds,
	}

	resp.Message = "success"
	resp.Data = respProduct
	return c.JSON(http.StatusOK, resp)
}

// CreateAdmin implements [IProductHandler].
func (p *productHandler) CreateAdmin(c echo.Context) error {
	var (
		resp = response.DefaultResponse{}
		ctx  = c.Request().Context()
		req  = request.ProductRequest{}
	)

	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[ProductHandler-1] CreateAdmin: %s", "data token not found")
		resp.Message = "data token not found"
		resp.Data = nil
		return c.JSON(http.StatusNotFound, resp)
	}

	if err := c.Bind(&req); err != nil {
		log.Errorf("[ProductHandler-2] CreateAdmin: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	if err := c.Validate(req); err != nil {
		log.Errorf("[ProductHandler-3] CreateAdmin: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	reqEntity := entities.ProductEntity{
		CategorySlug: req.CategorySlug,
		ParentID:     nil,
		Name:         req.ProductName,
		Image:        req.VariantDetail[0].ProductImage,
		Description:  req.ProductDescription,
		RegulerPrice: float64(req.VariantDetail[0].RegulerPrice),
		SalePrice:    float64(req.VariantDetail[0].SalePrice),
		Unit:         req.Unit,
		Weight:       req.VariantDetail[0].Weight,
		Stock:        req.VariantDetail[0].Stock,
		Variant:      req.Variant,
		Status:       req.Status,
	}

	productChilds := []entities.ProductEntity{}
	if len(req.VariantDetail) > 1 {
		for i := 1; i < len(req.VariantDetail); i++ {
			productChilds = append(productChilds, entities.ProductEntity{
				Image:        req.VariantDetail[i].ProductImage,
				RegulerPrice: float64(req.VariantDetail[i].RegulerPrice),
				SalePrice:    float64(req.VariantDetail[i].SalePrice),
				Weight:       req.VariantDetail[i].Weight,
				Stock:        req.VariantDetail[i].Stock,
			})
		}

		reqEntity.Child = productChilds
	}

	err := p.productService.Create(ctx, reqEntity)
	if err != nil {
		log.Errorf("[ProductHandler-4] CreateAdmin: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	resp.Message = "success"
	resp.Data = nil
	return c.JSON(http.StatusCreated, resp)
}

// GetAllAdmin implements [IProductHandler].
func (p *productHandler) GetAllAdmin(c echo.Context) error {
	var (
		resp         = response.DefaultResponseWithPaginations{}
		ctx          = c.Request().Context()
		respProducts = []response.ProductListResponse{}
	)

	search := c.QueryParam("search")
	orderBy := "created_at"
	if c.QueryParam("orderBy") != "" {
		orderBy = c.QueryParam("orderBy")
	}
	orderType := "desc"
	if c.QueryParam("orderType") != "" {
		orderType = c.QueryParam("orderType")
	}

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

	categorySlug := c.QueryParam("categorySlug")
	startPrice, err := conv.StringToInt64(c.QueryParam("startPrice"))
	if err != nil {
		startPrice = 0
	}

	endPrice, err := conv.StringToInt64(c.QueryParam("endPrice"))
	if err != nil {
		endPrice = 0
	}

	var status = ""
	if c.QueryParam("status") != "" {
		status = c.QueryParam("status")
	}

	reqEntity := entities.QueryStringProduct{
		Search:       search,
		OrderBy:      orderBy,
		OrderType:    orderType,
		Page:         int(page),
		Limit:        int(perPage),
		CategorySlug: categorySlug,
		StartPrice:   startPrice,
		EndPrice:     endPrice,
		Status:       status,
	}

	results, totalData, totalPage, err := p.productService.GetAll(ctx, reqEntity)

	if err != nil {
		log.Errorf("[ProductHandler-1] GetAll: %v", err)
		if err.Error() == "404" {
			resp.Message = "Data not found"
			resp.Data = nil
			return c.JSON(http.StatusNotFound, resp)
		}
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	for _, product := range results {
		respProducts = append(respProducts, response.ProductListResponse{
			ID:            product.ID,
			ProductName:   product.Name,
			ParentID:      conv.Int64PointerToInt64(product.ParentID),
			ProductImage:  product.Image,
			CategoryName:  product.CategoryName,
			ProductStatus: product.Status,
			SalePrice:     int64(product.SalePrice),
			CreatedAt:     product.CreatedAt,
		})
	}

	resp.Data = respProducts
	resp.Message = "success"
	resp.Pagination = &response.Pagination{
		Page:       page,
		TotalCount: totalData,
		TotalPage:  totalPage,
		PerPage:    perPage,
	}

	return c.JSON(http.StatusOK, resp)
}

// function
func NewProductHandler(e *echo.Echo, cfg *config.Config, productService service.IProductService) IProductHandler {
	productHandler := &productHandler{
		productService: productService,
	}

	e.Use(middleware.Recover())
	mid := adapter.NewMiddlewareAdapter(cfg)
	adminGroup := e.Group("/admin", mid.CheckToken())
	adminGroup.GET("/products", productHandler.GetAllAdmin)
	adminGroup.POST("/products", productHandler.CreateAdmin)
	adminGroup.GET("/products/:id", productHandler.GetByIDAdmin)
	adminGroup.PUT("/products/:id", productHandler.EditAdmin)
	adminGroup.DELETE("/products/:id", productHandler.DeleteAdmin)

	return productHandler
}
