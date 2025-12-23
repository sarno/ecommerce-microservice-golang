package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"order-service/config"
	httpclient "order-service/internal/adapter/http_client"
	"order-service/internal/adapter/message"
	"order-service/internal/adapter/repository"
	"order-service/internal/core/domain/entity"
	"order-service/utils/conv"
	"strconv"

	"github.com/labstack/gommon/log"
)

type IOrderService interface {
	GetAll(ctx context.Context, queryString entity.QueryStringEntity, accessToken string) ([]entity.OrderEntity, int64, int64, error)
	GetByID(ctx context.Context, orderID int64, accessToken string) (*entity.OrderEntity, error)
	CreateOrder(ctx context.Context, req entity.OrderEntity, accessToken string) (int64, error)
	GetDetailCustomer(ctx context.Context, orderID int64, accessToken string) (*entity.OrderEntity, error)
}

type orderService struct {
	repo              repository.IOrderRepository
	cfg               *config.Config
	httpClient        httpclient.IHttpClient
	publisherRabbitMQ message.IPublisherRabbitMQ
	elasticRepo       repository.IElasticRepository
}

// GetDetailCustomer implements [IOrderService].
func (o *orderService) GetDetailCustomer(ctx context.Context, orderID int64, accessToken string) (*entity.OrderEntity, error) {
	result, err := o.repo.GetByID(ctx, orderID)
	if err != nil {
		log.Errorf("[OrderService-1] GetByID: %v", err)
		return nil, err
	}

	var token map[string]interface{}
	err = json.Unmarshal([]byte(accessToken), &token)
	if err != nil {
		log.Errorf("[OrderService-2] GetByID: %v", err)
		return nil, err
	}

	userResponse, err := o.httpClientUserService(result.BuyerId, token["token"].(string), true)
	if err != nil {
		log.Errorf("[OrderService-3] GetByID: %v", err)
		return nil, err
	}

	result.BuyerName = userResponse.Name
	result.BuyerEmail = userResponse.Email
	result.BuyerPhone = userResponse.Phone
	result.BuyerAddress = userResponse.Address

	for key, val := range result.OrderItems {
		productResponse, err := o.httpClientProductService(val.ProductID, token["token"].(string), true)
		if err != nil {
			log.Errorf("[OrderService-3] GetByID: %v", err)
			return nil, err
		}

		result.OrderItems[key].ProductImage = productResponse.ProductImage
		if productResponse.Child != nil {
			result.OrderItems[key].ProductImage = productResponse.Child[0].Image
		}
		result.OrderItems[key].ProductName = productResponse.ProductName
		result.OrderItems[key].Price = int64(productResponse.SalePrice)
		result.OrderItems[key].ProductWeight = int64(productResponse.Weight)
		result.OrderItems[key].ProductUnit = productResponse.Unit
	}

	return result, nil
}

// CreateOrder implements [IOrderService].
func (o *orderService) CreateOrder(ctx context.Context, req entity.OrderEntity, accessToken string) (int64, error) {
	req.OrderCode = conv.GenerateOrderCode()
	shippingFee := 0
	if req.ShippingType == "Delivery" {
		shippingFee = 5000
	}
	req.ShippingFee = int64(shippingFee)
	req.Status = "Pending"

	orderID, err := o.repo.CreateOrder(ctx, req)
	if err != nil {
		log.Errorf("[OrderService-1] CreateOrder: %v", err)
		return 0, err
	}

	resultData, err := o.GetByID(ctx, orderID, accessToken)
	if err != nil {
		log.Errorf("[OrderService-2] CreateOrder: %v", err)
		return 0, err
	}

	if err := o.publisherRabbitMQ.PublishOrderToQueue(*resultData); err != nil {
		log.Errorf("[OrderService-3] CreateOrder: %v", err)
	}

	for _, orderItem := range req.OrderItems {
		o.publisherRabbitMQ.PublishUpdateStock(orderItem.ProductID, orderItem.Quantity)
	}

	return orderID, nil
}

// GetAll implements [IOrderService].
func (o *orderService) GetAll(ctx context.Context, queryString entity.QueryStringEntity, accessToken string) ([]entity.OrderEntity, int64, int64, error) {
	results, count, total, err := o.elasticRepo.SearchOrderElastic(ctx, queryString)
	if err == nil {
		return results, count, total, nil
	} else {
		log.Errorf("[OrderService-1] GetAll: %v", err)
	}

	results, count, total, err = o.repo.GetAll(ctx, queryString)
	if err != nil {
		log.Errorf("[OrderService-2] GetAll: %v", err)
		return nil, 0, 0, err
	}

	var token map[string]interface{}
	err = json.Unmarshal([]byte(accessToken), &token)
	if err != nil {
		log.Errorf("[OrderService-3] GetAll: %v", err)
		return nil, 0, 0, err
	}

	isCustomer := false
	if token["role_name"].(string) != "Super Admin" {
		isCustomer = true
	}

	for key, val := range results {
		userResponse, err := o.httpClientUserService(val.BuyerId, token["token"].(string), isCustomer)
		if err != nil {
			log.Errorf("[OrderService-4] GetAll: %v", err)
			return nil, 0, 0, err
		}
		results[key].BuyerName = userResponse.Name

		for key2, res := range val.OrderItems {

			productResponse, err := o.httpClientProductService(res.ProductID, token["token"].(string), isCustomer)
			if err != nil {
				log.Errorf("[OrderService-5] GetAll: %v", err)
				return nil, 0, 0, err
			}

			val.OrderItems[key2].ProductImage = productResponse.ProductImage
		}
	}

	return results, count, total, nil
}

