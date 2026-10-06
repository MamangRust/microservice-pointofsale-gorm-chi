package auth_test

import (
	"context"
	"net"
	"testing"

	mencache "github.com/MamangRust/microservice-point-of-sale-auth/cache"
	"github.com/MamangRust/microservice-point-of-sale-auth/handler"
	"github.com/MamangRust/microservice-point-of-sale-auth/repository"
	"github.com/MamangRust/microservice-point-of-sale-auth/service"
	pbauth "github.com/MamangRust/microservice-point-of-sale-pb/auth"
	pbrole "github.com/MamangRust/microservice-point-of-sale-pb/role"
	pbuser "github.com/MamangRust/microservice-point-of-sale-pb/user"
	pbuserrole "github.com/MamangRust/microservice-point-of-sale-pb/user_role"
	"github.com/MamangRust/microservice-point-of-sale-pkg/auth"
	"github.com/MamangRust/microservice-point-of-sale-pkg/hash"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/MamangRust/microservice-point-of-sale-shared/cache"
	"github.com/MamangRust/microservice-point-of-sale-shared/observability"
	tests "github.com/MamangRust/microservice-point-of-sale-test"

	role_cache "github.com/MamangRust/microservice-point-of-sale-role/cache"
	role_handler "github.com/MamangRust/microservice-point-of-sale-role/handler"
	role_repo "github.com/MamangRust/microservice-point-of-sale-role/repository"
	role_service "github.com/MamangRust/microservice-point-of-sale-role/service"
	user_cache "github.com/MamangRust/microservice-point-of-sale-user/cache"
	user_handler "github.com/MamangRust/microservice-point-of-sale-user/handler"
	user_repo "github.com/MamangRust/microservice-point-of-sale-user/repository"
	user_service "github.com/MamangRust/microservice-point-of-sale-user/service"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type AuthHandlerGapiTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	redisClient *redis.Client
	queryClient pbauth.AuthServiceClient
	cmdClient   pbauth.AuthServiceClient
	conn        *grpc.ClientConn
	grpcServer  *grpc.Server
	email       string
	password    string
	accessToken string
}

