package tests

import (
	"fmt"
	"time"

	"github.com/spf13/viper"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

// clickhouseImage is the ClickHouse server used by the stats integration tests.
// The reader/backfill suites seed it with real aggregates and assert the values
// that come back over the wire.
const clickhouseImage = "clickhouse/clickhouse-server:24.8-alpine"

// SetupStatsEnv starts a ClickHouse testcontainer and points the process-global
// viper config at it. pkg/clickhouse.NewClient / ApplySchema and stats-writer's
// backfill both read the CLICKHOUSE_* keys from viper, so the rest of the stats
// test setup can open a real connection without any extra wiring.
//
// viper.Set is used (rather than os.Setenv) so the values are resolved
// regardless of whether viper.AutomaticEnv has been enabled. Only the stats
// suites touch ClickHouse, so mutating the singleton here is safe.
//
// The container handle is stored on the suite so BaseTestSuite.TearDownSuite
// can terminate it.
func (s *BaseTestSuite) SetupStatsEnv() {
	ctx := s.Ctx

	container, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: testcontainers.ContainerRequest{
			Image:        clickhouseImage,
			ExposedPorts: []string{"9000/tcp", "8123/tcp"},
			Env: map[string]string{
				"CLICKHOUSE_DB":                        "default",
				"CLICKHOUSE_DEFAULT_ACCESS_MANAGEMENT": "1",
				"CLICKHOUSE_SKIP_USER_SETUP":           "1",
			},
			WaitingFor: wait.ForHTTP("/ping").
				WithPort("8123/tcp").
				WithStartupTimeout(2 * time.Minute),
		},
		Started: true,
	})
	s.Require().NoError(err)
	s.CHContainer = container

	host, err := container.Host(ctx)
	s.Require().NoError(err)
	mapped, err := container.MappedPort(ctx, "9000")
	s.Require().NoError(err)

	// Point viper at the container's native protocol port.
	viper.Set("CLICKHOUSE_ADDR", fmt.Sprintf("%s:%s", host, mapped.Port()))
	viper.Set("CLICKHOUSE_DATABASE", "default")
	viper.Set("CLICKHOUSE_USERNAME", "default")
	viper.Set("CLICKHOUSE_PASSWORD", "")
}
