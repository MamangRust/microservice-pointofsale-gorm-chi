package repository

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/category_errors"
	"gorm.io/gorm"
)

type categoryCommandRepository struct {
	db *gorm.DB
}

func NewCategoryCommandRepository(db *gorm.DB) CategoryCommandRepository {
	return &categoryCommandRepository{db: db}
}

func (r *categoryCommandRepository) CreateCategory(ctx context.Context, request *requests.CreateCategoryRequest) (*models.Category, error) {
	category := &models.Category{
		Name:         request.Name,
		Description:  &request.Description,
		SlugCategory: request.SlugCategory,
	}
	if err := r.db.WithContext(ctx).Create(category).Error; err != nil {
		return nil, category_errors.ErrCreateCategory
	}
	return category, nil
}

func (r *categoryCommandRepository) UpdateCategory(ctx context.Context, request *requests.UpdateCategoryRequest) (*models.Category, error) {
	var category models.Category
	if err := r.db.WithContext(ctx).Where("category_id = ?", *request.CategoryID).First(&category).Error; err != nil {
		return nil, category_errors.ErrUpdateCategory
	}
	category.Name = request.Name
	category.Description = &request.Description
	category.SlugCategory = request.SlugCategory
	if err := r.db.WithContext(ctx).Save(&category).Error; err != nil {
		return nil, category_errors.ErrUpdateCategory
	}
	return &category, nil
}

func (r *categoryCommandRepository) TrashedCategory(ctx context.Context, category_id int) (*models.Category, error) {
	var category models.Category
	if err := r.db.WithContext(ctx).Where("category_id = ?", category_id).First(&category).Error; err != nil {
		return nil, category_errors.ErrTrashedCategory
	}
	tm := time.Now()
	if err := r.db.WithContext(ctx).Model(&category).Update("deleted_at", tm).Error; err != nil {
		return nil, category_errors.ErrTrashedCategory
	}
	return &category, nil
}

func (r *categoryCommandRepository) RestoreCategory(ctx context.Context, category_id int) (*models.Category, error) {
	var category models.Category
	if err := r.db.WithContext(ctx).Unscoped().Where("category_id = ?", category_id).First(&category).Error; err != nil {
		return nil, category_errors.ErrRestoreCategory
	}
	if err := r.db.WithContext(ctx).Unscoped().Model(&category).Update("deleted_at", nil).Error; err != nil {
		return nil, category_errors.ErrRestoreCategory
	}
	return &category, nil
}

func (r *categoryCommandRepository) DeleteCategoryPermanently(ctx context.Context, category_id int) (bool, error) {
	if err := r.db.WithContext(ctx).Unscoped().Where("category_id = ?", category_id).Delete(&models.Category{}).Error; err != nil {
		return false, category_errors.ErrDeleteCategoryPermanently
	}
	return true, nil
}

func (r *categoryCommandRepository) RestoreAllCategories(ctx context.Context) (bool, error) {
	if err := r.db.WithContext(ctx).Unscoped().Model(&models.Category{}).Where("deleted_at IS NOT NULL").Update("deleted_at", nil).Error; err != nil {
		return false, category_errors.ErrRestoreAllCategories
	}
	return true, nil
}

func (r *categoryCommandRepository) DeleteAllPermanentCategories(ctx context.Context) (bool, error) {
	if err := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Delete(&models.Category{}).Error; err != nil {
		return false, category_errors.ErrDeleteAllPermanentCategories
	}
	return true, nil
}
