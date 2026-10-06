package auth_test

import (
	"context"
	"net"
	"testing"

	auth_cache "github.com/MamangRust/microservice-point-of-sale-auth/cache"
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
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// AuthPasswordResetTestSuite covers the three password-flow gRPC resolvers
// (VerifyCode, ForgotPassword, ResetPassword) end-to-end against a real
// PostgreSQL + Redis stack, and verifies the ResetPassword hashing fix:
// the persisted password must be a bcrypt hash (never the plaintext) and
// login must succeed with the new password afterwards.
type AuthPasswordResetTestSuite struct {
	suite.Suite
	ts          *tests.TestSuite
	redisClient *redis.Client
	queryClient pbauth.AuthServiceClient
	cmdClient   pbauth.AuthServiceClient
	conn        *grpc.ClientConn
	grpcServer  *grpc.Server
	roleServer  *grpc.Server
	roleConn    *grpc.ClientConn
	userServer  *grpc.Server
	userConn    *grpc.ClientConn
	hasher      hash.HashPassword
}

func (s *AuthPasswordResetTestSuite) SetupSuite() {
	ts, err := tests.SetupTestSuite()
	s.ts = ts

	s.Require().NoError(err)

	opts, err := redis.ParseURL(s.ts.RedisURL)
	s.Require().NoError(err)
	s.redisClient = redis.NewClient(opts)

	authQueries := s.ts.GormDB()

	log, _ := logger.NewLogger("test", nil)
	s.hasher = hash.NewHashingPassword()
	cacheMetrics, _ := observability.NewCacheMetrics("test")
	cacheStore := cache.NewCacheStore(s.redisClient, log, cacheMetrics)
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
		Hash:          s.hasher,
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
	tokenManager, _ := auth.NewManager("mysecret")

	svc := service.NewService(&service.Deps{
		Repositories:  repos,
		Logger:        log,
		Mencache:      auth_cache.NewMencache(cacheStore),
		Token:         tokenManager,
		Hash:          s.hasher,
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
}

func (s *AuthPasswordResetTestSuite) TearDownSuite() {
	if s.conn != nil {
		s.conn.Close()
	}
	if s.grpcServer != nil {
		s.grpcServer.Stop()
	}
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
	if s.redisClient != nil {
		s.redisClient.Close()
	}
	if s.ts != nil {
		if sqlDB, err := s.ts.GormDB().DB(); err == nil {
			sqlDB.Close()
		}
		s.ts.Teardown()
	}
}

// registerUser creates a fresh user through the register resolver and returns
// its user_id.
func (s *AuthPasswordResetTestSuite) registerUser(email, password string) int32 {
	ctx := context.Background()
	res, err := s.cmdClient.RegisterUser(ctx, &pbauth.RegisterRequest{
		Firstname:       "Reset",
		Lastname:        "Tester",
		Email:           email,
		Password:        password,
		ConfirmPassword: password,
	})
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Require().NotZero(res.Data.Id)
	return res.Data.Id
}

func (s *AuthPasswordResetTestSuite) markUserVerified(userID int32) {
	s.Require().NoError(s.ts.GormDB().Exec("UPDATE users SET is_verified = true WHERE user_id = $1", userID).Error)
}

func (s *AuthPasswordResetTestSuite) fetchVerificationCode(userID int32) string {
	var code string
	err := s.ts.GormDB().Raw("SELECT verification_code FROM users WHERE user_id = $1", userID).Scan(&code).Error
	s.Require().NoError(err)
	return code
}

func (s *AuthPasswordResetTestSuite) fetchResetToken(userID int32) string {
	var token string
	err := s.ts.GormDB().Raw("SELECT token FROM reset_tokens WHERE user_id = $1", userID).Scan(&token).Error
	s.Require().NoError(err)
	return token
}

func (s *AuthPasswordResetTestSuite) storedPassword(userID int32) string {
	var pw string
	err := s.ts.GormDB().Raw("SELECT password FROM users WHERE user_id = $1", userID).Scan(&pw).Error
	s.Require().NoError(err)
	return pw
}

// ---------------------------------------------------------------------------
// VerifyCode resolver
// ---------------------------------------------------------------------------

func (s *AuthPasswordResetTestSuite) Test1_VerifyCodeResolver_Success() {
	userID := s.registerUser("verifycode.success@example.com", "password123")
	code := s.fetchVerificationCode(userID)
	s.Require().NotEmpty(code)

	res, err := s.queryClient.VerifyCode(context.Background(), &pbauth.VerifyCodeRequest{Code: code})
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Equal("success", res.Status)

	var isVerified bool
	err = s.ts.GormDB().Raw("SELECT is_verified FROM users WHERE user_id = $1", userID).Scan(&isVerified).Error
	s.Require().NoError(err)
	s.True(isVerified)
}

func (s *AuthPasswordResetTestSuite) Test2_VerifyCodeResolver_InvalidCode() {
	_, err := s.queryClient.VerifyCode(context.Background(), &pbauth.VerifyCodeRequest{Code: "does-not-exist"})
	s.Require().Error(err)
	s.Equal(codes.NotFound, status.Code(err))
}

// ---------------------------------------------------------------------------
// ForgotPassword resolver
// ---------------------------------------------------------------------------

func (s *AuthPasswordResetTestSuite) Test3_ForgotPasswordResolver_Success() {
	userID := s.registerUser("forgot.success@example.com", "password123")

	res, err := s.cmdClient.ForgotPassword(context.Background(), &pbauth.ForgotPasswordRequest{Email: "forgot.success@example.com"})
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Equal("success", res.Status)

	// A reset token must be persisted for the user.
	s.Require().NotEmpty(s.fetchResetToken(userID))
}

func (s *AuthPasswordResetTestSuite) Test4_ForgotPasswordResolver_UnknownEmail_DoesNotRevealAccount() {
	res, err := s.cmdClient.ForgotPassword(context.Background(), &pbauth.ForgotPasswordRequest{Email: "ghost@example.com"})
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Equal("success", res.Status)
}

// ---------------------------------------------------------------------------
// ResetPassword resolver + hashing fix
// ---------------------------------------------------------------------------

func (s *AuthPasswordResetTestSuite) Test5_ResetPasswordResolver_Success_AndHashingFix() {
	const newPassword = "freshpassword456"

	userID := s.registerUser("reset.success@example.com", "password123")

	_, err := s.cmdClient.ForgotPassword(context.Background(), &pbauth.ForgotPasswordRequest{Email: "reset.success@example.com"})
	s.Require().NoError(err)

	token := s.fetchResetToken(userID)
	s.Require().NotEmpty(token)

	res, err := s.cmdClient.ResetPassword(context.Background(), &pbauth.ResetPasswordRequest{
		ResetToken:      token,
		Password:        newPassword,
		ConfirmPassword: newPassword,
	})
	s.Require().NoError(err)
	s.Require().NotNil(res)
	s.Equal("success", res.Status)

	// --- Hashing fix: the DB must hold a bcrypt hash, never the plaintext ---
	stored := s.storedPassword(userID)
	s.NotEqual(newPassword, stored, "plaintext password must never be persisted")
	s.NoError(s.hasher.ComparePassword(stored, newPassword),
		"stored password should verify against the new password via bcrypt")

	// The reset token must be consumed after a successful reset.
	var count int
	err = s.ts.GormDB().Raw("SELECT COUNT(*) FROM reset_tokens WHERE user_id = $1", userID).Scan(&count).Error
	s.Require().NoError(err)
	s.Equal(0, count)

	// --- Login with the new password must succeed after the reset ---
	s.markUserVerified(userID)
	loginRes, err := s.queryClient.LoginUser(context.Background(), &pbauth.LoginRequest{
		Email:    "reset.success@example.com",
		Password: newPassword,
	})
	s.Require().NoError(err)
	s.Require().NotNil(loginRes)
	s.NotEmpty(loginRes.Data.AccessToken)
}

func (s *AuthPasswordResetTestSuite) Test6_ResetPasswordResolver_InvalidToken() {
	_, err := s.cmdClient.ResetPassword(context.Background(), &pbauth.ResetPasswordRequest{
		ResetToken:      "bogus-token",
		Password:        "newpassword123",
		ConfirmPassword: "newpassword123",
	})
	s.Require().Error(err)
	s.Equal(codes.NotFound, status.Code(err))
}

func (s *AuthPasswordResetTestSuite) Test7_ResetPasswordResolver_PasswordMismatch() {
	userID := s.registerUser("reset.mismatch@example.com", "password123")

	_, err := s.cmdClient.ForgotPassword(context.Background(), &pbauth.ForgotPasswordRequest{Email: "reset.mismatch@example.com"})
	s.Require().NoError(err)

	token := s.fetchResetToken(userID)
	s.Require().NotEmpty(token)

	_, err = s.cmdClient.ResetPassword(context.Background(), &pbauth.ResetPasswordRequest{
		ResetToken:      token,
		Password:        "newpassword123",
		ConfirmPassword: "differentpassword",
	})
	s.Require().Error(err)
}

func TestAuthPasswordResetSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(AuthPasswordResetTestSuite))
}
