package repository

import (
	"context"

	roleadapter "github.com/MamangRust/microservice-point-of-sale-pkg/adapter/role"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/models"
)

type roleRepository struct {
	adapter roleadapter.QueryRepository
}

func NewRoleRepository(adapter roleadapter.QueryRepository) *roleRepository {
	return &roleRepository{adapter: adapter}
}

func (r *roleRepository) FindByName(ctx context.Context, name string) (*models.Role, error) {
	return r.adapter.FindByName(ctx, name)
}
