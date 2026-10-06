package repository

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/cashier_errors"
	"gorm.io/gorm"
)

type cashierQueryRepository struct {
	db *gorm.DB
}

func NewCashierQueryRepository(db *gorm.DB) CashierQueryRepository {
	return &cashierQueryRepository{db: db}
}

func (r *cashierQueryRepository) FindAllCashiers(ctx context.Context, req *requests.FindAllCashiers) ([]*CashierResult, *int, error) {
	offset := (req.Page - 1) * req.PageSize
	var totalCount int64

	var cashiers []models.Cashier
	q := r.db.WithContext(ctx).Model(&models.Cashier{})
	if req.Search != "" {
		q = q.Where("name LIKE ?", "%"+req.Search+"%")
	}
	q.Count(&totalCount)

	if err := q.Offset(offset).Limit(req.PageSize).Order("cashier_id DESC").Find(&cashiers).Error; err != nil {
		return nil, nil, cashier_errors.ErrFindAllCashiers
	}

	var results []*CashierResult
	for _, c := range cashiers {
		cr := &CashierResult{CashierID: c.CashierID, MerchantID: c.MerchantID, UserID: c.UserID, Name: c.Name, TotalCount: totalCount}
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

func (r *cashierQueryRepository) FindById(ctx context.Context, cashier_id int) (*models.Cashier, error) {
	var cashier models.Cashier
	if err := r.db.WithContext(ctx).Where("cashier_id = ?", cashier_id).First(&cashier).Error; err != nil {
		return nil, cashier_errors.ErrFindCashierById
	}
	return &cashier, nil
}

func (r *cashierQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllCashiers) ([]*CashierResultDeleteAt, *int, error) {
	offset := (req.Page - 1) * req.PageSize
	var totalCount int64

	var cashiers []models.Cashier
	q := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NULL").Model(&models.Cashier{})
	if req.Search != "" {
		q = q.Where("name LIKE ?", "%"+req.Search+"%")
	}
	q.Count(&totalCount)

	if err := q.Offset(offset).Limit(req.PageSize).Order("cashier_id DESC").Find(&cashiers).Error; err != nil {
		return nil, nil, cashier_errors.ErrFindActiveCashiers
	}

	var results []*CashierResultDeleteAt
	for _, c := range cashiers {
		cr := &CashierResultDeleteAt{CashierID: c.CashierID, MerchantID: c.MerchantID, UserID: c.UserID, Name: c.Name, TotalCount: totalCount}
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

func (r *cashierQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllCashiers) ([]*CashierResultDeleteAt, *int, error) {
	offset := (req.Page - 1) * req.PageSize
	var totalCount int64

	var cashiers []models.Cashier
	q := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Model(&models.Cashier{})
	if req.Search != "" {
		q = q.Where("name LIKE ?", "%"+req.Search+"%")
	}
	q.Count(&totalCount)

	if err := q.Offset(offset).Limit(req.PageSize).Order("cashier_id DESC").Find(&cashiers).Error; err != nil {
		return nil, nil, cashier_errors.ErrFindTrashedCashiers
	}

	var results []*CashierResultDeleteAt
	for _, c := range cashiers {
		cr := &CashierResultDeleteAt{CashierID: c.CashierID, MerchantID: c.MerchantID, UserID: c.UserID, Name: c.Name, TotalCount: totalCount}
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

func (r *cashierQueryRepository) FindByMerchant(ctx context.Context, req *requests.FindAllCashierMerchant) ([]*CashierResult, *int, error) {
	offset := (req.Page - 1) * req.PageSize
	var totalCount int64

	var cashiers []models.Cashier
	q := r.db.WithContext(ctx).Model(&models.Cashier{}).Where("merchant_id = ?", req.MerchantID)
	if req.Search != "" {
		q = q.Where("name LIKE ?", "%"+req.Search+"%")
	}
	q.Count(&totalCount)

	if err := q.Offset(offset).Limit(req.PageSize).Order("cashier_id DESC").Find(&cashiers).Error; err != nil {
		return nil, nil, cashier_errors.ErrFindCashiersByMerchant
	}

	var results []*CashierResult
	for _, c := range cashiers {
		cr := &CashierResult{CashierID: c.CashierID, MerchantID: c.MerchantID, UserID: c.UserID, Name: c.Name, TotalCount: totalCount}
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
