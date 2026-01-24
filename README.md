# DispatchGO 
DispatchGo solves the bottlenecks of synchronous webhook delivery by decoupling your fast API from slow, unreliable external networks.

Backed by a zero-contention PostgreSQL queue, Go channels, and a concurrent worker pool, it ensures high-throughput, race-condition-free job delivery, secured by HMAC, at scale.


## Features

- **RESTful API**: Clean HTTP endpoints for task management
- **Asynchronous Processing**: Worker pool with configurable concurrency
- **Retry Mechanism**: Automatic retries with exponential backoff
- **Task Status Tracking**: Real-time status updates (pending, processing, completed, failed)
- **Webhook Support**: HTTP callbacks for task completion/failure
- **PostgreSQL Storage**: Reliable data persistence
- **HMAC Authentication**: Secure API access
- **Database Migrations**: Version-controlled schema changes
- **Graceful Shutdown:** Intercepts `SIGINT`/`SIGTERM` to drain HTTP connections and complete in-flight DB transactions before termination.

## Architecture

```
DispatchGO
├── cmd/api/           # Main application entry point
|         ├── main.go       # HTTP server setup
|
├── internal/
|     ├── api/           # HTTP handlers and routing
|     ├── database/      # PostgreSQL connection
|     ├── models/        # Task and webhook structure
|     ├── store/         # Data access layer
|     ├── utils/         # Hmac
|     └── worker/        # Background task processing
|
├── config/            # Configuration management
├── migrations/        # Database schema files
├── .env              # enviornment variables
```

## Quick Start

### Prerequisites
- Go 1.19+
- PostgreSQL 12+

### Installation

1. Clone the repository:
```bash
git clone https://github.com/Bharat1Rajput/DispatchGo.git
cd DispatchGo
```

2. Install dependencies:
```bash
go mod download
```

3. Set up PostgreSQL database and update config

4. Run migrations:
```bash
# Apply database schema
psql -d yourDB -f migrations/001_create_task.sql
```

5. Start the server:
```bash
go run cmd/api/main.go
```

## API Endpoints

- `POST /tasks` - Create a new task

### Authentication
All endpoints require HMAC authentication in the `X-HMAC-Signature` header.
( HMAC (Hash-based Message Authentication Code) is a cryptographic security technique used to verify that data has not been tampered with and comes from a trusted sender.)

## Task Processing
Tasks are processed by a worker pool that:
1. Fetches pending tasks from the database
2. Processes them concurrently
3. Updates status and triggers webhooks
4. Handles failures with retry logic

### Retry Strategy
- Maximum 5 retries per task
- Exponential backoff: 2s, 4s, 8s delays
- Failed tasks marked as "dead letter"

## Database Schema
The system uses a single `tasks` table with:
- Unique ID 
- Type
- Payload data (webhook payload)
- Status (pending/processing/completed/failed)
- Error message
- Retry count and max retries
- Created/updated timestamps

- whats inside webhook payload - (destination url , method, event , data, headers, secret)

## 🧪 Testing

The system is tested using Go's native testing toolkit.

Run the test suite with verbose output:
```bash
go test ./... -v
```

## Live Testing with Webhook.site

You don't need to spin up a local receiver server to test DispatchGo. You can instantly verify the background worker and HMAC security using a free online webhook catcher.

**1. Get a test URL**
Go to [webhook.site](https://webhook.site/) and copy "Your unique URL".

**2. Dispatch a task**
Use the provided URL in your `POST /tasks` payload:

```bash
curl -X POST http://localhost:8080/tasks \
-H "Content-Type: application/json" \
-d '{
    "url": "[https://webhook.site/YOUR-UNIQUE-ID](https://webhook.site/YOUR-UNIQUE-ID)",
    "method": "POST",
    "event": "payment.success",
    "secret": "whsec_test_secret_123",
    "headers": {
    "X-Api-Key": "crm_key_88492"
               },
    "data": {
      "transaction_id": "txn_8923749823"
      }
}' 
```

## Development 

### Adding New Task Types
1. Extend the Task model in `internal/models/task.go`
2. Update the processor in `internal/worker/processor.go`
3. Add validation in the API handlers

### Database Changes
1. Create migration files in `migrations/`
2. Update the store layer in `internal/store/`
3. Run migrations before deployment

## Deployment

The application is stateless and can be deployed as:
- Docker container
- Kubernetes deployment
- Cloud service (App Engine, etc.)

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## License

MIT License - see LICENSE file for details
