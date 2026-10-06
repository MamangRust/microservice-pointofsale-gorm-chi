package repository

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/product_errors"
	"gorm.io/gorm"
)

type productCommandRepository struct {
	db *gorm.DB
}

func NewProductCommandRepository(db *gorm.DB) *productCommandRepository {
	return &productCommandRepository{db: db}
}

func (r *productCommandRepository) CreateProduct(ctx context.Context, request *requests.CreateProductRequest) (*models.Product, error) {
	now := time.Now()
	weight := int32(request.Weight)
	product := &models.Product{
		MerchantID:   int32(request.MerchantID),
		CategoryID:   int32(request.CategoryID),
		Name:         request.Name,
		Description:  &request.Description,
		Price:        int32(request.Price),
		CountInStock: int32(request.CountInStock),
		Brand:        &request.Brand,
		Weight:       &weight,
		SlugProduct:  request.SlugProduct,
		ImageProduct: &request.ImageProduct,
		CreatedAt:    &now,
		UpdatedAt:    &now,
	}
	if err := r.db.WithContext(ctx).Create(product).Error; err != nil {
		return nil, product_errors.ErrCreateProduct
	}
	return product, nil
}

func (r *productCommandRepository) UpdateProduct(ctx context.Context, request *requests.UpdateProductRequest) (*models.Product, error) {
	var product models.Product
	if err := r.db.WithContext(ctx).Where("product_id = ?", *request.ProductID).First(&product).Error; err != nil {
		return nil, product_errors.ErrUpdateProduct
	}
	weight := int32(request.Weight)
	product.CategoryID = int32(request.CategoryID)
	product.Name = request.Name
	product.Description = &request.Description
	product.Price = int32(request.Price)
	product.CountInStock = int32(request.CountInStock)
	product.Brand = &request.Brand
	product.Weight = &weight
	product.ImageProduct = &request.ImageProduct
	product.UpdatedAt = timePtr(time.Now())
	if err := r.db.WithContext(ctx).Save(&product).Error; err != nil {
		return nil, product_errors.ErrUpdateProduct
	}
	return &product, nil
}

func (r *productCommandRepository) UpdateProductCountStock(ctx context.Context, product_id int, stock int) (*models.Product, error) {
	var product models.Product
	if err := r.db.WithContext(ctx).Where("product_id = ?", product_id).First(&product).Error; err != nil {
		return nil, product_errors.ErrUpdateProductCountStock
	}
	product.CountInStock = int32(stock)
	product.UpdatedAt = timePtr(time.Now())
	if err := r.db.WithContext(ctx).Save(&product).Error; err != nil {
		return nil, product_errors.ErrUpdateProductCountStock
	}
	return &product, nil
}

func (r *productCommandRepository) DecrementProductCountStock(ctx context.Context, productID int, quantity int) (*models.Product, error) {
	var product models.Product
	if err := r.db.WithContext(ctx).Where("product_id = ?", productID).First(&product).Error; err != nil {
		return nil, product_errors.ErrUpdateProductCountStock
	}
	product.CountInStock -= int32(quantity)
	product.UpdatedAt = timePtr(time.Now())
	if err := r.db.WithContext(ctx).Save(&product).Error; err != nil {
		return nil, product_errors.ErrUpdateProductCountStock
	}
	return &product, nil
}

func (r *productCommandRepository) IncrementProductCountStock(ctx context.Context, productID int, quantity int) (*models.Product, error) {
	var product models.Product
	if err := r.db.WithContext(ctx).Where("product_id = ?", productID).First(&product).Error; err != nil {
		return nil, product_errors.ErrUpdateProductCountStock
	}
	product.CountInStock += int32(quantity)
	product.UpdatedAt = timePtr(time.Now())
	if err := r.db.WithContext(ctx).Save(&product).Error; err != nil {
		return nil, product_errors.ErrUpdateProductCountStock
	}
	return &product, nil
}

func (r *productCommandRepository) TrashedProduct(ctx context.Context, product_id int) (*models.Product, error) {
	var product models.Product
	if err := r.db.WithContext(ctx).Where("product_id = ? AND deleted_at IS NULL", product_id).First(&product).Error; err != nil {
		return nil, product_errors.ErrTrashedProduct
	}
	now := time.Now()
	product.DeletedAt = &now
	if err := r.db.WithContext(ctx).Save(&product).Error; err != nil {
		return nil, product_errors.ErrTrashedProduct
	}
	return &product, nil
}

func (r *productCommandRepository) RestoreProduct(ctx context.Context, product_id int) (*models.Product, error) {
	var product models.Product
	if err := r.db.WithContext(ctx).Unscoped().Where("product_id = ? AND deleted_at IS NOT NULL", product_id).First(&product).Error; err != nil {
		return nil, product_errors.ErrRestoreProduct
	}
	product.DeletedAt = nil
	product.UpdatedAt = timePtr(time.Now())
	if err := r.db.WithContext(ctx).Unscoped().Save(&product).Error; err != nil {
		return nil, product_errors.ErrRestoreProduct
	}
	return &product, nil
}

func (r *productCommandRepository) DeleteProductPermanent(ctx context.Context, product_id int) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Where("product_id = ? AND deleted_at IS NOT NULL", product_id).Delete(&models.Product{})
	return result.RowsAffected > 0, result.Error
}

func (r *productCommandRepository) RestoreAllProducts(ctx context.Context) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Model(&models.Product{}).Update("deleted_at", nil)
	return true, result.Error
}

func (r *productCommandRepository) DeleteAllProductPermanent(ctx context.Context) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Delete(&models.Product{})
	return true, result.Error
}

func timePtr(t time.Time) *time.Time { return &t }
