package response

type DefaultResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

type DefaultResponseWithPaginations struct {
	Message    string      `json:"message"`
	Data       interface{} `json:"data"`
	Pagination *Pagination `json:"pagination,omitempty"`
}

type Pagination struct {
	Page       int64 `json:"page"`
	TotalCount int64 `json:"total_count"`
	PerPage    int64 `json:"per_page"`
	TotalPage  int64 `json:"total_page"`
}

type ProductHomeListResponse struct {
	ID           int64  `json:"id"`
	ProductName  string `json:"product_name"`
	ProductImage string `json:"product_image"`
	CategoryName string `json:"category_name"`
	SalePrice    int64  `json:"sale_price"`
	RegulerPrice int64  `json:"reguler_price"`
}