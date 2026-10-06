package repository

import (
	"context"
	"strings"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/product_errors"
	"gorm.io/gorm"
)

type productQueryRepository struct {
	db *gorm.DB
}

func NewProductQueryRepository(db *gorm.DB) *productQueryRepository {
	return &productQueryRepository{db: db}
}

func (r *productQueryRepository) FindAllProducts(ctx context.Context, req *requests.FindAllProducts) ([]*ProductResult, *int, error) {
	offset := (req.Page - 1) * req.PageSize
	var results []*ProductResult
	err := r.db.WithContext(ctx).Raw(`
		SELECT p.product_id, p.merchant_id, p.category_id, p.name, p.description,
			p.price, p.count_in_stock, p.brand, p.weight, p.slug_product,
			COALESCE(TO_CHAR(p.created_at, 'YYYY-MM-DD HH24:MI:SS'), '') as created_at,
			COALESCE(TO_CHAR(p.updated_at, 'YYYY-MM-DD HH24:MI:SS'), '') as updated_at,
			COUNT(*) OVER() as total_count
		FROM products p
		WHERE p.deleted_at IS NULL
		AND (? = '' OR p.name ILIKE ?)
		ORDER BY p.product_id DESC LIMIT ? OFFSET ?
	`, req.Search, "%"+req.Search+"%", req.PageSize, offset).Scan(&results).Error
	if err != nil {
		return nil, nil, product_errors.ErrFindAllProducts
	}
	var total int
	if len(results) > 0 {
		total = int(results[0].TotalCount)
	}
	return results, &total, nil
}

func (r *productQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllProducts) ([]*ProductResultDeleteAt, *int, error) {
	offset := (req.Page - 1) * req.PageSize
	var results []*ProductResultDeleteAt
	err := r.db.WithContext(ctx).Raw(`
		SELECT p.product_id, p.merchant_id, p.category_id, p.name, p.description,
			p.price, p.count_in_stock, p.brand, p.weight, p.slug_product,
			COALESCE(TO_CHAR(p.created_at, 'YYYY-MM-DD HH24:MI:SS'), '') as created_at,
			COALESCE(TO_CHAR(p.updated_at, 'YYYY-MM-DD HH24:MI:SS'), '') as updated_at,
			COALESCE(TO_CHAR(p.deleted_at, 'YYYY-MM-DD HH24:MI:SS'), '') as deleted_at,
			COUNT(*) OVER() as total_count
		FROM products p
		WHERE p.deleted_at IS NULL
		AND (? = '' OR p.name ILIKE ?)
		ORDER BY p.product_id DESC LIMIT ? OFFSET ?
	`, req.Search, "%"+req.Search+"%", req.PageSize, offset).Scan(&results).Error
	if err != nil {
		return nil, nil, product_errors.ErrFindByActive
	}
	var total int
	if len(results) > 0 {
		total = int(results[0].TotalCount)
	}
	return results, &total, nil
}

func (r *productQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllProducts) ([]*ProductResultDeleteAt, *int, error) {
	offset := (req.Page - 1) * req.PageSize
	var results []*ProductResultDeleteAt
	err := r.db.WithContext(ctx).Raw(`
		SELECT p.product_id, p.merchant_id, p.category_id, p.name, p.description,
			p.price, p.count_in_stock, p.brand, p.weight, p.slug_product,
			COALESCE(TO_CHAR(p.created_at, 'YYYY-MM-DD HH24:MI:SS'), '') as created_at,
			COALESCE(TO_CHAR(p.updated_at, 'YYYY-MM-DD HH24:MI:SS'), '') as updated_at,
			COALESCE(TO_CHAR(p.deleted_at, 'YYYY-MM-DD HH24:MI:SS'), '') as deleted_at,
			COUNT(*) OVER() as total_count
		FROM products p
		WHERE p.deleted_at IS NOT NULL
		AND (? = '' OR p.name ILIKE ?)
		ORDER BY p.product_id DESC LIMIT ? OFFSET ?
	`, req.Search, "%"+req.Search+"%", req.PageSize, offset).Scan(&results).Error
	if err != nil {
		return nil, nil, product_errors.ErrFindByTrashed
	}
	var total int
	if len(results) > 0 {
		total = int(results[0].TotalCount)
	}
	return results, &total, nil
}

func (r *productQueryRepository) FindByMerchant(ctx context.Context, req *requests.ProductByMerchantRequest) ([]*ProductByMerchantResult, *int, error) {
	offset := (req.Page - 1) * req.PageSize
	var results []*ProductByMerchantResult
	err := r.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) OVER() as total_count, p.product_id, p.category_id, p.name, p.description,
			p.price, p.count_in_stock, p.brand, p.image_product,
			COALESCE(TO_CHAR(p.created_at, 'YYYY-MM-DD HH24:MI:SS'), '') as created_at
		FROM products p
		WHERE p.deleted_at IS NULL AND p.merchant_id = ?
		ORDER BY p.product_id DESC LIMIT ? OFFSET ?
	`, req.MerchantID, req.PageSize, offset).Scan(&results).Error
	if err != nil {
		return nil, nil, product_errors.ErrFindByMerchant
	}
	var total int
	if len(results) > 0 {
		total = int(results[0].TotalCount)
	}
	return results, &total, nil
}

func (r *productQueryRepository) FindByCategoryID(ctx context.Context, categoryID int32, req *requests.ProductByCategoryRequest) ([]*ProductByCategoryResult, *int, error) {
	offset := (req.Page - 1) * req.PageSize
	var results []*ProductByCategoryResult
	err := r.db.WithContext(ctx).Raw(`
		SELECT COUNT(*) OVER() as total_count, p.product_id, p.merchant_id, p.category_id,
			p.slug_product, p.weight, p.name, p.description, p.price, p.count_in_stock
		FROM products p
		WHERE p.deleted_at IS NULL AND p.category_id = ?
		ORDER BY p.product_id DESC LIMIT ? OFFSET ?
	`, categoryID, req.PageSize, offset).Scan(&results).Error
	if err != nil {
		return nil, nil, product_errors.ErrFindByCategory
	}
	var total int
	if len(results) > 0 {
		total = int(results[0].TotalCount)
	}
	return results, &total, nil
}

func (r *productQueryRepository) FindById(ctx context.Context, product_id int) (*models.Product, error) {
	var product models.Product
	if err := r.db.WithContext(ctx).Where("product_id = ? AND deleted_at IS NULL", product_id).First(&product).Error; err != nil {
		if strings.Contains(err.Error(), "record not found") {
			return nil, product_errors.ErrFindById
		}
		return nil, product_errors.ErrFindById
	}
	return &product, nil
}

func (r *productQueryRepository) FindByIdTrashed(ctx context.Context, id int) (*models.Product, error) {
	var product models.Product
	if err := r.db.WithContext(ctx).Unscoped().Where("product_id = ? AND deleted_at IS NOT NULL", id).First(&product).Error; err != nil {
		return nil, product_errors.ErrFindById
	}
	return &product, nil
}
