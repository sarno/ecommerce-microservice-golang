package service

import (
	"context"
	"encoding/json"
	"errors"
	"payment-service/config"
	httpclient "payment-service/internal/adapter/http_client"
	"payment-service/internal/adapter/message"
	"payment-service/internal/adapter/repository"
	"payment-service/internal/core/domain/entity"

	"github.com/labstack/gommon/log"
)

type IPaymentService interface {
	ProcessPayment(ctx context.Context, payment entity.PaymentEntity, accessToken string) (*entity.PaymentEntity, error)
	UpdateStatusByOrderCode(ctx context.Context, orderCode, status string) error
	GetAll(ctx context.Context, req entity.PaymentQueryStringRequest, accessToken string) ([]entity.PaymentEntity, int64, int64, error)
	GetDetail(ctx context.Context, paymentID uint, accessToken string) (*entity.PaymentEntity, error)
}

type paymentService struct {
	paymentRepo       repository.IPaymentRepository
	publisherRabbitMQ message.IPublishRabbitMQ
	cfg               *config.Config
	serviceCall       httpclient.IServiceCall
	midtrans          httpclient.IMidtransClient
}

// GetDetail implements [IPaymentService].
func (p *paymentService) GetDetail(ctx context.Context, paymentID uint, accessToken string) (*entity.PaymentEntity, error) {
	result, err := p.paymentRepo.GetDetail(ctx, paymentID)
	if err != nil {
		log.Errorf("[PaymentService] GetDetail-1: %v", err)
		return nil, err
	}

	var token map[string]interface{}
	err = json.Unmarshal([]byte(accessToken), &token)

	if err != nil {
		log.Errorf("[PaymentService] GetDetail-2: %v", err)
		return nil, err
	}

	userID := int64(result.UserID)
	if token["role_name"].(string) == "admin" {
		userID = 0
	}

	orderDetail, err := p.serviceCall.OrderService(int64(result.OrderID), token["token"].(string))
	if err != nil {
		log.Errorf("[PaymentService] GetDetail-3: %v", err)
		return nil, err
	}

	isAdmin := false
	if token["role_name"].(string) == "admin" {
		isAdmin = true
	}

	userDetail, err := p.serviceCall.UserService(token["token"].(string), userID, isAdmin)
	if err != nil {
		log.Errorf("[PaymentService] GetDetail-4: %v", err)
		return nil, err
	}

	result.CustomerName = userDetail.Name
	result.CustomerEmail = userDetail.Email
	result.CustomerAddress = userDetail.Address

	result.OrderCode = orderDetail.OrderCode
	result.OrderShippingType = orderDetail.ShippingType
	result.OrderAt = orderDetail.OrderDatetime
	result.OrderRemarks = orderDetail.Remarks

	return result, nil
}

// GetAll implements [IPaymentService].
func (p *paymentService) GetAll(ctx context.Context, req entity.PaymentQueryStringRequest, accessToken string) ([]entity.PaymentEntity, int64, int64, error) {
	results, count, total, err := p.paymentRepo.GetAll(ctx, req)
	if err != nil {
		log.Errorf("[PaymentService] GetAll-1: %v", err)
		return nil, 0, 0, err
	}

	var token map[string]interface{}
	err = json.Unmarshal([]byte(accessToken), &token)

	if err != nil {
		log.Errorf("[PaymentService] GetAll-2: %v", err)
		return nil, 0, 0, err
	}

	for key, val := range results {
		orderDetail, err := p.serviceCall.OrderService(int64(val.OrderID), token["token"].(string))
		if err != nil {
			log.Errorf("[PaymentService] GetAll-3: %v", err)
			return nil, 0, 0, err
		}
		results[key].OrderCode = orderDetail.OrderCode
		results[key].OrderShippingType = orderDetail.ShippingType
	}

	return results, count, total, nil
}

// UpdateStatusByOrderCode implements [IPaymentService].
func (p *paymentService) UpdateStatusByOrderCode(ctx context.Context, orderCode string, status string) error {
	orderDetailID, err := p.serviceCall.PublicOrderIDByCodeService(orderCode)
	if err != nil {
		log.Errorf("[PaymentService] UpdateStatusByOrderCode-1: %v", err)
		return err
	}

	if err = p.paymentRepo.UpdateStatusByOrderCode(ctx, uint(orderDetailID), status); err != nil {
		log.Errorf("[PaymentService] UpdateStatusByOrderCode-2: %v", err)
		return err
	}

	return nil
}

// ProcessPayment implements [IPaymentService].
func (p *paymentService) ProcessPayment(ctx context.Context, payment entity.PaymentEntity, accessToken string) (*entity.PaymentEntity, error) {
	err := p.paymentRepo.GetByOrderID(ctx, uint(payment.OrderID))
	if err == nil {
		log.Infof("[PaymentService] ProcessPayment-1: Payment already exists")
		return nil, errors.New("Payment already exists")
	}

	if payment.PaymentMethod == "cod" {
		payment.PaymentStatus = "Success"
		if err := p.paymentRepo.CreatePayment(ctx, payment); err != nil {
			log.Errorf("[PaymentService] ProcessPayment-2: %v", err)
			return nil, err
		}

		if err := p.publisherRabbitMQ.PublishPaymentSuccess(payment); err != nil {
			log.Errorf("[PaymentService] ProcessPayment-3: %v", err)
		}

		return &payment, nil
	}

	if payment.PaymentMethod == "midtrans" {
		var token map[string]interface{}
		err := json.Unmarshal([]byte(accessToken), &token)
		if err != nil {
			log.Errorf("[PaymentService] ProcessPayment-4: %v", err)
			return nil, err
		}

		isAdmin := false
		if token["role_name"].(string) == "Super Admin" {
			isAdmin = true
		}

		userResponse, err := p.serviceCall.UserService(token["token"].(string), int64(payment.UserID), isAdmin)
		if err != nil {
			log.Errorf("[PaymentService] ProcessPayment-5: %v", err)
			return nil, err
		}

		orderDetail, err := p.serviceCall.OrderService(int64(payment.OrderID), token["token"].(string))
		if err != nil {
			log.Errorf("[PaymentService] ProcessPayment-6: %v", err)
			return nil, err
		}

		transactionID, err := p.midtrans.CreateTransaction(orderDetail.OrderCode, int64(payment.GrossAmount), userResponse.Name, userResponse.Email)
		if err != nil {
			log.Errorf("[PaymentService] ProcessPayment-7: %v", err)
			return nil, err
		}

		payment.PaymentStatus = "Pending"
		payment.PaymentGatewayID = transactionID

		if err := p.paymentRepo.CreatePayment(ctx, payment); err != nil {
			log.Errorf("[PaymentService] ProcessPayment-8: %v", err)
			return nil, err
		}

		if err := p.publisherRabbitMQ.PublishPaymentSuccess(payment); err != nil {
			log.Errorf("[PaymentService] ProcessPayment-9: %v", err)
		}

		return &payment, nil
	}

	return nil, errors.New("Invalid payment method")
}

func NewPaymentService(paymentRepo repository.IPaymentRepository, cfg *config.Config, serviceCall httpclient.IServiceCall, midtrans httpclient.IMidtransClient, publisherRabbitMQ message.IPublishRabbitMQ) IPaymentService {
	return &paymentService{
		paymentRepo:       paymentRepo,
		cfg:               cfg,
		serviceCall:       serviceCall,
		midtrans:          midtrans,
		publisherRabbitMQ: publisherRabbitMQ,
	}
}