package db

import (
	"database/sql"
	"fmt"

	"github.com/google/uuid"
	_ "github.com/lib/pq" // sql driver
)

const (
	host     = "localhost"
	port     = 5432
	user     = "postgres"
	password = "postgres"
	dbname   = "postgres"
)

// GetConnection returns a database connection
func GetConnection() (*sql.DB, error) {
	connStr := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s",
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

func RunMaintenance(db *sql.DB) error {
	_, err := db.Exec(`
		CALL run_maintenance_proc();
	`)
	return err
}
