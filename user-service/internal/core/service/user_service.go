package service

import (
	"context"
	"errors"
	"log"
	"user-service/internal/adapter/repository"
	"user-service/internal/core/domain/entity"
	"user-service/utils/conv"
)

type IUserService interface {
	SignIn(ctx context.Context, req entity.UserEntity) (*entity.UserEntity, string, error)
}

type UserService struct {
	repo repository.IUserRepository
}

// SignIn implements IUserService.
func (u *UserService) SignIn(ctx context.Context, req entity.UserEntity) (*entity.UserEntity, string, error) {
	user, err := u.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, "", err
	}

	if !conv.CheckPasswordHash(req.Password, user.Password) {
		err := errors.New("password incorrect")
		log.Println(err)
		return nil, "", err
	}

	return user, "", nil
}

func NewUserService(repo repository.IUserRepository) IUserService {
	return &UserService{
		repo: repo,
	}
}
