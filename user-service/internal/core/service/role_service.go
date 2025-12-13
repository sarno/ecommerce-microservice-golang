package service

import (
	"context"
	"user-service/internal/adapter/repository"
	"user-service/internal/core/domain/entity"
)

type IRoleService interface {
	CreateRole(ctx context.Context, req entity.RoleEntity) error
	GetRoleByID(ctx context.Context, id int) (*entity.RoleEntity, error)
	GetAllRole(ctx context.Context, search string) ([]entity.RoleEntity, error)
	UpdateRole(ctx context.Context, req entity.RoleEntity) error
	DeleteRole(ctx context.Context, id int) error
}

type RoleService struct {
	repo repository.IRoleRepository
}

// CreateRole implements IRoleService.
func (r *RoleService) CreateRole(ctx context.Context, req entity.RoleEntity) error {
	return r.repo.CreateRole(ctx, req)
}

// DeleteRole implements IRoleService.
func (r *RoleService) DeleteRole(ctx context.Context, id int) error {
	return r.repo.DeleteRole(ctx, id)
}

// GetAllRole implements IRoleService.
func (r *RoleService) GetAllRole(ctx context.Context, search string) ([]entity.RoleEntity, error) {
	return r.repo.GetAllRole(ctx,  search)
}

// GetRoleByID implements IRoleService.
func (r *RoleService) GetRoleByID(ctx context.Context, id int) (*entity.RoleEntity, error) {
	return r.repo.GetRoleByID(ctx, id)
}

// UpdateRole implements IRoleService.
func (r *RoleService) UpdateRole(ctx context.Context, req entity.RoleEntity) error {
	return r.repo.UpdateRole(ctx, req)
}

func NewRoleService(repo repository.IRoleRepository) IRoleService {
	return &RoleService{
		repo: repo,
	}
}
