package order_item_test

import (
	"context"
	"testing"

	item_cache "github.com/MamangRust/microservice-point-of-sale-order-item/cache"
	item_handler "github.com/MamangRust/microservice-point-of-sale-order-item/handler"
	item_repo "github.com/MamangRust/microservice-point-of-sale-order-item/repository"
	item_service "github.com/MamangRust/microservice-point-of-sale-order-item/service"
	pborderitem "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	tests "github.com/MamangRust/microservice-point-of-sale-test"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
)

type OrderItemGapiTestSuite struct {
	tests.BaseTestSuite
	client *grpc.ClientConn
}

func (s *OrderItemGapiTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()

	// Setup dependencies
	s.SetupRoleService()
	s.SetupUserService()
	s.SetupCategoryService()
	s.SetupMerchantService()
	s.SetupProductService()
	s.SetupOrderItemService()
	s.SetupOrderService()

	// Infrastructure
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.RedisClient(), s.Log, cacheMetrics)
	order_itemQueries := s.GormDB()

	// Item dependencies
	mencache := item_cache.NewMencache(cacheStore)
	repos := item_repo.NewRepositories(order_itemQueries)
	svc := item_service.NewService(&item_service.Deps{
		Mencache:      mencache,
		Repositories:  repos,
		Logger:        s.Log,
		Observability: s.Obs,
	})

	// Handler
	handlers := item_handler.NewHandler(svc, s.Log)

	// Server
	server := grpc.NewServer()
	pborderitem.RegisterOrderItemQueryServiceServer(server, handlers)
	pborderitem.RegisterOrderItemCommandServiceServer(server, handlers)

	addr := s.RegisterServer(server)
	conn := s.GetConnection(addr)

	s.client = conn
}

func (s *OrderItemGapiTestSuite) TestOrderItemGapiLifecycle() {
	ctx := context.Background()

	// 1. Seed dependencies
	userID := s.SeedUser(ctx)
	catID := s.SeedCategory(ctx)
	merchantID := s.SeedMerchant(ctx, userID)
	productID := s.SeedProduct(ctx, merchantID, catID)
	orderID := s.SeedOrder(ctx, userID, merchantID, productID)

	// 2. Create an order item directly in DB for query testing
	var orderItemID int
	err := s.GormDB().Raw(`INSERT INTO order_items (order_id, product_id, quantity, price) VALUES ($1, $2, $3, $4) RETURNING order_item_id`,
		orderID, productID, 10, 700,
	).Scan(&orderItemID).Error

	queryClient := pborderitem.NewOrderItemQueryServiceClient(s.client)

	// 3. FindOrderItemByOrder
	findByOrderRes, err := queryClient.FindOrderItemByOrder(ctx, &pborderitem.FindByIdOrderItemRequest{Id: int32(orderID)})
	s.Require().NoError(err)
	s.NotEmpty(findByOrderRes.Data)

	// 4. FindAll
	allRes, err := queryClient.FindAll(ctx, &pborderitem.FindAllOrderItemRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(allRes.Data)

	// 5. FindByActive
	activeRes, err := queryClient.FindByActive(ctx, &pborderitem.FindAllOrderItemRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(activeRes.Data)

	// 6. Trash the order item directly in DB
	s.Require().NoError(s.GormDB().Exec(`UPDATE order_items SET deleted_at = NOW() WHERE order_item_id = $1`, orderItemID).Error)

	// 7. FindByTrashed
	trashedRes, err := queryClient.FindByTrashed(ctx, &pborderitem.FindAllOrderItemRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(trashedRes.Data)

	// 8. FindAll again to verify
	_, err = queryClient.FindAll(ctx, &pborderitem.FindAllOrderItemRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
}

func TestOrderItemGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(OrderItemGapiTestSuite))
}
