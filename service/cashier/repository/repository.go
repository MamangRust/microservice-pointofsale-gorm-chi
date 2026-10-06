package repository

import (
	pbmerchant "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	pbuser "github.com/MamangRust/microservice-point-of-sale-pb/user"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	merchantadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/merchant"
	useradapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/user"
	"gorm.io/gorm"
)

type Repositories struct {
	UserQuery      useradapter.QueryRepository
	MerchantQuery  merchantadapter.Repository
	CashierQuery   CashierQueryRepository
	CashierCommand CashierCommandRepository
}

// GuardOptions collects the dependency guards for each remote dependency.
type GuardOptions struct {
	User     []adapter.GuardOption
	Merchant []adapter.GuardOption
}

func NewRepositories(
	db *gorm.DB,
	userQueryClient pbuser.UserQueryServiceClient,
	userCommandClient pbuser.UserCommandServiceClient,
	merchantClient pbmerchant.MerchantQueryServiceClient,
	guards ...GuardOptions,
) *Repositories {
	var g GuardOptions
	if len(guards) > 0 {
		g = guards[0]
	}

	return &Repositories{
		UserQuery:      useradapter.New(userQueryClient, userCommandClient, g.User...),
		MerchantQuery:  merchantadapter.New(merchantClient, g.Merchant...),
		CashierQuery:   NewCashierQueryRepository(db),
		CashierCommand: NewCashierCommandRepository(db),
	}
}
