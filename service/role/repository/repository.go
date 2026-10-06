package repository

import "gorm.io/gorm"

type Repositories struct {
	RoleCommand RoleCommandRepository
	RoleQuery   RoleQueryRepository
	UserRole    UserRoleRepository
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		RoleCommand: NewRoleCommandRepository(db),
		RoleQuery:   NewRoleQueryRepository(db),
		UserRole:    NewUserRoleRepository(db),
	}
}
