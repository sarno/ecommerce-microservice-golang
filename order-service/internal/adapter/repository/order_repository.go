package repository

import (
	"context"
	"errors"
	"math"
	"order-service/internal/core/domain/entity"
	"order-service/internal/core/domain/model"

	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
)

type IOrderRepository interface {
	GetAll(ctx context.Context, queryString entity.QueryStringEntity) ([]entity.OrderEntity, int64, int64, error)
}

type OrderRepository struct {
	db *gorm.DB
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
		log.Errorf("[OrderRepository-1] GetAll: %v", err)
		return nil, 0, 0, err
	}

	totalPage := int(math.Ceil(float64(countData) / float64(queryString.Limit)))
	if err := sqlMain.Order("order_date DESC").Limit(int(queryString.Limit)).Offset(int(offset)).Find(&modelOrders).Error; err != nil {
		log.Errorf("[OrderRepository-2] GetAll: %v", err)
		return nil, 0, 0, err
	}

	if len(modelOrders) == 0 {
		err := errors.New("404")
		log.Infof("[OrderRepository-3] GetAll: No order found")
		return nil, 0, 0, err
	}

	entities := []entity.OrderEntity{}

	for _, val := range modelOrders {
		orderItemEntities := []entity.OrderItemEntity{}
		for _, item := range val.OrderItems {
			orderItemEntities = append(orderItemEntities, entity.OrderItemEntity{
				ID:        item.ID,
				ProductID: item.ProductID,
				Quantity:  item.Quantity,
			})
		}
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
