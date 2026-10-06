package repository

import (
	"context"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	orderitem_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors/order_item_errors"
	"gorm.io/gorm"
)

type orderItemCommandRepository struct {
	db *gorm.DB
}

func NewOrderItemCommandRepository(db *gorm.DB) OrderItemCommandRepository {
	return &orderItemCommandRepository{db: db}
}

func (r *orderItemCommandRepository) CreateOrderItem(ctx context.Context, req *requests.CreateOrderItemRecordRequest) (*models.OrderItem, error) {
	now := time.Now()
	item := &models.OrderItem{
		OrderID:   int32(req.OrderID),
		ProductID: int32(req.ProductID),
		Quantity:  int32(req.Quantity),
		Price:     int64(req.Price),
		CreatedAt: &now,
		UpdatedAt: &now,
	}
	if err := r.db.WithContext(ctx).Create(item).Error; err != nil {
		return nil, orderitem_errors.ErrCreateOrderItem
	}
	return item, nil
}

func (r *orderItemCommandRepository) UpdateOrderItem(ctx context.Context, req *requests.UpdateOrderItemRecordRequest) (*models.OrderItem, error) {
	var item models.OrderItem
	if err := r.db.WithContext(ctx).Where("order_item_id = ?", req.OrderItemID).First(&item).Error; err != nil {
		return nil, orderitem_errors.ErrUpdateOrderItem
	}

	item.ProductID = int32(req.ProductID)
	item.Quantity = int32(req.Quantity)
	item.Price = int64(req.Price)
	now := time.Now()
	item.UpdatedAt = &now

	if err := r.db.WithContext(ctx).Save(&item).Error; err != nil {
		return nil, orderitem_errors.ErrUpdateOrderItem
	}
	return &item, nil
}

func (r *orderItemCommandRepository) DeleteOrderItem(ctx context.Context, orderItemID int) error {
	if err := r.db.WithContext(ctx).Where("order_item_id = ?", orderItemID).Delete(&models.OrderItem{}).Error; err != nil {
		return orderitem_errors.ErrDeleteOrderItemPermanent
	}
	return nil
}
