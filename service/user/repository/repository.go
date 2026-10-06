package repository

import (
	pbrole "github.com/MamangRust/microservice-point-of-sale-pb/role"
	pbuserrole "github.com/MamangRust/microservice-point-of-sale-pb/user_role"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	roleadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/role"
	userroleadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/user_role"
	"gorm.io/gorm"
)

type Repositories struct {
	UserCommand UserCommandRepository
	UserQuery   UserQueryRepository
	Role        RoleQueryRepository
	UserRole    UserRoleRepository
}

// GuardOptions lets the wiring layer attach resilience guards to the gRPC
// adapters consumed by this service.
type GuardOptions struct {
	Role     []adapter.GuardOption
	UserRole []adapter.GuardOption
}

// Compile-time assertion: the shared user-role adapter must satisfy the user
// service's contract so the wiring can inject it directly.
var _ UserRoleRepository = (*userroleadapter.Repository)(nil)

func NewRepositories(
	db *gorm.DB,
	roleClient pbrole.RoleQueryServiceClient,
	userRoleClient pbuserrole.UserRoleServiceClient,
	guards ...GuardOptions,
) *Repositories {
	var g GuardOptions
	if len(guards) > 0 {
		g = guards[0]
	}

	roleRepo := roleadapter.New(roleClient, g.Role...)

	return &Repositories{
		UserCommand: NewUserCommandRepository(db),
		UserQuery:   NewUserQueryRepository(db),
		Role:        NewRoleRepository(roleRepo),
		UserRole:    userroleadapter.New(userRoleClient, g.UserRole...),
	}
}
