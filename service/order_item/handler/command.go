package handler

import (
	"context"

	pborderitem "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/errors"
	orderitem_errors "github.com/MamangRust/microservice-point-of-sale-shared/errors/order_item_errors"
	"go.uber.org/zap"
)

func (s *orderItemQueryHandleGrpc) CreateOrderItem(ctx context.Context, request *pborderitem.CreateOrderItemRecordRequest) (*pborderitem.ApiResponseOrderItem, error) {
	orderID := int(request.GetOrderId())
	productID := int(request.GetProductId())
	quantity := int(request.GetQuantity())

	if orderID <= 0 || productID <= 0 {
		return nil, orderitem_errors.ErrGrpcInvalidID
	}
	if quantity <= 0 {
		return nil, orderitem_errors.ErrFailedInvalidQuantity
	}

	req := &requests.CreateOrderItemRecordRequest{
		OrderID:   orderID,
		ProductID: productID,
		Quantity:  quantity,
		Price:     int(request.GetPrice()),
	}

	item, err := s.orderItemCommandService.CreateOrderItem(ctx, req)
	if err != nil {
		s.logger.Error("CreateOrderItem failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	return &pborderitem.ApiResponseOrderItem{
		Status:  "success",
		Message: "Successfully created order item",
		Data:    mapOrderItemModel(item),
	}, nil
}

func (s *orderItemQueryHandleGrpc) UpdateOrderItem(ctx context.Context, request *pborderitem.UpdateOrderItemRecordRequest) (*pborderitem.ApiResponseOrderItem, error) {
	orderItemID := int(request.GetOrderItemId())
	quantity := int(request.GetQuantity())

	if orderItemID <= 0 {
		return nil, orderitem_errors.ErrGrpcInvalidID
	}
	if quantity <= 0 {
		return nil, orderitem_errors.ErrFailedInvalidQuantity
	}

	req := &requests.UpdateOrderItemRecordRequest{
		OrderItemID: orderItemID,
		OrderID:     int(request.GetOrderId()),
		ProductID:   int(request.GetProductId()),
		Quantity:    quantity,
		Price:       int(request.GetPrice()),
	}

	item, err := s.orderItemCommandService.UpdateOrderItem(ctx, req)
	if err != nil {
		s.logger.Error("UpdateOrderItem failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	return &pborderitem.ApiResponseOrderItem{
		Status:  "success",
		Message: "Successfully updated order item",
		Data:    mapOrderItemModel(item),
	}, nil
}

func (s *orderItemQueryHandleGrpc) DeleteOrderItem(ctx context.Context, request *pborderitem.DeleteOrderItemRecordRequest) (*pborderitem.ApiResponseOrderItemDelete, error) {
	orderItemID := int(request.GetOrderItemId())
	if orderItemID <= 0 {
		return nil, orderitem_errors.ErrGrpcInvalidID
	}

	if err := s.orderItemCommandService.DeleteOrderItem(ctx, orderItemID); err != nil {
		s.logger.Error("DeleteOrderItem failed", zap.Error(err))
		return nil, errors.ToGrpcError(err)
	}

	return &pborderitem.ApiResponseOrderItemDelete{
		Status:  "success",
		Message: "Successfully deleted order item",
	}, nil
}
