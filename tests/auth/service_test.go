package auth_test

import (
	"context"
	"testing"

	mencache "github.com/MamangRust/microservice-point-of-sale-auth/cache"
	"github.com/MamangRust/microservice-point-of-sale-auth/repository"
	"github.com/MamangRust/microservice-point-of-sale-auth/service"
	pbrole "github.com/MamangRust/microservice-point-of-sale-pb/role"
	pbuser "github.com/MamangRust/microservice-point-of-sale-pb/user"
	pbuserrole "github.com/MamangRust/microservice-point-of-sale-pb/user_role"
	"github.com/MamangRust/microservice-point-of-sale-pkg/auth"
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

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"net"
)

type AuthServiceTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	authService *service.Service
	email       string
	password    string
	roleServer  *grpc.Server
	roleConn    *grpc.ClientConn
	userServer  *grpc.Server
	userConn    *grpc.ClientConn
}

func (s *AuthServiceTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.ts = ts

	s.Require().NoError(err)

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	redisClient := redis.NewClient(opts)

	authQueries := s.ts.GormDB()

	log, _ := logger.NewLogger("test", nil)
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(redisClient, log, cacheMetrics)
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

	hasher := hash.NewHashingPassword()

	userMencache := user_cache.NewMencache(cacheStore)
	userRepos := user_repo.NewRepositories(authQueries, roleClient, userRoleClient)
	userSvc := user_service.NewService(&user_service.Deps{
		Repositories:  userRepos,
		Logger:        log,
		Hash:          hasher,
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

	repos := repository.NewRepositories(authQueries, userQueryClient, userCommandClient, roleClient, userRoleClient)

	mencacheService := mencache.NewMencache(cacheStore)

	tokenManager, _ := auth.NewManager("mysecretkey")

	s.authService = service.NewService(&service.Deps{
		Repositories:  repos,
		Logger:        log,
		Mencache:      mencacheService,
		Token:         tokenManager,
		Hash:          hasher,
		Kafka:         nil,
		Observability: obs,
	})

	s.email = "auth.service.test@example.com"
	s.password = "password123"
}

func (s *AuthServiceTestSuite) TearDownSuite() {
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

func (s *AuthServiceTestSuite) TestAuthLifecycle() {
	ctx := context.Background()

	// 1. Register
	regReq := &requests.RegisterRequest{
		FirstName:       "Auth",
		LastName:        "Service",
		Email:           s.email,
		Password:        s.password,
		ConfirmPassword: s.password,
	}

	created, err := s.authService.Register.Register(ctx, regReq)
	s.Require().NoError(err)
	s.Require().NotNil(created)
	s.Equal(s.email, created.Email)

	// 1b. Verify email (login hanya menerima user is_verified = true)
	var verifyCode string
	err = s.ts.GormDB().Raw("SELECT verification_code FROM users WHERE email = $1", s.email).Scan(&verifyCode).Error
	s.Require().NoError(err)

	verified, err := s.authService.PasswordReset.VerifyCode(ctx, verifyCode)
	s.Require().NoError(err)
	s.True(verified)

	// 2. Login
	loginReq := &requests.AuthRequest{
		Email:    s.email,
		Password: s.password,
	}

	tokenRes, err := s.authService.Login.Login(ctx, loginReq)
	s.Require().NoError(err)
	s.Require().NotNil(tokenRes)
	s.NotEmpty(tokenRes.AccessToken)
	s.NotEmpty(tokenRes.RefreshToken)

	// 3. ForgotPassword
	success, err := s.authService.PasswordReset.ForgotPassword(ctx, s.email)
	s.Require().NoError(err)
	s.True(success)
}

func TestAuthServiceSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(AuthServiceTestSuite))
}
