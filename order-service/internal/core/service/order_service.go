package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"order-service/config"
	httpclient "order-service/internal/adapter/http_client"
	"order-service/internal/adapter/repository"
	"order-service/internal/core/domain/entity"
	"strconv"

	"github.com/labstack/gommon/log"
)

type IOrderService interface {
	GetAll(ctx context.Context, queryString entity.QueryStringEntity, accessToken string) ([]entity.OrderEntity, int64, int64, error)
}

type orderService struct {
	repo repository.IOrderRepository
	cfg *config.Config
	httpClient        httpclient.IHttpClient
}



// GetAll implements [IOrderService].
func (o *orderService) GetAll(ctx context.Context, queryString entity.QueryStringEntity, accessToken string) ([]entity.OrderEntity, int64, int64, error) {
	results, count, total, err := o.repo.GetAll(ctx, queryString)
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

func NewOrderService(orderRepo repository.IOrderRepository, cfg *config.Config, httpClient httpclient.IHttpClient) IOrderService {
	return &orderService{
		repo: orderRepo,
		cfg: cfg,
		httpClient: httpClient,
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

	return &productResponse.Data, nil
}