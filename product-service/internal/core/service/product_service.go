package service

import (
	"context"
	"errors"
	"product-service/internal/adapter/message"
	"product-service/internal/adapter/repository"
	"product-service/internal/core/domain/entities"

	"github.com/labstack/gommon/log"
)

// interface
type IProductService interface {
	GetAll(ctx context.Context, query entities.QueryStringProduct) ([]entities.ProductEntity, int64, int64, error)
	GetByID(ctx context.Context, productID int64) (*entities.ProductEntity, error)
	Create(ctx context.Context, req entities.ProductEntity) error
	Update(ctx context.Context, req entities.ProductEntity) error
	Delete(ctx context.Context, productID int64) error
}

// struct
type productService struct {
	repo              repository.IProductRepository
	repoCat           repository.ICategoryRepository
	publisherRabbitMQ message.IPublishRabbitMQ
}

// Delete implements [IProductService].
func (p *productService) Delete(ctx context.Context, productID int64) error {
	err := p.repo.Delete(ctx, productID)
	if err != nil {
		log.Errorf("[ProductService-1] Delete: %v", err)
		return err
	}
	
	if err := p.publisherRabbitMQ.DeleteProductFromQueue(productID); err != nil {
		log.Errorf("[ProductService-2] Delete: %v", err)
	}

	return nil
	
}

// Update implements [IProductService].
func (p *productService) Update(ctx context.Context, req entities.ProductEntity) error {
	err := p.repo.Update(ctx, req)
	if err != nil {
		log.Errorf("[ProductService-1] Update: %v", err)
		return err
	}

	getProductByID, err := p.GetByID(ctx, req.ID)
	if err != nil {
		log.Errorf("[ProductService-2] Update: %v", err)
	}

	if err := p.publisherRabbitMQ.PublishProductToQueue(*getProductByID); err != nil {
		log.Errorf("[ProductService-3] Update: %v", err)
	}

	return nil
}

// GetByID implements [IProductService].
func (p *productService) GetByID(ctx context.Context, productID int64) (*entities.ProductEntity, error) {
	result, err := p.repo.GetByID(ctx, productID)
	if err != nil {
		log.Errorf("[ProductService-1] GetByID: %v", err)
		return nil, err
	}

	resultCat, err := p.repoCat.GetCategoryBySlug(ctx, result.CategorySlug)
	if err != nil {
		log.Errorf("[ProductService-2] GetByID: %v", err)
		return nil, err
	}

	if resultCat == nil {
		return nil, errors.New("category not found")
	}
	result.CategoryName = resultCat.Name
	return result, nil
}

// Create implements [IProductService].
func (p *productService) Create(ctx context.Context, req entities.ProductEntity) error {
	productID, err := p.repo.Create(ctx, req)
	if err != nil {
		log.Errorf("[ProductService-1] Create: %v", err)
		return err
	}

	getProductByID, err := p.GetByID(ctx, productID)
	if err != nil {
		log.Errorf("[ProductService-2] Create: %v", err)
	}

	if err := p.publisherRabbitMQ.PublishProductToQueue(*getProductByID); err != nil {
		log.Errorf("[ProductService-3] Create: %v", err)
	}

	return nil
}

// GetAll implements [IProductService].
func (p *productService) GetAll(ctx context.Context, query entities.QueryStringProduct) ([]entities.ProductEntity, int64, int64, error) {
	return p.repo.GetAll(ctx, query)
}

func NewProductService(repo repository.IProductRepository, repoCat repository.ICategoryRepository, publisherRabbitMQ message.IPublishRabbitMQ) IProductService {
	return &productService{
		repo:              repo,
		repoCat:           repoCat,
		publisherRabbitMQ: publisherRabbitMQ,
	}
}
