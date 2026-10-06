package merchant_test

import (
	"context"
	"testing"

	merchant_cache "github.com/MamangRust/microservice-point-of-sale-merchant/cache"
	"github.com/MamangRust/microservice-point-of-sale-merchant/handler"
	"github.com/MamangRust/microservice-point-of-sale-merchant/repository"
	"github.com/MamangRust/microservice-point-of-sale-merchant/service"
	pbmerchant "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	pbuser "github.com/MamangRust/microservice-point-of-sale-pb/user"
	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	tests "github.com/MamangRust/microservice-point-of-sale-test"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type MerchantGapiTestSuite struct {
	tests.BaseTestSuite
	client     *grpc.ClientConn
	userID     int
	merchantID int
}

func (s *MerchantGapiTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()
	s.SetupUserService()
	merchantQueries := s.GormDB()
	repos := repository.NewRepositories(merchantQueries,
		pbuser.NewUserQueryServiceClient(s.Conns["user"]),
		pbuser.NewUserCommandServiceClient(s.Conns["user"]),
	)

	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.RedisClient(), s.Log, cacheMetrics)
	mencache := merchant_cache.NewMencache(cacheStore)

	svc := service.NewService(&service.Deps{
		Kafka:         nil,
		Repositories:  repos,
		Logger:        s.Log,
		Mencache:      mencache,
		Observability: s.Obs,
	})

	merchantHandlers, merchantDocHandlers := handler.NewHandler(svc)
	server := grpc.NewServer()
	pbmerchant.RegisterMerchantQueryServiceServer(server, merchantHandlers)
	pbmerchant.RegisterMerchantCommandServiceServer(server, merchantHandlers)
	_ = merchantDocHandlers

	addr := s.RegisterServer(server)
	conn := s.GetConnection(addr)
	s.client = conn
	s.userID = s.SeedUser(context.Background())
}

func (s *MerchantGapiTestSuite) TestMerchantGapiLifecycle() {
	ctx := context.Background()

	cmdClient := pbmerchant.NewMerchantCommandServiceClient(s.client)
	queryClient := pbmerchant.NewMerchantQueryServiceClient(s.client)

	// 1. Create
	createReq := &pbmerchant.CreateMerchantRequest{
		UserId:       int32(s.userID),
		Name:         "Gapi Merchant",
		Description:  "Detailed description of the merchant.",
		Address:      "Merchant Street No. 1",
		ContactEmail: "gapi.merchant@example.com",
		ContactPhone: "08123456789",
		Status:       "active",
	}
	res, err := cmdClient.Create(ctx, createReq)
	s.NoError(err)
	s.Equal(createReq.Name, res.Data.Name)
	merchantID := res.Data.Id

	// 2. FindById
	found, err := queryClient.FindById(ctx, &pbmerchant.FindByIdMerchantRequest{Id: merchantID})
	s.NoError(err)
	s.Equal(merchantID, found.Data.Id)

	// 3. FindAll
	allRes, err := queryClient.FindAll(ctx, &pbmerchant.FindAllMerchantRequest{Page: 1, PageSize: 10})
	s.NoError(err)
	s.NotEmpty(allRes.Data)

	// 4. FindByActive
	activeRes, err := queryClient.FindByActive(ctx, &pbmerchant.FindAllMerchantRequest{Page: 1, PageSize: 10})
	s.NoError(err)
	s.NotEmpty(activeRes.Data)

	// 5. Update
	updateReq := &pbmerchant.UpdateMerchantRequest{
		MerchantId:   merchantID,
		UserId:       int32(s.userID),
		Name:         "Gapi Merchant Updated",
		Description:  "Updated description.",
		Address:      "New Street 2",
		ContactEmail: "updated@example.com",
		ContactPhone: "08987654321",
		Status:       "waiting",
	}
	updateRes, err := cmdClient.Update(ctx, updateReq)
	s.NoError(err)
	s.Equal(updateReq.Name, updateRes.Data.Name)

	// 6. Update Status
	statusRes, err := cmdClient.UpdateMerchantStatus(ctx, &pbmerchant.UpdateMerchantStatusRequest{
		MerchantId: merchantID,
		Status:     "active",
	})
	s.NoError(err)
	s.Equal("active", statusRes.Data.Status)

	// 7. Trash
	_, err = cmdClient.TrashedMerchant(ctx, &pbmerchant.FindByIdMerchantRequest{Id: merchantID})
	s.NoError(err)

	// 8. FindByTrashed
	trashedRes, err := queryClient.FindByTrashed(ctx, &pbmerchant.FindAllMerchantRequest{Page: 1, PageSize: 10})
	s.NoError(err)
	s.NotEmpty(trashedRes.Data)

	// 9. Restore
	_, err = cmdClient.RestoreMerchant(ctx, &pbmerchant.FindByIdMerchantRequest{Id: merchantID})
	s.NoError(err)

	// 10. DeletePermanent
	_, _ = cmdClient.TrashedMerchant(ctx, &pbmerchant.FindByIdMerchantRequest{Id: merchantID})
	_, err = cmdClient.DeleteMerchantPermanent(ctx, &pbmerchant.FindByIdMerchantRequest{Id: merchantID})
	s.NoError(err)

	// 11. RestoreAll
	_, err = cmdClient.RestoreAllMerchant(ctx, &emptypb.Empty{})
	s.NoError(err)

	// 12. DeleteAll
	_, err = cmdClient.DeleteAllMerchantPermanent(ctx, &emptypb.Empty{})
	s.NoError(err)
}

func TestMerchantGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(MerchantGapiTestSuite))
}
