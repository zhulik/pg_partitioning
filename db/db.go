package db

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"pg_partitioning/db/models"

	"github.com/google/uuid"
	"github.com/stephenafamo/bob"
	"github.com/stephenafamo/bob/dialect/psql"
	"github.com/stephenafamo/bob/dialect/psql/sm"
	"github.com/stephenafamo/bob/types"
	"github.com/stephenafamo/scan"
)

var Events = psql.NewTablex[*models.EventsDefault, models.EventsDefaultSlice, *models.EventsDefaultSetter]("", "events")

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "postgres"
)

var (
	Logger = slog.New(slog.NewTextHandler(os.Stdout, nil))
)

type SlogPrinter struct {
	logger *slog.Logger
}

func (s SlogPrinter) PrintQuery(query string, args ...any) {
	s.logger.Info(query, "args", args)
}

// GetConnection returns a database connection
func GetConnection() (bob.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
		host, port, user, password, dbname)

	db, err := bob.Open("postgres", connStr)
	if err != nil {
		return bob.DB{}, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		return bob.DB{}, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// InsertEvent inserts a new event with the given payload
func InsertEvent(ctx context.Context, db bob.DB, name string, actorID, aggregateID uuid.UUID, payload string) error {
	evtUUID := uuid.New()
	_, err := Events.Insert(&models.EventsDefaultSetter{
		UUID:        &evtUUID,
		Name:        &name,
		ActorID:     &actorID,
		AggregateID: &aggregateID,
		Payload:     &types.JSON[json.RawMessage]{Val: json.RawMessage(payload)},
		// CreatedAt:   nil,
	}).One(ctx, bob.DebugToPrinter(db, SlogPrinter{Logger}))

	if err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}

	return nil
}

// GetPartitionCount returns the number of partitions for the events table
func GetPartitionCount(ctx context.Context, db bob.DB) (int, error) {
	count, err := bob.One(ctx, bob.DebugToPrinter(db, SlogPrinter{Logger}),
		psql.Select(
			sm.Columns(psql.F("count", "*")),
			sm.From("pg_inherits"),
			sm.LeftJoin("pg_class parent").On(psql.Quote("pg_inherits", "inhparent").EQ(psql.Quote("parent", "oid"))),
			sm.LeftJoin("pg_class child").On(psql.Quote("pg_inherits", "inhrelid").EQ(psql.Quote("child", "oid"))),
			sm.Where(psql.Quote("parent", "relname").EQ(psql.S("events"))),
		),
		scan.SingleColumnMapper[int])
	return count, err
}

func RunMaintenance(ctx context.Context, db bob.DB) error {
	_, err := bob.DebugToPrinter(db, SlogPrinter{Logger}).ExecContext(ctx, "call run_maintenance_proc()")
	return err
}
