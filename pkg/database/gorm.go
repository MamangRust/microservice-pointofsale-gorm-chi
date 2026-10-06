package database

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/MamangRust/microservice-point-of-sale-pkg/logger"
	"github.com/spf13/viper"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"
)

// NewGormClient connects to the base database using the generic DB_* keys
// and returns a *gorm.DB instance.
func NewGormClient(log logger.LoggerInterface) (*gorm.DB, error) {
	return NewGormClientWithPrefix(log, "DB")
}

// NewGormClientWithPrefix connects to the database configured via the given
// prefix keys (e.g. DB_ROLE_HOST, DB_ROLE_NAME) with fallback to the base
// DB_* keys.
func NewGormClientWithPrefix(log logger.LoggerInterface, prefix string) (*gorm.DB, error) {
	if prefix == "" {
		prefix = "DB"
	}

	host := viper.GetString(fmt.Sprintf("%s_HOST", prefix))
	if host == "" {
		host = viper.GetString("DB_HOST")
	}
	port := viper.GetString(fmt.Sprintf("%s_PORT", prefix))
	if port == "" {
		port = viper.GetString("DB_PORT")
	}
	user := viper.GetString(fmt.Sprintf("%s_USERNAME", prefix))
	if user == "" {
		user = viper.GetString("DB_USERNAME")
	}
	dbname := viper.GetString(fmt.Sprintf("%s_NAME", prefix))
	if dbname == "" {
		dbname = viper.GetString("DB_NAME")
	}
	password := viper.GetString(fmt.Sprintf("%s_PASSWORD", prefix))
	if password == "" {
		password = viper.GetString("DB_PASSWORD")
	}

	// Fail fast when the resolved connection target is incomplete. Each service
	// must target its own bounded-context instance via the DB_<CONTEXT>_* keys;
	// silently falling back to a generic/shared instance would let a service
	// write to the wrong database (FKs cannot cross instances anyway).
	if host == "" || port == "" || dbname == "" {
		return nil, fmt.Errorf(
			"incomplete database config for prefix %q: host=%q port=%q name=%q (set DB_%s_HOST/PORT/NAME)",
			prefix, host, port, dbname, strings.ToUpper(prefix),
		)
	}

	dsn := fmt.Sprintf(
		"host=%s port=%s user=%s dbname=%s password=%s sslmode=disable",
		host, port, user, dbname, password,
	)

	maxOpenConns := viper.GetInt("DB_MAX_OPEN_CONNS")
	if maxOpenConns <= 0 {
		maxOpenConns = 100
	}
	maxIdleConns := viper.GetInt("DB_MIN_IDLE_CONNS")
	if maxIdleConns <= 0 {
		maxIdleConns = 50
	}
	connMaxLifetime := viper.GetDuration("DB_CONN_MAX_LIFETIME")
	if connMaxLifetime == 0 {
		connMaxLifetime = time.Hour
	}
	connMaxIdleTime := viper.GetDuration("DB_CONN_MAX_IDLE_TIME")
	if connMaxIdleTime == 0 {
		connMaxIdleTime = 30 * time.Minute
	}

	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: gormlogger.Default.LogMode(gormlogger.Silent),
		NowFunc: func() time.Time {
			return time.Now().UTC()
		},
		PrepareStmt: true,
	})
	if err != nil {
		log.Error("Failed to connect to database via GORM", zap.Error(err))
		return nil, fmt.Errorf("failed to connect to database via GORM: %w", err)
	}

	sqlDB, err := gormDB.DB()
	if err != nil {
		log.Error("Failed to get underlying sql.DB from GORM", zap.Error(err))
		return nil, fmt.Errorf("failed to get underlying sql.DB: %w", err)
	}

	sqlDB.SetMaxOpenConns(maxOpenConns)
	sqlDB.SetMaxIdleConns(maxIdleConns)
	sqlDB.SetConnMaxLifetime(connMaxLifetime)
	sqlDB.SetConnMaxIdleTime(connMaxIdleTime)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := sqlDB.PingContext(ctx); err != nil {
		log.Error("Failed to ping database via GORM", zap.Error(err))
		return nil, fmt.Errorf("failed to ping database via GORM: %w", err)
	}

	log.Debug("GORM database connection established",
		zap.String("prefix", prefix),
		zap.String("dbname", dbname),
	)

	return gormDB, nil
}
