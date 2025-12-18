package repository

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"product-service/internal/core/domain/entities"
	"product-service/internal/core/domain/models"
	"strconv"
	"strings"

	"github.com/elastic/go-elasticsearch/v7"
	"github.com/labstack/gommon/log"
	"gorm.io/gorm"
)

// interface

type IProductRepository interface {
	GetAll(ctx context.Context, query entities.QueryStringProduct) ([]entities.ProductEntity, int64, int64, error)
	GetByID(ctx context.Context, productID int64) (*entities.ProductEntity, error)
	Create(ctx context.Context, req entities.ProductEntity) (int64, error)
	Update(ctx context.Context, req entities.ProductEntity) error
	Delete(ctx context.Context, productID int64) error

	SearchProducts(ctx context.Context, query entities.QueryStringProduct) ([]entities.ProductEntity, int64, int64, error)
}

// struct
type productRepository struct {
	db       *gorm.DB
	esClient *elasticsearch.Client
}

// SearchProducts implements [IProductRepository].
func (p *productRepository) SearchProducts(ctx context.Context, query entities.QueryStringProduct) ([]entities.ProductEntity, int64, int64, error) {
	var mainQueries []string
	var filterQueries []string
	from := (query.Page - 1) * query.Limit

	sortField := "id"
	if query.OrderBy != "" {
		sortField = query.OrderBy
	}

	// Menentukan urutan sorting (asc atau desc)
	sortOrder := "asc"
	if query.OrderType == "desc" {
		sortOrder = "desc"
	}

	// Menyusun bagian sort query
	sortQuery := fmt.Sprintf(`{ "%s": "%s" }`, sortField, sortOrder)

	if query.CategorySlug != "" {
		filterQueries = append(filterQueries, fmt.Sprintf(`{ "term": { "category_slug.keyword": "%s" } }`, query.CategorySlug))
	}

	if query.StartPrice > 0 && query.EndPrice > 0 {
		filterQueries = append(filterQueries, fmt.Sprintf(`{ "range": { "reguler_price": { "gte": %d, "lte": %d } } }`, query.StartPrice, query.EndPrice))
	}

	if query.Search != "" {
		mainQueries = append(mainQueries, fmt.Sprintf(`{ "multi_match": { "query": "%s", "fields": ["name", "description", "category_name"] } }`, query.Search))
	}

	// Query Elasticsearch dengan filtering dan pagination
	mainQuery := fmt.Sprintf(`{
		"from": %d,
		"size": %d,
		"query": {
			"bool": {
				"must": [
					%s
				],
				"filter": [ 
					%s
				]
			}
		},
		"sort": [
			%s
		]
	}`, from, query.Limit, strings.Join(mainQueries, ","), strings.Join(filterQueries, ","), sortQuery)

	// Query Elasticsearch dengan filtering dan pagination
	// Kirim query ke Elasticsearch
	res, err := p.esClient.Search(
		p.esClient.Search.WithContext(ctx),
		p.esClient.Search.WithIndex("products"),
		p.esClient.Search.WithBody(strings.NewReader(mainQuery)),
		p.esClient.Search.WithPretty(),
	)

	if err != nil {
		log.Printf("Error searching Elasticsearch: %s", err)
		return nil, 0, 0, err
	}
	defer res.Body.Close()

	// Decode response
	var result map[string]interface{}
	if err := json.NewDecoder(res.Body).Decode(&result); err != nil {
		log.Printf("Error decoding response: %s", err)
		return nil, 0, 0, err
	}
	
	// Ambil total data
	totalData := 0
	var totalPage int64
	products := []entities.ProductEntity{}

	if hits, ok := result["hits"].(map[string]interface{}); ok {
		// Ambil total data
		if total, ok := hits["total"].(map[string]interface{}); ok {
			if value, ok := total["value"].(float64); ok {
				totalData = int(value)
			}
		}

		// Hitung total halaman
		if query.Limit > 0 {
			totalPage = int64(math.Ceil(float64(totalData) / float64(query.Limit)))
		}

		// Parsing hasil pencarian ke struct domain.Product
		if hits, ok := hits["hits"].([]interface{}); ok {
			for _, hit := range hits {
				if source, ok := hit.(map[string]interface{})["_source"]; ok {
					data, _ := json.Marshal(source)
					var product entities.ProductEntity
					json.Unmarshal(data, &product)
					products = append(products, product)
				}
			}
		}
	}

	return products, int64(totalData), int64(totalPage), nil
}

