package entity

// BulkUserHttpClientResponse is expected to match the JSON response from the bulk user service endpoint.
type BulkUserHttpClientResponse struct {
	Message string                   `json:"message"`
	Data    []CustomerResponseEntity `json:"data"`
}

// UserHttpClientResponse is expected to match the JSON response from the single user service endpoint.
type UserHttpClientResponse struct {
	Message string                 `json:"message"`
	Data    CustomerResponseEntity `json:"data"`
}
