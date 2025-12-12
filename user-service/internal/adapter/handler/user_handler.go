package handler

import (
	"errors"
	"net/http"
	"user-service/config"
	"user-service/internal/adapter"
	"user-service/internal/adapter/handler/request"
	"user-service/internal/adapter/handler/response"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/service"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

type IUserHandler interface {
	SignIn(ctx echo.Context) error
	CreateUserAccount(ctx echo.Context) error
}

type userHandler struct {
	UserService service.IUserService
}

// CreateUserAccount implements IUserHandler.
func (u *userHandler) CreateUserAccount(c echo.Context) error {
	var (
		req  = request.SignUpRequest{}
		resp = response.DefaultResponse{}
		ctx  = c.Request().Context()
	)

	if err = c.Bind(&req); err != nil {
		return response.RespondWithError(c, http.StatusUnprocessableEntity, "[UserHandler-1] CreateUserAccount", err)
	}

	if err = c.Validate(req); err != nil {
		return response.RespondWithError(c, http.StatusUnprocessableEntity, "[UserHandler-2] CreateUserAccount", err)
	}

	if req.Password != req.PasswordConfirm {
		err = errors.New("passwords do not match")
		return response.RespondWithError(c, http.StatusUnprocessableEntity, "[UserHandler-3] CreateUserAccount", err)
	}

	reqEntity := entity.UserEntity{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	}

	err = u.UserService.CreateUserAccount(ctx, reqEntity)
	if err != nil {
		return response.RespondWithError(c, http.StatusInternalServerError, "[UserHandler-4] CreateUserAccount", err)
	}

	resp.Message = "Success"
	
	return c.JSON(http.StatusCreated, resp)

}

var err error

// SignIn implements IUserHandler.
func (u *userHandler) SignIn(c echo.Context) error {
	var (
		req      = request.SignInRequest{}
		resp     = response.DefaultResponse{}
		respSign = response.SignResponse{}
		ctx      = c.Request().Context()
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

func NewUserHandler(e *echo.Echo, userService service.IUserService, cfg *config.Config, jwtService service.IJWTService) IUserHandler {
	userHandler := &userHandler{
		UserService: userService,
	}

	e.Use(middleware.Recover())
	e.POST("/sign-in", userHandler.SignIn)
	e.POST("/signup", userHandler.CreateUserAccount)

	mid := adapter.NewMiddlewareAdapter(cfg, jwtService)
	adminGroup := e.Group("/admin", mid.CheckToken())
	adminGroup.GET("/check", func(c echo.Context) error {
		return c.String(200, "OK")
	})

	return userHandler
}
