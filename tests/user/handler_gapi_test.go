package user_test

import (
	"context"
	"testing"

	pbrole "github.com/MamangRust/microservice-point-of-sale-pb/role"
	pbuser "github.com/MamangRust/microservice-point-of-sale-pb/user"
	pbuserrole "github.com/MamangRust/microservice-point-of-sale-pb/user_role"
	"github.com/MamangRust/microservice-point-of-sale-pkg/hash"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	tests "github.com/MamangRust/microservice-point-of-sale-test"
	gapi "github.com/MamangRust/microservice-point-of-sale-user/handler"
	"github.com/MamangRust/microservice-point-of-sale-user/repository"
	"github.com/MamangRust/microservice-point-of-sale-user/service"

	user_cache "github.com/MamangRust/microservice-point-of-sale-user/cache"
	"github.com/stretchr/testify/suite"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

type UserGapiTestSuite struct {
	tests.BaseTestSuite
	client *grpc.ClientConn
	userID int
}

func (s *UserGapiTestSuite) SetupSuite() {
	s.BaseTestSuite.SetupSuite()

	s.SetupRoleService()

	// Seed default role required by user service CreateUser
	s.GormDB().Exec(`INSERT INTO roles (role_name, created_at, updated_at)
		 VALUES ('Admin Access 1', current_timestamp, current_timestamp)
		 ON CONFLICT (role_name) DO NOTHING`)

	userQueries := s.GormDB()
	repos := repository.NewRepositories(userQueries, pbrole.NewRoleQueryServiceClient(s.Conns["role"]), pbuserrole.NewUserRoleServiceClient(s.Conns["role"]))

	log, _ := logger.NewLogger("test", nil)
	hasher := hash.NewHashingPassword()
	cacheStore := s.GetCacheStore()
	mencache := user_cache.NewMencache(cacheStore)

	userService := service.NewService(&service.Deps{
		Repositories:  repos,
		Logger:        log,
		Hash:          hasher,
		Mencache:      mencache,
		Observability: s.Obs,
	})

	// Start gRPC Server
	userHandler := gapi.NewHandler(userService)
	server := grpc.NewServer()
	pbuser.RegisterUserQueryServiceServer(server, userHandler)
	pbuser.RegisterUserCommandServiceServer(server, userHandler)

	addr := s.RegisterServer(server)
	conn := s.GetConnection(addr)

	s.client = conn
}

func (s *UserGapiTestSuite) TestUserGapiLifecycle() {
	ctx := context.Background()

	cmdClient := pbuser.NewUserCommandServiceClient(s.client)
	queryClient := pbuser.NewUserQueryServiceClient(s.client)

	// 1. Create
	createReq := &pbuser.CreateUserRequest{
		Firstname:       "Gapi",
		Lastname:        "User",
		Email:           "gapi.user@example.com",
		Password:        "password123",
		ConfirmPassword: "password123",
	}
	res, err := cmdClient.Create(ctx, createReq)
	s.Require().NoError(err)
	s.Equal(createReq.Email, res.Data.Email)
	userID := res.Data.Id

	// 2. FindById
	getRes, err := queryClient.FindById(ctx, &pbuser.FindByIdUserRequest{Id: userID})
	s.Require().NoError(err)
	s.Equal(userID, getRes.Data.Id)

	// 3. FindAll
	allRes, err := queryClient.FindAll(ctx, &pbuser.FindAllUserRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(allRes.Data)

	// 4. FindByActive
	activeRes, err := queryClient.FindByActive(ctx, &pbuser.FindAllUserRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(activeRes.Data)

	// 5. Update
	updateRes, _ := cmdClient.Update(ctx, &pbuser.UpdateUserRequest{
		Id:              userID,
		Firstname:       "GapiUpdated",
		Lastname:        "UserUpdated",
		Email:           "gapi.updated@example.com",
		Password:        "password123",
		ConfirmPassword: "password123",
	})
	s.Require().NoError(err)
	s.Equal("GapiUpdated", updateRes.Data.Firstname)

	// 6. Trash
	_, err = cmdClient.TrashedUser(ctx, &pbuser.FindByIdUserRequest{Id: userID})
	s.Require().NoError(err)

	// 7. FindByTrashed
	trashedRes, err := queryClient.FindByTrashed(ctx, &pbuser.FindAllUserRequest{Page: 1, PageSize: 10})
	s.Require().NoError(err)
	s.NotEmpty(trashedRes.Data)

	// 8. Restore
	_, err = cmdClient.RestoreUser(ctx, &pbuser.FindByIdUserRequest{Id: userID})
	s.Require().NoError(err)

	// 9. DeletePermanent
	_, _ = cmdClient.TrashedUser(ctx, &pbuser.FindByIdUserRequest{Id: userID})
	_, err = cmdClient.DeleteUserPermanent(ctx, &pbuser.FindByIdUserRequest{Id: userID})
	s.Require().NoError(err)

	// 10. RestoreAll
	_, err = cmdClient.RestoreAllUser(ctx, &emptypb.Empty{})
	s.Require().NoError(err)

	// 11. DeleteAll
	_, err = cmdClient.DeleteAllUserPermanent(ctx, &emptypb.Empty{})
	s.Require().NoError(err)
}

func TestUserGapiSuite(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping integration test")
	}
	suite.Run(t, new(UserGapiTestSuite))
}
