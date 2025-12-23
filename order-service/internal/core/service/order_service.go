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
	"strings"
	"sync"

	"github.com/labstack/gommon/log"
)

type IOrderService interface {
	GetAll(ctx context.Context, queryString entity.QueryStringEntity, accessToken string) ([]entity.OrderEntity, int64, int64, error)
	GetByID(ctx context.Context, orderID int64, accessToken string) (*entity.OrderEntity, error)
	CreateOrder(ctx context.Context, req entity.OrderEntity, accessToken string) (int64, error)
	GetDetailCustomer(ctx context.Context, orderID int64, accessToken string) (*entity.OrderEntity, error)
	GetAllCustomer(ctx context.Context, queryString entity.QueryStringEntity, accessToken string) ([]entity.OrderEntity, int64, int64, error)
}

type orderService struct {
	repo              repository.IOrderRepository
	cfg               *config.Config
	httpClient        httpclient.IHttpClient
	publisherRabbitMQ message.IPublisherRabbitMQ
	elasticRepo       repository.IElasticRepository
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
		usersMap, userErr = o.httpClientBulkUserService(buyerIDList, token["token"].(string))
	}()

	go func() {
		defer wg.Done()
		// Assuming a bulk product endpoint exists
		productsMap, productErr = o.httpClientBulkProductService(productIDList, token["token"].(string), true)
	}()

	wg.Wait()

	if userErr != nil {
		log.Errorf("[OrderService-4] GetAllCustomer (Bulk User): %v", userErr)
		// Decide if you want to fail the whole request or proceed with partial data
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
				// Quantity is already in the order item
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

	// Efficiently fetch user and product data
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
		usersMap, userErr = o.httpClientBulkUserService(buyerIDList, token["token"].(string))
	}()
	go func() {
		defer wg.Done()
		productsMap, productErr = o.httpClientBulkProductService(productIDList, token["token"].(string), isCustomer)
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

	// Populate from maps
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
	baseUrlUser := fmt.Sprintf("%s/admin/customers/%d", o.cfg.App.UserServiceUrl, userID)
	if isCustomer {
		baseUrlUser = fmt.Sprintf("%s/auth/profile", o.cfg.App.UserServiceUrl)
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

func (o *orderService) httpClientBulkUserService(userIDs []int64, accessToken string) (map[int64]entity.CustomerResponseEntity, error) {
	if len(userIDs) == 0 {
		return nil, nil
	}

	idStrs := make([]string, len(userIDs))
	for i, id := range userIDs {
		idStrs[i] = strconv.FormatInt(id, 10)
	}
	idsQueryParam := strings.Join(idStrs, ",")
	
	// Assuming the bulk endpoint is /admin/customers/bulk
	baseUrlUser := fmt.Sprintf("%s/admin/customers/bulk?ids=%s", o.cfg.App.UserServiceUrl, idsQueryParam)

	header := map[string]string{
		"Authorization": "Bearer " + accessToken,
		"Accept":        "application/json",
	}

	resp, err := o.httpClient.CallURL("GET", baseUrlUser, header, nil)
	if err != nil {
		log.Errorf("[OrderService-1] httpClientBulkUserService: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("[OrderService-2] httpClientBulkUserService: %v", err)
		return nil, err
	}

	var bulkResponse entity.BulkUserHttpClientResponse
	err = json.Unmarshal(body, &bulkResponse)
	if err != nil {
		log.Errorf("[OrderService-3] httpClientBulkUserService: %v", err)
		return nil, err
	}

	// Convert slice to map for efficient lookup
	userMap := make(map[int64]entity.CustomerResponseEntity, len(bulkResponse.Data))
	for _, user := range bulkResponse.Data {
		userMap[user.ID] = user
	}

	return userMap, nil
}

func (o *orderService) httpClientProductService(productID int64, accessToken string, isCustomer bool) (*entity.ProductResponseEntity, error) {
	baseUrlProduct := fmt.Sprintf("%s/admin/products/%d", o.cfg.App.ProductServiceUrl, productID)
	if isCustomer {
		baseUrlProduct = fmt.Sprintf("%s/products/home/%d", o.cfg.App.ProductServiceUrl, productID)
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

	return &productResponse.Data, nil
}

func (o *orderService) httpClientBulkProductService(productIDs []int64, accessToken string, isCustomer bool) (map[int64]entity.ProductResponseEntity, error) {
	if len(productIDs) == 0 {
		return nil, nil
	}

	idStrs := make([]string, len(productIDs))
	for i, id := range productIDs {
		idStrs[i] = strconv.FormatInt(id, 10)
	}
	idsQueryParam := strings.Join(idStrs, ",")

	var baseUrlProduct string
	if isCustomer {
		baseUrlProduct = fmt.Sprintf("%s/products/home/bulk?ids=%s", o.cfg.App.ProductServiceUrl, idsQueryParam)
	} else {
		baseUrlProduct = fmt.Sprintf("%s/admin/products/bulk?ids=%s", o.cfg.App.ProductServiceUrl, idsQueryParam)
	}

	header := map[string]string{
		"Authorization": "Bearer " + accessToken,
		"Accept":        "application/json",
	}

	resp, err := o.httpClient.CallURL("GET", baseUrlProduct, header, nil)
	if err != nil {
		log.Errorf("[OrderService-1] httpClientBulkProductService: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("[OrderService-2] httpClientBulkProductService: %v", err)
		return nil, err
	}

	var bulkResponse entity.BulkProductHttpClientResponse
	if err := json.Unmarshal(body, &bulkResponse); err != nil {
		log.Errorf("[OrderService-3] httpClientBulkProductService: %v. Body: %s", err, string(body))
		return nil, err
	}

	productMap := make(map[int64]entity.ProductResponseEntity, len(bulkResponse.Data))
	for _, product := range bulkResponse.Data {
		productMap[product.ID] = product
	}

	return productMap, nil
}
