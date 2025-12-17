package repository

import (
	"context"
	"errors"
	"fmt"
	"math"
	"product-service/internal/core/domain/entities"
	"product-service/internal/core/domain/models"

	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
)

// interface
type ICategoryRepository interface {
	GetAll(ctx context.Context, queryString entities.QueryStringEntity) ([]entities.CategoryEntity, int64, int64, error)
	GetCategoryByID(ctx context.Context, id int64) (*entities.CategoryEntity, error)
	GetCategoryBySlug(ctx context.Context, slug string) (*entities.CategoryEntity, error)
	CreateCategory(ctx context.Context, req entities.CategoryEntity) error
	UpdateCategory(ctx context.Context, req entities.CategoryEntity) error
	DeleteCategory(ctx context.Context, id int64) error
}

type categoryRepository struct {
	db *gorm.DB
}

// CreateCategory implements [ICategoryRepository].
func (c *categoryRepository) CreateCategory(ctx context.Context, req entities.CategoryEntity) error {
	status  := true
	if req.Status == entities.UnpublishedStatus {
		status = false	
	}

	categoryMdl := models.Category{
		ParentID:    req.ParentID,
		Name:        req.Name,
		Icon:        req.Icon,
		Status:      status,
		Slug:        req.Slug,
		Description: req.Description,
	}

	if req.ParentID != nil {
		categoryMdl.ParentID = req.ParentID
	}

	if err := c.db.Create(&categoryMdl).Error; err != nil {
		log.Errorf("[CategoryRepository-1] CreateCategory: %v", err)
		return err
	}

	return nil
}

// DeleteCategory implements [ICategoryRepository].
func (c *categoryRepository) DeleteCategory(ctx context.Context, id int64) error {
	categoryMdl := models.Category{}

	// find product by category slug
	if err := c.db.WithContext(ctx).Preload("Products").Where("id = ?", id).First(&categoryMdl).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("404")
			log.Errorf("[CategoryRepository-1] DeleteCategory: %v", err)
			return err
		}
		log.Errorf("[CategoryRepository-2] DeleteCategory: %v", err)
		return err
	}

	if len(categoryMdl.Products) > 0 {
		err := errors.New("304")
		log.Errorf("[CategoryRepository-3] DeleteCategory: %v", "category has products")
		return err
	}

	if err := c.db.WithContext(ctx).Delete(&categoryMdl).Error; err != nil {
		log.Errorf("[CategoryRepository-1] DeleteCategory: %v", err)
		return err
	}

	return nil
}

// GetAllCategory implements [ICategoryRepository].
func (c *categoryRepository) GetAll(ctx context.Context, query entities.QueryStringEntity) ([]entities.CategoryEntity, int64, int64, error) {
	modelCategories := []models.Category{}
	var countData int64

	order := fmt.Sprintf("%s %s", query.OrderBy, query.OrderType)
	offset := (query.Page - 1) * query.Limit

	sqlMain := c.db.Preload("Products").
		Where("name ILIKE ? OR slug ILIKE ?", "%"+query.Search+"%", "%"+query.Search+"%")
	if err := sqlMain.Model(&modelCategories).Count(&countData).Error; err != nil {
		log.Errorf("[CategoryRepository-1] GetAll: %v", err)
		return nil, 0, 0, err
	}

	totalPage := int(math.Ceil(float64(countData) / float64(query.Limit)))
	if err := sqlMain.Order(order).Limit(int(query.Limit)).Offset(int(offset)).Find(&modelCategories).Error; err != nil {
		log.Errorf("[CategoryRepository-2] GetAll: %v", err)
		return nil, 0, 0, err
	}

	if len(modelCategories) == 0 {
		err := errors.New("404")
		log.Infof("[CategoryRepository-3] GetAll: No category found")
		return nil, 0, 0, err
	}

	categoryEntities := []entities.CategoryEntity{}
	for _, val := range modelCategories {
		productEntities := []entities.ProductEntity{}
		for _, prd := range val.Products {
			productEntities = append(productEntities, entities.ProductEntity{
				ID:           prd.ID,
				CategorySlug: val.Slug,
				ParentID:     prd.ParentID,
				Name:         prd.Name,
				Image:        prd.Image,
			})
		}
		
		status := entities.PublishedStatus
		if !val.Status {
			status = entities.UnpublishedStatus
		}

		categoryEntities = append(categoryEntities, entities.CategoryEntity{
			ID:          val.ID,
			ParentID:    val.ParentID,
			Name:        val.Name,
			Icon:        val.Icon,
			Status:      status,
			Slug:        val.Slug,
			Description: val.Description,
			Products:    productEntities,
		})
	}

	return categoryEntities, countData, int64(totalPage), nil
}

