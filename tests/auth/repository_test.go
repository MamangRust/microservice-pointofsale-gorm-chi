package auth_test

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-auth/repository"
	pbrole "github.com/MamangRust/microservice-point-of-sale-pb/role"
	pbuser "github.com/MamangRust/microservice-point-of-sale-pb/user"
	pbuserrole "github.com/MamangRust/microservice-point-of-sale-pb/user_role"
	"github.com/MamangRust/microservice-point-of-sale-pkg/hash"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	role_cache "github.com/MamangRust/microservice-point-of-sale-role/cache"
	role_handler "github.com/MamangRust/microservice-point-of-sale-role/handler"
	role_repo "github.com/MamangRust/microservice-point-of-sale-role/repository"
	role_service "github.com/MamangRust/microservice-point-of-sale-role/service"
	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	tests "github.com/MamangRust/microservice-point-of-sale-test"
	user_cache "github.com/MamangRust/microservice-point-of-sale-user/cache"
	user_handler "github.com/MamangRust/microservice-point-of-sale-user/handler"
	user_repo "github.com/MamangRust/microservice-point-of-sale-user/repository"
	user_service "github.com/MamangRust/microservice-point-of-sale-user/service"

	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthRepositoryTestSuite struct {
	suite.Suite
	ts         *tests.TestSuite
	repo       *repository.Repositories
	userID     int
	email      string
	roleServer *grpc.Server
	roleConn   *grpc.ClientConn
	userServer *grpc.Server
	userConn   *grpc.ClientConn
}

