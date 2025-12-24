package client

import (
	"encoding/json"
	"fmt"
	"io"
	"order-service/config"
	httpclient "order-service/internal/adapter/http_client"
	"order-service/internal/core/domain/entity"
	"strconv"
	"strings"

	"github.com/labstack/gommon/log"
)

type IProductClient interface {
	GetProduct(productID int64, accessToken string, isCustomer bool) (*entity.ProductResponseEntity, error)
	GetProductsBulk(productIDs []int64, accessToken string, isCustomer bool) (map[int64]entity.ProductResponseEntity, error)
}

type productClient struct {
	cfg        *config.Config
	httpClient httpclient.IHttpClient
}

func NewProductClient(cfg *config.Config, httpClient httpclient.IHttpClient) IProductClient {
	return &productClient{
		cfg:        cfg,
		httpClient: httpClient,
	}
}

func (c *productClient) GetProduct(productID int64, accessToken string, isCustomer bool) (*entity.ProductResponseEntity, error) {
	baseUrlProduct := fmt.Sprintf("%s/admin/products/%d", c.cfg.App.ProductServiceUrl, productID)
	if isCustomer {
		baseUrlProduct = fmt.Sprintf("%s/products/home/%d", c.cfg.App.ProductServiceUrl, productID)
	}

	header := map[string]string{
		"Authorization": "Bearer " + accessToken,
		"Accept":        "application/json",
	}

	dataProduct, err := c.httpClient.CallURL("GET", baseUrlProduct, header, nil)
	if err != nil {
		log.Errorf("[ProductClient-1] GetProduct: %v", err)
		return nil, err
	}
	defer dataProduct.Body.Close()

	body, err := io.ReadAll(dataProduct.Body)
	if err != nil {
		log.Errorf("[ProductClient-2] GetProduct: %v", err)
		return nil, err
	}

	var productResponse entity.ProductHttpClientResponse
	err = json.Unmarshal(body, &productResponse)
	if err != nil {
		log.Errorf("[ProductClient-3] GetProduct: %v", err)
		return nil, err
	}

	return &productResponse.Data, nil
}

func (c *productClient) GetProductsBulk(productIDs []int64, accessToken string, isCustomer bool) (map[int64]entity.ProductResponseEntity, error) {
	if len(productIDs) == 0 {
		return make(map[int64]entity.ProductResponseEntity), nil
	}

	idStrs := make([]string, len(productIDs))
	for i, id := range productIDs {
		idStrs[i] = strconv.FormatInt(id, 10)
	}
	idsQueryParam := strings.Join(idStrs, ",")

	var baseUrlProduct string
	if isCustomer {
		baseUrlProduct = fmt.Sprintf("%s/products/home/bulk?ids=%s", c.cfg.App.ProductServiceUrl, idsQueryParam)
	} else {
		baseUrlProduct = fmt.Sprintf("%s/admin/products/bulk?ids=%s", c.cfg.App.ProductServiceUrl, idsQueryParam)
	}

	header := map[string]string{
		"Authorization": "Bearer " + accessToken,
		"Accept":        "application/json",
	}

	resp, err := c.httpClient.CallURL("GET", baseUrlProduct, header, nil)
	if err != nil {
		log.Errorf("[ProductClient-1] GetProductsBulk: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("[ProductClient-2] GetProductsBulk: %v", err)
		return nil, err
	}

	var bulkResponse entity.BulkProductHttpClientResponse
	if err := json.Unmarshal(body, &bulkResponse); err != nil {
		log.Errorf("[ProductClient-3] GetProductsBulk: %v. Body: %s", err, string(body))
		return nil, err
	}

	productMap := make(map[int64]entity.ProductResponseEntity, len(bulkResponse.Data))
	for _, product := range bulkResponse.Data {
		productMap[product.ID] = product
	}

	return productMap, nil
}
