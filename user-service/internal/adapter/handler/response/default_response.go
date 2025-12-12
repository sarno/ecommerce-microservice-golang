package response

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type DefaultResponse struct {
	Message string `json:"message"`
	Data    interface{} `json:"data"`
}


func RespondWithError(c echo.Context, code int, context string, err error) error {
	log.Errorf("%s: %v", context, err)
	resp := DefaultResponse{
		Message: err.Error(),
		Data:    nil,
	}
	return c.JSON(code, resp)
}
