package repository

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors/order_errors"
	"gorm.io/gorm"
)

type orderCommandRepository struct {
	db *gorm.DB
}

func NewOrderCommandRepository(db *gorm.DB) OrderCommandRepository {
	return &orderCommandRepository{db: db}
}

func (r *orderCommandRepository) DeleteOrder(ctx context.Context, orderID int) error {
	now := time.Now()
	if err := r.db.WithContext(ctx).Unscoped().Model(&models.Order{}).Where("order_id = ?", orderID).Update("deleted_at", now).Error; err != nil {
		return order_errors.ErrDeleteOrderPermanent
	}
	return nil
}

func (r *orderCommandRepository) CreateOrder(ctx context.Context, request *requests.CreateOrderRecordRequest) (*models.Order, error) {
	order := &models.Order{
		MerchantID: int32(request.MerchantID),
		CashierID:  int32(request.CashierID),
		TotalPrice: int64(request.TotalPrice),
		CreatedAt:  ptrTime(time.Now()),
		UpdatedAt:  ptrTime(time.Now()),
	}

	if err := r.db.WithContext(ctx).Create(order).Error; err != nil {
		return nil, order_errors.ErrCreateOrder
	}

	return order, nil
}

func (r *orderCommandRepository) FindAllTrashed(ctx context.Context) ([]*models.Order, error) {
	var orders []*models.Order
	if err := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Find(&orders).Error; err != nil {
		return nil, order_errors.ErrFindAllTrashed
	}
	return orders, nil
}

func (r *orderCommandRepository) UpdateOrder(ctx context.Context, request *requests.UpdateOrderRecordRequest) (*models.Order, error) {
	var order models.Order
	if err := r.db.WithContext(ctx).Where("order_id = ?", request.OrderID).First(&order).Error; err != nil {
		return nil, order_errors.ErrUpdateOrder
	}

	order.TotalPrice = int64(request.TotalPrice)
	order.UpdatedAt = ptrTime(time.Now())

	if err := r.db.WithContext(ctx).Save(&order).Error; err != nil {
		return nil, order_errors.ErrUpdateOrder
	}

	return &order, nil
}

func (r *orderCommandRepository) TrashedOrder(ctx context.Context, orderID int) (*models.Order, error) {
	var order models.Order
	if err := r.db.WithContext(ctx).Where("order_id = ? AND deleted_at IS NULL", orderID).First(&order).Error; err != nil {
		return nil, order_errors.ErrTrashedOrder.WithInternal(err)
	}

	now := time.Now()
	order.DeletedAt = &now
	if err := r.db.WithContext(ctx).Save(&order).Error; err != nil {
		return nil, order_errors.ErrTrashedOrder.WithInternal(err)
	}

	return &order, nil
}

func (r *orderCommandRepository) RestoreOrder(ctx context.Context, orderID int) (*models.Order, error) {
	var order models.Order
	if err := r.db.WithContext(ctx).Unscoped().Where("order_id = ? AND deleted_at IS NOT NULL", orderID).First(&order).Error; err != nil {
		return nil, order_errors.ErrRestoreOrderNotFound
	}

	order.DeletedAt = nil
	order.UpdatedAt = ptrTime(time.Now())

	if err := r.db.WithContext(ctx).Unscoped().Save(&order).Error; err != nil {
		return nil, order_errors.ErrRestoreOrder.WithInternal(err)
	}

	return &order, nil
}

func (r *orderCommandRepository) DeleteOrderPermanent(ctx context.Context, orderID int) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Where("order_id = ? AND deleted_at IS NOT NULL", orderID).Delete(&models.Order{})
	if result.Error != nil {
		return false, order_errors.ErrDeleteOrderPermanent.WithInternal(result.Error)
	}
	if result.RowsAffected == 0 {
		return false, order_errors.ErrDeleteOrderPermanentNotFound
	}
	return true, nil
}

func (r *orderCommandRepository) DeleteAllOrderPermanent(ctx context.Context) (bool, error) {
	result := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Delete(&models.Order{})
	if result.Error != nil {
		return false, order_errors.ErrDeleteAllOrderPermanent.WithInternal(result.Error)
	}
	return true, nil
}

func ptrTime(t time.Time) *time.Time { return &t }
