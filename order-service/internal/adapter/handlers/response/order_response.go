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

type OrderAdminDetail struct {
	ID            int64         `json:"id"`
	OrderCode     string        `json:"order_code"`
	ProductImage  string        `json:"product_image"`
	OrderDatetime string        `json:"order_datetime"`
	Status        string        `json:"order_status"`
	PaymentMethod string        `json:"payment_method"`
	ShippingFee   int64         `json:"shipping_fee"`
	ShippingType  string        `json:"shipping_type"`
	Remarks       string        `json:"remarks"`
	TotalAmount   int64         `json:"total_amount"`
	Customer      CustomerOrder `json:"customer"`
	OrderDetail   []OrderDetail `json:"order_detail"`
}

type CustomerOrder struct {
	CustomerName    string `json:"customer_name"`
	CustomerPhone   string `json:"customer_phone"`
	CustomerAddress string `json:"customer_address"`
	CustomerEmail   string `json:"customer_email"`
	CustomerID      int64  `json:"customer_id"`
}

type OrderDetail struct {
	ProductName  string `json:"product_name"`
	ProductImage string `json:"product_image"`
	ProductPrice int64  `json:"product_price"`
	Quantity     int64  `json:"quantity"`
}