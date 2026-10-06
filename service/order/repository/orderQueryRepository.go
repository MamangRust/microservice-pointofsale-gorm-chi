package repository

import (
	"context"
	"strings"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/order_errors"
	"gorm.io/gorm"
)

type orderQueryRepository struct {
	db *gorm.DB
}

func NewOrderQueryRepository(db *gorm.DB) OrderQueryRepository {
	return &orderQueryRepository{db: db}
}

func (r *orderQueryRepository) FindAllOrders(ctx context.Context, req *requests.FindAllOrders) ([]*OrderResult, *int, error) {
	offset := (req.Page - 1) * req.PageSize
	var results []*OrderResult

	query := r.db.WithContext(ctx).Raw(`
		SELECT o.order_id, o.merchant_id, o.cashier_id, o.total_price,
			COALESCE(TO_CHAR(o.created_at, 'YYYY-MM-DD HH24:MI:SS'), '') as created_at,
			COALESCE(TO_CHAR(o.updated_at, 'YYYY-MM-DD HH24:MI:SS'), '') as updated_at,
			COUNT(*) OVER() as total_count
		FROM orders o
		WHERE o.deleted_at IS NULL
		AND (? = '' OR o.order_id::TEXT LIKE ? OR o.total_price::TEXT LIKE ?)
		ORDER BY o.order_id DESC
		LIMIT ? OFFSET ?
	`, req.Search, "%"+req.Search+"%", "%"+req.Search+"%", req.PageSize, offset)

	if err := query.Scan(&results).Error; err != nil {
		return nil, nil, order_errors.ErrFindAllOrders
	}

	var totalCount int
	if len(results) > 0 {
		totalCount = int(results[0].TotalCount)
	}

	return results, &totalCount, nil
}

func (r *orderQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllOrders) ([]*OrderResultDeleteAt, *int, error) {
	offset := (req.Page - 1) * req.PageSize
	var results []*OrderResultDeleteAt

	query := r.db.WithContext(ctx).Raw(`
		SELECT o.order_id, o.merchant_id, o.cashier_id, o.total_price,
			COALESCE(TO_CHAR(o.created_at, 'YYYY-MM-DD HH24:MI:SS'), '') as created_at,
			COALESCE(TO_CHAR(o.updated_at, 'YYYY-MM-DD HH24:MI:SS'), '') as updated_at,
			COALESCE(TO_CHAR(o.deleted_at, 'YYYY-MM-DD HH24:MI:SS'), '') as deleted_at,
			COUNT(*) OVER() as total_count
		FROM orders o
		WHERE o.deleted_at IS NULL
		AND (? = '' OR o.order_id::TEXT LIKE ? OR o.total_price::TEXT LIKE ?)
		ORDER BY o.order_id DESC
		LIMIT ? OFFSET ?
	`, req.Search, "%"+req.Search+"%", "%"+req.Search+"%", req.PageSize, offset)

	if err := query.Scan(&results).Error; err != nil {
		return nil, nil, order_errors.ErrFindByActive
	}

	var totalCount int
	if len(results) > 0 {
		totalCount = int(results[0].TotalCount)
	}

	return results, &totalCount, nil
}

func (r *orderQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllOrders) ([]*OrderResultDeleteAt, *int, error) {
	offset := (req.Page - 1) * req.PageSize
	var results []*OrderResultDeleteAt

	query := r.db.WithContext(ctx).Raw(`
		SELECT o.order_id, o.merchant_id, o.cashier_id, o.total_price,
			COALESCE(TO_CHAR(o.created_at, 'YYYY-MM-DD HH24:MI:SS'), '') as created_at,
			COALESCE(TO_CHAR(o.updated_at, 'YYYY-MM-DD HH24:MI:SS'), '') as updated_at,
			COALESCE(TO_CHAR(o.deleted_at, 'YYYY-MM-DD HH24:MI:SS'), '') as deleted_at,
			COUNT(*) OVER() as total_count
		FROM orders o
		WHERE o.deleted_at IS NOT NULL
		AND (? = '' OR o.order_id::TEXT LIKE ? OR o.total_price::TEXT LIKE ?)
		ORDER BY o.order_id DESC
		LIMIT ? OFFSET ?
	`, req.Search, "%"+req.Search+"%", "%"+req.Search+"%", req.PageSize, offset)

	if err := query.Scan(&results).Error; err != nil {
		return nil, nil, order_errors.ErrFindByTrashed
	}

	var totalCount int
	if len(results) > 0 {
		totalCount = int(results[0].TotalCount)
	}

	return results, &totalCount, nil
}

func (r *orderQueryRepository) FindByMerchant(ctx context.Context, req *requests.FindAllOrderMerchant) ([]*OrderResult, *int, error) {
	offset := (req.Page - 1) * req.PageSize
	var results []*OrderResult

	merchantID := req.MerchantID
	query := r.db.WithContext(ctx).Raw(`
		SELECT o.order_id, o.merchant_id, o.cashier_id, o.total_price,
			COALESCE(TO_CHAR(o.created_at, 'YYYY-MM-DD HH24:MI:SS'), '') as created_at,
			COALESCE(TO_CHAR(o.updated_at, 'YYYY-MM-DD HH24:MI:SS'), '') as updated_at,
			COUNT(*) OVER() as total_count
		FROM orders o
		WHERE o.deleted_at IS NULL
		AND o.merchant_id = ?
		AND (? = '' OR o.order_id::TEXT LIKE ? OR o.total_price::TEXT LIKE ?)
		ORDER BY o.order_id DESC
		LIMIT ? OFFSET ?
	`, merchantID, req.Search, "%"+req.Search+"%", "%"+req.Search+"%", req.PageSize, offset)

	if err := query.Scan(&results).Error; err != nil {
		return nil, nil, order_errors.ErrFindByMerchant
	}

	var totalCount int
	if len(results) > 0 {
		totalCount = int(results[0].TotalCount)
	}

	return results, &totalCount, nil
}

func (r *orderQueryRepository) FindById(ctx context.Context, orderID int) (*models.Order, error) {
	var order models.Order
	if err := r.db.WithContext(ctx).Where("order_id = ? AND deleted_at IS NULL", orderID).First(&order).Error; err != nil {
		if strings.Contains(err.Error(), "record not found") {
			return nil, order_errors.ErrFindById
		}
		return nil, order_errors.ErrFindById
	}
	return &order, nil
}

func (r *orderQueryRepository) FindByTrashedId(ctx context.Context, orderID int) (*models.Order, error) {
	var order models.Order
	if err := r.db.WithContext(ctx).Unscoped().Where("order_id = ? AND deleted_at IS NOT NULL", orderID).First(&order).Error; err != nil {
		if strings.Contains(err.Error(), "record not found") {
			return nil, order_errors.ErrFindByTrashedId
		}
		return nil, order_errors.ErrFindByTrashedId
	}
	return &order, nil
}