func (o *orderService) GetByID(ctx context.Context, orderID int64, accessToken string) (*entity.OrderEntity, error) {
	result, err := o.repo.GetByID(ctx, orderID)
	if err != nil {
		log.Errorf("[OrderService-1] GetByID: %v", err)
		return nil, err
	}

	var token map[string]interface{}
	err = json.Unmarshal([]byte(accessToken), &token)
	if err != nil {
		log.Errorf("[OrderService-2] GetByID: %v", err)
		return nil, err
	}
	isCustomer := false
	if token["role_name"].(string) != "admin" {
		isCustomer = true
	}

	userResponse, err := o.httpClientUserService(result.BuyerId, token["token"].(string), isCustomer)
	if err != nil {
		log.Errorf("[OrderService-2] GetByID: %v", err)
		return nil, err
	}

	result.BuyerName = userResponse.Name
	result.BuyerEmail = userResponse.Email
	result.BuyerPhone = userResponse.Phone
	result.BuyerAddress = userResponse.Address

	for key, val := range result.OrderItems {
		productResponse, err := o.httpClientProductService(val.ProductID, token["token"].(string), isCustomer)
		if err != nil {
			log.Errorf("[OrderService-3] GetByID: %v", err)
			return nil, err
		}

		result.OrderItems[key].ProductImage = productResponse.ProductImage
		result.OrderItems[key].ProductName = productResponse.ProductName
		result.OrderItems[key].Price = int64(productResponse.SalePrice)
	}

	return result, nil
}

func NewOrderService(orderRepo repository.IOrderRepository, cfg *config.Config, httpClient httpclient.IHttpClient, publisher message.IPublisherRabbitMQ, elasticRepo repository.IElasticRepository) IOrderService {
	return &orderService{
		repo:              orderRepo,
		cfg:               cfg,
		httpClient:        httpClient,
		publisherRabbitMQ: publisher,
		elasticRepo:       elasticRepo,
	}
}

func (o *orderService) httpClientUserService(userID int64, accessToken string, isCustomer bool) (*entity.CustomerResponseEntity, error) {
	baseUrlUser := fmt.Sprintf("%s/%s", o.cfg.App.UserServiceUrl, "admin/customers/"+strconv.FormatInt(userID, 10))
	if isCustomer {
		baseUrlUser = fmt.Sprintf("%s/%s", o.cfg.App.UserServiceUrl, "auth/profile")
	}

	header := map[string]string{
		"Authorization": "Bearer " + accessToken,
		"Accept":        "application/json",
	}
	dataUser, err := o.httpClient.CallURL("GET", baseUrlUser, header, nil)
	if err != nil {
		log.Errorf("[OrderService-1] httpClientUserService: %v", err)
		return nil, err
	}

	defer dataUser.Body.Close()
	body, err := io.ReadAll(dataUser.Body)
	if err != nil {
		log.Errorf("[OrderService-2] httpClientUserService: %v", err)
		return nil, err
	}
	var userResponse entity.UserHttpClientResponse
	err = json.Unmarshal(body, &userResponse)
	if err != nil {
		log.Errorf("[OrderService-3] httpClientUserService: %v", err)
		return nil, err
	}

	return &userResponse.Data, nil
}

func (o *orderService) httpClientProductService(productID int64, accessToken string, isCustomer bool) (*entity.ProductResponseEntity, error) {
	baseUrlProduct := fmt.Sprintf("%s/%s", o.cfg.App.ProductServiceUrl, "admin/products/"+strconv.FormatInt(productID, 10))
	if isCustomer {
		baseUrlProduct = fmt.Sprintf("%s/%s", o.cfg.App.ProductServiceUrl, "products/home/"+strconv.FormatInt(productID, 10))
	}

	header := map[string]string{
		"Authorization": "Bearer " + accessToken,
		"Accept":        "application/json",
	}

	dataProduct, err := o.httpClient.CallURL("GET", baseUrlProduct, header, nil)
	if err != nil {
		log.Errorf("[OrderService-1] httpClientProductService: %v", err)
		return nil, err
	}
	
	defer dataProduct.Body.Close()

	body, err := io.ReadAll(dataProduct.Body)

	if err != nil {
		log.Errorf("[OrderService-2] httpClientProductService: %v", err)
		return nil, err
	}

	
	var productResponse entity.ProductHttpClientResponse
	err = json.Unmarshal(body, &productResponse)
	if err != nil {
		log.Errorf("[OrderService-3] httpClientProductService: %v", err)
		return nil, err
	}

	log.Infof("Web service Product Response: %+v", baseUrlProduct)

	return &productResponse.Data, nil
}
