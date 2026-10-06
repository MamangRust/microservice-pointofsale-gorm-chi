package repository

import (
	pbcashier "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	pbmerchant "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	pborderitem "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
	pbproduct "github.com/MamangRust/microservice-point-of-sale-pb/product"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	cashieradapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/cashier"
	merchantadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/merchant"
	orderitemadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/orderitem"
	productadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/product"
	"gorm.io/gorm"
)

type Repositories struct {
	OrderQuery       OrderQueryRepository
	OrderCommand     OrderCommandRepository
	CashierQuery     CashierQueryRepository
	MerchantQuery    MerchantQueryRepository
	ProductQuery     ProductQueryRepository
	ProductCommand   ProductCommandRepository
	OrderItemQuery   OrderItemQueryRepository
	OrderItemCommand OrderItemCommandRepository
}

type GuardOptions struct {
	Cashier   []adapter.GuardOption
	Merchant  []adapter.GuardOption
	Product   []adapter.GuardOption
	OrderItem []adapter.GuardOption
}

func NewRepositories(
	db *gorm.DB,
	cashierClient pbcashier.CashierQueryServiceClient,
	merchantClient pbmerchant.MerchantQueryServiceClient,
	productQueryClient pbproduct.ProductQueryServiceClient,
	productCommandClient pbproduct.ProductCommandServiceClient,
	orderItemQueryClient pborderitem.OrderItemQueryServiceClient,
	orderItemCommandClient pborderitem.OrderItemCommandServiceClient,
	guards ...GuardOptions,
) *Repositories {
	var g GuardOptions
	if len(guards) > 0 {
		g = guards[0]
	}
	productRepo := productadapter.New(productQueryClient, productCommandClient, g.Product...)
	orderItemRepo := orderitemadapter.New(orderItemQueryClient, orderItemCommandClient, g.OrderItem...)
	return &Repositories{
		OrderQuery:       NewOrderQueryRepository(db),
		OrderCommand:     NewOrderCommandRepository(db),
		CashierQuery:     cashieradapter.New(cashierClient, g.Cashier...),
		MerchantQuery:    merchantadapter.New(merchantClient, g.Merchant...),
		ProductQuery:     productRepo,
		ProductCommand:   productRepo,
		OrderItemQuery:   orderItemRepo,
		OrderItemCommand: orderItemRepo,
	}
}
