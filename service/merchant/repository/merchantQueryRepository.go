package repository

import (
	"context"
	"strings"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	sharedErrors "github.com/MamangRust/microservice-point-of-sale-shared/errors"
	"gorm.io/gorm"
)

type merchantQueryRepository struct {
	db *gorm.DB
}

func NewMerchantQueryRepository(db *gorm.DB) MerchantQueryRepository {
	return &merchantQueryRepository{db: db}
}

func (r *merchantQueryRepository) FindAllMerchants(ctx context.Context, req *requests.FindAllMerchants) ([]*MerchantResult, *int, error) {
	offset := (req.Page - 1) * req.PageSize
	var results []*MerchantResult

	err := r.db.WithContext(ctx).Raw(`
		SELECT m.merchant_id, m.user_id, m.name, m.description, m.address,
			m.contact_email, m.contact_phone, m.status,
			COALESCE(TO_CHAR(m.created_at, 'YYYY-MM-DD HH24:MI:SS'), '') as created_at,
			COALESCE(TO_CHAR(m.updated_at, 'YYYY-MM-DD HH24:MI:SS'), '') as updated_at,
			COUNT(*) OVER() as total_count
		FROM merchants m
		WHERE m.deleted_at IS NULL
		AND (? = '' OR m.name ILIKE ? OR m.merchant_id::TEXT LIKE ?)
		ORDER BY m.merchant_id DESC
		LIMIT ? OFFSET ?
	`, req.Search, "%"+req.Search+"%", "%"+req.Search+"%", req.PageSize, offset).Scan(&results).Error
	if err != nil {
		return nil, nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	var totalCount int
	if len(results) > 0 {
		totalCount = int(results[0].TotalCount)
	}
	return results, &totalCount, nil
}

func (r *merchantQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllMerchants) ([]*MerchantResultDeleteAt, *int, error) {
	offset := (req.Page - 1) * req.PageSize
	var results []*MerchantResultDeleteAt

	err := r.db.WithContext(ctx).Raw(`
		SELECT m.merchant_id, m.user_id, m.name, m.description, m.address,
			m.contact_email, m.contact_phone, m.status,
			COALESCE(TO_CHAR(m.created_at, 'YYYY-MM-DD HH24:MI:SS'), '') as created_at,
			COALESCE(TO_CHAR(m.updated_at, 'YYYY-MM-DD HH24:MI:SS'), '') as updated_at,
			COALESCE(TO_CHAR(m.deleted_at, 'YYYY-MM-DD HH24:MI:SS'), '') as deleted_at,
			COUNT(*) OVER() as total_count
		FROM merchants m
		WHERE m.deleted_at IS NULL
		AND (? = '' OR m.name ILIKE ? OR m.merchant_id::TEXT LIKE ?)
		ORDER BY m.merchant_id DESC
		LIMIT ? OFFSET ?
	`, req.Search, "%"+req.Search+"%", "%"+req.Search+"%", req.PageSize, offset).Scan(&results).Error
	if err != nil {
		return nil, nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	var totalCount int
	if len(results) > 0 {
		totalCount = int(results[0].TotalCount)
	}
	return results, &totalCount, nil
}

func (r *merchantQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllMerchants) ([]*MerchantResultDeleteAt, *int, error) {
	offset := (req.Page - 1) * req.PageSize
	var results []*MerchantResultDeleteAt

	err := r.db.WithContext(ctx).Raw(`
		SELECT m.merchant_id, m.user_id, m.name, m.description, m.address,
			m.contact_email, m.contact_phone, m.status,
			COALESCE(TO_CHAR(m.created_at, 'YYYY-MM-DD HH24:MI:SS'), '') as created_at,
			COALESCE(TO_CHAR(m.updated_at, 'YYYY-MM-DD HH24:MI:SS'), '') as updated_at,
			COALESCE(TO_CHAR(m.deleted_at, 'YYYY-MM-DD HH24:MI:SS'), '') as deleted_at,
			COUNT(*) OVER() as total_count
		FROM merchants m
		WHERE m.deleted_at IS NOT NULL
		AND (? = '' OR m.name ILIKE ? OR m.merchant_id::TEXT LIKE ?)
		ORDER BY m.merchant_id DESC
		LIMIT ? OFFSET ?
	`, req.Search, "%"+req.Search+"%", "%"+req.Search+"%", req.PageSize, offset).Scan(&results).Error
	if err != nil {
		return nil, nil, sharedErrors.ErrInternal.WithInternal(err)
	}

	var totalCount int
	if len(results) > 0 {
		totalCount = int(results[0].TotalCount)
	}
	return results, &totalCount, nil
}

func (r *merchantQueryRepository) FindById(ctx context.Context, userID int) (*models.Merchant, error) {
	var merchant models.Merchant
	if err := r.db.WithContext(ctx).Where("merchant_id = ? AND deleted_at IS NULL", userID).First(&merchant).Error; err != nil {
		if strings.Contains(err.Error(), "record not found") {
			return nil, sharedErrors.ErrInternal
		}
		return nil, sharedErrors.ErrInternal.WithInternal(err)
	}
	return &merchant, nil
}
