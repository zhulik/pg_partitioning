package main

import (
	"encoding/json"
	"log"
	"math/rand"
	"time"

	"pg_partitioning/db"

	"github.com/google/uuid"
)

const (
	// Events per second to write
	eventsPerSecond = 1000

	// Sleep duration between batches of events
	sleepDuration = time.Second / eventsPerSecond
)

// RandomPayload generates a random JSON payload
type RandomPayload struct {
	ID        int       `json:"id"`
	Timestamp time.Time `json:"timestamp"`
	Value     float64   `json:"value"`
	Message   string    `json:"message"`
	Tags      []string  `json:"tags"`
}

// Generate random messages for variety
var messages = []string{
	"System startup",
	"User login",
	"Data processing",
	"Transaction completed",
	"Error occurred",
	"Warning detected",
	"Scheduled task",
	"Background job",
	"Cache invalidated",
	"Resource allocated",
}

var names = []string{
	"user.created", "user.updated", "user.deleted",
	"transaction.started", "transaction.completed", "transaction.failed",
	"task.started", "task.completed", "task.failed",
	"job.started", "job.completed", "job.failed",
	"cache.invalidated", "cache.updated", "cache.cleared",
	"resource.allocated", "resource.released",
}

// Generate random tags for variety
var tags = []string{
	"system", "user", "data", "transaction", "error",
	"warning", "task", "job", "cache", "resource",
}

func randomName() string {
	return names[rand.Intn(len(messages))]
}

func main() {
	log.Println("Starting event writer...")

	// Connect to the database
	database, err := db.GetConnection()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer database.Close()

	eventCount := 0
	startTime := time.Now()

	for {
		// Generate a random payload
		payload := generateRandomPayload(eventCount)

		// Convert to JSON
		jsonPayload, err := json.Marshal(payload)
		if err != nil {
			log.Printf("Error marshaling JSON: %v", err)
			continue
		}

		// Insert the event
		if err := db.InsertEvent(database, randomName(), uuid.NewString(), uuid.NewString(), string(jsonPayload)); err != nil {
			log.Printf("Error inserting event: %v", err)
		} else {
			eventCount++

			if eventCount%eventsPerSecond == 0 {
				elapsed := time.Since(startTime)
				rate := float64(eventCount) / elapsed.Seconds()
				log.Printf("Inserted %d events (%.2f events/sec)", eventCount, rate)
			}
		}

		// Sleep to maintain the desired rate
		time.Sleep(sleepDuration)
	}
}

func generateRandomPayload(id int) RandomPayload {
	// Select random message and tags
	message := messages[rand.Intn(len(messages))]

	// Generate 1-3 random tags
	numTags := rand.Intn(3) + 1
	selectedTags := make([]string, numTags)
	for i := 0; i < numTags; i++ {
		selectedTags[i] = tags[rand.Intn(len(tags))]
	}

	return RandomPayload{
		ID:        id,
		Timestamp: time.Now(),
		Value:     rand.Float64() * 100,
		Message:   message,
		Tags:      selectedTags,
	}
}
