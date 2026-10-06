package server

import "time"

// Config defines the common configuration for all gRPC services
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	Port           int
	OtelEndpoint   string

	// DBCluster is the env key prefix of the PostgreSQL instance this service
	// owns (e.g. "DB_SALES" for the order/order_item/transaction services). The
	// server opens the connection via database.NewGormClientWithPrefix, which
	// reads DB_<PREFIX>_HOST/PORT/NAME/USERNAME/PASSWORD. Each bounded context
	// has its own instance, so this must be set to route the service to its
	// own database rather than a shared/generic one.
	DBCluster string
}

// Default constants for gRPC server
const (
	DefaultMaxConcurrentConn = 1024
	DefaultWindowSize        = 16 * 1024 * 1024
	DefaultKeepaliveTime     = 20 * time.Second
	DefaultKeepaliveTimeout  = 5 * time.Second
	DefaultMinKeepaliveTime  = 5 * time.Second

	MonitoringInterval     = 30 * time.Second
	CleanupInterval        = 120 * time.Second
	CacheRefCountThreshold = 500

	ShutdownTimeout = 30 * time.Second

	RedisDialTimeout  = 5 * time.Second
	RedisReadTimeout  = 3 * time.Second
	RedisWriteTimeout = 3 * time.Second
	RedisPoolSize     = 10
	RedisMinIdleConns = 3
)
