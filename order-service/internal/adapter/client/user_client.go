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

type IUserClient interface {
	GetUser(userID int64, accessToken string, isCustomer bool) (*entity.CustomerResponseEntity, error)
	GetUsersBulk(userIDs []int64, accessToken string) (map[int64]entity.CustomerResponseEntity, error)
}

type userClient struct {
	cfg        *config.Config
	httpClient httpclient.IHttpClient
}

func NewUserClient(cfg *config.Config, httpClient httpclient.IHttpClient) IUserClient {
	return &userClient{
		cfg:        cfg,
		httpClient: httpClient,
	}
}

func (c *userClient) GetUser(userID int64, accessToken string, isCustomer bool) (*entity.CustomerResponseEntity, error) {
	baseUrlUser := fmt.Sprintf("%s/admin/customers/%d", c.cfg.App.UserServiceUrl, userID)
	if isCustomer {
		baseUrlUser = fmt.Sprintf("%s/auth/profile", c.cfg.App.UserServiceUrl)
	}

	header := map[string]string{
		"Authorization": "Bearer " + accessToken,
		"Accept":        "application/json",
	}
	dataUser, err := c.httpClient.CallURL("GET", baseUrlUser, header, nil)
	if err != nil {
		log.Errorf("[UserClient-1] GetUser: %v", err)
		return nil, err
	}

	defer dataUser.Body.Close()
	body, err := io.ReadAll(dataUser.Body)
	if err != nil {
		log.Errorf("[UserClient-2] GetUser: %v", err)
		return nil, err
	}
	var userResponse entity.UserHttpClientResponse
	err = json.Unmarshal(body, &userResponse)
	if err != nil {
		log.Errorf("[UserClient-3] GetUser: %v", err)
		return nil, err
	}

	return &userResponse.Data, nil
}

func (c *userClient) GetUsersBulk(userIDs []int64, accessToken string) (map[int64]entity.CustomerResponseEntity, error) {
	if len(userIDs) == 0 {
		return make(map[int64]entity.CustomerResponseEntity), nil
	}

	idStrs := make([]string, len(userIDs))
	for i, id := range userIDs {
		idStrs[i] = strconv.FormatInt(id, 10)
	}
	idsQueryParam := strings.Join(idStrs, ",")

	baseUrlUser := fmt.Sprintf("%s/admin/customers/bulk?ids=%s", c.cfg.App.UserServiceUrl, idsQueryParam)

	header := map[string]string{
		"Authorization": "Bearer " + accessToken,
		"Accept":        "application/json",
	}

	resp, err := c.httpClient.CallURL("GET", baseUrlUser, header, nil)
	if err != nil {
		log.Errorf("[UserClient-1] GetUsersBulk: %v", err)
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Errorf("[UserClient-2] GetUsersBulk: %v", err)
		return nil, err
	}

	var bulkResponse entity.BulkUserHttpClientResponse
	err = json.Unmarshal(body, &bulkResponse)
	if err != nil {
		log.Errorf("[UserClient-3] GetUsersBulk: %v", err)
		return nil, err
	}

	userMap := make(map[int64]entity.CustomerResponseEntity, len(bulkResponse.Data))
	for _, user := range bulkResponse.Data {
		userMap[user.ID] = user
	}

	return userMap, nil
}
