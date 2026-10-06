package repository

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/category_errors"
	"gorm.io/gorm"
)

type categoryQueryRepository struct {
	db *gorm.DB
}

func NewCategoryQueryRepository(db *gorm.DB) CategoryQueryRepository {
	return &categoryQueryRepository{db: db}
}

func (r *categoryQueryRepository) FindAllCategory(ctx context.Context, req *requests.FindAllCategory) ([]*CategoryResult, *int, error) {
	offset := (req.Page - 1) * req.PageSize
	var totalCount int64

	var categories []models.Category
	q := r.db.WithContext(ctx).Model(&models.Category{})
	if req.Search != "" {
		q = q.Where("name LIKE ?", "%"+req.Search+"%")
	}
	q.Count(&totalCount)

	if err := q.Offset(offset).Limit(req.PageSize).Order("category_id DESC").Find(&categories).Error; err != nil {
		return nil, nil, category_errors.ErrFindAllCategory
	}

	var results []*CategoryResult
	for _, c := range categories {
		r := &CategoryResult{
			CategoryID:   c.CategoryID,
			Name:         c.Name,
			Description:  c.Description,
			SlugCategory: c.SlugCategory,
			TotalCount:   totalCount,
		}
		if c.CreatedAt != nil {
			s := c.CreatedAt.Format("2006-01-02 15:04:05")
			r.CreatedAt = &s
		}
		if c.UpdatedAt != nil {
			s := c.UpdatedAt.Format("2006-01-02 15:04:05")
			r.UpdatedAt = &s
		}
		results = append(results, r)
	}
	total := int(totalCount)
	return results, &total, nil
}

func (r *categoryQueryRepository) FindById(ctx context.Context, category_id int) (*models.Category, error) {
	var category models.Category
	if err := r.db.WithContext(ctx).Where("category_id = ?", category_id).First(&category).Error; err != nil {
		return nil, category_errors.ErrFindById
	}
	return &category, nil
}

func (r *categoryQueryRepository) FindByIdTrashed(ctx context.Context, category_id int) (*models.Category, error) {
	var category models.Category
	if err := r.db.WithContext(ctx).Unscoped().Where("category_id = ?", category_id).First(&category).Error; err != nil {
		return nil, category_errors.ErrFindById
	}
	return &category, nil
}

func (r *categoryQueryRepository) FindByNameAndId(ctx context.Context, req *requests.CategoryNameAndId) (*models.Category, error) {
	var category models.Category
	if err := r.db.WithContext(ctx).Where("name = ? AND category_id != ?", req.Name, req.CategoryID).First(&category).Error; err != nil {
		return nil, category_errors.ErrFindByNameAndId
	}
	return &category, nil
}

func (r *categoryQueryRepository) FindByName(ctx context.Context, name string) (*models.Category, error) {
	var category models.Category
	if err := r.db.WithContext(ctx).Where("name = ? AND deleted_at IS NULL", name).First(&category).Error; err != nil {
		return nil, category_errors.ErrFindByName
	}
	return &category, nil
}

func (r *categoryQueryRepository) FindByIds(ctx context.Context, ids []int) ([]*models.Category, error) {
	if len(ids) == 0 {
		return nil, nil
	}
	var categories []*models.Category
	if err := r.db.WithContext(ctx).Where("category_id IN ? AND deleted_at IS NULL", ids).Find(&categories).Error; err != nil {
		return nil, category_errors.ErrFindByIds
	}
	return categories, nil
}

func (r *categoryQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllCategory) ([]*CategoryResultDeleteAt, *int, error) {
	offset := (req.Page - 1) * req.PageSize
	var totalCount int64

	var categories []models.Category
	q := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NULL").Model(&models.Category{})
	if req.Search != "" {
		q = q.Where("name LIKE ?", "%"+req.Search+"%")
	}
	q.Count(&totalCount)

	if err := q.Offset(offset).Limit(req.PageSize).Order("category_id DESC").Find(&categories).Error; err != nil {
		return nil, nil, category_errors.ErrFindByActive
	}

	var results []*CategoryResultDeleteAt
	for _, c := range categories {
		cr := &CategoryResultDeleteAt{
			CategoryID:   c.CategoryID,
			Name:         c.Name,
			Description:  c.Description,
			SlugCategory: c.SlugCategory,
			TotalCount:   totalCount,
		}
		if c.CreatedAt != nil {
			s := c.CreatedAt.Format("2006-01-02 15:04:05")
			cr.CreatedAt = &s
		}
		if c.UpdatedAt != nil {
			s := c.UpdatedAt.Format("2006-01-02 15:04:05")
			cr.UpdatedAt = &s
		}
		results = append(results, cr)
	}
	total := int(totalCount)
	return results, &total, nil
}

func (r *categoryQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllCategory) ([]*CategoryResultDeleteAt, *int, error) {
	offset := (req.Page - 1) * req.PageSize
	var totalCount int64

	var categories []models.Category
	q := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Model(&models.Category{})
	if req.Search != "" {
		q = q.Where("name LIKE ?", "%"+req.Search+"%")
	}
	q.Count(&totalCount)

	if err := q.Offset(offset).Limit(req.PageSize).Order("category_id DESC").Find(&categories).Error; err != nil {
		return nil, nil, category_errors.ErrFindByTrashed
	}

	var results []*CategoryResultDeleteAt
	for _, c := range categories {
		cr := &CategoryResultDeleteAt{
			CategoryID:   c.CategoryID,
			Name:         c.Name,
			Description:  c.Description,
			SlugCategory: c.SlugCategory,
			TotalCount:   totalCount,
		}
		if c.CreatedAt != nil {
			s := c.CreatedAt.Format("2006-01-02 15:04:05")
			cr.CreatedAt = &s
		}
		if c.UpdatedAt != nil {
			s := c.UpdatedAt.Format("2006-01-02 15:04:05")
			cr.UpdatedAt = &s
		}
		if c.DeletedAt != nil {
			s := c.DeletedAt.Format("2006-01-02 15:04:05")
			cr.DeletedAt = &s
		}
		results = append(results, cr)
	}
	total := int(totalCount)
	return results, &total, nil
}
