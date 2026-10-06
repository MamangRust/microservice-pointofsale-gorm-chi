package repository

import (
	"context"

	userroleadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/user_role"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
	"github.com/MamangRust/microservice-point-of-sale-shared/domain/requests"
)

type UserResult struct {
	UserID     int32
	Firstname  string
	Lastname   string
	Email      string
	CreatedAt  *string
	UpdatedAt  *string
	DeletedAt  *string
	TotalCount int64
}

type UserQueryRepository interface {
	FindAllUsers(ctx context.Context, req *requests.FindAllUsers) ([]*UserResult, error)
	FindById(ctx context.Context, user_id int) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByEmailAndVerify(ctx context.Context, email string) (*models.User, error)
	FindByVerificationCode(ctx context.Context, code string) (*models.User, error)
	FindByActive(ctx context.Context, req *requests.FindAllUsers) ([]*UserResult, error)
	FindByTrashed(ctx context.Context, req *requests.FindAllUsers) ([]*UserResult, error)
}

type UserCommandRepository interface {
	CreateUser(ctx context.Context, request *requests.CreateUserRequest) (*models.User, error)
	CreateUserRecord(ctx context.Context, request *requests.RegisterRequest) (*models.User, error)
	UpdateUser(ctx context.Context, request *requests.UpdateUserRequest) (*models.User, error)
	UpdateUserIsVerified(ctx context.Context, user_id int, is_verified bool) (*models.User, error)
	UpdateUserPassword(ctx context.Context, user_id int, password string) (*models.User, error)
	TrashedUser(ctx context.Context, user_id int) (*models.User, error)
	RestoreUser(ctx context.Context, user_id int) (*models.User, error)
	DeleteUserPermanent(ctx context.Context, user_id int) (bool, error)
	RestoreAllUser(ctx context.Context) (bool, error)
	DeleteAllUserPermanent(ctx context.Context) (bool, error)
}

type RoleQueryRepository interface {
	FindByName(ctx context.Context, name string) (*models.Role, error)
}

// UserRoleRepository is the user service's view of the role service's hosted
// user-role assignments. The role service piggybacks the user-role gRPC server,
// so the user service reaches it through the shared userroleadapter.
type UserRoleRepository interface {
	userroleadapter.QueryRepository
	userroleadapter.CommandRepository
}