func (s *AuthRepositoryTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.ts = ts

	s.Require().NoError(err)

	authQueries := s.ts.GormDB()

	log, _ := logger.NewLogger("test", nil)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.ts.RedisClient(), log, cacheMetrics)
	obs, _ := observability.NewObservability("test", log)

	roleMencache := role_cache.NewMencache(cacheStore)
	roleRepos := role_repo.NewRepositories(authQueries)
	roleSvc := role_service.NewService(&role_service.Deps{
		Repositories:  roleRepos,
		Logger:        log,
		Mencache:      roleMencache,
		Observability: obs,
	})
	roleGapi := role_handler.NewHandler(roleSvc)
	s.roleServer = grpc.NewServer()
	pbrole.RegisterRoleQueryServiceServer(s.roleServer, roleGapi)
	pbrole.RegisterRoleCommandServiceServer(s.roleServer, roleGapi)
	pbuserrole.RegisterUserRoleServiceServer(s.roleServer, roleGapi)
	roleLis, err := net.Listen("tcp", "localhost:0")
	s.Require().NoError(err)
	go func() {
		_ = s.roleServer.Serve(roleLis)
	}()
	s.roleConn, err = grpc.NewClient(roleLis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	roleClient := pbrole.NewRoleQueryServiceClient(s.roleConn)
	userRoleClient := pbuserrole.NewUserRoleServiceClient(s.roleConn)

	userMencache := user_cache.NewMencache(cacheStore)
	userRepos := user_repo.NewRepositories(authQueries, roleClient, userRoleClient)
	userSvc := user_service.NewService(&user_service.Deps{
		Repositories:  userRepos,
		Logger:        log,
		Hash:          hash.NewHashingPassword(),
		Mencache:      userMencache,
		Observability: obs,
	})
	userGapi := user_handler.NewHandler(userSvc)
	s.userServer = grpc.NewServer()
	pbuser.RegisterUserQueryServiceServer(s.userServer, userGapi)
	pbuser.RegisterUserCommandServiceServer(s.userServer, userGapi)
	userLis, err := net.Listen("tcp", "localhost:0")
	s.Require().NoError(err)
	go func() {
		_ = s.userServer.Serve(userLis)
	}()
	s.userConn, err = grpc.NewClient(userLis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	userQueryClient := pbuser.NewUserQueryServiceClient(s.userConn)
	userCommandClient := pbuser.NewUserCommandServiceClient(s.userConn)

	s.repo = repository.NewRepositories(authQueries, userQueryClient, userCommandClient, roleClient, userRoleClient)
	s.email = "auth.repo.test@example.com"
}

func (s *AuthRepositoryTestSuite) TearDownSuite() {
	if s.roleServer != nil {
		s.roleServer.Stop()
	}
	if s.roleConn != nil {
		s.roleConn.Close()
	}
	if s.userServer != nil {
		s.userServer.Stop()
	}
	if s.userConn != nil {
		s.userConn.Close()
	}
	s.ts.Teardown()
}

func (s *AuthRepositoryTestSuite) Test1_CreateUser() {
	ctx := context.Background()

	req := &requests.RegisterRequest{
		FirstName:       "Auth",
		LastName:        "Repo",
		Email:           s.email,
		Password:        "password123",
		ConfirmPassword: "password123",
		VerifiedCode:    "123456",
		IsVerified:      false,
	}

	res, err := s.repo.User.CreateUser(ctx, req)
	s.NoError(err)
	s.NotNil(res)
	s.Equal(s.email, res.Email)
	s.userID = int(res.UserID)
}

func (s *AuthRepositoryTestSuite) Test2_FindByEmail() {
	s.Require().NotEmpty(s.email)
	ctx := context.Background()

	found, err := s.repo.User.FindByEmail(ctx, s.email)
	s.NoError(err)
	s.NotNil(found)
	s.Equal(int32(s.userID), found.UserID)
}

func (s *AuthRepositoryTestSuite) Test3_FindById() {
	s.Require().NotZero(s.userID)
	ctx := context.Background()

	found, err := s.repo.User.FindById(ctx, s.userID)
	s.NoError(err)
	s.NotNil(found)
	s.Equal(int32(s.userID), found.UserID)
}

func (s *AuthRepositoryTestSuite) Test4_UpdateVerification() {
	s.Require().NotZero(s.userID)
	ctx := context.Background()

	updated, err := s.repo.User.UpdateUserIsVerified(ctx, s.userID, true)
	s.NoError(err)
	s.NotNil(updated)
	s.Equal(int32(s.userID), updated.UserID)
}

func (s *AuthRepositoryTestSuite) Test5_UpdatePassword() {
	s.Require().NotZero(s.userID)
	ctx := context.Background()

	updated, err := s.repo.User.UpdateUserPassword(ctx, s.userID, "newpassword123")
	s.NoError(err)
	s.NotNil(updated)
	s.Equal(int32(s.userID), updated.UserID)

	// The update RPC does not echo the password back; verify via a fresh read.
	found, err := s.repo.User.FindByEmail(ctx, s.email)
	s.NoError(err)
	s.NotNil(found)
	s.Equal("newpassword123", found.Password)
}

func (s *AuthRepositoryTestSuite) Test6_FindByVerificationCode() {
	ctx := context.Background()

	found, err := s.repo.User.FindByVerificationCode(ctx, "123456")
	s.NoError(err)
	s.NotNil(found)
}

func (s *AuthRepositoryTestSuite) Test7_RefreshToken() {
	s.Require().NotZero(s.userID)
	ctx := context.Background()

	token := "test-refresh-token"
	expiresAt := time.Now().Add(24 * time.Hour).Format("2006-01-02 15:04:05")

	req := &requests.CreateRefreshToken{
		UserId:    s.userID,
		Token:     token,
		ExpiresAt: expiresAt,
	}

	res, err := s.repo.RefreshToken.CreateRefreshToken(ctx, req)
	s.NoError(err)
	s.NotNil(res)
	s.Equal(token, res.Token)

	found, err := s.repo.RefreshToken.FindByToken(ctx, token)
	s.NoError(err)
	s.NotNil(found)

	foundByUser, err := s.repo.RefreshToken.FindByUserId(ctx, s.userID)
	s.NoError(err)
	s.NotNil(foundByUser)

	err = s.repo.RefreshToken.DeleteRefreshToken(ctx, token)
	s.NoError(err)
}

func (s *AuthRepositoryTestSuite) Test8_ResetToken() {
	s.Require().NotZero(s.userID)
	ctx := context.Background()

	token := "reset-token-123"
	expiresAt := time.Now().Add(1 * time.Hour).Format("2006-01-02 15:04:05")

	req := &requests.CreateResetTokenRequest{
		UserID:     s.userID,
		ResetToken: token,
		ExpiredAt:  expiresAt,
	}

	res, err := s.repo.ResetToken.CreateResetToken(ctx, req)
	s.NoError(err)
	s.NotNil(res)
	s.Equal(token, res.Token)

	found, err := s.repo.ResetToken.FindByToken(ctx, token)
	s.NoError(err)
	s.NotNil(found)

	err = s.repo.ResetToken.DeleteResetToken(ctx, s.userID)
	s.NoError(err)
}

func TestAuthRepositorySuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(AuthRepositoryTestSuite))
}