func (s *AuthHandlerGapiTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.ts = ts

	s.Require().NoError(err)

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	s.redisClient = redis.NewClient(opts)

	roleQueries := s.ts.GormDB()
	userQueries := s.ts.GormDB()
	authQueries := s.ts.GormDB()

	log, _ := logger.NewLogger("test", nil)
	hasher := hash.NewHashingPassword()
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)
	obs, _ := observability.NewObservability("test", log)

	// 1. Setup Role Service & gRPC Server
	roleMencache := role_cache.NewMencache(cacheStore)
	roleRepos := role_repo.NewRepositories(roleQueries)
	roleSvc := role_service.NewService(&role_service.Deps{
		Repositories:  roleRepos,
		Logger:        log,
		Mencache:      roleMencache,
		Observability: obs,
	})
	roleGapi := role_handler.NewHandler(roleSvc)
	roleServer := grpc.NewServer()
	pbrole.RegisterRoleQueryServiceServer(roleServer, roleGapi)
	pbrole.RegisterRoleCommandServiceServer(roleServer, roleGapi)
	pbuserrole.RegisterUserRoleServiceServer(roleServer, roleGapi)
	roleLis, _ := net.Listen("tcp", "localhost:0")
	go roleServer.Serve(roleLis)

	roleConn, _ := grpc.NewClient(roleLis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	roleClient := pbrole.NewRoleQueryServiceClient(roleConn)
	userRoleClient := pbuserrole.NewUserRoleServiceClient(roleConn)

	// 2. Setup User Service & gRPC Server
	userMencache := user_cache.NewMencache(cacheStore)
	userRepos := user_repo.NewRepositories(userQueries, roleClient, userRoleClient)
	userSvc := user_service.NewService(&user_service.Deps{
		Repositories:  userRepos,
		Logger:        log,
		Hash:          hasher,
		Mencache:      userMencache,
		Observability: obs,
	})
	userGapi := user_handler.NewHandler(userSvc)
	userServer := grpc.NewServer()
	pbuser.RegisterUserQueryServiceServer(userServer, userGapi)
	pbuser.RegisterUserCommandServiceServer(userServer, userGapi)
	userLis, _ := net.Listen("tcp", "localhost:0")
	go userServer.Serve(userLis)

	userConn, _ := grpc.NewClient(userLis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	userQueryClient := pbuser.NewUserQueryServiceClient(userConn)
	userCommandClient := pbuser.NewUserCommandServiceClient(userConn)

	// 3. Setup Auth Service
	repos := repository.NewRepositories(authQueries, userQueryClient, userCommandClient, roleClient, userRoleClient)

	tokenManager, _ := auth.NewManager("mysecret")
	svc := service.NewService(&service.Deps{
		Repositories:  repos,
		Logger:        log,
		Mencache:      mencache.NewMencache(cacheStore),
		Token:         tokenManager,
		Hash:          hasher,
		Kafka:         nil,
		Observability: obs,
	})

	h := handler.NewAuthHandleGrpc(svc, log)

	s.grpcServer = grpc.NewServer()
	pbauth.RegisterAuthServiceServer(s.grpcServer, h)

	lis, err := net.Listen("tcp", "localhost:0")
	s.Require().NoError(err)

	go func() {
		_ = s.grpcServer.Serve(lis)
	}()

	conn, err := grpc.NewClient(lis.Addr().String(), grpc.WithTransportCredentials(insecure.NewCredentials()))
	s.Require().NoError(err)
	s.conn = conn
	s.queryClient = pbauth.NewAuthServiceClient(conn)
	s.cmdClient = pbauth.NewAuthServiceClient(conn)

	s.email = "auth.handler.gapi.test@example.com"
	s.password = "password123"

	_ = s.ts.GormDB().Exec("INSERT INTO roles (role_name) VALUES ('ROLE_ADMIN')").Error
}

func (s *AuthHandlerGapiTestSuite) TearDownSuite() {
	if s.conn != nil {
		s.conn.Close()
	}
	if s.grpcServer != nil {
		s.grpcServer.Stop()
	}
	if s.redisClient != nil {
		s.redisClient.Close()
	}
	if s.ts != nil {
		if sqlDB, err := s.ts.GormDB().DB(); err == nil {
			sqlDB.Close()
		}
	}
	s.ts.Teardown()
}

func (s *AuthHandlerGapiTestSuite) verifyUser(email string) {
	err := s.ts.GormDB().Exec("UPDATE users SET is_verified = true WHERE email = $1", email).Error
	s.Require().NoError(err)
}

func (s *AuthHandlerGapiTestSuite) Test1_Register() {
	ctx := context.Background()
	req := &pbauth.RegisterRequest{
		Firstname:       "Auth",
		Lastname:        "Handler",
		Email:           s.email,
		Password:        s.password,
		ConfirmPassword: s.password,
	}

	res, err := s.cmdClient.RegisterUser(ctx, req)
	s.NoError(err)
	s.NotNil(res)
	s.Equal("success", res.Status)
	s.Equal(s.email, res.Data.Email)
}

func (s *AuthHandlerGapiTestSuite) Test2_Login() {
	ctx := context.Background()

	s.verifyUser(s.email)

	req := &pbauth.LoginRequest{
		Email:    s.email,
		Password: s.password,
	}

	res, err := s.queryClient.LoginUser(ctx, req)
	s.NoError(err)
	s.NotNil(res)
	s.Equal("success", res.Status)
	s.NotEmpty(res.Data.AccessToken)
	s.accessToken = res.Data.AccessToken
}

func (s *AuthHandlerGapiTestSuite) Test4_LoginLockout() {
	ctx := context.Background()
	email := "locked.gapi@example.com"
	password := "wrongpassword"

	regReq := &pbauth.RegisterRequest{
		Firstname:       "Locked",
		Lastname:        "Gapi",
		Email:           email,
		Password:        "correctpassword",
		ConfirmPassword: "correctpassword",
	}
	_, err := s.cmdClient.RegisterUser(ctx, regReq)
	s.NoError(err)
	s.verifyUser(email)

	loginReq := &pbauth.LoginRequest{
		Email:    email,
		Password: password,
	}

	for i := 0; i < 5; i++ {
		_, err := s.queryClient.LoginUser(ctx, loginReq)
		s.Error(err)
	}

	_, err = s.queryClient.LoginUser(ctx, loginReq)
	s.Error(err)
	s.Contains(err.Error(), "Account is locked")
}

func (s *AuthHandlerGapiTestSuite) Test3_GetMe() {
	s.Require().NotEmpty(s.accessToken)
	ctx := context.Background()

	res, err := s.queryClient.GetMe(ctx, &pbauth.GetMeRequest{AccessToken: s.accessToken})
	s.NoError(err)
	s.NotNil(res)
	s.Equal("success", res.Status)
	s.Equal(s.email, res.Data.Email)
}

func TestAuthHandlerGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(AuthHandlerGapiTestSuite))
}
