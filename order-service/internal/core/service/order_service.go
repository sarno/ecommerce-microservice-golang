package service

import (
	"context"
	"encoding/json"
	"fmt"
	"order-service/config"
	"order-service/internal/adapter/client"
	"order-service/internal/adapter/message"
	"order-service/internal/adapter/repository"
	"order-service/internal/core/domain/entity"
	"order-service/utils"
	"order-service/utils/conv"
	"sync"

	"github.com/labstack/gommon/log"
)

type IOrderService interface {
	GetAll(ctx context.Context, queryString entity.QueryStringEntity, accessToken string) ([]entity.OrderEntity, int64, int64, error)
	GetByID(ctx context.Context, orderID int64, accessToken string) (*entity.OrderEntity, error)
	CreateOrder(ctx context.Context, req entity.OrderEntity, accessToken string) (int64, error)
	GetDetailCustomer(ctx context.Context, orderID int64, accessToken string) (*entity.OrderEntity, error)
	GetAllCustomer(ctx context.Context, queryString entity.QueryStringEntity, accessToken string) ([]entity.OrderEntity, int64, int64, error)

	GetOrderByOrderCode(ctx context.Context, orderCode, accessToken string) (*entity.OrderEntity, error)
	UpdateStatus(ctx context.Context, req entity.OrderEntity, accessToken string) error
}

type orderService struct {
	repo              repository.IOrderRepository
	cfg               *config.Config
	publisherRabbitMQ message.IPublisherRabbitMQ
	elasticRepo       repository.IElasticRepository
	userClient        client.IUserClient
	productClient     client.IProductClient
}

// UpdateStatus implements [IOrderService].
func (o *orderService) UpdateStatus(ctx context.Context, req entity.OrderEntity, accessToken string) error {
	buyerID, statusOrder, orderCode, err := o.repo.UpdateStatus(ctx, req)
	if err != nil {
		log.Errorf("[OrderService-1] UpdateStatus: %v", err)
		return err
	}

	var token map[string]interface{}
	err = json.Unmarshal([]byte(accessToken), &token)
	if err != nil {
		log.Errorf("[OrderService-2] UpdateStatus: %v", err)
		return err
	}

	userResponse, err := o.userClient.GetUser(buyerID, token["token"].(string), false)
	if err != nil {
		log.Errorf("[OrderService-3] UpdateStatus: %v", err)
		return err
	}
	message := fmt.Sprintf("Hello,\n\nYour order with ID %s has been updated to status: %s.\n\nThank you for shopping with us!", orderCode, statusOrder)
	go o.publisherRabbitMQ.PublishSendEmailUpdateStatus(userResponse.Email, message, o.cfg.PublisherName.EmailUpdateStatus, buyerID)
	go o.publisherRabbitMQ.PublishSendPushNotifUpdateStatus(message, utils.PUSH_NOTIF, buyerID)
	go o.publisherRabbitMQ.PublishUpdateStatus(o.cfg.PublisherName.PublisherUpdateStatus, req.ID, req.Status)

	return nil
}

// GetOrderByOrderCode implements [IOrderService].
func (o *orderService) GetOrderByOrderCode(ctx context.Context, orderCode string, accessToken string) (*entity.OrderEntity, error) {
	result, err := o.repo.GetOrderByOrderCode(ctx, orderCode)
	if err != nil {
		log.Errorf("[OrderService-1] GetOrderByOrderCode: %v", err)
		return nil, err
	}

	var token map[string]interface{}
	err = json.Unmarshal([]byte(accessToken), &token)
	if err != nil {
		log.Errorf("[OrderService-2] GetOrderByOrderCode: %v", err)
		return nil, err
	}

	isCustomer := false
	if token["role_name"].(string) != "admin" {
		isCustomer = true
	}

	userResponse, err := o.userClient.GetUser(result.BuyerId, token["token"].(string), isCustomer)
	if err != nil {
		log.Errorf("[OrderService-3] GetOrderByOrderCode: %v", err)
		return nil, err
	}

	result.BuyerName = userResponse.Name
	result.BuyerEmail = userResponse.Email
	result.BuyerPhone = userResponse.Phone
	result.BuyerAddress = userResponse.Address

	if len(result.OrderItems) == 0 {
		return result, nil
	}

	productIDList := make([]int64, 0, len(result.OrderItems))
	for _, item := range result.OrderItems {
		productIDList = append(productIDList, item.ProductID)
	}

	productsMap, err := o.productClient.GetProductsBulk(productIDList, token["token"].(string), isCustomer)
	if err != nil {
		log.Errorf("[OrderService-4] GetByID (Bulk Product): %v", err)
		return nil, err
	}

	for i := range result.OrderItems {
		if product, ok := productsMap[result.OrderItems[i].ProductID]; ok {
			result.OrderItems[i].ProductImage = product.ProductImage
			result.OrderItems[i].ProductName = product.ProductName
			result.OrderItems[i].Price = int64(product.SalePrice)
		}
	}

	return result, nil
}

