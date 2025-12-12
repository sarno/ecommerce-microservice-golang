package repository

import (
	"context"
	"errors"
	"time"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/domain/models"

	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
)

type IUserRepository interface {
	GetUserByEmail(ctx context.Context, email string) (*entity.UserEntity, error)
	CreateUserAccount(ctx context.Context, req entity.UserEntity) (int, error)
	UpdateUserVerified(ctx context.Context, userID int) (*entity.UserEntity, error)
}

type UserRepository struct {
	db *gorm.DB
}

// UpdateUserVerified implements IUserRepository.
func (u *UserRepository) UpdateUserVerified(ctx context.Context, userID int) (*entity.UserEntity, error) {
	modelUser := models.User{}

	if err := u.db.Where("id = ?", userID).Preload("Roles").First(&modelUser).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("404")
			log.Errorf("[UserRepository-1] UpdateUserVerified: %v", err)
			return nil, err
		}
		log.Errorf("[UserRepository-2] UpdateUserVerified: %v", err)
		return nil, err
	}

	modelUser.IsVerified = true

	if err := u.db.Save(&modelUser).Error; err != nil {
		log.Errorf("[UserRepository-3] UpdateUserVerified: %v", err)
		return nil, err
	}

	return &entity.UserEntity{
		ID:         userID,
		Name:       modelUser.Name,
		Email:      modelUser.Email,
		RoleName:   modelUser.Roles[0].Name,
		Address:    modelUser.Address,
		Lat:        modelUser.Lat,
		Lng:        modelUser.Lng,
		Phone:      modelUser.Phone,
		Photo:      modelUser.Photo,
		IsVerified: modelUser.IsVerified,
	}, nil
}

// CreateUserAccount implements IUserRepository.
func (u *UserRepository) CreateUserAccount(ctx context.Context, req entity.UserEntity) (int, error) {
	var roleId int

	if err := u.db.Select("id").
		Where("name = ?", "user").
		Model(&models.Role{}).
		Scan(&roleId).
		Error; err != nil {
		log.Errorf("[UserRepository-1] CreateUserAccount : %v", err)
		return 0, err
	}

	userMdl := models.User{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
		Roles:    []models.Role{{ID: roleId}},
	}

	if err := u.db.Create(&userMdl).Error; err != nil {
		log.Errorf("[UserRepository-2] CreateUserAccount : %v", err)
		return 0, err
	}

	verifyMdl := models.VerificationToken{
		UserID:    userMdl.ID,
		Token:     req.Token,
		TokenType: "email_verification",
		ExpiresAt: time.Now().Add(time.Hour),
	}

	if err := u.db.Create(&verifyMdl).Error; err != nil {
		log.Errorf("[UserRepository-3] CreateUserAccount : %v", err)
		return 0, err
	}

	return userMdl.ID, nil

}

func (u *UserRepository) GetUserByEmail(ctx context.Context, email string) (*entity.UserEntity, error) {
	userMdl := models.User{}

	if err := u.db.Where("email = ? and is_verified = ?", email, true).Preload("Roles").First(&userMdl).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("404")
			log.Infof("[UserRepository-1] GetUserByEmail : %v", err)
			return nil, err
		}

		log.Errorf("[UserRepository-1] GetUserByEmail : %v", err)
		return nil, err
	}

	userE := entity.UserEntity{
		ID:         userMdl.ID,
		Name:       userMdl.Name,
		Email:      userMdl.Email,
		Password:   userMdl.Password,
		Phone:      userMdl.Phone,
		Photo:      userMdl.Photo,
		Address:    userMdl.Address,
		Lat:        userMdl.Lat,
		Lng:        userMdl.Lng,
		IsVerified: userMdl.IsVerified,
		RoleName:   userMdl.Roles[0].Name,
	}

	return &userE, nil
}

func NewUserRepository(db *gorm.DB) IUserRepository {
	return &UserRepository{
		db: db,
	}
}
