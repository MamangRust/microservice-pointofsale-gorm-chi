package repository

import (
	pbcashier "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	pbmerchant "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	pborder "github.com/MamangRust/microservice-point-of-sale-pb/order"
	pborderitem "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	cashieradapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/cashier"
	merchantadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/merchant"
	orderadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/order"
	orderitemadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/orderitem"
	"gorm.io/gorm"
)

type Repositories struct {
	CashierQuery                 CashierQueryRepository
	MerchantQuery                MerchantQueryRepository
	OrderQuery                   OrderQueryRepository
	OrderItemQuery               OrderItemQueryRepository
	TransactionCommandRepository TransactionCommandRepository
	TransactionQueryRepository   TransactionQueryRepository
}

type GuardOptions struct {
	Cashier   []adapter.GuardOption
	Merchant  []adapter.GuardOption
	Order     []adapter.GuardOption
	OrderItem []adapter.GuardOption
}

func NewRepositories(
	db *gorm.DB,
	cashierClient pbcashier.CashierQueryServiceClient,
	merchantClient pbmerchant.MerchantQueryServiceClient,
	orderClient pborder.OrderQueryServiceClient,
	orderItemQueryClient pborderitem.OrderItemQueryServiceClient,
	orderItemCommandClient pborderitem.OrderItemCommandServiceClient,
	guards ...GuardOptions,
) *Repositories {
	var g GuardOptions
	if len(guards) > 0 {
		g = guards[0]
	}
	return &Repositories{
		CashierQuery:                 cashieradapter.New(cashierClient, g.Cashier...),
		MerchantQuery:                merchantadapter.New(merchantClient, g.Merchant...),
		OrderQuery:                   orderadapter.New(orderClient, g.Order...),
		OrderItemQuery:               orderitemadapter.New(orderItemQueryClient, orderItemCommandClient, g.OrderItem...),
		TransactionCommandRepository: NewTransactionCommandRepository(db),
		TransactionQueryRepository:   NewTransactionQueryRepository(db),
	}
}
