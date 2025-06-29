# PostgreSQL Partitioning Demo

This project demonstrates PostgreSQL table partitioning with a practical example. It creates a partitioned table and continuously writes events to it while managing partitions automatically.

## Features

- PostgreSQL 16 running in Docker with persistent storage
- Table partitioned by time range (`created_at` field)
- Automatic partition management (creates 6 partitions every 60 seconds)
- Continuous event generation (10 events per second)
- JSONB payload support

## Components

1. **PostgreSQL Database**: Runs in a Docker container with a volume for persistence
2. **Partition Manager**: Creates 6 partitions (10 seconds each) every 60 seconds
3. **Event Writer**: Continuously writes 10 events per second with random JSON payloads

## Project Structure

```
.
├── cmd
│   ├── event_writer      # Writes events to the database
│   └── partition_manager # Creates and manages partitions
├── db                    # Database connection and operations
├── docker-compose.yml    # Docker configuration
├── go.mod                # Go module definition
├── init.sql              # SQL initialization script
├── main.go               # Main orchestration code
└── README.md             # This file
```

## How It Works

1. The main application starts PostgreSQL using Docker Compose
2. The initialization script creates a partitioned table
3. The partition manager runs every 60 seconds, creating 6 partitions (10 seconds each)
4. The event writer continuously writes events at a rate of 10 per second
5. Each event has a timestamp and a random JSON payload

## Running the Demo

```bash
# Make sure Docker is running
go run main.go
```

The application will:
1. Start PostgreSQL
2. Wait for it to be ready
3. Start the partition manager and event writer
4. Run until you press Ctrl+C
5. Clean up resources on exit

## Technical Details

### Table Structure

```sql
CREATE TABLE events (
    id SERIAL,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    payload JSONB NOT NULL,
    PRIMARY KEY (id, created_at)
) PARTITION BY RANGE (created_at);
```

### Partitioning Strategy

- Range partitioning by `created_at` timestamp
- Each partition covers a 10-second time range
- New partitions are created automatically every 60 seconds

### Event Payload

Each event contains a JSON payload with:
- ID
- Timestamp
- Random value
- Random message
- Random tags

## Requirements

- Go 1.24 or later
- Docker and Docker Compose
- PostgreSQL client (for manual inspection if desired)