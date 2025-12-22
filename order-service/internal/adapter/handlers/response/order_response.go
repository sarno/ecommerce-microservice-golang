package response

import "order-service/internal/core/domain/entity"

type OrderAdminList struct {
	ID            int64  `json:"id"`
	OrderCode     string `json:"order_code"`
	ProductImage  string `json:"product_image"`
	CustomerName  string `json:"customer_name"`
	Status        string `json:"status"`
	PaymentMethod string `json:"payment_method"`
	TotalAmount   int64  `json:"total_amount"`
}

func NewOrderAdminList(e entity.OrderEntity) OrderAdminList {
	var productImage string
	for _, val := range e.OrderItems {
		productImage = val.ProductImage
	}

	return OrderAdminList{
		ID:            e.ID,
		OrderCode:     e.OrderCode,
		ProductImage:  productImage,
		CustomerName:  e.BuyerName,
		Status:        e.Status,
		PaymentMethod: e.PaymentMethod,
		TotalAmount:   e.TotalAmount,
	}
}

