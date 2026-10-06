package repository

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	orderitem_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors/order_item_errors"
	"gorm.io/gorm"
)

type orderItemQueryRepository struct {
	db *gorm.DB
}

func NewOrderItemQueryRepository(db *gorm.DB) OrderItemQueryRepository {
	return &orderItemQueryRepository{db: db}
}

func (r *orderItemQueryRepository) FindAllOrderItems(ctx context.Context, req *requests.FindAllOrderItems) ([]*OrderItemResult, *int, error) {
	offset := (req.Page - 1) * req.PageSize
	var totalCount int64

	var items []models.OrderItem
	q := r.db.WithContext(ctx).Model(&models.OrderItem{})
	q.Count(&totalCount)

	if err := q.Offset(offset).Limit(req.PageSize).Order("order_item_id DESC").Find(&items).Error; err != nil {
		return nil, nil, orderitem_errors.ErrFindAllOrderItems
	}

	var results []*OrderItemResult
	for _, item := range items {
		cr := &OrderItemResult{
			OrderItemID: item.OrderItemID, OrderID: item.OrderID, ProductID: item.ProductID,
			Quantity: item.Quantity, Price: int32(item.Price), TotalCount: totalCount,
		}
		if item.CreatedAt != nil {
			s := item.CreatedAt.Format("2006-01-02 15:04:05")
			cr.CreatedAt = &s
		}
		if item.UpdatedAt != nil {
			s := item.UpdatedAt.Format("2006-01-02 15:04:05")
			cr.UpdatedAt = &s
		}
		results = append(results, cr)
	}
	total := int(totalCount)
	return results, &total, nil
}

func (r *orderItemQueryRepository) FindByActive(ctx context.Context, req *requests.FindAllOrderItems) ([]*OrderItemResultDeleteAt, *int, error) {
	offset := (req.Page - 1) * req.PageSize
	var totalCount int64

	var items []models.OrderItem
	q := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NULL").Model(&models.OrderItem{})
	q.Count(&totalCount)

	if err := q.Offset(offset).Limit(req.PageSize).Order("order_item_id DESC").Find(&items).Error; err != nil {
		return nil, nil, orderitem_errors.ErrFindAllOrderItems
	}

	var results []*OrderItemResultDeleteAt
	for _, item := range items {
		cr := &OrderItemResultDeleteAt{
			OrderItemID: item.OrderItemID, OrderID: item.OrderID, ProductID: item.ProductID,
			Quantity: item.Quantity, Price: int32(item.Price), TotalCount: totalCount,
		}
		if item.CreatedAt != nil {
			s := item.CreatedAt.Format("2006-01-02 15:04:05")
			cr.CreatedAt = &s
		}
		if item.UpdatedAt != nil {
			s := item.UpdatedAt.Format("2006-01-02 15:04:05")
			cr.UpdatedAt = &s
		}
		results = append(results, cr)
	}
	total := int(totalCount)
	return results, &total, nil
}

func (r *orderItemQueryRepository) FindByTrashed(ctx context.Context, req *requests.FindAllOrderItems) ([]*OrderItemResultDeleteAt, *int, error) {
	offset := (req.Page - 1) * req.PageSize
	var totalCount int64

	var items []models.OrderItem
	q := r.db.WithContext(ctx).Unscoped().Where("deleted_at IS NOT NULL").Model(&models.OrderItem{})
	q.Count(&totalCount)

	if err := q.Offset(offset).Limit(req.PageSize).Order("order_item_id DESC").Find(&items).Error; err != nil {
		return nil, nil, orderitem_errors.ErrFindAllOrderItems
	}

	var results []*OrderItemResultDeleteAt
	for _, item := range items {
		cr := &OrderItemResultDeleteAt{
			OrderItemID: item.OrderItemID, OrderID: item.OrderID, ProductID: item.ProductID,
			Quantity: item.Quantity, Price: int32(item.Price), TotalCount: totalCount,
		}
		if item.CreatedAt != nil {
			s := item.CreatedAt.Format("2006-01-02 15:04:05")
			cr.CreatedAt = &s
		}
		if item.UpdatedAt != nil {
			s := item.UpdatedAt.Format("2006-01-02 15:04:05")
			cr.UpdatedAt = &s
		}
		if item.DeletedAt != nil {
			s := item.DeletedAt.Format("2006-01-02 15:04:05")
			cr.DeletedAt = &s
		}
		results = append(results, cr)
	}
	total := int(totalCount)
	return results, &total, nil
}

func (r *orderItemQueryRepository) FindOrderItemByOrder(ctx context.Context, orderID int) ([]*models.OrderItem, error) {
	var items []models.OrderItem
	if err := r.db.WithContext(ctx).Where("order_id = ?", orderID).Find(&items).Error; err != nil {
		return nil, orderitem_errors.ErrFindOrderItemByOrder
	}
	var result []*models.OrderItem
	for i := range items {
		result = append(result, &items[i])
	}
	return result, nil
}
