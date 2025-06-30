package main

import (
	"context"
	"log"
	"time"

	"pg_partitioning/db"
)

const (
	// How often to run the partition manager (60 seconds)
	managerInterval = 10 * time.Second
)

func main() {
	log.Println("Starting partition manager...")

	for {
		if err := createPartitions(); err != nil {
			log.Printf("Error creating partitions: %v", err)
		}

		log.Printf("Sleeping for %s before next partition creation cycle", managerInterval)
		time.Sleep(managerInterval)
	}
}

func createPartitions() error {
	// Connect to the database
	database, err := db.GetConnection()
	if err != nil {
		return err
	}
	defer database.Close()

	err = db.RunMaintenance(context.Background(), database)
	if err != nil {
		log.Printf("Warning: Failed to create partition index: %v", err)
	}

	count, err := db.GetPartitionCount(context.Background(), database)
	if err != nil {
		log.Printf("Warning: Failed to get partition count: %v", err)
	} else {
		log.Printf("Total partitions: %d", count)
	}

	return nil
}
