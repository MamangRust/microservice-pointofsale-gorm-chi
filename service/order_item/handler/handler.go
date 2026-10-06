package handler

import (
	"github.com/MamangRust/microservice-point-of-sale-order-item/service"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
)

// Handler groups all order_item gRPC handlers.
type Handler interface {
	OrderItemQueryHandleGrpc
}

type handler struct {
	OrderItemQueryHandleGrpc
}

// NewHandler initializes order_item gRPC handlers.
func NewHandler(svc *service.Service, logger logger.LoggerInterface) Handler {
	return &handler{
		OrderItemQueryHandleGrpc: NewOrderItemQueryHandleGrpc(svc.OrderItemQuery, svc.OrderItemCommand, logger),
	}
}
