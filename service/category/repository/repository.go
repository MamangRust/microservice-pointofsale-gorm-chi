package repository

import "gorm.io/gorm"

type Repositories struct {
	CategoryQuery   CategoryQueryRepository
	CategoryCommand CategoryCommandRepository
}

func NewRepositories(DB *gorm.DB) *Repositories {
	return &Repositories{
		CategoryQuery:   NewCategoryQueryRepository(DB),
		CategoryCommand: NewCategoryCommandRepository(DB),
	}
}
