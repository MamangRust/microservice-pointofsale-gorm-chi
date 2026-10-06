package repository

import "gorm.io/gorm"

type Repositories struct {
	OrderItemQuery   OrderItemQueryRepository
	OrderItemCommand OrderItemCommandRepository
}

func NewRepositories(DB *gorm.DB) *Repositories {
	return &Repositories{
		OrderItemQuery:   NewOrderItemQueryRepository(DB),
		OrderItemCommand: NewOrderItemCommandRepository(DB),
	}
}
