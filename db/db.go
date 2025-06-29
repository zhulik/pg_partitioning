package db

import (
	"database/sql"
	"fmt"
	"github.com/google/uuid"
	_ "github.com/lib/pq"
	"log"
	"strings"
	"time"
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "partitioning_demo"
)

// GetConnection returns a database connection
func GetConnection() (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Test the connection
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return db, nil
}

// CreatePartition creates a partition for the events table for the given time range
func CreatePartition(db *sql.DB, startTime, endTime time.Time) error {
	partitionName := fmt.Sprintf("events_%s", startTime.Format("20060102_150405"))

	query := fmt.Sprintf("CREATE TABLE IF NOT EXISTS %s PARTITION OF events FOR VALUES FROM ('%s') TO ('%s')",
		partitionName,
		startTime.Format(time.RFC3339Nano),
		endTime.Format(time.RFC3339Nano))

	_, err := db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create partition %s: %w", partitionName, err)
	}

	log.Printf("Created partition %s for time range %s to %s",
		partitionName, startTime.Format(time.RFC3339), endTime.Format(time.RFC3339))

	return nil
}

// InsertEvent inserts a new event with the given payload
func InsertEvent(db *sql.DB, name, actorID, aggregateID, payload string) error {
	query := `INSERT INTO events (uuid, name, actor_id, aggregate_id, payload) VALUES ($1, $2, $3, $4, $5)`

	_, err := db.Exec(query, uuid.NewString(), name, actorID, aggregateID, payload)
	if err != nil {
		return fmt.Errorf("failed to insert event: %w", err)
	}

	return nil
}

// GetPartitionCount returns the number of partitions for the events table
func GetPartitionCount(db *sql.DB) (int, error) {
	query := `
		SELECT count(*) 
		FROM pg_inherits 
		JOIN pg_class parent ON pg_inherits.inhparent = parent.oid 
		JOIN pg_class child ON pg_inherits.inhrelid = child.oid 
		WHERE parent.relname = 'events'
	`

	var count int
	err := db.QueryRow(query).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get partition count: %w", err)
	}

	return count, nil
}

// CreatePartitionIndex creates an index on name column for historical partitions
func CreatePartitionIndex(db *sql.DB, from time.Time, partitionDuration time.Duration) error {
	for from.Before(time.Now().Add(-20 * time.Second)) {
		partitionName := fmt.Sprintf("events_%s", from.Format("20060102_150405"))

		indexQuery := fmt.Sprintf("CREATE INDEX CONCURRENTLY %s_name_idx ON %s(name)",
			partitionName, partitionName)

		if _, err := db.Exec(indexQuery); err != nil {
			if !isDuplicateIndexError(err) {
				return fmt.Errorf("failed to create index on partition %s: %w", partitionName, err)
			}
			return nil
		}

		indexQuery = fmt.Sprintf("CREATE INDEX CONCURRENTLY %s_payload_idx ON %s USING gin (payload)",
			partitionName, partitionName)

		if _, err := db.Exec(indexQuery); err != nil {
			if !isDuplicateIndexError(err) {
				return fmt.Errorf("failed to create index on partition %s: %w", partitionName, err)
			}
			return nil
		}

		log.Printf("Created indexes on partition %s", partitionName)
		from = from.Add(partitionDuration)
	}

	return nil
}

// isDuplicateIndexError checks if the error is a duplicate index error
func isDuplicateIndexError(err error) bool {
	return strings.Contains(err.Error(), `relation "events_`) &&
		strings.Contains(err.Error(), `_idx" already exists`)
}

// GetLastPartitionEndTime returns the end time of the last partition
func GetLastPartitionEndTime(db *sql.DB) (time.Time, error) {
	query := `
		SELECT max(end_at) as end_at from (
			SELECT
				((regexp_matches(pg_get_expr(child.relpartbound, child.oid),
								$$FROM \('([^']+)'\) TO \('([^']+)'\)$$))[2])::timestamptz as end_at
			FROM
				pg_inherits
					JOIN
				pg_class parent ON pg_inherits.inhparent = parent.oid
					JOIN
				pg_class child ON pg_inherits.inhrelid = child.oid
			WHERE
				parent.relname = 'events'
		)
	`

	var partitionBound time.Time
	err := db.QueryRow(query).Scan(&partitionBound)
	if err != nil {
		return time.Time{}, fmt.Errorf("failed to get last partition bound: %w", err)
	}

	return partitionBound, nil
}
