package repository

import (
	"context"
	"errors"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/domain/models"

	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
)

type IUserRepository interface {
	GetUserByEmail(ctx context.Context, email string) (*entity.UserEntity, error)
	FindAll() ([]entity.UserEntity, error)
	FindById(id int) (entity.UserEntity, error)
	Save(user entity.UserEntity) (entity.UserEntity, error)
	Update(user entity.UserEntity) (entity.UserEntity, error)
	Delete(user entity.UserEntity) error
}

type UserRepository struct {
	db *gorm.DB
}


func (u *UserRepository) Delete(user entity.UserEntity) error {
	panic("unimplemented")
}


func (u *UserRepository) FindAll() ([]entity.UserEntity, error) {
 	panic("unimplemented")
}


func (u *UserRepository) FindById(id int) (entity.UserEntity, error) {
	panic("unimplemented")
}

// GetUserByEmail implements IUserRepository.
func (u *UserRepository) GetUserByEmail(ctx context.Context, email string) (*entity.UserEntity, error) {
	userMdl := models.User{}

	if err := u.db.Where("email = ? and is_verified = ?", email , true).Preload("Roles").First(&userMdl).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("404")
			log.Infof("[UserRepository-1] GetUserByEmail : %v", err)
			return nil, err
		}

		log.Errorf("[UserRepository-1] GetUserByEmail : %v", err)
		return nil, err
	}

	userE := entity.UserEntity{
		ID:        userMdl.ID,
		Name:      userMdl.Name,
		Email:     userMdl.Email,
		Password:  userMdl.Password,
		Phone:     userMdl.Phone,
		Photo:     userMdl.Photo,
		Address:   userMdl.Address,
		Lat:       userMdl.Lat,
		Lng:       userMdl.Lng,
		IsVerified: userMdl.IsVerified,
		RoleName:  userMdl.Roles[0].Name,
	}

	return &userE, nil
}

// Save implements IUserRepository.
func (u *UserRepository) Save(user entity.UserEntity) (entity.UserEntity, error) {
	panic("unimplemented")
}

// Update implements IUserRepository.
func (u *UserRepository) Update(user entity.UserEntity) (entity.UserEntity, error) {
	panic("unimplemented")
}

func NewUserRepository(db *gorm.DB) IUserRepository {
	return &UserRepository{
		db: db,
	}
}
