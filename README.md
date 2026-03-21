# DispatchGo – Distributed Webhook Dispatcher

Originally built as a monolith with a PostgreSQL-backed queue, DispatchGo was re-architected into decoupled microservices to remove delivery bottlenecks, isolate failures, and scale API ingestion independently from webhook execution.

DispatchGo is a production‑style, distributed webhook dispatcher built in Go.  
It turns slow, unreliable outbound webhooks into a **fast, authenticated API + durable background workers**, using **RabbitMQ** for jobs and **PostgreSQL** for persistence.

This is my flagship project: it showcases clean microservice design, robust error handling, and real‑world infrastructure wiring (Docker, healthchecks, graceful shutdowns).

---

## Architecture (High Level)

```text
Client → api-service (HTTP)
          ├─ HMAC-SHA256 auth middleware
          ├─ Validates payload + client_url
          └─ Publishes WebhookJob → RabbitMQ (topic exchange, durable queue)

worker-service
  ├─ Consumes jobs from RabbitMQ with a worker pool
  ├─ Posts payload to client_url (HTTP)
  ├─ Retries with exponential backoff
  └─ Persists status in Postgres (webhook_jobs table)
```

**Key technologies**
- Go 1.22, `chi` router, `zap` logging
- RabbitMQ (`amqp091-go`) for async jobs
- PostgreSQL for durable job history
- Docker Compose for a full local stack

---

## Quick Start (Full Stack in One Command)

From the project root:

```bash
docker-compose up --build
```

This starts:
- `postgres` (with `webhook_jobs` table migrated)
- `rabbitmq` (with management UI on `http://localhost:15672`)
- `api-service` on `http://localhost:8080`
- `worker-service` consuming and dispatching webhooks

Check API health:

```bash
curl http://localhost:8080/health
# -> {"status":"ok"}
```

---

## Send a Test Webhook

The API accepts signed webhooks over `POST /webhooks`.  
The body is signed with `HMAC-SHA256(body, HMAC_SECRET)` and sent in the `X-Signature` header as `sha256=<hex>`.

With `HMAC_SECRET=supersecretkey`:

```bash
BODY='{"payload":"{\"event\":\"user.created\",\"id\":\"42\"}","client_url":"https://httpbin.org/post"}'
SECRET="supersecretkey"
SIG="sha256=$(echo -n "$BODY" | openssl dgst -sha256 -hmac "$SECRET" | awk '{print $2}')"

curl -X POST http://localhost:8080/webhooks \
  -H "Content-Type: application/json" \
  -H "X-Signature: $SIG" \
  -d "$BODY"
```

Expected response:

```json
{
  "job_id": "uuid",
  "status": "pending",
  "message": "job accepted and queued for processing"
}
```

---

## Observe the System Working

- **RabbitMQ UI**:  
  Open `http://localhost:15672` (guest/guest) → Queues → `webhook.jobs`  
  You’ll see jobs appear as they are published and consumed.

- **Postgres job history**:

```bash
docker exec -it dispatchgo-postgres \
  psql -U dispatch_user -d dispatchgo \
  -c "SELECT id, status, retry_count, error, created_at FROM webhook_jobs ORDER BY created_at DESC LIMIT 10;"
```

You should see jobs flow through `pending → processing → success` (or `failed` with retries and error details).

---

## Why This Project Matters

- **Realistic architecture**: Two independent Go services, message broker, and database, wired together with Docker and healthchecks.
- **Robustness**: HMAC auth, durable queues, exponential backoff, idempotent DB writes, and graceful shutdown of HTTP and workers.
- **Clarity**: Clean package boundaries (`config`, `middleware`, `handler`, `broker`, `consumer`, `processor`, `repository`) designed for testing and extension.

This codebase is my reference implementation for how I like to design and ship reliable backend systems in Go. 
