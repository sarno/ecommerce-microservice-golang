package entity

// BulkProductHttpClientResponse is expected to match the JSON response from the bulk product service endpoint.
type BulkProductHttpClientResponse struct {
	Message string                `json:"message"`
	Data    []ProductResponseEntity `json:"data"`
}
