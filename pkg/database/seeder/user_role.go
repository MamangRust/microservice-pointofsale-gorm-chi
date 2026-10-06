package seeder

import (
	"context"
	"fmt"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"go.uber.org/zap"
	"golang.org/x/exp/rand"
	"gorm.io/gorm"
)

type userRoleSeeder struct {
	db     *gorm.DB
	ctx    context.Context
	logger logger.LoggerInterface
}

func NewUserRoleSeeder(db *gorm.DB, ctx context.Context, logger logger.LoggerInterface) *userRoleSeeder {
	return &userRoleSeeder{db: db, ctx: ctx, logger: logger}
}

type userRow struct {
	UserID int
	Email  string
}

type roleRow struct {
	RoleID   int
	RoleName string
}

func (r *userRoleSeeder) Seed() error {
	var users []userRow
	err := r.db.WithContext(r.ctx).Raw(`SELECT user_id, email FROM users WHERE deleted_at IS NULL LIMIT 20`).Scan(&users).Error
	if err != nil {
		r.logger.Error("failed to fetch users", zap.Error(err))
		return fmt.Errorf("failed to fetch users: %w", err)
	}

	var roles []roleRow
	err = r.db.WithContext(r.ctx).Raw(`SELECT role_id, role_name FROM roles LIMIT 4`).Scan(&roles).Error
	if err != nil {
		r.logger.Error("failed to fetch roles", zap.Error(err))
		return fmt.Errorf("failed to fetch roles: %w", err)
	}

	if len(users) == 0 || len(roles) == 0 {
		r.logger.Debug("no users or roles available for seeding")
		return nil
	}

	rand.Seed(uint64(time.Now().UnixNano()))
	now := time.Now()

	for _, user := range users {
		role := roles[rand.Intn(len(roles))]
		err := r.db.WithContext(r.ctx).Exec(
			`INSERT INTO user_roles (user_id, role_id, created_at, updated_at) VALUES (?, ?, ?, ?) ON CONFLICT DO NOTHING`,
			user.UserID, role.RoleID, now, now,
		).Error
		if err != nil {
			r.logger.Error("failed to assign role to user", zap.String("user", user.Email), zap.String("role", role.RoleName), zap.Error(err))
			return fmt.Errorf("failed to assign role %s to user %s: %w", role.RoleName, user.Email, err)
		}
	}

	r.logger.Info("user roles assigned successfully", zap.Int("users", len(users)), zap.Int("roles", len(roles)))
	return nil
}
