package repository

import (
	"context"
	"errors"
	"user-service/internal/core/domain/entity"
	"user-service/internal/core/domain/models"

	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
)

type IRoleRepository interface {
	GetAllRole(ctx context.Context, search string) ([]entity.RoleEntity, error)
	GetRoleByID(ctx context.Context, id int) (*entity.RoleEntity, error)
	CreateRole(ctx context.Context, req entity.RoleEntity) error
	UpdateRole(ctx context.Context, req entity.RoleEntity) error
	DeleteRole(ctx context.Context, id int) error
}

type RoleRepository struct {
	db *gorm.DB
}

// CreateRole implements IRoleRepository.
func (r *RoleRepository) CreateRole(ctx context.Context, req entity.RoleEntity) error {
	roleMdl := models.Role{
		Name: req.Name,
	}

	if err := r.db.WithContext(ctx).Create(&roleMdl).Error; err != nil {
		log.Errorf("[RoleRepository-1] Create: %v", err)
		return err
	}

	return nil
}

// DeleteRole implements IRoleRepository.
func (r *RoleRepository) DeleteRole(ctx context.Context, id int) error {
	roleMdl := models.Role{}

	//find

	if err := r.db.WithContext(ctx).Where("id = ?", id).Preload("Users").First(&roleMdl).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("404")
			log.Errorf("[RoleRepository-1] Delete: %v", err)
			return err
		}
		log.Errorf("[RoleRepository-2] Delete: %v", err)
		return err
	}

	//check users
	if len(roleMdl.Users) > 0 {
		err := errors.New("role has user")
		log.Errorf("[RoleRepository-3] Delete: %v", err)
		return err
	}

	if err := r.db.WithContext(ctx).Delete(&roleMdl).Error; err != nil {
		log.Errorf("[RoleRepository-4] Delete: %v", err)
		return err
	}

	return nil
}

// GetAllRole implements IRoleRepository.
func (r *RoleRepository) GetAllRole(ctx context.Context, search string) ([]entity.RoleEntity, error) {
	roleMdl := []models.Role{}

	if err := r.db.Where("name ILIKE ?", "%"+search+"%").Find(&roleMdl).Error; err != nil {
		log.Errorf("[RoleRepository-1] GetAll: %v", err)
		return nil, err
	}

	if len(roleMdl) == 0 {
		err := errors.New("404")
		log.Infof("[RoleRepository-2] GetAll: No role found")
		return nil, err
	}

	roleEntity := []entity.RoleEntity{}
	for _, v := range roleMdl {
		roleEntity = append(roleEntity, entity.RoleEntity{
			ID:    v.ID,
			Name:  v.Name,
		})
	}

	return roleEntity, nil

}

// GetRoleByID implements IRoleRepository.
func (r *RoleRepository) GetRoleByID(ctx context.Context, id int) (*entity.RoleEntity, error) {
	roleMdl := models.Role{}

	if err := r.db.WithContext(ctx).Where("id = ?", id).First(&roleMdl).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("404")
			log.Errorf("[RoleRepository-1] GetByID: %v", err)
			return nil, err
		}
		log.Errorf("[RoleRepository-2] GetByID: %v", err)
		return nil, err
	}

	return &entity.RoleEntity{
		ID:   roleMdl.ID,
		Name: roleMdl.Name,
	}, nil
}

// UpdateRole implements IRoleRepository.
func (r *RoleRepository) UpdateRole(ctx context.Context, req entity.RoleEntity) error {
	roleMdl := models.Role{
		Name: req.Name,
	}

	if err := r.db.WithContext(ctx).Model(&models.Role{}).Where("id = ?", req.ID).Updates(roleMdl).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("404")
			log.Errorf("[RoleRepository-1] Update: %v", err)
			return err
		}
		log.Errorf("[RoleRepository-2] Update: %v", err)
		return err
	}

	return nil
}

func NewRoleRepository(db *gorm.DB) IRoleRepository {
	return &RoleRepository{
		db: db,
	}
}