// Delete implements [IProductRepository].
func (p *productRepository) Delete(ctx context.Context, productID int64) error {
	modelProduct := models.Product{}
	if err := p.db.WithContext(ctx).Preload("Childs").First(&modelProduct, "id = ?", productID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("404")
		}
		log.Errorf("[ProductRepository-1] Delete: %v", err)
		return err
	}

	if err := p.db.WithContext(ctx).Select("Childs").Delete(&modelProduct).Error; err != nil {
		log.Errorf("[ProductRepository-2] Delete: %v", err)
		return err
	}

	res, err := p.esClient.Delete(
		"products",
		strconv.Itoa(int(productID)),
		p.esClient.Delete.WithRefresh("true"),
	)

	if err != nil {
		log.Errorf("[ProductRepository-3] Delete: %v", err)
		return err
	}

	defer res.Body.Close()
	log.Infof("[ProductRepository-4] Delete Product Elasticsearch: %d", productID)

	return nil
}

// Update implements [IProductRepository].
func (p *productRepository) Update(ctx context.Context, req entities.ProductEntity) error {
	modelProduct := models.Product{}

	if err := p.db.Where("id = ?", req.ID).First(&modelProduct).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("404")
		}
		log.Errorf("[ProductRepository-1] Update: %v", err)
		return err
	}

	modelProduct.CategorySlug = req.CategorySlug
	modelProduct.ParentID = req.ParentID
	modelProduct.Name = req.Name
	modelProduct.Image = req.Image
	modelProduct.Description = req.Description
	modelProduct.RegulerPrice = req.RegulerPrice
	modelProduct.SalePrice = req.SalePrice
	modelProduct.Unit = req.Unit
	modelProduct.Weight = req.Weight
	modelProduct.Stock = req.Stock
	modelProduct.Variant = req.Variant
	modelProduct.Status = req.Status

	if err := p.db.Save(&modelProduct).Error; err != nil {
		log.Errorf("[ProductRepository-2] Update: %v", err)
		return err
	}

	if len(req.Child) > 0 {
		if err := p.db.Where("parent_id = ?", modelProduct.ID).Delete(&models.Product{}).Error; err != nil {
			log.Errorf("[ProductRepository-3] Update: %v", err)
			return err
		}

		modelProductChild := []models.Product{}

		for _, val := range req.Child {
			modelProductChild = append(modelProductChild, models.Product{
				CategorySlug: req.CategorySlug,
				ParentID:     &modelProduct.ID,
				Name:         req.Name,
				Image:        val.Image,
				Description:  req.Description,
				RegulerPrice: val.RegulerPrice,
				SalePrice:    val.SalePrice,
				Unit:         req.Unit,
				Weight:       val.Weight,
				Stock:        val.Stock,
				Variant:      req.Variant,
				Status:       req.Status,
			})
		}

		if err := p.db.Create(&modelProductChild).Error; err != nil {
			log.Errorf("[ProductRepository-3] Update: %v", err)
			return err
		}
	}

	return nil
}

// GetByID implements [IProductRepository].
func (p *productRepository) GetByID(ctx context.Context, productID int64) (*entities.ProductEntity, error) {
	modelProduct := models.Product{}

	if err := p.db.WithContext(ctx).Preload("Category").First(&modelProduct, "id = ?", productID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			err = errors.New("404")
		}
		log.Errorf("[ProductRepository-1] GetByID: %v", err)
		return nil, err
	}

	modelParent := []models.Product{}

	err := p.db.WithContext(ctx).Preload("Category").Where("parent_id = ?", modelProduct.ID).Find(&modelParent).Error
	if err != nil {
		log.Errorf("[ProductRepository-2] GetByID: %v", err)
		return nil, err
	}

	childEntities := []entities.ProductEntity{}

	for _, val := range modelParent {
		childEntities = append(childEntities, entities.ProductEntity{
			ID:           val.ID,
			CategorySlug: val.CategorySlug,
			ParentID:     val.ParentID,
			Name:         val.Name,
			Image:        val.Image,
			Description:  val.Description,
			RegulerPrice: val.RegulerPrice,
			SalePrice:    val.SalePrice,
			Unit:         val.Unit,
			Weight:       val.Weight,
			Stock:        val.Stock,
			Variant:      val.Variant,
			Status:       val.Status,
			CategoryName: val.Category.Name,
			Child:        childEntities,
			CreatedAt:    val.CreatedAt,
		})
	}

	return &entities.ProductEntity{
		ID:           modelProduct.ID,
		CategorySlug: modelProduct.CategorySlug,
		ParentID:     modelProduct.ParentID,
		Name:         modelProduct.Name,
		Image:        modelProduct.Image,
		Description:  modelProduct.Description,
		RegulerPrice: modelProduct.RegulerPrice,
		SalePrice:    modelProduct.SalePrice,
		Unit:         modelProduct.Unit,
		Weight:       modelProduct.Weight,
		Stock:        modelProduct.Stock,
		Variant:      modelProduct.Variant,
		Status:       modelProduct.Status,
		CategoryName: modelProduct.Category.Name,
		Child:        childEntities,
		CreatedAt:    modelProduct.CreatedAt,
	}, nil
}

