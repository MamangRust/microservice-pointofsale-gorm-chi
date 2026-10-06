package seeder

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/hash"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"gorm.io/gorm"
)

type Deps struct {
	// DBs maps a bounded context to its own PostgreSQL instance
	// (keys: "identity", "merchant", "catalog", "sales", "email"). Each domain
	// seeder receives the instance(s) it actually reads/writes, since foreign
	// keys cannot cross instances.
	DBs    map[string]*gorm.DB
	Ctx    context.Context
	Logger logger.LoggerInterface
	Hash   hash.HashPassword
}

type Seeder struct {
	User        *userSeeder
	Role        *roleSeeder
	UserRole    *userRoleSeeder
	Cashier     *cashierSeeder
	Category    *categorySeeder
	Product     *productSeeder
	Merchant    *merchantSeeder
	Order       *orderSeeder
	Transaction *transactionSeeder
}

func NewSeeder(deps Deps) *Seeder {
	dbs := deps.DBs
	return &Seeder{
		User:        NewUserSeeder(dbs["identity"], deps.Hash, deps.Ctx, deps.Logger),
		Role:        NewRoleSeeder(dbs["identity"], deps.Ctx, deps.Logger),
		UserRole:    NewUserRoleSeeder(dbs["identity"], deps.Ctx, deps.Logger),
		Merchant:    NewMerchantSeeder(dbs["merchant"], dbs["identity"], deps.Ctx, deps.Logger),
		Cashier:     NewCashierSeeder(dbs["merchant"], dbs["identity"], deps.Ctx, deps.Logger),
		Category:    NewCategorySeeder(dbs["catalog"], deps.Ctx, deps.Logger),
		Product:     NewProductSeeder(dbs["catalog"], dbs["merchant"], deps.Ctx, deps.Logger),
		Order:       NewOrderSeeder(dbs["sales"], dbs["merchant"], dbs["catalog"], deps.Ctx, deps.Logger),
		Transaction: NewTransactionSeeder(dbs["sales"], dbs["merchant"], deps.Ctx, deps.Logger),
	}
}

func (s *Seeder) Run() error {
	if err := s.seedWithDelay("users", s.User.Seed); err != nil {
		return err
	}

	if err := s.seedWithDelay("roles", s.Role.Seed); err != nil {
		return err
	}

	if err := s.seedWithDelay("user_roles", s.UserRole.Seed); err != nil {
		return err
	}

	if err := s.seedWithDelay("merchant", s.Merchant.Seed); err != nil {
		return nil
	}

	if err := s.seedWithDelay("cashier", s.Cashier.Seed); err != nil {
		return err
	}

	if err := s.seedWithDelay("category", s.Category.Seed); err != nil {
		return err
	}

	if err := s.seedWithDelay("product", s.Product.Seed); err != nil {
		return err
	}

	if err := s.seedWithDelay("order", s.Order.Seed); err != nil {
		return err
	}

	if err := s.seedWithDelay("transaction", s.Transaction.Seed); err != nil {
		return err
	}

	return nil
}

func (s *Seeder) seedWithDelay(entityName string, seedFunc func() error) error {
	if err := seedFunc(); err != nil {
		return fmt.Errorf("failed to seed %s: %w", entityName, err)
	}
	time.Sleep(seedDelay())
	return nil
}

func seedDelay() time.Duration {
	if raw := os.Getenv("SEED_DELAY_SECONDS"); raw != "" {
		if secs, err := strconv.Atoi(raw); err == nil && secs >= 0 {
			return time.Duration(secs) * time.Second
		}
	}
	return 30 * time.Second
}

func ptrString(s string) *string {
	return &s
}

func ptrInt32(i int32) *int32 {
	return &i
}
