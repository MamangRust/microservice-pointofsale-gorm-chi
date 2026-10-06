package repository

import (
	pbcategory "github.com/MamangRust/microservice-point-of-sale-pb/category"
	pbmerchant "github.com/MamangRust/microservice-point-of-sale-pb/merchant"
	"github.com/MamangRust/microservice-point-of-sale-pkg/adapter"
	categoryadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/category"
	merchantadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/merchant"
	"gorm.io/gorm"
)

type Repositories struct {
	ProductQuery   ProductQueryRepository
	ProductCommand ProductCommandRepository
	CategoryQuery  CategoryQueryRepository
	MerchantQuery  MerchantQueryRepository
}

type GuardOptions struct {
	Category []adapter.GuardOption
	Merchant []adapter.GuardOption
}

func NewRepositories(
	db *gorm.DB,
	categoryClient pbcategory.CategoryQueryServiceClient,
	merchantClient pbmerchant.MerchantQueryServiceClient,
	guards ...GuardOptions,
) *Repositories {
	var g GuardOptions
	if len(guards) > 0 {
		g = guards[0]
	}
	return &Repositories{
		ProductQuery:   NewProductQueryRepository(db),
		ProductCommand: NewProductCommandRepository(db),
		CategoryQuery:  categoryadapter.New(categoryClient, g.Category...),
		MerchantQuery:  merchantadapter.New(merchantClient, g.Merchant...),
	}
}
