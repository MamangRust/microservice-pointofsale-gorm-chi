package outbox

import (
	"context"
	"errors"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// stubExecutor satisfies InboxExecutor without a backing database. It is only
// used for argument-validation tests, which short-circuit before any SQL runs.
type stubExecutor struct{}

func (stubExecutor) Exec(context.Context, string, ...any) *gorm.DB { return nil }
func (stubExecutor) Raw(context.Context, string, ...any) *gorm.DB  { return nil }

// newMockDB returns a GORM session backed by sqlmock so queries are validated
// against expectations without a live PostgreSQL server.
func newMockDB(t *testing.T) (*gorm.DB, sqlmock.Sqlmock) {
	t.Helper()
	sqlDB, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to open sqlmock: %v", err)
	}
	db, err := gorm.Open(postgres.New(postgres.Config{Conn: sqlDB}), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open gorm db: %v", err)
	}
	return db, mock
}

func TestReserveValidatesKeys(t *testing.T) {
	for name, tc := range map[string]struct {
		tx           InboxExecutor
		consumerName string
		eventKey     string
	}{
		"nil executor":    {tx: nil, consumerName: "email-service-group", eventKey: "topic:evt-1"},
		"empty consumer":  {tx: stubExecutor{}, consumerName: "", eventKey: "topic:evt-1"},
		"empty event key": {tx: stubExecutor{}, consumerName: "email-service-group", eventKey: ""},
	} {
		t.Run(name, func(t *testing.T) {
			if _, _, _, err := Reserve(context.Background(), tc.tx, tc.consumerName, tc.eventKey, "topic", 0, 1); !errors.Is(err, ErrInvalidInboxKey) {
				t.Fatalf("expected ErrInvalidInboxKey, got %v", err)
			}
		})
	}
}

func TestMarkProcessedValidatesKeys(t *testing.T) {
	if err := MarkProcessed(context.Background(), nil, "email-service-group", "topic:evt-1", 1); !errors.Is(err, ErrInvalidInboxKey) {
		t.Fatalf("expected ErrInvalidInboxKey, got %v", err)
	}
	if err := MarkProcessed(context.Background(), stubExecutor{}, "", "topic:evt-1", 1); !errors.Is(err, ErrInvalidInboxKey) {
		t.Fatalf("expected ErrInvalidInboxKey, got %v", err)
	}
}

func TestReleaseValidatesKeys(t *testing.T) {
	if err := Release(context.Background(), nil, "email-service-group", "topic:evt-1", 1, nil); !errors.Is(err, ErrInvalidInboxKey) {
		t.Fatalf("expected ErrInvalidInboxKey, got %v", err)
	}
}

func TestReleaseRecordsProcessingError(t *testing.T) {
	db, mock := newMockDB(t)
	// The third argument must carry the processing error into last_error.
	mock.ExpectExec(`UPDATE consumer_inbox`).
		WithArgs("email-service-group", "topic:evt-1", "smtp down", sqlmock.AnyArg()).
		WillReturnResult(sqlmock.NewResult(0, 1))

	inbox, err := NewPostgresInbox(db)
	if err != nil {
		t.Fatalf("NewPostgresInbox returned error: %v", err)
	}
	if err := inbox.Release(context.Background(), "email-service-group", "topic:evt-1", 1, errors.New("smtp down")); err != nil {
		t.Fatalf("Release returned error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestReserveUsesLeaseFencedUpsert(t *testing.T) {
	db, mock := newMockDB(t)
	// The reserve path must run the lease-fenced upsert (ON CONFLICT with an
	// expired-lease guard) rather than a plain insert.
	mock.ExpectQuery(`ON CONFLICT \(consumer_name, event_key\)`).
		WithArgs("email-service-group", "topic:evt-1", "topic", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"reservation_version", "reserved"}).AddRow(int64(1), true))

	inbox, err := NewPostgresInbox(db)
	if err != nil {
		t.Fatalf("NewPostgresInbox returned error: %v", err)
	}
	reserved, processed, version, err := inbox.Reserve(context.Background(), "email-service-group", "topic:evt-1", "topic", 0, 1)
	if err != nil {
		t.Fatalf("Reserve returned error: %v", err)
	}
	if !reserved || processed || version != 1 {
		t.Fatalf("expected reserved=true processed=false version=1, got reserved=%v processed=%v version=%d", reserved, processed, version)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}

func TestReserveDetectsAlreadyProcessed(t *testing.T) {
	db, mock := newMockDB(t)
	// The upsert did not reserve (event already processed), so the fallback
	// status lookup must report the processed state.
	mock.ExpectQuery(`ON CONFLICT \(consumer_name, event_key\)`).
		WithArgs("email-service-group", "topic:evt-1", "topic", sqlmock.AnyArg(), sqlmock.AnyArg()).
		WillReturnRows(sqlmock.NewRows([]string{"reservation_version", "reserved"}).AddRow(int64(0), false))
	mock.ExpectQuery(`SELECT status FROM consumer_inbox`).
		WithArgs("email-service-group", "topic:evt-1").
		WillReturnRows(sqlmock.NewRows([]string{"status"}).AddRow("processed"))

	inbox, err := NewPostgresInbox(db)
	if err != nil {
		t.Fatalf("NewPostgresInbox returned error: %v", err)
	}
	reserved, processed, version, err := inbox.Reserve(context.Background(), "email-service-group", "topic:evt-1", "topic", 0, 1)
	if err != nil {
		t.Fatalf("Reserve returned error: %v", err)
	}
	if reserved || !processed || version != 0 {
		t.Fatalf("expected reserved=false processed=true version=0, got reserved=%v processed=%v version=%d", reserved, processed, version)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("unmet sqlmock expectations: %v", err)
	}
}
