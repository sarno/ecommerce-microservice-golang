package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"
	"user-service/config"
	"user-service/internal/adapter/message"
	"user-service/internal/adapter/repository"
	"user-service/internal/core/domain/entity"
	"user-service/utils"
	"user-service/utils/conv"

	"github.com/go-redis/redis/v8"
	"github.com/google/uuid"
	"github.com/labstack/gommon/log"
)

type IUserService interface {
	SignIn(ctx context.Context, req entity.UserEntity) (*entity.UserEntity, string, error)
	CreateUserAccount(ctx context.Context, req entity.UserEntity) error
	ForgotPassword(ctx context.Context, req entity.UserEntity) error
	VerifyToken(ctx context.Context, token string) (*entity.UserEntity, error)
	UpdatePassword(ctx context.Context, req entity.UserEntity) error
	GetProfileUser(ctx context.Context, userID int) (*entity.UserEntity, error)
	UpdateDataUser(ctx context.Context, req entity.UserEntity) error

	// modul user
	GetCustomerAll(ctx context.Context, query entity.QueryStringCustomer) ([]entity.UserEntity, int, int, error)
	GetCustomerByID(ctx context.Context, customerID int) (*entity.UserEntity, error)
	CreateCustomer(ctx context.Context, req entity.UserEntity) error
	UpdateCustomer(ctx context.Context, req entity.UserEntity) error
	DeleteCustomer(ctx context.Context, customerID int) error
	GetUsersByIDs(ctx context.Context, userIDs []int) ([]entity.UserEntity, error)
}

type UserService struct {
	repo        repository.IUserRepository
	cfg         *config.Config
	jwtService  IJWTService
	repoToken   repository.IVerificationTokenRepository
	redisClient *redis.Client
}

// GetUsersByIDs implements [IUserService].
func (u *UserService) GetUsersByIDs(ctx context.Context, userIDs []int) ([]entity.UserEntity, error) {
	return u.repo.GetUsersByIDs(ctx, userIDs)
}

// CreateCustomer implements IUserService.
func (u *UserService) CreateCustomer(ctx context.Context, req entity.UserEntity) error {
	userID, err := u.repo.CreateCustomer(ctx, req)
	if err != nil {
		log.Errorf("[UserService-2] CreateCustomer: %v", err)
		return err
	}

	messageparam := "You have been registered in Sayur Project. Please login with the email and password you provided."
	go message.PublishMessage(userID,
		req.Email,
		messageparam,
		utils.NOTIF_EMAIL_CREATE_CUSTOMER,
		"Account Exists")

	return nil
}

// DeleteCustomer implements IUserService.
func (u *UserService) DeleteCustomer(ctx context.Context, customerID int) error {
	return u.repo.DeleteCustomer(ctx, customerID)
}

// UpdateCustomer implements IUserService.
func (u *UserService) UpdateCustomer(ctx context.Context, req entity.UserEntity) error {
	// The password from the handler is already hashed.
	// If it's empty, the repository will ignore it.
	// If it's not empty, we pass the hash directly.
	err := u.repo.UpdateCustomer(ctx, req)
	if err != nil {
		log.Errorf("[UserService-2] UpdateCustomer: %v", err)
		return err
	}

	if req.Password != "" {
		messageparam := "Your account password has been updated. Please login using your new password."
		go message.PublishMessage(req.ID,
			req.Email,
			messageparam,
			utils.NOTIF_EMAIL_UPDATE_CUSTOMER,
			"Password Updated")
	}

	return nil
}

// GetCustomerByID implements IUserService.
func (u *UserService) GetCustomerByID(ctx context.Context, customerID int) (*entity.UserEntity, error) {
	return u.repo.GetCustomerByID(ctx, customerID)
}

// GetCustomerAll implements IUserService.
func (u *UserService) GetCustomerAll(ctx context.Context, query entity.QueryStringCustomer) ([]entity.UserEntity, int, int, error) {
	return u.repo.GetCustomerAll(ctx, query)
}

// UpdateDataUser implements IUserService.
func (u *UserService) UpdateDataUser(ctx context.Context, req entity.UserEntity) error {
	return u.repo.UpdateDataUser(ctx, req)
}

// GetProfileUser implements IUserService.
func (u *UserService) GetProfileUser(ctx context.Context, userID int) (*entity.UserEntity, error) {
	return u.repo.GetUserByID(ctx, userID)
}

// UpdatePassword implements IUserService.
func (u *UserService) UpdatePassword(ctx context.Context, req entity.UserEntity) error {
	token, err := u.repoToken.GetDataByToken(ctx, req.Token)

	if err != nil {
		log.Errorf("[UserService-1] UpdatePassword: %v", err)
		return err
	}

	if token.TokenType != utils.NOTIF_EMAIL_FORGOT_PASSWORD {
		err = errors.New("invalid token type")
		log.Errorf("[UserService-2] UpdatePassword: %v", err)
		return err
	}

	password, err := conv.HashPassword(req.Password)
	if err != nil {
		log.Errorf("[UserService-3] UpdatePassword: %v", err)
		return err
	}

	req.Password = password
	req.ID = token.UserID

	if err := u.repo.UpdatePasswordByID(ctx, req); err != nil {
		log.Errorf("[UserService-4] UpdatePassword: %v", err)
		return err
	}

	return nil
}

