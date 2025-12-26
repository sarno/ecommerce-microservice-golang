package httpclient

import (
	"encoding/json"
	"fmt"
	"io"
	"payment-service/config"
	"payment-service/internal/core/domain/entity"
	"strconv"

	"github.com/labstack/gommon/log"
)

type IServiceCall interface {
	UserService(accessToken string, userID int64, isAdmin bool) (*entity.ProfileHttpResponse, error)
	OrderService(orderId int64, accessToken string) (*entity.OrderDetailHttpResponse, error)
	PublicOrderIDByCodeService(orderCode string) (int64, error)
}

type serviceCall struct {
	cfg                 *config.Config
	httpClientToService HttpClientToService
}

func NewServiceCall(cfg *config.Config, httpClientToService HttpClientToService) IServiceCall {
	return &serviceCall{
		cfg:                 cfg,
		httpClientToService: httpClientToService,
	}
}

func (s *serviceCall) UserService(accessToken string, userID int64, isAdmin bool) (*entity.ProfileHttpResponse, error) {
	baseUrlUser := fmt.Sprintf("%s/%s", s.cfg.App.UserServiceUrl, "auth/profile")
	if isAdmin {
		baseUrlUser = fmt.Sprintf("%s/%s", s.cfg.App.UserServiceUrl, "admin/customers/"+strconv.FormatInt(userID, 10))
	}

	header := map[string]string{
		"Authorization": "Bearer " + accessToken,
		"Accept":        "application/json",
	}
	dataUser, err := s.httpClientToService.CallURL("GET", baseUrlUser, header, nil)
	if err != nil {
		log.Errorf("[ServiceCall] UserService-1: %v", err)
		return nil, err
	}

	defer dataUser.Body.Close()

	body, err := io.ReadAll(dataUser.Body)
	if err != nil {
		log.Errorf("[ServiceCall] UserService-2: %v", err)
		return nil, err
	}

	var userResponse entity.UserHttpClientResponse
	err = json.Unmarshal([]byte(body), &userResponse)
	if err != nil {
		log.Errorf("[ServiceCall] UserService-3: %v", err)
		return nil, err
	}

	return &userResponse.Data, nil
}

func (s *serviceCall) OrderService(orderId int64, accessToken string) (*entity.OrderDetailHttpResponse, error) {
	baseUrlOrder := fmt.Sprintf("%s/%s", s.cfg.App.OrderServiceUrl, "auth/orders/"+strconv.FormatInt(orderId, 10))
	header := map[string]string{
		"Authorization": "Bearer " + accessToken,
		"Accept":        "application/json",
	}

	dataOrder, err := s.httpClientToService.CallURL("GET", baseUrlOrder, header, nil)
	if err != nil {
		log.Errorf("[ServiceCall] OrderService-1: %v", err)
		return nil, err
	}

	defer dataOrder.Body.Close()
	body, err := io.ReadAll(dataOrder.Body)
	if err != nil {
		log.Errorf("[ServiceCall] OrderService-2: %v", err)
		return nil, err
	}

	var orderDetail entity.OrderHttpClientResponse
	err = json.Unmarshal([]byte(body), &orderDetail)
	if err != nil {
		log.Errorf("[ServiceCall] OrderService-3: %v", err)
		return nil, err
	}

	return &orderDetail.Data, nil
}

func (s *serviceCall) PublicOrderIDByCodeService(orderCode string) (int64, error) {
	baseUrlOrder := fmt.Sprintf("%s/%s", s.cfg.App.OrderServiceUrl, "public/orders/"+orderCode+"/code")
	header := map[string]string{
		"Accept": "application/json",
	}

	dataOrder, err := s.httpClientToService.CallURL("GET", baseUrlOrder, header, nil)
	if err != nil {
		log.Errorf("[ServiceCall] PublicOrderIDByCodeService-1: %v", err)
		return 0, err
	}

	defer dataOrder.Body.Close()
	if dataOrder.StatusCode != 200 {
		return 0, fmt.Errorf("Order not found")
	}

	body, err := io.ReadAll(dataOrder.Body)
	if err != nil {
		log.Errorf("[ServiceCall] PublicOrderIDByCodeService-2: %v", err)
		return 0, err
	}

	var orderDetail entity.GetOrderIDByCodeResponse
	err = json.Unmarshal([]byte(body), &orderDetail)

	if err != nil {
		log.Errorf("[ServiceCall] PublicOrderIDByCodeService-4: %v", err)
		return 0, err
	}

	return int64(orderDetail.Data.OrderID), nil
}
