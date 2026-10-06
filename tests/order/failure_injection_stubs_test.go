package order_test

import (
	"context"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

// stubCashierRepo implements repository.CashierQueryRepository. In the
// failure-injection suite it is never reached (the merchant dependency fails
// first), so it returns a minimal valid row to keep the flow honest if the
// ordering ever changes.
type stubCashierRepo struct {
	cashierID int
}

func (s *stubCashierRepo) FindById(ctx context.Context, cashierID int) (*models.Cashier, error) {
	return &models.Cashier{CashierID: int32(cashierID), MerchantID: int32(s.cashierID), Name: "stub"}, nil
}

// stubProductRepo implements repository.ProductQueryRepository. Never reached
// in the failure-injection suite (merchant fails first).
type stubProductRepo struct{}

func (s *stubProductRepo) FindById(ctx context.Context, productID int) (*models.Product, error) {
	return &models.Product{ProductID: int32(productID), Price: 10000, CountInStock: 100}, nil
}

// stubProductCommandRepo implements repository.ProductCommandRepository. It is
// never exercised in the failure-injection suite (the merchant dependency
// fails before any stock mutation), so it returns a minimal valid product.
type stubProductCommandRepo struct{}

func (s *stubProductCommandRepo) DecrementProductCountStock(ctx context.Context, productID int, quantity int) (*models.Product, error) {
	return &models.Product{ProductID: int32(productID), CountInStock: 100}, nil
}

func (s *stubProductCommandRepo) IncrementProductCountStock(ctx context.Context, productID int, quantity int) (*models.Product, error) {
	return &models.Product{ProductID: int32(productID), CountInStock: 100}, nil
}

// stubOrderItemCommandRepo implements repository.OrderItemCommandRepository.
// Likewise never reached in the failure-injection suite; it returns minimal
// valid rows so the flow stays honest if the ordering ever changes.
type stubOrderItemCommandRepo struct{}

func (s *stubOrderItemCommandRepo) DeleteOrderItem(ctx context.Context, orderItemID int) error {
	return nil
}

func (s *stubOrderItemCommandRepo) CreateOrderItem(ctx context.Context, req *requests.CreateOrderItemRecordRequest) (*models.OrderItem, error) {
	return &models.OrderItem{
		OrderItemID: 1,
		OrderID:     int32(req.OrderID),
		ProductID:   int32(req.ProductID),
		Quantity:    int32(req.Quantity),
		Price:       int64(req.Price),
	}, nil
}

func (s *stubOrderItemCommandRepo) UpdateOrderItem(ctx context.Context, req *requests.UpdateOrderItemRecordRequest) (*models.OrderItem, error) {
	return &models.OrderItem{OrderItemID: int32(req.OrderItemID)}, nil
}