// GetCategoryByID implements [ICategoryRepository].
func (c *categoryRepository) GetCategoryByID(ctx context.Context, id int64) (*entities.CategoryEntity, error) {
	cetegoryMdl := models.Category{}

	err := c.db.WithContext(ctx).Where("id = ?", id).First(&cetegoryMdl).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("404")
			log.Errorf("[CategoryRepository-1] GetCategoryByID: %v", err)
			return nil, err
		}
		log.Errorf("[CategoryRepository-2] GetCategoryByID: %v", err)
		return nil, err
	}

	status := entities.PublishedStatus
	if !cetegoryMdl.Status {
		status = entities.UnpublishedStatus
	}

	return &entities.CategoryEntity{
		ID:          cetegoryMdl.ID,
		ParentID:    cetegoryMdl.ParentID,
		Name:        cetegoryMdl.Name,
		Icon:        cetegoryMdl.Icon,
		Status:      status,
		Slug:        cetegoryMdl.Slug,
		Description: cetegoryMdl.Description,
	}, nil
}

func (c *categoryRepository) GetCategoryBySlug(ctx context.Context, slug string) (*entities.CategoryEntity, error) {
	cetegoryMdl := models.Category{}

	if err := c.db.Where("slug =?", slug).First(&cetegoryMdl).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("404")
			log.Infof("[CategoryRepository-1] GetBySlug: Category not found")
			return nil, err
		}
		log.Errorf("[CategoryRepository-2] GetBySlug: %v", err)
		return nil, err
	}

	status := entities.PublishedStatus
	if !cetegoryMdl.Status {
		status = entities.UnpublishedStatus
	}

	return &entities.CategoryEntity{
		ID:          cetegoryMdl.ID,
		ParentID:    cetegoryMdl.ParentID,
		Name:        cetegoryMdl.Name,
		Icon:        cetegoryMdl.Icon,
		Status:      status,
		Slug:        cetegoryMdl.Slug,
		Description: cetegoryMdl.Description,
	}, nil
}

// UpdateCategory implements [ICategoryRepository].
func (c *categoryRepository) UpdateCategory(ctx context.Context, req entities.CategoryEntity) error {
	categoryMdl := models.Category{}

	err := c.db.WithContext(ctx).Where("id = ?", req.ID).First(&categoryMdl).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("404")
			log.Errorf("[CategoryRepository-1] UpdateCategory: %v", err)
			return err
		}
		log.Errorf("[CategoryRepository-2] UpdateCategory: %v", err)
		return err
	}

	status  := true
	if req.Status == entities.UnpublishedStatus {
		status = false	
	}

	categoryMdl.ParentID = req.ParentID
	categoryMdl.Name = req.Name
	categoryMdl.Icon = req.Icon
	categoryMdl.Status = status
	categoryMdl.Slug = req.Slug
	categoryMdl.Description = req.Description

	err = c.db.WithContext(ctx).Save(&categoryMdl).Error
	if err != nil {
		log.Errorf("[CategoryRepository-3] UpdateCategory: %v", err)
		return err
	}

	return nil
}

func NewCategoryRepository(db *gorm.DB) ICategoryRepository {
	return &categoryRepository{
		db: db,
	}
}
