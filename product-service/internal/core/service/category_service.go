package service

import (
	"context"
	"errors"
	"product-service/internal/adapter/repository"
	"product-service/internal/core/domain/entities"
	"product-service/internal/utils/conv"

	"github.com/labstack/gommon/log"
)

// interface

type ICategoryService interface {
	GetAll(ctx context.Context, queryString entities.QueryStringEntity) ([]entities.CategoryEntity, int64, int64, error)
	GetCategoryByID(ctx context.Context, id int64) (*entities.CategoryEntity, error)
	GetCategoryBySlug(ctx context.Context, slug string) (*entities.CategoryEntity, error)
	CreateCategory(ctx context.Context, req entities.CategoryEntity) error
	UpdateCategory(ctx context.Context, req entities.CategoryEntity) error
	DeleteCategory(ctx context.Context, id int64) error

	GetAllPublished(ctx context.Context) ([]entities.CategoryEntity, error)
}

//struct

type categoryService struct {
	repo repository.ICategoryRepository
}

// GetAllPublished implements [ICategoryService].
func (c *categoryService) GetAllPublished(ctx context.Context) ([]entities.CategoryEntity, error) {
	return c.repo.GetAllPublished(ctx)
}

// CreateCategory implements [ICategoryService].
func (c *categoryService) CreateCategory(ctx context.Context, req entities.CategoryEntity) error {
	slug := conv.GenerateSlug(req.Name)
	// GetBySlug
	result, err := c.repo.GetCategoryBySlug(ctx, slug)
	if err != nil {
		if err.Error() == "404" {
			log.Errorf("[CategoryService-1] CreateCategory: %v", err)
			req.Slug = slug
			return c.repo.CreateCategory(ctx, req)
		}
		return err
	}

	if result != nil {
		err := errors.New("409")
		log.Errorf("[CategoryService-2] CreateCategory: %v", err)
		return err
	}

	req.Slug = slug

	err = c.repo.CreateCategory(ctx, req)
	if err != nil {
		log.Errorf("[CategoryService-3] CreateCategory: %v", err)
		return err
	}
	return nil
}

// DeleteCategory implements [ICategoryService].
func (c *categoryService) DeleteCategory(ctx context.Context, id int64) error {
	return c.repo.DeleteCategory(ctx, id)
}

// GetAll implements [ICategoryService].
func (c *categoryService) GetAll(ctx context.Context, queryString entities.QueryStringEntity) ([]entities.CategoryEntity, int64, int64, error) {
	return c.repo.GetAll(ctx, queryString)
}

// GetCategoryByID implements [ICategoryService].
func (c *categoryService) GetCategoryByID(ctx context.Context, id int64) (*entities.CategoryEntity, error) {
	return c.repo.GetCategoryByID(ctx, id)
}

func (c *categoryService) GetCategoryBySlug(ctx context.Context, slug string) (*entities.CategoryEntity, error) {
	return c.repo.GetCategoryBySlug(ctx, slug)
}

// UpdateCategory implements [ICategoryService].
func (c *categoryService) UpdateCategory(ctx context.Context, req entities.CategoryEntity) error {
	slug := conv.GenerateSlug(req.Name)
	result, err := c.repo.GetCategoryByID(ctx, req.ID)
	if err != nil {
		log.Errorf("[CategoryService-1] EditCategory : %v", err)
		return err
	}

	if slug != result.Slug {
		resSlug, err := c.repo.GetCategoryBySlug(ctx, slug)
		if err != nil && err.Error() != "404" {
			log.Errorf("[CategoryService-2] EditCategory : %v", err)
			return err
		}

		if resSlug != nil {
			err := errors.New("409")
			log.Infof("[CategoryService-3] EditCategory: Category already exists")
			return err
		}
	}

	req.Slug = slug
	err = c.repo.UpdateCategory(ctx, req)
	if err != nil {
		log.Errorf("[CategoryService-4] EditCategory : %v", err)
		return err
	}

	return nil

}

func NewCategoryService(repo repository.ICategoryRepository) ICategoryService {
	return &categoryService{
		repo: repo,
	}
}
