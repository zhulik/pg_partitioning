package main

import (
	"log"
	"time"

	"pg_partitioning/db"
)

const (
	// How often to run the partition manager (60 seconds)
	managerInterval = 10 * time.Second

	// Duration of each partition (10 seconds)
	partitionDuration = 10 * time.Second

	// Number of partitions to create in each run (6 partitions)
	partitionsPerRun = 10
)

var (
	firstPartitionTime = time.Date(2025, time.June, 29, 22, 0, 0, 0, time.UTC)
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

	// Get the end time of the last partition as the starting point
	origLastPartitionEndTime, err := db.GetLastPartitionEndTime(database)
	lastPartitionEndTime := origLastPartitionEndTime
	if err != nil {
		log.Printf("Warning: Failed to get last partition end time: %v", err)
		lastPartitionEndTime = firstPartitionTime
	}

	for lastPartitionEndTime.Before(time.Now().Add(partitionsPerRun * partitionDuration)) {
		endTime := lastPartitionEndTime.Add(partitionDuration)
		if err := db.CreatePartition(database, lastPartitionEndTime, endTime); err != nil {
			panic(err)
		}

		lastPartitionEndTime = endTime
	}

	indexStart := firstPartitionTime

	for indexStart.Before(time.Now().Add(-2 * time.Minute)) {
		indexStart = indexStart.Add(partitionDuration)
	}

	err = db.CreatePartitionIndex(database, indexStart, partitionDuration)
	if err != nil {
		log.Printf("Warning: Failed to create partition index: %v", err)
	}

	count, err := db.GetPartitionCount(database)
	if err != nil {
		log.Printf("Warning: Failed to get partition count: %v", err)
	} else {
		log.Printf("Total partitions: %d", count)
	}

	return nil
}
