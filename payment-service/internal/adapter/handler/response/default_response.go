package response

type DefaultResponse struct {
	Message string      `json:"message"`
	Data    interface{} `json:"data"`
}

func ResponseDefault(message string, data interface{}) DefaultResponse {
	return DefaultResponse{
		Message: message,
		Data:    data,
	}
}