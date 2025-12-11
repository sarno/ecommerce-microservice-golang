package service

import (
	"context"
	"encoding/json"
	"errors"
	"time"
	"user-service/config"
	"user-service/internal/adapter/repository"
	"user-service/internal/core/domain/entity"
	"user-service/utils/conv"

	"github.com/labstack/gommon/log"
)

type IUserService interface {
	SignIn(ctx context.Context, req entity.UserEntity) (*entity.UserEntity, string, error)
}

type UserService struct {
	repo repository.IUserRepository
	cfg  *config.Config
	jwtService IJWTService
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

	redisConn := config.NewConfig().NewRedisClient()
	err = redisConn.Set(ctx, token, jsonData, time.Hour*23).Err()

	if err != nil {
		log.Errorf("[UserService-4] SignIn: %v", err)
		return nil, "",  err
	}

	return user, token, nil
}

func NewUserService(repo repository.IUserRepository, cfg *config.Config, jwtService IJWTService ) IUserService {
	return &UserService{
		repo: repo,
		cfg:  cfg,
		jwtService: jwtService,
	}
}