// GetAllCustomer implements [IOrderService].
func (o *orderService) GetAllCustomer(ctx context.Context, queryString entity.QueryStringEntity, accessToken string) ([]entity.OrderEntity, int64, int64, error) {
	results, count, total, err := o.elasticRepo.SearchOrderElasticByBuyerId(ctx, queryString, queryString.BuyerID)
	if err != nil {
		log.Errorf("[OrderService-1] GetAllCustomer (Elasticsearch): %v, falling back to DB", err)
		// Fallback to database if elastic fails
		results, count, total, err = o.repo.GetAll(ctx, queryString)
		if err != nil {
			log.Errorf("[OrderService-2] GetAllCustomer (DB): %v", err)
			return nil, 0, 0, err
		}
	}

	if len(results) == 0 {
		return results, count, total, nil
	}

	var token map[string]interface{}
	err = json.Unmarshal([]byte(accessToken), &token)
	if err != nil {
		log.Errorf("[OrderService-3] GetAllCustomer: %v", err)
		return nil, 0, 0, err
	}

	// Step 1: Collect all unique IDs
	buyerIDs := make(map[int64]struct{})
	productIDs := make(map[int64]struct{})
	for _, order := range results {
		buyerIDs[order.BuyerId] = struct{}{}
		for _, item := range order.OrderItems {
			productIDs[item.ProductID] = struct{}{}
		}
	}

	// Convert map keys to slices
	buyerIDList := make([]int64, 0, len(buyerIDs))
	for id := range buyerIDs {
		buyerIDList = append(buyerIDList, id)
	}
	productIDList := make([]int64, 0, len(productIDs))
	for id := range productIDs {
		productIDList = append(productIDList, id)
	}

	// Step 2: Fetch data in bulk concurrently
	var usersMap map[int64]entity.CustomerResponseEntity
	var productsMap map[int64]entity.ProductResponseEntity
	var wg sync.WaitGroup
	var userErr, productErr error

	wg.Add(2)

	go func() {
		defer wg.Done()
		usersMap, userErr = o.userClient.GetUsersBulk(buyerIDList, token["token"].(string))
	}()

	go func() {
		defer wg.Done()
		productsMap, productErr = o.productClient.GetProductsBulk(productIDList, token["token"].(string), true)
	}()

	wg.Wait()

	if userErr != nil {
		log.Errorf("[OrderService-4] GetAllCustomer (Bulk User): %v", userErr)
		return nil, 0, 0, userErr
	}
	if productErr != nil {
		log.Errorf("[OrderService-5] GetAllCustomer (Bulk Product): %v", productErr)
		return nil, 0, 0, productErr
	}

	// Step 3: Populate results from in-memory maps
	for i := range results {
		if user, ok := usersMap[results[i].BuyerId]; ok {
			results[i].BuyerName = user.Name
			results[i].BuyerEmail = user.Email
			results[i].BuyerPhone = user.Phone
			results[i].BuyerAddress = user.Address
		}

		for j := range results[i].OrderItems {
			if product, ok := productsMap[results[i].OrderItems[j].ProductID]; ok {
				results[i].OrderItems[j].ProductImage = product.ProductImage
				results[i].OrderItems[j].ProductName = product.ProductName
				results[i].OrderItems[j].Price = int64(product.SalePrice)
				results[i].OrderItems[j].ProductUnit = product.Unit
				results[i].OrderItems[j].ProductWeight = int64(product.Weight)
			}
		}
	}

	return results, count, total, nil
}

