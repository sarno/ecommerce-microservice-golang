package handler

import (
	"net/http"
	"user-service/internal/adapter/handler/request"
	"user-service/internal/adapter/handler/response"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/service"

	"github.com/labstack/echo/v4"
	"github.com/labstack/gommon/log"
)

type IUserHandler interface {
	SignIn(ctx echo.Context) error
}

type userHandler struct {
	UserService service.IUserService
}

var err error

// SignIn implements IUserHandler.
func (u *userHandler) SignIn(c echo.Context) error {
	var (
		req = request.SignInRequest{}
		resp = response.DefaultResponse{}
		respSign = response.SignResponse{}
		ctx = c.Request().Context()
	)

	if err = c.Bind(&req); err != nil {
		log.Errorf("[UserHandler-1] SignIn : %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	if err = c.Validate(&req); err != nil {
		log.Errorf("[UserHandler-1] SignIn : %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	reqEntity := entity.UserEntity{
		Email:    req.Email,
		Password: req.Password,
	}

	user, token, err := u.UserService.SignIn(ctx, reqEntity)
	if err != nil {
		if err.Error() == "404" {
			log.Errorf("[UserHandler-1] SignIn : %v", "User not found")
			resp.Message = "User not found"
			resp.Data = nil
			return c.JSON(http.StatusNotFound, resp)
		}

		log.Errorf("[UserHandler-1] SignIn : %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	respSign.Id = user.ID
	respSign.Name = user.Name
	respSign.Email = user.Email
	respSign.Phone = user.Phone
	respSign.Address = user.Address
	respSign.Lat = user.Lat
	respSign.Lng = user.Lng
	respSign.AccessToken = token

	resp.Message = "success"
	resp.Data = respSign

	return c.JSON(http.StatusOK, resp)


	
}

func NewUserHandler(e *echo.Echo, userService service.IUserService) IUserHandler {
	userHandler := &userHandler{
		UserService: userService,
	}

	e.POST("/sign-in", userHandler.SignIn)
	// e.POST("/user/sign-up", userHandler.SignIn)
	// e.GET("/user/profile", userHandler.SignIn)
	// e.POST("/user/sign-out", userHandler.SignIn)
	

	return userHandler
}