// Create implements [IProductRepository].
func (p *productRepository) Create(ctx context.Context, req entities.ProductEntity) (int64, error) {
	modelProduct := models.Product{
		CategorySlug: req.CategorySlug,
		ParentID:     req.ParentID,
		Name:         req.Name,
		Image:        req.Image,
		Description:  req.Description,
		RegulerPrice: req.RegulerPrice,
		SalePrice:    req.SalePrice,
		Unit:         req.Unit,
		Weight:       req.Weight,
		Stock:        req.Stock,
		Variant:      req.Variant,
		Status:       req.Status,
	}

	if err := p.db.Create(&modelProduct).Error; err != nil {
		log.Errorf("[ProductRepository-1] Create: %v", err)
		return 0, err
	}

	if len(req.Child) > 0 {
		modelProductChild := []models.Product{}
		for _, val := range req.Child {
			modelProductChild = append(modelProductChild, models.Product{
				CategorySlug: req.CategorySlug,
				ParentID:     &modelProduct.ID,
				Name:         req.Name,
				Image:        val.Image,
				Description:  req.Description,
				RegulerPrice: val.RegulerPrice,
				SalePrice:    val.SalePrice,
				Unit:         req.Unit,
				Weight:       val.Weight,
				Stock:        val.Stock,
				Variant:      req.Variant,
				Status:       req.Status,
			})
		}

		if err := p.db.Create(&modelProductChild).Error; err != nil {
			log.Errorf("[ProductRepository-2] Create: %v", err)
			return 0, err
		}
	}

	return modelProduct.ID, nil
}

// GetAll implements [IProductRepository].
func (p *productRepository) GetAll(ctx context.Context, query entities.QueryStringProduct) ([]entities.ProductEntity, int64, int64, error) {
	modelProducts := []models.Product{}
	var countData int64

	order := fmt.Sprintf("%s %s", query.OrderBy, query.OrderType)
	offset := (query.Page - 1) * query.Limit

	sqlMain := p.db.WithContext(ctx).Preload("Category").Where("parent_id IS NULL")

	if query.Status != "" {
		sqlMain = sqlMain.Where("status = ?", query.Status)
	}

	if query.Search != "" {
		sqlMain = sqlMain.Where("name ILIKE ? OR description ILIKE ? OR category_slug ILIKE ?", "%"+query.Search+"%", "%"+query.Search+"%", "%"+query.Search+"%")
	}

	if query.CategorySlug != "" {
		sqlMain = sqlMain.Where("category_slug = ?", query.CategorySlug)
	}

	if query.StartPrice > 0 {
		sqlMain = sqlMain.Where("sale_price >= ?", query.StartPrice)
	}

	if query.EndPrice > 0 {
		sqlMain = sqlMain.Where("sale_price <= ?", query.EndPrice)
	}

	if err := sqlMain.Model(&modelProducts).Count(&countData).Error; err != nil {
		log.Errorf("[ProductRepository-1] GetAll: %v", err)
		return nil, 0, 0, err
	}

	totalPage := int(math.Ceil(float64(countData) / float64(query.Limit)))
	if err := sqlMain.Order(order).Limit(int(query.Limit)).Offset(int(offset)).Find(&modelProducts).Error; err != nil {
		log.Errorf("[ProductRepository-2] GetAll: %v", err)
		return nil, 0, 0, err
	}

	if len(modelProducts) == 0 {
		log.Errorf("[ProductRepository-3] GetAll: %v", "Data not found")
		return nil, 0, 0, errors.New("404")
	}

	respProducts := []entities.ProductEntity{}
	for _, val := range modelProducts {
		respProducts = append(respProducts, entities.ProductEntity{
			ID:           val.ID,
			CategorySlug: val.CategorySlug,
			ParentID:     val.ParentID,
			Name:         val.Name,
			Image:        val.Image,
			Description:  val.Description,
			RegulerPrice: val.RegulerPrice,
			SalePrice:    val.SalePrice,
			Unit:         val.Unit,
			Weight:       val.Weight,
			Stock:        val.Stock,
			Variant:      val.Variant,
			Status:       val.Status,
			CategoryName: val.Category.Name,
			CreatedAt:    val.CreatedAt,
		})
	}

	return respProducts, countData, int64(totalPage), nil
}

func NewProductRepository(db *gorm.DB, es *elasticsearch.Client) IProductRepository {
	return &productRepository{
		db:       db,
		esClient: es,
	}
}