// GetDetailCustomer implements [IOrderService].
func (o *orderService) GetDetailCustomer(ctx context.Context, orderID int64, accessToken string) (*entity.OrderEntity, error) {
	result, err := o.repo.GetByID(ctx, orderID)
	if err != nil {
		log.Errorf("[OrderService-1] GetDetailCustomer: %v", err)
		return nil, err
	}

	var token map[string]interface{}
	err = json.Unmarshal([]byte(accessToken), &token)
	if err != nil {
		log.Errorf("[OrderService-2] GetDetailCustomer: %v", err)
		return nil, err
	}

	userResponse, err := o.userClient.GetUser(result.BuyerId, token["token"].(string), true)
	if err != nil {
		log.Errorf("[OrderService-3] GetDetailCustomer (User): %v", err)
		return nil, err
	}

	result.BuyerName = userResponse.Name
	result.BuyerEmail = userResponse.Email
	result.BuyerPhone = userResponse.Phone
	result.BuyerAddress = userResponse.Address

	if len(result.OrderItems) == 0 {
		return result, nil
	}

	productIDList := make([]int64, 0, len(result.OrderItems))
	for _, item := range result.OrderItems {
		productIDList = append(productIDList, item.ProductID)
	}

	productsMap, err := o.productClient.GetProductsBulk(productIDList, token["token"].(string), true)
	if err != nil {
		log.Errorf("[OrderService-4] GetDetailCustomer (Bulk Product): %v", err)
		return nil, err
	}

	for i := range result.OrderItems {
		if product, ok := productsMap[result.OrderItems[i].ProductID]; ok {
			result.OrderItems[i].ProductImage = product.ProductImage
			if len(product.Child) > 0 {
				result.OrderItems[i].ProductImage = product.Child[0].Image
			}
			result.OrderItems[i].ProductName = product.ProductName
			result.OrderItems[i].Price = int64(product.SalePrice)
			result.OrderItems[i].ProductWeight = int64(product.Weight)
			result.OrderItems[i].ProductUnit = product.Unit
		}
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

	if err != nil {
		log.Errorf("[OrderService-1] GetAll (Elasticsearch): %v, falling back to DB", err)
		results, count, total, err = o.repo.GetAll(ctx, queryString)
		if err != nil {
			log.Errorf("[OrderService-2] GetAll (DB): %v", err)
			return nil, 0, 0, err
		}
	}

	if len(results) == 0 {
		return results, count, total, nil
	}

	var token map[string]interface{}
	err = json.Unmarshal([]byte(accessToken), &token)
	if err != nil {
		log.Errorf("[OrderService-3] GetAll: %v", err)
		return nil, 0, 0, err
	}

	isCustomer := false
	if val, ok := token["role_name"]; ok && val.(string) != "Super Admin" {
		isCustomer = true
	}

	buyerIDs := make(map[int64]struct{})
	productIDs := make(map[int64]struct{})
	for _, order := range results {
		buyerIDs[order.BuyerId] = struct{}{}
		for _, item := range order.OrderItems {
			productIDs[item.ProductID] = struct{}{}
		}
	}

	buyerIDList := make([]int64, 0, len(buyerIDs))
	for id := range buyerIDs {
		buyerIDList = append(buyerIDList, id)
	}
	productIDList := make([]int64, 0, len(productIDs))
	for id := range productIDs {
		productIDList = append(productIDList, id)
	}

	var usersMap map[int64]entity.CustomerResponseEntity
	var productsMap map[int64]entity.ProductResponseEntity
	var wg sync.WaitGroup
	var userErr, productErr error

	wg.Add(2)
	go func() {
		defer wg.Done()
		usersMap, userErr = o.userClient.GetUsersBulk(buyerIDList, token["token"].(string))
	}()
	go func() {
		defer wg.Done()
		productsMap, productErr = o.productClient.GetProductsBulk(productIDList, token["token"].(string), isCustomer)
	}()
	wg.Wait()

	if userErr != nil {
		log.Errorf("[OrderService-4] GetAll (Bulk User): %v", userErr)
		return nil, 0, 0, userErr
	}
	if productErr != nil {
		log.Errorf("[OrderService-5] GetAll (Bulk Product): %v", productErr)
		return nil, 0, 0, productErr
	}

	for i := range results {
		if user, ok := usersMap[results[i].BuyerId]; ok {
			results[i].BuyerName = user.Name
		}
		for j := range results[i].OrderItems {
			if product, ok := productsMap[results[i].OrderItems[j].ProductID]; ok {
				results[i].OrderItems[j].ProductImage = product.ProductImage
			}
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
	if role, ok := token["role_name"]; ok && role.(string) != "admin" {
		isCustomer = true
	}

	userResponse, err := o.userClient.GetUser(result.BuyerId, token["token"].(string), isCustomer)
	if err != nil {
		log.Errorf("[OrderService-3] GetByID (User): %v", err)
		return nil, err
	}

	result.BuyerName = userResponse.Name
	result.BuyerEmail = userResponse.Email
	result.BuyerPhone = userResponse.Phone
	result.BuyerAddress = userResponse.Address

	if len(result.OrderItems) == 0 {
		return result, nil
	}

	productIDList := make([]int64, 0, len(result.OrderItems))
	for _, item := range result.OrderItems {
		productIDList = append(productIDList, item.ProductID)
	}

	productsMap, err := o.productClient.GetProductsBulk(productIDList, token["token"].(string), isCustomer)
	if err != nil {
		log.Errorf("[OrderService-4] GetByID (Bulk Product): %v", err)
		return nil, err
	}

	for i := range result.OrderItems {
		if product, ok := productsMap[result.OrderItems[i].ProductID]; ok {
			result.OrderItems[i].ProductImage = product.ProductImage
			result.OrderItems[i].ProductName = product.ProductName
			result.OrderItems[i].Price = int64(product.SalePrice)
		}
	}

	return result, nil
}

func NewOrderService(
	orderRepo repository.IOrderRepository,
	cfg *config.Config,
	publisher message.IPublisherRabbitMQ,
	elasticRepo repository.IElasticRepository,
	userClient client.IUserClient,
	productClient client.IProductClient,
) IOrderService {
	return &orderService{
		repo:              orderRepo,
		cfg:               cfg,
		publisherRabbitMQ: publisher,
		elasticRepo:       elasticRepo,
		userClient:        userClient,
		productClient:     productClient,
	}
}
