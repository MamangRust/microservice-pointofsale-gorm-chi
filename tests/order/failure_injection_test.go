package order_test

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	mencache "github.com/MamangRust/microservice-point-of-sale-order/cache"
	"github.com/MamangRust/microservice-point-of-sale-order/handler"
	"github.com/MamangRust/microservice-point-of-sale-order/repository"
	"github.com/MamangRust/microservice-point-of-sale-order/service"
	pbmerchant "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	pborder "github.com/MamangRust/microservice-point-of-sale-pb/order"
	pborderitem "github.com/MamangRust/microservice-point-of-sale-pb/order_item"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	merchantadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/merchant"
	orderitemadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/orderitem"
	"github.com/MamangRust/microservice-point-of-sale-pkg/resilience"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	tests "github.com/MamangRust/microservice-point-of-sale-test"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

type faultMerchantServer struct {
	pbmerchant.MerchantQueryServiceServer
	calls int32
}

func (f *faultMerchantServer) FindById(ctx context.Context, req *pbmerchant.FindByIdMerchantRequest) (*pbmerchant.ApiResponseMerchant, error) {
	atomic.AddInt32(&f.calls, 1)
	return nil, status.Error(codes.Unavailable, "merchant service unavailable (injected)")
}

func (f *faultMerchantServer) callCount() int {
	return int(atomic.LoadInt32(&f.calls))
}

type OrderFailureInjectionTestSuite struct {
	tests.BaseTestSuite
	merchantServer *faultMerchantServer
	svc            *service.Service
	client         *grpc.ClientConn
	orderID        int
	cashierID      int
}

func (s *OrderFailureInjectionTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()

	queries := s.GormDB()

	var userID int
	err := s.GormDB().Raw(`INSERT INTO users (firstname, lastname, email, password, verification_code, is_verified) VALUES ($1, $2, $3, $4, 'test-verify', true) RETURNING user_id`,
		"Fail", "Inject", "fail.inject@example.com", "password123",
	).Scan(&userID).Error

	err = s.GormDB().Raw(`INSERT INTO merchants (user_id, name, description, address, contact_email, contact_phone, status) VALUES ($1, $2, $3, $4, $5, $6, $7) RETURNING merchant_id`,
		userID, "Failure Inject Merchant", "Desc", "Addr", "fi@example.com", "123", "active",
	).Scan(&s.orderID).Error
	s.Require().NoError(err)

	err = s.GormDB().Raw(`INSERT INTO cashiers (merchant_id, user_id, name) VALUES ($1, $2, 'Failure Cashier') RETURNING cashier_id`,
		s.orderID, userID,
	).Scan(&s.cashierID).Error
	s.Require().NoError(err)

	s.merchantServer = &faultMerchantServer{}
	merchantGRPC := grpc.NewServer()
	pbmerchant.RegisterMerchantQueryServiceServer(merchantGRPC, s.merchantServer)
	addr, err := tests.RunGRPCServer(merchantGRPC)
	s.Require().NoError(err)
	merchantConn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.Servers = append(s.Servers, merchantGRPC)
	s.Conns["merchant"] = merchantConn

	merchantGuard := resilience.NewDependencyGuard("merchant", 5, 30, 100, 3*time.Second, s.Log)

	repos := &repository.Repositories{
		CashierQuery: &stubCashierRepo{cashierID: s.cashierID},
		MerchantQuery: merchantadapter.New(
			pbmerchant.NewMerchantQueryServiceClient(merchantConn),
			adapter.WithDependencyGuard(merchantGuard),
		),
		ProductQuery:   &stubProductRepo{},
		ProductCommand: &stubProductCommandRepo{},
		OrderQuery:     repository.NewOrderQueryRepository(queries),
		OrderCommand:   repository.NewOrderCommandRepository(queries),
		OrderItemQuery: orderitemadapter.New(
			pborderitem.NewOrderItemQueryServiceClient(merchantConn),
			pborderitem.NewOrderItemCommandServiceClient(merchantConn),
		),
		OrderItemCommand: &stubOrderItemCommandRepo{},
	}

	mencacheObj := mencache.NewMencache(s.GetCacheStore())
	s.svc = service.NewService(&service.Deps{
		Repositories:  repos,
		Logger:        s.Log,
		Mencache:      mencacheObj,
		Observability: s.Obs,
	})

	orderGapi := handler.NewHandler(s.svc)
	orderGRPC := grpc.NewServer()
	pborder.RegisterOrderQueryServiceServer(orderGRPC, orderGapi)
	pborder.RegisterOrderCommandServiceServer(orderGRPC, orderGapi)
	orderAddr, err := tests.RunGRPCServer(orderGRPC)
	s.Require().NoError(err)
	s.Servers = append(s.Servers, orderGRPC)
	orderConn, err := grpc.NewClient(orderAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.Conns["order"] = orderConn
	s.client = orderConn
}

func (s *OrderFailureInjectionTestSuite) TearDownSuite() {
	s.BaseTestSuite.TearDownSuite()
}

func (s *OrderFailureInjectionTestSuite) TestF1_InvalidRequest_RejectedWith400_BeforeRemoteCalls() {
	ctx := context.Background()

	orderCmdClient := pborder.NewOrderCommandServiceClient(s.client)
	_, err := orderCmdClient.Create(ctx, &pborder.CreateOrderRequest{
		MerchantId: int32(s.orderID),
		CashierId:  int32(s.cashierID),
	})

	s.Require().Error(err, "empty items must be rejected")
	s.Equal(codes.InvalidArgument, status.Code(err), "validation failure must map to InvalidArgument (400)")
	s.Zero(s.merchantServer.callCount(), "no remote call may happen before validation")
}

func (s *OrderFailureInjectionTestSuite) TestF2_DependencyDown_OpensCircuit_ThenFailsFast() {
	ctx := context.Background()
	req := &requests.CreateOrderRequest{
		MerchantID: s.orderID,
		CashierID:  s.cashierID,
		Items: []requests.CreateOrderItemRequest{
			{ProductID: 1, Quantity: 1},
		},
	}

	baseline := s.merchantServer.callCount()

	for i := 0; i < 5; i++ {
		_, err := s.svc.OrderCommand.CreateOrder(ctx, req)
		s.Require().Error(err, "dependency down must fail the create (%d)", i+1)
	}
	s.Equal(baseline+5, s.merchantServer.callCount(), "all 5 failures must reach the dependency")

	afterOpen := s.merchantServer.callCount()
	_, err := s.svc.OrderCommand.CreateOrder(ctx, req)
	s.Require().Error(err, "open circuit must fail fast with an error")
	s.Equal(afterOpen, s.merchantServer.callCount(), "fail-fast must not reach the dependency")
}

func TestOrderFailureInjectionSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(OrderFailureInjectionTestSuite))
}
