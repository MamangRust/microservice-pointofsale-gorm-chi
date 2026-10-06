package apps

import (
	"context"
	"os"
	"time"

	mencache "github.com/MamangRust/microservice-point-of-sale-cashier/cache"
	"github.com/MamangRust/microservice-point-of-sale-cashier/handler"
	"github.com/MamangRust/microservice-point-of-sale-cashier/repository"
	"github.com/MamangRust/microservice-point-of-sale-cashier/service"
	pbcashier "github.com/MamangRust/microservice-point-of-sale-pb/cashier"
	pbmerchant "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	pbuser "github.com/MamangRust/microservice-point-of-sale-pb/user"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	"github.com/MamangRust/microservice-point-of-sale-pkg/resilience"
	"github.com/MamangRust/microservice-point-of-sale-pkg/server"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

func NewServer(cfg *server.Config) (*server.GRPCServer, error) {
	srv, err := server.New(cfg)
	if err != nil {
		return nil, err
	}

	userAddr := os.Getenv("GRPC_USER_ADDR")
	if userAddr == "" {
		userAddr = "localhost:50053"
	}
	merchantAddr := os.Getenv("GRPC_MERCHANT_ADDR")
	if merchantAddr == "" {
		merchantAddr = "localhost:50056"
	}

	srv.Logger.Info("Connecting to User service via gRPC", zap.String("addr", userAddr))
	userConn, err := server.NewGRPCClient(userAddr)
	if err != nil {
		return nil, err
	}

	srv.Logger.Info("Connecting to Merchant service via gRPC", zap.String("addr", merchantAddr))
	merchantConn, err := server.NewGRPCClient(merchantAddr)
	if err != nil {
		userConn.Close()
		return nil, err
	}

	go func() {
		<-srv.Ctx.Done()
		srv.Logger.Info("Closing cashier service remote gRPC connections")
		userConn.Close()
		merchantConn.Close()
	}()

	userQueryClient := pbuser.NewUserQueryServiceClient(userConn)
	userCommandClient := pbuser.NewUserCommandServiceClient(userConn)
	merchantClient := pbmerchant.NewMerchantQueryServiceClient(merchantConn)

	guardUser := resilience.NewDependencyGuard("user", 5, 30, 100, 3*time.Second, srv.Logger)
	guardMerchant := resilience.NewDependencyGuard("merchant", 5, 30, 100, 3*time.Second, srv.Logger)

	repos := repository.NewRepositories(srv.GormDB, userQueryClient, userCommandClient, merchantClient,
		repository.GuardOptions{
			User:     []adapter.GuardOption{adapter.WithDependencyGuard(guardUser)},
			Merchant: []adapter.GuardOption{adapter.WithDependencyGuard(guardMerchant)},
		},
	)
	traceLoggerObservability := observability.NewTraceLoggerObservability(srv.Logger)

	mencacheObj := mencache.NewMencache(srv.CacheStore)

	services := service.NewService(&service.Deps{
		Ctx:           context.Background(),
		Mencache:      mencacheObj,
		Repositories:  repos,
		Logger:        srv.Logger,
		Observability: traceLoggerObservability,
	})

	handlers := handler.NewHandler(services)

	srv.RegisterServices = func(gs *grpc.Server) {
		pbcashier.RegisterCashierQueryServiceServer(gs, handlers)
		pbcashier.RegisterCashierCommandServiceServer(gs, handlers)
	}

	return srv, nil
}
