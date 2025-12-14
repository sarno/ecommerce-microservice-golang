package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"
	"user-service/config"
	"user-service/internal/adapter"
	"user-service/internal/adapter/handler/request"
	"user-service/internal/adapter/handler/response"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/service"
	"user-service/utils/conv"

	"github.com/go-redis/redis/v8"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"github.com/labstack/gommon/log"
)

type IUserHandler interface {
	SignIn(ctx echo.Context) error
	CreateUserAccount(ctx echo.Context) error
	ForgotPassword(ctx echo.Context) error
	VerifyAccount(c echo.Context) error
	UpdatePassword(ctx echo.Context) error
	GetProfileUser(c echo.Context) error
	UpdateDataUser(ctx echo.Context) error

	//modul user
	GetCustomerAll(c echo.Context) error
	GetCustomerByID(c echo.Context) error
	CreateCustomer(c echo.Context) error
	UpdateCustomer(c echo.Context) error
	DeleteCustomer(c echo.Context) error
}

type userHandler struct {
	UserService service.IUserService
	RedisClient *redis.Client
}

// GetCustomerAll implements IUserHandler.

func (u *userHandler) SignIn(c echo.Context) error {
	var (
		req      = request.SignInRequest{}
		resp     = response.DefaultResponse{}
		respSign = response.SignInResponse{}
		ctx      = c.Request().Context()
		err      error
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

// UpdateDataUser implements IUserHandler.
func (u *userHandler) UpdateDataUser(c echo.Context) error {
	var (
		req         = request.UpdateDataUserRequest{}
		resp        = response.DefaultResponse{}
		ctx         = c.Request().Context()
		jwtUserData = entity.JwtUserData{}
	)

	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[UserHandler-1] UpdateDataUser: %s", "data token not found")
		resp.Message = "data token not found"
		resp.Data = nil
		return c.JSON(http.StatusNotFound, resp)
	}

	err := json.Unmarshal([]byte(user), &jwtUserData)
	if err != nil {
		log.Errorf("[UserHandler-2] UpdateDataUser: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	userID := jwtUserData.UserID

	if err = c.Bind(&req); err != nil {
		log.Errorf("[UserHandler-3] UpdateDataUser: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	if err = c.Validate(&req); err != nil {
		log.Errorf("[UserHandler-4] UpdateDataUser: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	reqEntity := entity.UserEntity{
		ID:      userID,
		Name:    req.Name,
		Email:   req.Email,
		Address: req.Address,
		Lat:     req.Lat,
		Lng:     req.Lng,
		Phone:   req.Phone,
		Photo:   req.Photo,
	}

	err = u.UserService.UpdateDataUser(ctx, reqEntity)
	if err != nil {
		log.Errorf("[UserHandler-5] UpdateDataUser: %v", err)
		if err.Error() == "404" {
			resp.Message = "User not found"
			resp.Data = nil
			return c.JSON(http.StatusNotFound, resp)
		}
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	resp.Message = "success"
	resp.Data = nil
	return c.JSON(http.StatusOK, resp)
}

// GetProfileUser implements IUserHandler.
func (u *userHandler) GetProfileUser(c echo.Context) error {
	var (
		resp        = response.DefaultResponse{}
		respProfile = response.ProfileResponse{}
		ctx         = c.Request().Context()
		jwtUserData = entity.JwtUserData{}
	)

	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[UserHandler-1] GetProfileUser: %s", "data token not found")
		resp.Message = "data token not found"
		resp.Data = nil
		return c.JSON(http.StatusNotFound, resp)
	}

	err := json.Unmarshal([]byte(user), &jwtUserData)
	if err != nil {
		log.Errorf("[UserHandler-2] GetProfileUser: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	userID := jwtUserData.UserID

	dataUser, err := u.UserService.GetProfileUser(ctx, userID)
	if err != nil {
		log.Errorf("[UserHandler-3] GetProfileUser: %v", err)
		if err.Error() == "404" {
			resp.Message = "user not found"
			resp.Data = nil
			return c.JSON(http.StatusNotFound, resp)
		}
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	respProfile.Address = dataUser.Address
	respProfile.Name = dataUser.Name
	respProfile.Email = dataUser.Email
	respProfile.ID = dataUser.ID
	respProfile.Lat = dataUser.Lat
	respProfile.Lng = dataUser.Lng
	respProfile.Phone = dataUser.Phone
	respProfile.Photo = dataUser.Photo
	respProfile.RoleName = dataUser.RoleName

	resp.Message = "success"
	resp.Data = respProfile

	return c.JSON(http.StatusOK, resp)
}

// UpdatePassword implements IUserHandler.
func (u *userHandler) UpdatePassword(c echo.Context) error {
	var (
		resp = response.DefaultResponse{}
		req  = request.UpdatePasswordRequest{}
		ctx  = c.Request().Context()
		err  error
	)

	tokenString := c.QueryParam("token")

	if tokenString == "" {
		log.Infof("[UserHandler-1] UpdatePassword: %s", "missing or invalid token")
		resp.Message = "missing or invalid token"
		resp.Data = nil
		return c.JSON(http.StatusUnauthorized, resp)
	}

	if err = c.Bind(&req); err != nil {
		log.Infof("[UserHandler-2] UpdatePassword: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	if err = c.Validate(req); err != nil {
		log.Errorf("[UserHandler-3] UpdatePassword: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	if req.NewPassword != req.ConfirmPassword {
		log.Infof("[UserHandler-4] UpdatePassword: %s", "new password and confirm password does not match")
		resp.Message = "new password and confirm password does not match"
		resp.Data = nil
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	reqEntity := entity.UserEntity{
		Password: req.NewPassword,
		Token:    tokenString,
	}

	err = u.UserService.UpdatePassword(ctx, reqEntity)

	if err != nil {
		log.Errorf("[UserHandler-5] UpdatePassword: %v", err)
		if err.Error() == "404" {
			resp.Message = "User not found"
			resp.Data = nil
			return c.JSON(http.StatusNotFound, resp)
		}

		if err.Error() == "401" {
			resp.Message = "Token expired or invalid"
			resp.Data = nil
			return c.JSON(http.StatusUnauthorized, resp)
		}
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	resp.Data = nil
	resp.Message = "Password updated successfully"

	return c.JSON(http.StatusOK, resp)
}

// VerifyAccount implements IUserHandler.
func (u *userHandler) VerifyAccount(c echo.Context) error {
	var (
		resp       = response.DefaultResponse{}
		respSignIn = response.SignInResponse{}
		ctx        = c.Request().Context()
	)

	tokenString := c.QueryParam("token")
	if tokenString == "" {
		log.Infof("[UserHandler-1] VerifyAccount: %s", "missing or invalid token")
		resp.Message = "missing or invalid token"
		resp.Data = nil
		return c.JSON(http.StatusUnauthorized, resp)
	}

	user, err := u.UserService.VerifyToken(ctx, tokenString)
	if err != nil {
		log.Errorf("[UserHandler-2] VerifyAccount: %v", err)
		if err.Error() == "404" {
			resp.Message = "User not found"
			resp.Data = nil
			return c.JSON(http.StatusNotFound, resp)
		}

		if err.Error() == "401" {
			resp.Message = "Token expired or invalid"
			resp.Data = nil
			return c.JSON(http.StatusUnauthorized, resp)
		}
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	respSignIn.Id = user.ID
	respSignIn.Name = user.Name
	respSignIn.Email = user.Email
	respSignIn.Role = user.RoleName
	respSignIn.Lat = user.Lat
	respSignIn.Lng = user.Lng
	respSignIn.Phone = user.Phone
	respSignIn.AccessToken = user.Token

	resp.Message = "Success"
	resp.Data = respSignIn

	return c.JSON(http.StatusOK, resp)
}

// ForgotPassword implements IUserHandler.
func (u *userHandler) ForgotPassword(c echo.Context) error {
	var (
		req  = request.ForgotPasswordRequest{}
		resp = response.DefaultResponse{}
		ctx  = c.Request().Context()
		err  error
	)

	if err = c.Bind(&req); err != nil {
		log.Errorf("[UserHandler-1] ForgotPassword: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	if err = c.Validate(req); err != nil {
		log.Errorf("[UserHandler-2] ForgotPassword: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	reqEntity := entity.UserEntity{
		Email: req.Email,
	}

	err = u.UserService.ForgotPassword(ctx, reqEntity)
	if err != nil {
		log.Errorf("[UserHandler-3] ForgotPassword: %v", err)
		if err.Error() == "404" {
			resp.Message = "User not found"
			resp.Data = nil
			return c.JSON(http.StatusNotFound, resp)
		}
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	resp.Message = "Success"
	resp.Data = nil
	return c.JSON(http.StatusOK, resp)
}

// CreateUserAccount implements IUserHandler.
func (u *userHandler) CreateUserAccount(c echo.Context) error {
	var (
		req  = request.SignUpRequest{}
		resp = response.DefaultResponse{}
		ctx  = c.Request().Context()
		err  error
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

func (u *userHandler) GetCustomerAll(c echo.Context) error {
	var (
		resp     = response.DefaultResponseWithPaginations{}
		ctx      = c.Request().Context()
		respUser = []response.CustomerListResponse{}
	)

	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[UserHandler-1] GetCustomerAll: %s", "data token not found")
		resp.Message = "data token not found"
		resp.Data = nil
		return c.JSON(http.StatusNotFound, resp)
	}

	search := c.QueryParam("search")
	orderBy := "created_at"
	if c.QueryParam("order_by") != "" {
		orderBy = c.QueryParam("order_by")
	}

	orderType := c.QueryParam("order_type")
	if orderType != "asc" && orderType != "desc" {
		orderType = "desc"
	}

	pageStr := c.QueryParam("page")
	var page int = 1
	if pageStr != "" {
		page, _ = conv.StringToInt(pageStr)
		if page <= 0 {
			page = 1
		}
	}

	limitStr := c.QueryParam("limit")
	var limit int = 10
	if limitStr != "" {
		limit, _ = conv.StringToInt(limitStr)
		if limit <= 0 {
			limit = 10
		}
	}

	// Membuat cache key yang unik berdasarkan parameter query
	cacheKey := fmt.Sprintf("customers:%s:%s:%s:page_%d:limit_%d", search, orderBy, orderType, page, limit)

	// Coba ambil dari Redis dulu
	cachedData, err := u.RedisClient.Get(ctx, cacheKey).Result()
	if err == nil {
		// Cache hit!
		log.Infof("Cache hit for key: %s", cacheKey)
		c.Response().Header().Set("X-Cache-Status", "HIT")
		c.Response().Header().Set(echo.HeaderContentType, echo.MIMEApplicationJSONCharsetUTF8)
		return c.String(http.StatusOK, cachedData)
	} else if err != redis.Nil {
		// Jika ada error selain cache miss, log errornya tapi tetap lanjut ke db
		log.Errorf("[UserHandler] GetCustomerAll: Redis error: %v", err)
	}

	// Cache miss atau error, ambil dari database
	c.Response().Header().Set("X-Cache-Status", "MISS")
	log.Infof("Cache miss for key: %s. Fetching from database.", cacheKey)
	reqEntity := entity.QueryStringCustomer{
		Search:    search,
		Page:      page,
		Limit:     limit,
		OrderBy:   orderBy,
		OrderType: orderType,
	}

	results, countData, totalPages, err := u.UserService.GetCustomerAll(ctx, reqEntity)
	if err != nil {
		if err.Error() == "404" {
			resp.Message = "Data not found"
			resp.Data = nil
			return c.JSON(http.StatusNotFound, resp)
		}

		log.Errorf("[UserHandler-2] GetCustomerAll: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	for _, val := range results {
		respUser = append(respUser, response.CustomerListResponse{
			ID:    val.ID,
			Name:  val.Name,
			Email: val.Email,
			Photo: val.Photo,
			Phone: val.Phone,
		})
	}

	resp.Message = "Data retrieved successfully"
	resp.Data = respUser
	resp.Pagination = &response.Pagination{
		Page:       page,
		TotalCount: countData,
		PerPage:    limit,
		TotalPage:  totalPages,
	}

	// Simpan hasil ke Redis dengan expiration time 5 menit
	jsonData, err := json.Marshal(resp)
	if err != nil {
		log.Errorf("[UserHandler] GetCustomerAll: Failed to marshal response for caching: %v", err)
	} else {
		err := u.RedisClient.Set(ctx, cacheKey, jsonData, 5*time.Minute).Err()
		if err != nil {
			log.Errorf("[UserHandler] GetCustomerAll: Failed to set cache in Redis: %v", err)
		}
	}

	return c.JSON(http.StatusOK, resp)
}

func (u *userHandler) GetCustomerByID(c echo.Context) error {
	var (
		resp     = response.DefaultResponseWithPaginations{}
		ctx      = c.Request().Context()
		respUser = response.CustomerResponse{}
	)

	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[UserHandler-1] GetCustomerByID: %s", "data token not found")
		resp.Message = "data token not found"
		resp.Data = nil
		return c.JSON(http.StatusNotFound, resp)
	}

	idStr := c.Param("id")
	if idStr == "" {
		log.Errorf("[UserHandler-2] GetCustomerByID: %s", "id invalid")
		resp.Message = "id invalid"
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	id, err := conv.StringToInt(idStr)
	if err != nil {
		log.Errorf("[UserHandler-2] GetCustomerByID: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	result, err := u.UserService.GetCustomerByID(ctx, id)
	if err != nil {
		if err.Error() == "404" {
			resp.Message = "Data not found"
			resp.Data = nil
			return c.JSON(http.StatusNotFound, resp)
		}

		log.Errorf("[UserHandler-3] GetCustomerByID: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	respUser.ID = result.ID
	respUser.RoleID = result.RoleID
	respUser.Name = result.Name
	respUser.Email = result.Email
	respUser.Phone = result.Phone
	respUser.Address = result.Address
	respUser.Photo = result.Photo
	respUser.Lat = result.Lat
	respUser.Lng = result.Lng

	resp.Message = "Data retrieved successfully"
	resp.Data = respUser
	resp.Pagination = nil

	return c.JSON(http.StatusOK, resp)
}

func (u *userHandler) CreateCustomer(c echo.Context) error {
	var (
		resp = response.DefaultResponseWithPaginations{}
		ctx  = c.Request().Context()
		req  = request.CustomerRequest{}
	)

	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[UserHandler-1] CreateCustomer: %s", "data token not found")
		resp.Message = "data token not found"
		resp.Data = nil
		return c.JSON(http.StatusNotFound, resp)
	}

	err := c.Bind(&req)
	if err != nil {
		log.Errorf("[UserHandler-2] CreateCustomer: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	if err = c.Validate(&req); err != nil {
		log.Errorf("[UserHandler-3] CreateCustomer: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	if req.Password != req.PasswordConfirmation {
		log.Infof("[UserHandler-4] CreateCustomer: %s", "password and confirm password does not match")
		resp.Message = "password and confirm password does not match"
		resp.Data = nil
		return c.JSON(http.StatusUnprocessableEntity, resp)
	}

	hashedPassword, err := conv.HashPassword(req.Password)
	if err != nil {
		log.Errorf("[UserHandler-5] CreateCustomer: failed to hash password %v", err)
		resp.Message = "internal server error"
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	latString := strconv.FormatFloat(req.Lat, 'g', -1, 64)
	lngString := strconv.FormatFloat(req.Lng, 'g', -1, 64)

	reqEntity := entity.UserEntity{
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword,
		Phone:    req.Phone,
		Address:  req.Address,
		Lat:      latString,
		Lng:      lngString,
		Photo:    req.Photo,
		RoleID:   req.RoleID,
	}

	err = u.UserService.CreateCustomer(ctx, reqEntity)
	if err != nil {
		if strings.Contains(err.Error(), "violates unique constraint") {
			log.Warnf("[UserHandler-6] CreateCustomer: Duplicate entry attempt: %v", err)
			resp.Message = "Email already registered."
			resp.Data = nil
			return c.JSON(http.StatusConflict, resp)
		}

		log.Errorf("[UserHandler-7] CreateCustomer: %v", err)
		resp.Message = "An internal server error occurred."
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	resp.Message = "success"
	resp.Data = nil
	resp.Pagination = nil

	return c.JSON(http.StatusCreated, resp)
}

func (u *userHandler) UpdateCustomer(c echo.Context) error {
	var (
		resp = response.DefaultResponseWithPaginations{}
		ctx  = c.Request().Context()
		req  = request.UpdateCustomerRequest{}
	)

	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[UserHandler-1] UpdateCustomer: %s", "data token not found")
		resp.Message = "data token not valid"
		resp.Data = nil
		return c.JSON(http.StatusUnauthorized, resp)
	}

	err := c.Bind(&req)
	if err != nil {
		log.Errorf("[UserHandler-2] UpdateCustomer: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	if err = c.Validate(&req); err != nil {
		log.Errorf("[UserHandler-3] UpdateCustomer: %v", err)
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	var hashedPassword string
	if req.Password != "" {
		if req.Password != req.PasswordConfirmation {
			log.Infof("[UserHandler-4] UpdateCustomer: password and confirm password do not match")
			resp.Message = "password and confirm password do not match"
			resp.Data = nil
			return c.JSON(http.StatusUnprocessableEntity, resp)
		}

		hashedPassword, err = conv.HashPassword(req.Password)
		if err != nil {
			log.Errorf("[UserHandler-5] UpdateCustomer: failed to hash password: %v", err)
			resp.Message = "internal server error"
			resp.Data = nil
			return c.JSON(http.StatusInternalServerError, resp)
		}
	}

	latString := ""
	lngString := ""

	if req.Lat != 0 {
		latString = strconv.FormatFloat(req.Lat, 'g', -1, 64)
	}

	if req.Lng != 0 {
		lngString = strconv.FormatFloat(req.Lng, 'g', -1, 64)
	}

	phoneString := fmt.Sprintf("%v", req.Phone)

	idParamStr := c.Param("id")
	if idParamStr == "" {
		log.Infof("[UserHandler-6] UpdateCustomer: missing or invalid customer ID")
		resp.Message = "missing or invalid customer ID"
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	id, err := conv.StringToInt(idParamStr)
	if err != nil {
		log.Infof("[UserHandler-7] UpdateCustomer: invalid customer ID")
		resp.Message = "invalid customer ID"
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	reqEntity := entity.UserEntity{
		ID:       id,
		Name:     req.Name,
		Email:    req.Email,
		Password: hashedPassword, // Will be empty if not provided, and ignored by the repository
		Phone:    phoneString,
		Address:  req.Address,
		Lat:      latString,
		Lng:      lngString,
		Photo:    req.Photo,
		RoleID:   req.RoleID,
	}

	err = u.UserService.UpdateCustomer(ctx, reqEntity)
	if err != nil {
		if strings.Contains(err.Error(), "violates unique constraint") {
			log.Warnf("[UserHandler-8] UpdateCustomer: Duplicate entry attempt: %v", err)
			resp.Message = "Email already registered."
			resp.Data = nil
			return c.JSON(http.StatusConflict, resp)
		}

		if err.Error() == "404" {
			log.Warnf("[UserHandler-9] UpdateCustomer: Customer not found: %v", err)
			resp.Message = "Customer not found"
			resp.Data = nil
			return c.JSON(http.StatusNotFound, resp)
		}

		log.Errorf("[UserHandler-10] UpdateCustomer: %v", err)
		resp.Message = "An internal server error occurred."
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	resp.Message = "success"
	resp.Data = nil

	return c.JSON(http.StatusOK, resp)
}

func (u *userHandler) DeleteCustomer(c echo.Context) error {
	var (
		resp = response.DefaultResponseWithPaginations{}
		ctx  = c.Request().Context()
	)

	user := c.Get("user").(string)
	if user == "" {
		log.Errorf("[UserHandler-1] DeleteCustomer: %s", "data token not found")
		resp.Message = "data token not valid"
		resp.Data = nil
		return c.JSON(http.StatusUnauthorized, resp)
	}

	idParamStr := c.Param("id")
	if idParamStr == "" {
		log.Infof("[UserHandler-2] DeleteCustomer: %s", "missing or invalid customer ID")
		resp.Message = "missing or invalid customer ID"
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	id, err := conv.StringToInt(idParamStr)
	if err != nil {
		log.Infof("[UserHandler-3] DeleteCustomer: %s", "invalid customer ID")
		resp.Message = "invalid customer ID"
		resp.Data = nil
		return c.JSON(http.StatusBadRequest, resp)
	}

	err = u.UserService.DeleteCustomer(ctx, id)
	if err != nil {
		log.Errorf("[UserHandler-4] DeleteCustomer: %v", err)
		if err.Error() == "404" {
			resp.Message = "Customer not found"
			resp.Data = nil
			return c.JSON(http.StatusNotFound, resp)
		}
		resp.Message = err.Error()
		resp.Data = nil
		return c.JSON(http.StatusInternalServerError, resp)
	}

	resp.Message = "success"
	resp.Data = nil

	return c.JSON(http.StatusOK, resp)

}

// SignIn implements IUserHandler.
func NewUserHandler(e *echo.Echo, userService service.IUserService, cfg *config.Config, jwtService service.IJWTService, redisClient *redis.Client) IUserHandler {
	userHandler := &userHandler{
		UserService: userService,
		RedisClient: redisClient,
	}

	e.Use(middleware.Recover())
	e.POST("/sign-in", userHandler.SignIn)
	e.POST("/signup", userHandler.CreateUserAccount)
	e.POST("/forgot-password", userHandler.ForgotPassword)
	e.GET("/verify-account", userHandler.VerifyAccount)
	e.PUT("/reset-password", userHandler.UpdatePassword)

	mid := adapter.NewMiddlewareAdapter(cfg, jwtService)
	adminGroup := e.Group("/admin", mid.CheckToken())
	adminGroup.GET("/customers", userHandler.GetCustomerAll)
	adminGroup.POST("/customers", userHandler.CreateCustomer)
	adminGroup.PUT("/customers/:id", userHandler.UpdateCustomer)
	adminGroup.GET("/customers/:id", userHandler.GetCustomerByID)
	adminGroup.DELETE("/customers/:id", userHandler.DeleteCustomer)
	adminGroup.GET("/check", func(c echo.Context) error {
		return c.String(200, "OK")
	})

	authGroup := e.Group("/auth", mid.CheckToken())
	authGroup.GET("/profile", userHandler.GetProfileUser)
	authGroup.PUT("/profile", userHandler.UpdateDataUser)

	return userHandler
}