// VerifyToken implements IUserService.
func (u *UserService) VerifyToken(ctx context.Context, token string) (*entity.UserEntity, error) {
	verifyToken, err := u.repoToken.GetDataByToken(ctx, token)

	if err != nil {
		log.Errorf("[UserService-1] VerifyToken: %v", err)
		return nil, err
	}

	user, err := u.repo.UpdateUserVerified(ctx, verifyToken.UserID)
	if err != nil {
		log.Errorf("[UserService-2] VerifyToken: %v", err)
		return nil, err
	}

	accessToken, err := u.jwtService.GenerateToken(user.ID)
	if err != nil {
		log.Errorf("[UserService-3] VerifyToken: %v", err)
		return nil, err
	}

	sessionData := map[string]interface{}{
		"user_id":    user.ID,
		"name":       user.Name,
		"email":      user.Email,
		"logged_in":  true,
		"created_at": time.Now().String(),
		"token":      token,
		"role_name":  user.RoleName,
	}

	jsonData, err := json.Marshal(sessionData)

	if err != nil {
		fmt.Println("Error encoding JSON:", err)
		return nil, err
	}

	err = u.redisClient.Set(ctx, token, jsonData, time.Hour*23).Err()
	if err != nil {
		log.Errorf("[UserService-4] VerifyToken: %v", err)
		return nil, err
	}

	user.Token = accessToken

	return user, nil
}

// ForgotPassword implements IUserService.
func (u *UserService) ForgotPassword(ctx context.Context, req entity.UserEntity) error {
	user, err := u.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		log.Errorf("[UserService-1] ForgotPassword: %v", err)
		return err
	}

	if user == nil {
		err = errors.New("user not found")
		log.Errorf("[UserService-2] ForgotPassword: %v", err)
		return err
	}

	token := uuid.New().String()
	reqEntity := entity.VerificationTokenEntity{
		UserID:    user.ID,
		Token:     token,
		TokenType: utils.NOTIF_EMAIL_FORGOT_PASSWORD,
	}

	err = u.repoToken.CreateVerificationToken(ctx, reqEntity)
	if err != nil {
		log.Errorf("[UserService-3] ForgotPassword: %v", err)
		return err
	}

	urlForgot := fmt.Sprintf("%s/auth/reset-password?token=%s", u.cfg.App.UrlFrontFE, token)
	forgotMessage := fmt.Sprintf("Please reset your password by clicking the link below: %s", urlForgot)

	go message.PublishMessage(
		user.ID,
		req.Email,
		forgotMessage,
		utils.NOTIF_EMAIL_FORGOT_PASSWORD,
		"Reset Your Password",
	)

	return nil
}

// CreateUserAccount implements IUserService.
func (u *UserService) CreateUserAccount(ctx context.Context, req entity.UserEntity) error {
	password, err := conv.HashPassword(req.Password)

	if err != nil {
		log.Errorf("[UserService-1] CreateUserAccount: %v", err)
		return err
	}

	req.Password = password
	req.Token = uuid.New().String()

	userId, err := u.repo.CreateUserAccount(ctx, req)
	if err != nil {
		log.Errorf("[UserService-2] CreateUserAccount: %v", err)
		return err
	}

	verifyURL := fmt.Sprintf("%s/auth/verify-account?token=%s", u.cfg.App.UrlFrontFE, req.Token)
	veryfyMessage := fmt.Sprintf("Please verify your account by clicking the link below: %s", verifyURL)

	go message.PublishMessage(
		userId,
		req.Email,
		veryfyMessage,
		utils.NOTIF_EMAIL_VERIFICATION,
		"Verify Your Account",
	)

	return nil
}

// SignIn implements IUserService.
func (u *UserService) SignIn(ctx context.Context, req entity.UserEntity) (*entity.UserEntity, string, error) {
	user, err := u.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, "", err
	}

	if !conv.CheckPasswordHash(req.Password, user.Password) {
		err := errors.New("password incorrect")
		log.Errorf("[UserService-2] SignIn: %v", err)
		return nil, "", err
	}

	token, err := u.jwtService.GenerateToken(user.ID)
	if err != nil {
		return nil, "", err
	}

	sessionData := map[string]interface{}{
		"user_id":    user.ID,
		"name":       user.Name,
		"email":      user.Email,
		"logged_in":  true,
		"created_at": time.Now().String(),
		"token":      token,
		"role_name":  user.RoleName,
	}

	jsonData, err := json.Marshal(sessionData)
	if err != nil {
		return nil, "", err
	}

	err = u.redisClient.Set(ctx, token, jsonData, time.Hour*23).Err()

	if err != nil {
		log.Errorf("[UserService-4] SignIn: %v", err)
		return nil, "", err
	}

	return user, token, nil
}

func NewUserService(repo repository.IUserRepository, cfg *config.Config, jwtService IJWTService, repoToken repository.IVerificationTokenRepository, redisClient *redis.Client) IUserService {
	return &UserService{
		repo:        repo,
		cfg:         cfg,
		jwtService:  jwtService,
		repoToken:   repoToken,
		redisClient: redisClient,
	}
}
