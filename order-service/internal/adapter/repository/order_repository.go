package repository

import (
	"context"
	"errors"
	"math"
	"order-service/internal/core/domain/entity"
	"order-service/internal/core/domain/model"
	"time"

	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
)

type IOrderRepository interface {
	GetAll(ctx context.Context, queryString entity.QueryStringEntity) ([]entity.OrderEntity, int64, int64, error)
	GetByID(ctx context.Context, orderID int64) (*entity.OrderEntity, error)
	CreateOrder(ctx context.Context, req entity.OrderEntity) (int64, error)
}

type OrderRepository struct {
	db *gorm.DB
}

// GetByID implements [IOrderRepository].
func (o *OrderRepository) GetByID(ctx context.Context, orderID int64) (*entity.OrderEntity, error) {
	var modelOrder model.Order
	if err := o.db.Preload("OrderItems").Where("id =?", orderID).First(&modelOrder).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("404")
			log.Infof("[OrderRepository-1] GetByID: Order not found")
			return nil, err
		}
		return nil, o.logAndReturnError(err, "OrderRepository-2", "GetByID")
	}

	orderItemEntities := o.mapOrderItemModelsToEntities(modelOrder.OrderItems)

	return &entity.OrderEntity{
		ID:           modelOrder.ID,
		OrderCode:    modelOrder.OrderCode,
		Status:       modelOrder.Status,
		BuyerId:      modelOrder.BuyerId,
		OrderDate:    modelOrder.OrderDate.Format("2006-01-02 15:04:05"),
		TotalAmount:  int64(modelOrder.TotalAmount),
		OrderItems:   orderItemEntities,
		Remarks:      modelOrder.Remarks,
		ShippingType: modelOrder.ShippingType,
		ShippingFee:  int64(modelOrder.ShippingFee),
	}, nil
}

// CreateOrder implements [IOrderRepository].
func (o *OrderRepository) CreateOrder(ctx context.Context, req entity.OrderEntity) (int64, error) {
	orderDate, err := time.Parse("2006-01-02", req.OrderDate) // YYYY-MM-DD
	if err != nil {
		return 0, o.logAndReturnError(err, "OrderRepository-1", "CreateOrder")
	}

	var orderItems []model.OrderItem
	orderItems = o.mapOrderItemEntitiesToModels(req.OrderItems)

	modelOrder := model.Order{
		OrderCode:    req.OrderCode,
		BuyerId:      req.BuyerId,
		OrderDate:    orderDate,
		OrderTime:    req.OrderTime,
		Status:       req.Status,
		TotalAmount:  float64(req.TotalAmount),
		ShippingType: req.ShippingType,
		ShippingFee:  float64(req.ShippingFee),
		Remarks:      req.Remarks,
		OrderItems:   orderItems,
	}

	if err := o.db.Create(&modelOrder).Error; err != nil {
		return 0, o.logAndReturnError(err, "OrderRepository-3", "CreateOrder")
	}

	return modelOrder.ID, nil
}

// GetAll implements [IOrderRepository].
func (o *OrderRepository) GetAll(ctx context.Context, queryString entity.QueryStringEntity) ([]entity.OrderEntity, int64, int64, error) {
	var modelOrders []model.Order
	var countData int64
	offset := (queryString.Page - 1) * queryString.Limit

	sqlMain := o.db.Preload("OrderItems").
		Where("order_code ILIKE ? OR status ILIKE ?", "%"+queryString.Search+"%", "%"+queryString.Status+"%")

	if queryString.BuyerID != 0 {
		sqlMain = sqlMain.Where("buyer_id = ?", queryString.BuyerID)
	}

	if err := sqlMain.Model(&modelOrders).Count(&countData).Error; err != nil {
		return nil, 0, 0, o.logAndReturnError(err, "OrderRepository-1", "GetAll")
	}

	totalPage := int(math.Ceil(float64(countData) / float64(queryString.Limit)))
	if err := sqlMain.Order("order_date DESC").Limit(int(queryString.Limit)).Offset(int(offset)).Find(&modelOrders).Error; err != nil {
		return nil, 0, 0, o.logAndReturnError(err, "OrderRepository-2", "GetAll")
	}

	if len(modelOrders) == 0 {
		err := errors.New("404")
		log.Infof("[OrderRepository-3] GetAll: No order found")
		return nil, 0, 0, err
	}

	entities := []entity.OrderEntity{}

	for _, val := range modelOrders {
		orderItemEntities := o.mapOrderItemModelsToEntities(val.OrderItems)

		entities = append(entities, entity.OrderEntity{
			ID:          val.ID,
			OrderCode:   val.OrderCode,
			Status:      val.Status,
			OrderDate:   val.OrderDate.Format("2006-01-02 15:04:05"),
			TotalAmount: int64(val.TotalAmount),
			OrderItems:  orderItemEntities,
			BuyerId:     val.BuyerId,
		})
	}

	return entities, countData, int64(totalPage), nil
}

func NewOrderRepository(db *gorm.DB) IOrderRepository {
	return &OrderRepository{db: db}
}

func (o *OrderRepository) logAndReturnError(err error, code string, method string) error {
	log.Errorf("[%s] %s: %v", code, method, err)
	return err
}

func (o *OrderRepository) mapOrderItemEntitiesToModels(itemEntities []entity.OrderItemEntity) []model.OrderItem {
	orderItems := make([]model.OrderItem, 0, len(itemEntities))
	for _, item := range itemEntities {
		orderItems = append(orderItems, model.OrderItem{
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}
	return orderItems
}

func (o *OrderRepository) mapOrderItemModelsToEntities(itemModels []model.OrderItem) []entity.OrderItemEntity {
	orderItemEntities := make([]entity.OrderItemEntity, 0, len(itemModels))
	for _, item := range itemModels {
		orderItemEntities = append(orderItemEntities, entity.OrderItemEntity{
			ID:        item.ID,
			ProductID: item.ProductID,
			Quantity:  item.Quantity,
		})
	}
	return orderItemEntities
}
