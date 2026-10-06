package main

import (
	"context"
	"log"

	"github.com/MamangRust/microservice-point-of-sale-pkg/database"
	"github.com/MamangRust/microservice-point-of-sale-pkg/database/seeder"
	"github.com/MamangRust/microservice-point-of-sale-pkg/dotenv"
	"github.com/MamangRust/microservice-point-of-sale-pkg/hash"
	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"gorm.io/gorm"
)

func main() {
	if err := dotenv.Viper(); err != nil {
		log.Fatalf("Error loading environment variables: %v", err)
	}

	l, err := logger.NewLogger("seeder", nil)
	if err != nil {
		log.Fatalf("Error creating logger: %v", err)
	}

	ctx := context.Background()

	// Each bounded context owns a separate PostgreSQL instance, so the seeder
	// opens one connection per context and routes each domain's seed data to
	// its owning instance.
	dbs := make(map[string]*gorm.DB, len(database.BoundedContexts))
	for _, c := range database.BoundedContexts {
		db, err := database.NewGormClientWithPrefix(l, database.ContextPrefix[c])
		if err != nil {
			log.Fatalf("Error connecting to %s database: %v", c, err)
		}
		dbs[c] = db
	}
	defer func() {
		for _, db := range dbs {
			if sqlDB, err := db.DB(); err == nil {
				_ = sqlDB.Close()
			}
		}
	}()

	s := seeder.NewSeeder(seeder.Deps{
		DBs:    dbs,
		Ctx:    ctx,
		Logger: l,
		Hash:   hash.NewHashingPassword(),
	})

	if err := s.Run(); err != nil {
		log.Fatalf("Seeding failed: %v", err)
	}

	l.Info("Seeding completed successfully.")
}
