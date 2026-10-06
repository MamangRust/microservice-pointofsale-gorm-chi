package tests

import (
	"context"
	"database/sql"
	"reflect"

	"github.com/testcontainers/testcontainers-go"

	pbcategory "github.com/MamangRust/microservice-point-of-sale-pb/category"
	pbmerchant "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	pborder "github.com/MamangRust/microservice-point-of-sale-pb/order"
	pbproduct "github.com/MamangRust/microservice-point-of-sale-pb/product"
	pbuser "github.com/MamangRust/microservice-point-of-sale-pb/user"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	goredis "github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"gorm.io/gorm"
)

type BaseTestSuite struct {
	suite.Suite
	ts          *TestSuite
	Log         logger.LoggerInterface
	Obs         observability.TraceLoggerObservability
	Conns       map[string]*grpc.ClientConn
	Servers     []*grpc.Server
	Ctx         context.Context
	Cancel      context.CancelFunc
	CHContainer testcontainers.Container
}

func (s *BaseTestSuite) SetupSuite() {
	s.Ctx, s.Cancel = context.WithCancel(context.Background())

	ts, err := SetupTestSuite()
	s.ts = ts

	s.Log, _ = logger.NewLogger("test", nil)

	if s.Log == nil || (reflect.ValueOf(s.Log).Kind() == reflect.Ptr && reflect.ValueOf(s.Log).IsNil()) {
		z, _ := zap.NewDevelopment()
		s.Log = &logger.Logger{Log: z}
	}

	s.Obs, err = observability.NewObservability("test", s.Log)
	s.Require().NoError(err)
	s.Require().NotNil(s.Obs)
	s.Conns = make(map[string]*grpc.ClientConn)
}

func (s *BaseTestSuite) TearDownSuite() {
	for _, conn := range s.Conns {
		conn.Close()
	}
	for _, server := range s.Servers {
		server.GracefulStop()
	}
	if s.CHContainer != nil {
		if err := s.CHContainer.Terminate(s.Ctx); err != nil {
			s.Log.Error("failed to terminate ClickHouse container", zap.Error(err))
		}
	}
	if s.ts != nil {
		s.ts.Teardown()
	}
	if s.Cancel != nil {
		s.Cancel()
	}
}

func (s *BaseTestSuite) GormDB() *gorm.DB {
	return s.ts.GormDB()
}

// SQLxDB returns a *sql.DB (database/sql) for raw SQL fixtures that need
// QueryRowContext/ExecContext rather than the GORM wrapper. The lib/pq driver
// is registered by the test_setup package, so the "postgres" driver is valid.
func (s *BaseTestSuite) SQLxDB() *sql.DB {
	db, err := sql.Open("postgres", s.ts.DBURL)
	s.Require().NoError(err)
	return db
}

func (s *BaseTestSuite) RedisClient() *goredis.Client {
	return s.ts.RedisClient()
}

func (s *BaseTestSuite) RegisterServer(server *grpc.Server) string {
	addr, err := RunGRPCServer(server)
	s.Require().NoError(err)
	s.Servers = append(s.Servers, server)
	return addr
}

func (s *BaseTestSuite) GetConnection(addr string) *grpc.ClientConn {
	conn, err := grpc.Dial(addr, grpc.WithInsecure())
	s.Require().NoError(err)
	return conn
}

func (s *BaseTestSuite) SeedUser(ctx context.Context) int {
	err := s.GormDB().Exec(`
		INSERT INTO roles (role_name, created_at, updated_at) 
		VALUES ('Admin Access 1', current_timestamp, current_timestamp),
		       ('ROLE_ADMIN', current_timestamp, current_timestamp)
		ON CONFLICT (role_name) DO NOTHING
	`).Error
	s.Require().NoError(err)

	res, err := pbuser.NewUserCommandServiceClient(s.Conns["user"]).Create(ctx, &pbuser.CreateUserRequest{
		Firstname:       "Seed",
		Lastname:        "User",
		Email:           "seed.user@example.com",
		Password:        "password123",
		ConfirmPassword: "password123",
	})
	s.Require().NoError(err)
	return int(res.Data.Id)
}

func (s *BaseTestSuite) SeedCategory(ctx context.Context) int {
	res, err := pbcategory.NewCategoryCommandServiceClient(s.Conns["category"]).Create(ctx, &pbcategory.CreateCategoryRequest{
		Name:        "Seed Category",
		Description: "Seed Description",
	})
	s.Require().NoError(err)
	return int(res.Data.Id)
}

func (s *BaseTestSuite) SeedMerchant(ctx context.Context, userID int) int {
	res, err := pbmerchant.NewMerchantCommandServiceClient(s.Conns["merchant"]).Create(ctx, &pbmerchant.CreateMerchantRequest{
		UserId:       int32(userID),
		Name:         "Seed Merchant",
		Description:  "Seed Description",
		Address:      "Seed Address",
		ContactEmail: "merchant@example.com",
		ContactPhone: "08123456789",
		Status:       "active",
	})
	s.Require().NoError(err)
	return int(res.Data.Id)
}

func (s *BaseTestSuite) SeedProduct(ctx context.Context, merchantID int, categoryID int) int {
	res, err := pbproduct.NewProductCommandServiceClient(s.Conns["product"]).Create(ctx, &pbproduct.CreateProductRequest{
		MerchantId:   int32(merchantID),
		CategoryId:   int32(categoryID),
		Name:         "Seed Product",
		Description:  "Seed Description",
		Price:        10000,
		CountInStock: 100,
		Brand:        "Seed Brand",
		Weight:       1000,
		ImageProduct: "seed.jpg",
	})
	s.Require().NoError(err)
	return int(res.Data.Id)
}

func (s *BaseTestSuite) SeedOrder(ctx context.Context, userID int, merchID int, prodID int) int {
	var cashierID int
	lookup := s.GormDB().Raw(`
		SELECT cashier_id FROM cashiers WHERE user_id = $1 AND merchant_id = $2 AND deleted_at IS NULL LIMIT 1
	`, userID, merchID).Scan(&cashierID)
	s.Require().NoError(lookup.Error)
	if lookup.RowsAffected == 0 {
		err := s.GormDB().Raw(`
			INSERT INTO cashiers (merchant_id, user_id, name, created_at, updated_at)
			VALUES ($1, $2, 'Seed Cashier', current_timestamp, current_timestamp)
			RETURNING cashier_id
		`, merchID, userID).Scan(&cashierID).Error
		s.Require().NoError(err)
	}

	res, err := pborder.NewOrderCommandServiceClient(s.Conns["order"]).Create(ctx, &pborder.CreateOrderRequest{
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
	return int(res.Data.Id)
}

func (s *BaseTestSuite) SeedOrderItem(ctx context.Context, orderID int, productID int) int {
	var id int
	err := s.GormDB().Raw(`INSERT INTO "order_items" (order_id, product_id, quantity, price) VALUES ($1, $2, 1, 1000) RETURNING order_item_id`,
		orderID, productID,
	).Scan(&id).Error
	s.Require().NoError(err)
	return id
}
