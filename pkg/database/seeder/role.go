package seeder

import (
	"context"
	"fmt"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type roleSeeder struct {
	db     *gorm.DB
	ctx    context.Context
	logger logger.LoggerInterface
}

func NewRoleSeeder(db *gorm.DB, ctx context.Context, logger logger.LoggerInterface) *roleSeeder {
	return &roleSeeder{db: db, ctx: ctx, logger: logger}
}

func (r *roleSeeder) Seed() error {
	randomRoles := []string{
		"ROLE_ADMIN", "Admin Access 1", "Super Admin", "Admin",
		"Store Manager", "Cashier", "Inventory Staff", "Support", "Auditor", "Viewer",
	}

	now := time.Now()
	for i, roleName := range randomRoles {
		err := r.db.WithContext(r.ctx).Exec(
			`INSERT INTO roles (role_name, created_at, updated_at) VALUES (?, ?, ?) ON CONFLICT (role_name) DO NOTHING`,
			roleName, now, now,
		).Error
		if err != nil {
			r.logger.Error("failed to seed role", zap.Int("role", i+1), zap.String("roleName", roleName), zap.Error(err))
			return fmt.Errorf("failed to seed role %d (%s): %w", i+1, roleName, err)
		}
	}

	r.logger.Info("role seeded successfully", zap.Int("totalRoles", len(randomRoles)))
	return nil
}
