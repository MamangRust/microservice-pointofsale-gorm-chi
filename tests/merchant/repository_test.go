package merchant_test

import (
	"context"
	"testing"

	"github.com/MamangRust/microservice-point-of-sale-merchant/repository"
	pbuser "github.com/MamangRust/microservice-point-of-sale-pb/user"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	tests "github.com/MamangRust/microservice-point-of-sale-test"

	"net"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type MerchantRepositoryTestSuite struct {
	suite.Suite
	ts         *tests.TestSuite
	repo       *repository.Repositories
	merchantID int
	userID     int
}

func (s *MerchantRepositoryTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.ts = ts

	s.Require().NoError(err)

	userLis, _ := net.Listen("tcp", "localhost:0")
	userConn, _ := grpc.NewClient(userLis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))

	merchantQueries := s.ts.GormDB()
	s.repo = repository.NewRepositories(merchantQueries,
		pbuser.NewUserQueryServiceClient(userConn),
		pbuser.NewUserCommandServiceClient(userConn),
	)

	err = s.ts.GormDB().Raw(`INSERT INTO users (firstname, lastname, email, password, verification_code, is_verified) VALUES ($1, $2, $3, $4, 'test-verify', true) RETURNING user_id`,
		"Merchant", "RepoTest", "merchant.repo@example.com", "password123",
	).Scan(&s.userID).Error
	s.Require().NoError(err)
}

func (s *MerchantRepositoryTestSuite) TearDownSuite() {
	s.ts.Teardown()
}

func (s *MerchantRepositoryTestSuite) Test1_CreateMerchant() {
	ctx := context.Background()

	req := &requests.CreateMerchantRequest{
		UserID:       s.userID,
		Name:         "Test Merchant",
		Description:  "Detailed description",
		Address:      "Merchant Street No. 1",
		ContactEmail: "merchant@example.com",
		ContactPhone: "08123456789",
		Status:       "active",
	}

	res, err := s.repo.MerchantCommand.CreateMerchant(ctx, req)
	s.NoError(err)
	s.NotNil(res)
	s.Equal(req.Name, res.Name)
	s.Equal(int32(s.userID), res.UserID)
	s.merchantID = int(res.MerchantID)
}

func (s *MerchantRepositoryTestSuite) Test2_FindById() {
	s.Require().NotZero(s.merchantID)
	ctx := context.Background()

	found, err := s.repo.MerchantQuery.FindById(ctx, s.merchantID)
	s.NoError(err)
	s.NotNil(found)
	s.Equal(s.merchantID, int(found.MerchantID))
}

func (s *MerchantRepositoryTestSuite) Test3_UpdateMerchant() {
	s.Require().NotZero(s.merchantID)
	ctx := context.Background()

	req := &requests.UpdateMerchantRequest{
		MerchantID:   &s.merchantID,
		UserID:       s.userID,
		Name:         "Updated Merchant",
		Description:  "Updated description",
		Address:      "Updated Street",
		ContactEmail: "updated@example.com",
		ContactPhone: "08987654321",
		Status:       "active",
	}

	res, err := s.repo.MerchantCommand.UpdateMerchant(ctx, req)
	s.NoError(err)
	s.NotNil(res)
	s.Equal("Updated Merchant", res.Name)
}

func (s *MerchantRepositoryTestSuite) Test4_TrashAndRestore() {
	s.Require().NotZero(s.merchantID)
	ctx := context.Background()

	trashed, err := s.repo.MerchantCommand.TrashedMerchant(ctx, s.merchantID)
	s.NoError(err)
	s.NotNil(trashed)

	restored, err := s.repo.MerchantCommand.RestoreMerchant(ctx, s.merchantID)
	s.NoError(err)
	s.NotNil(restored)

	found, err := s.repo.MerchantQuery.FindById(ctx, s.merchantID)
	s.NoError(err)
	s.NotNil(found)
}

func (s *MerchantRepositoryTestSuite) Test5_DeletePermanent() {
	s.Require().NotZero(s.merchantID)
	ctx := context.Background()

	_, err := s.repo.MerchantCommand.TrashedMerchant(ctx, s.merchantID)
	s.NoError(err)

	success, err := s.repo.MerchantCommand.DeleteMerchantPermanent(ctx, s.merchantID)
	s.NoError(err)
	s.True(success)
}

func TestMerchantRepositorySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(MerchantRepositoryTestSuite))
}
