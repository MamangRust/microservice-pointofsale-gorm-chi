package seeder

import (
	"context"
	"fmt"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/hash"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type userSeeder struct {
	db     *gorm.DB
	hash   hash.HashPassword
	ctx    context.Context
	logger logger.LoggerInterface
}

func NewUserSeeder(db *gorm.DB, hash hash.HashPassword, ctx context.Context, logger logger.LoggerInterface) *userSeeder {
	return &userSeeder{db: db, hash: hash, ctx: ctx, logger: logger}
}

func (r *userSeeder) Seed() error {
	for i := 1; i <= 10; i++ {
		email := fmt.Sprintf("user_%s@example.com", uuid.NewString())
		rawPassword := fmt.Sprintf("password%d", i)

		hashedPassword, err := r.hash.HashPassword(rawPassword)
		if err != nil {
			r.logger.Error("failed to hash password", zap.Int("user", i), zap.Error(err))
			return fmt.Errorf("failed to hash password for user %d: %w", i, err)
		}

		now := time.Now()
		var userID int
		err = r.db.WithContext(r.ctx).Raw(`
			INSERT INTO users (firstname, lastname, email, password, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?)
			RETURNING user_id`,
			fmt.Sprintf("User%d", i), fmt.Sprintf("Last%d", i), email, hashedPassword, now, now,
		).Scan(&userID).Error
		if err != nil {
			r.logger.Error("failed to seed user", zap.Int("user", i), zap.Error(err))
			return fmt.Errorf("failed to seed user %d: %w", i, err)
		}

		if i > 5 {
			deletedAt := time.Now()
			err = r.db.WithContext(r.ctx).Exec(
				`UPDATE users SET deleted_at = ? WHERE user_id = ?`, deletedAt, userID,
			).Error
			if err != nil {
				r.logger.Error("failed to trash user", zap.Int("user", i), zap.Error(err))
				return fmt.Errorf("failed to trash user %d: %w", i, err)
			}
		}
	}

	r.logger.Info("User seeding completed successfully", zap.Int("totalUsers", 10))
	return nil
}
