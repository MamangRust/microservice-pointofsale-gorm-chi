package order_test

import (
	"context"
	"testing"

	order_cache "github.com/MamangRust/microservice-point-of-sale-order/cache"
	"github.com/MamangRust/microservice-point-of-sale-order/repository"
	"github.com/MamangRust/microservice-point-of-sale-order/service"
	pbcashier "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	pbmerchant "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	pborderitem "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
	pbproduct "github.com/MamangRust/microservice-point-of-sale-pb/product"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	tests "github.com/MamangRust/microservice-point-of-sale-test"

	"github.com/stretchr/testify/suite"
)

type OrderServiceTestSuite struct {
	tests.BaseTestSuite
	svc       *service.Service
	orderID   int
	cashierID int
	productID int
}

func (s *OrderServiceTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()

	queries := s.GormDB()

	s.SetupOrderService()

	var userID int
	err := s.GormDB().Raw(`INSERT INTO users (firstname, lastname, email, password, verification_code, is_verified) VALUES ($1, $2, $3, $4, 'test-verify', true) RETURNING user_id`,
		"Order", "Svc", "order.svc@example.com", "password123",
	).Scan(&userID).Error

	err = s.GormDB().Raw(`INSERT INTO merchants (user_id, name, description, address, contact_email, contact_phone, status) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING merchant_id`,
		userID, "Order Svc Merchant", "Desc", "Addr", "os@example.com", "123", "active",
	).Scan(&s.orderID).Error
	s.Require().NoError(err)

	err = s.GormDB().Raw(`INSERT INTO cashiers (merchant_id, user_id, name) VALUES ($1, $2, 'Order Cashier') RETURNING cashier_id`,
		s.orderID, userID,
	).Scan(&s.cashierID).Error
	s.Require().NoError(err)

	var categoryID int
	err = s.GormDB().Raw(`INSERT INTO categories (name, description) VALUES ('Order Category', 'Desc') RETURNING category_id`).Scan(&categoryID).Error
	s.Require().NoError(err)

	err = s.GormDB().Raw(`INSERT INTO products (merchant_id, category_id, name, description, price, count_in_stock, brand, weight, image_product)
		 VALUES ($1, $2, 'Order Product', 'Desc', 10000, 100, 'Brand', 100, 'img') RETURNING product_id`,
		s.orderID, categoryID,
	).Scan(&s.productID).Error
	s.Require().NoError(err)

	mencache := order_cache.NewMencache(s.GetCacheStore())
	repos := repository.NewRepositories(
		queries,
		pbcashier.NewCashierQueryServiceClient(s.Conns["cashier"]),
		pbmerchant.NewMerchantQueryServiceClient(s.Conns["merchant"]),
		pbproduct.NewProductQueryServiceClient(s.Conns["product"]),
		pbproduct.NewProductCommandServiceClient(s.Conns["product"]),
		pborderitem.NewOrderItemQueryServiceClient(s.Conns["order-item"]),
		pborderitem.NewOrderItemCommandServiceClient(s.Conns["order-item"]),
	)

	s.svc = service.NewService(&service.Deps{
		Repositories:  repos,
		Logger:        s.Log,
		Mencache:      mencache,
		Observability: s.Obs,
	})
}

func (s *OrderServiceTestSuite) TearDownSuite() {
	s.BaseTestSuite.TearDownSuite()
}

func (s *OrderServiceTestSuite) TestOrderLifecycle() {
	ctx := context.Background()

	req := &requests.CreateOrderRequest{
		MerchantID: s.orderID,
		CashierID:  s.cashierID,
		Items: []requests.CreateOrderItemRequest{
			{ProductID: s.productID, Quantity: 1},
		},
	}
	created, err := s.svc.OrderCommand.CreateOrder(ctx, req)
	s.Require().NoError(err)
	s.Require().NotNil(created)
	orderID := int(created.OrderID)

	var orderItemID int
	err = s.GormDB().Raw(`SELECT order_item_id FROM order_items WHERE order_id = $1 LIMIT 1`, orderID).Scan(&orderItemID).Error
	s.Require().NoError(err)

	found, err := s.svc.OrderQuery.FindById(ctx, orderID)
	s.Require().NoError(err)
	s.NotNil(found)

	updateReq := &requests.UpdateOrderRequest{
		OrderID: &orderID,
		Items: []requests.UpdateOrderItemRequest{
			{OrderItemID: orderItemID, ProductID: s.productID, Quantity: 2},
		},
	}
	updated, err := s.svc.OrderCommand.UpdateOrder(ctx, updateReq)
	s.Require().NoError(err)
	s.NotNil(updated)

	_, total, err := s.svc.OrderQuery.FindAll(ctx, &requests.FindAllOrders{Search: "", Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.GreaterOrEqual(*total, 1)

	_, err = s.svc.OrderCommand.TrashedOrder(ctx, orderID)
	s.Require().NoError(err)

	_, totalTrashed, err := s.svc.OrderQuery.FindByTrashed(ctx, &requests.FindAllOrders{Search: "", Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.GreaterOrEqual(*totalTrashed, 1)

	active, _, err := s.svc.OrderQuery.FindByActive(ctx, &requests.FindAllOrders{Search: "", Page: 1, PageSize: 10})
	s.Require().NoError(err)
	for _, o := range active {
		s.NotEqual(orderID, int(o.OrderID))
	}

	_, err = s.svc.OrderCommand.RestoreOrder(ctx, orderID)
	s.Require().NoError(err)

	_, err = s.svc.OrderCommand.TrashedOrder(ctx, orderID)
	s.Require().NoError(err)
	success, err := s.svc.OrderCommand.DeleteOrderPermanent(ctx, orderID)
	s.Require().NoError(err)
	s.True(success)

	o1, _ := s.svc.OrderCommand.CreateOrder(ctx, &requests.CreateOrderRequest{
		MerchantID: s.orderID, CashierID: s.cashierID,
		Items: []requests.CreateOrderItemRequest{{ProductID: s.productID, Quantity: 1}},
	})
	o2, _ := s.svc.OrderCommand.CreateOrder(ctx, &requests.CreateOrderRequest{
		MerchantID: s.orderID, CashierID: s.cashierID,
		Items: []requests.CreateOrderItemRequest{{ProductID: s.productID, Quantity: 2}},
	})

	s.svc.OrderCommand.TrashedOrder(ctx, int(o1.OrderID))
	s.svc.OrderCommand.TrashedOrder(ctx, int(o2.OrderID))

	resRestoreAll, err := s.svc.OrderCommand.RestoreAllOrder(ctx)
	s.Require().NoError(err)
	s.True(resRestoreAll)

	s.svc.OrderCommand.TrashedOrder(ctx, int(o1.OrderID))
	s.svc.OrderCommand.TrashedOrder(ctx, int(o2.OrderID))

	resDeleteAll, err := s.svc.OrderCommand.DeleteAllOrderPermanent(ctx)
	s.Require().NoError(err)
	s.True(resDeleteAll)
}

func TestOrderServiceSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(OrderServiceTestSuite))
}
