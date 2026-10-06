package order_test

import (
	"context"
	"testing"

	order_cache "github.com/MamangRust/microservice-point-of-sale-order/cache"
	order_handler "github.com/MamangRust/microservice-point-of-sale-order/handler"
	order_repo "github.com/MamangRust/microservice-point-of-sale-order/repository"
	order_service "github.com/MamangRust/microservice-point-of-sale-order/service"
	pbcashier "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	pbmerchant "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	pborder "github.com/MamangRust/microservice-point-of-sale-pb/order"
	pborderitem "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
	pbproduct "github.com/MamangRust/microservice-point-of-sale-pb/product"
	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	tests "github.com/MamangRust/microservice-point-of-sale-test"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type OrderGapiTestSuite struct {
	tests.BaseTestSuite
	client *grpc.ClientConn
}

func (s *OrderGapiTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()

	// Setup dependencies
	s.SetupUserService()
	s.SetupCategoryService()
	s.SetupMerchantService()
	s.SetupProductService()
	s.SetupOrderItemService()
	s.SetupCashierService()
	s.SetupTransactionService()

	// Infrastructure
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.RedisClient(), s.Log, cacheMetrics)
	queries := s.GormDB()

	// Order dependencies
	mencache := order_cache.NewMencache(cacheStore)
	repos := order_repo.NewRepositories(
		queries,
		pbcashier.NewCashierQueryServiceClient(s.Conns["cashier"]),
		pbmerchant.NewMerchantQueryServiceClient(s.Conns["merchant"]),
		pbproduct.NewProductQueryServiceClient(s.Conns["product"]),
		pbproduct.NewProductCommandServiceClient(s.Conns["product"]),
		pborderitem.NewOrderItemQueryServiceClient(s.Conns["order-item"]),
		pborderitem.NewOrderItemCommandServiceClient(s.Conns["order-item"]),
	)
	svc := order_service.NewService(&order_service.Deps{
		Mencache:      mencache,
		Repositories:  repos,
		Logger:        s.Log,
		Observability: s.Obs,
	})

	// Handler
	handlers := order_handler.NewHandler(svc)

	// Server
	server := grpc.NewServer()
	pborder.RegisterOrderQueryServiceServer(server, handlers)
	pborder.RegisterOrderCommandServiceServer(server, handlers)

	addr := s.RegisterServer(server)
	conn := s.GetConnection(addr)

	s.client = conn
}

func (s *OrderGapiTestSuite) TestOrderGapiLifecycle() {
	ctx := context.Background()

	// 1. Seed dependencies
	userID := s.SeedUser(ctx)
	catID := s.SeedCategory(ctx)
	merchID := s.SeedMerchant(ctx, userID)
	prodID := s.SeedProduct(ctx, merchID, catID)

	// Seed a cashier (orders.cashier_id references cashiers.cashier_id)
	var cashierID int
	err := s.GormDB().Raw(`INSERT INTO cashiers (merchant_id, user_id, name) VALUES ($1, $2, 'Order Gapi Cashier') RETURNING cashier_id`,
		merchID, userID,
	).Scan(&cashierID).Error

	cmdClient := pborder.NewOrderCommandServiceClient(s.client)
	queryClient := pborder.NewOrderQueryServiceClient(s.client)

	// 2. Create
	createRes, err := cmdClient.Create(ctx, &pborder.CreateOrderRequest{
		MerchantId: int32(merchID),
		CashierId:  int32(cashierID),
		Items: []*pborder.CreateOrderItemRequest{
			{
				ProductId: int32(prodID),
				Quantity:  1,
			},
		},
	})
	s.Require().NoError(err)
	s.Require().NotNil(createRes)
	orderID := createRes.Data.Id

	// 3. FindById
	getRes, err := queryClient.FindById(ctx, &pborder.FindByIdOrderRequest{Id: orderID})
	s.Require().NoError(err)
	s.Equal(int32(userID), getRes.Data.CashierId)

	// 4. FindAll
	allRes, err := queryClient.FindAll(ctx, &pborder.FindAllOrderRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(allRes.Data)

	// 5. FindByActive
	activeRes, err := queryClient.FindByActive(ctx, &pborder.FindAllOrderRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(activeRes.Data)

	// 6. Update
	// Fetch order items first
	itemClient := pborderitem.NewOrderItemQueryServiceClient(s.Conns["order-item"])
	itemsRes, err := itemClient.FindOrderItemByOrder(ctx, &pborderitem.FindByIdOrderItemRequest{Id: orderID})
	s.Require().NoError(err)
	s.NotEmpty(itemsRes.Data)
	orderItemID := itemsRes.Data[0].Id

	_, err = cmdClient.Update(ctx, &pborder.UpdateOrderRequest{
		OrderId: orderID,
		Items: []*pborder.UpdateOrderItemRequest{
			{
				OrderItemId: orderItemID,
				ProductId:   int32(prodID),
				Quantity:    1,
			},
		},
	})
	s.Require().NoError(err)

	// 7. Trash
	_, err = cmdClient.TrashedOrder(ctx, &pborder.FindByIdOrderRequest{Id: orderID})
	s.Require().NoError(err)

	// 8. FindByTrashed
	trashedRes, err := queryClient.FindByTrashed(ctx, &pborder.FindAllOrderRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(trashedRes.Data)

	// 9. Restore
	_, err = cmdClient.RestoreOrder(ctx, &pborder.FindByIdOrderRequest{Id: orderID})
	s.Require().NoError(err)

	// 10. DeletePermanent
	_, _ = cmdClient.TrashedOrder(ctx, &pborder.FindByIdOrderRequest{Id: orderID})
	_, err = cmdClient.DeleteOrderPermanent(ctx, &pborder.FindByIdOrderRequest{Id: orderID})
	s.Require().NoError(err)

	// 11. RestoreAll
	_, err = cmdClient.RestoreAllOrder(ctx, &emptypb.Empty{})
	s.Require().NoError(err)

	// 12. DeleteAll
	_, err = cmdClient.DeleteAllOrderPermanent(ctx, &emptypb.Empty{})
	s.Require().NoError(err)
}

func TestOrderGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(OrderGapiTestSuite))
}
