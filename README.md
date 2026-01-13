## shorten.it Go API

# Architecture Overview

This project implements a **high-throughput URL shortener** optimized for low-latency read operations and horizontal scalability. The architecture follows a **cache-first, write-through** design and leverages Go’s lightweight concurrency model to parallelize IO-bound operations.


---

## Core Components

### 1. API Layer (Go)

- Built using Go’s `net/http`
- Handles URL shortening and redirection
- Uses **goroutines** to parallelize:
  - Writes to Redis
  - Writes to CassandraDB
- Designed to be stateless for horizontal scaling

---

### 2. ID Generation (Base62 Encoding)

- Each long URL is mapped to a **Base62-encoded identifier**
- Produces short, URL-safe strings (`[a-zA-Z0-9]`)
- Variable-length encoding allows:
  - Better space utilization
  - Higher key capacity without collisions

---

### 3. Redis Cache

- Acts as the **first read layer**
- Stores short → long URL mappings
- Configured with:
  - High read throughput
  - TTL policies tuned for hot URLs
- Achieves ~95% cache hit rate under load

**Access Pattern**
- Read: Cache-first
- Write: Write-through (Redis + Cassandra)

---

### 4. CassandraDB (Primary Data Store)

- Acts as the **source of truth**
- Optimized for:
  - Write-heavy workloads
  - Constant-time primary-key reads
- Data model:
  - Partition key: `short_url`
  - Value: `long_url`, metadata (timestamps, counters, etc.)
- Supports linear horizontal scaling

---

### 5. Concurrency Model

- Go **goroutines** used for:
  - Parallel IO operations
  - Non-blocking request handling
- Enables high throughput with minimal thread overhead

---

## Request Flow

### URL Shortening

1. Client sends long URL
2. Server generates Base62 short ID
3. In parallel:
   - Store mapping in CassandraDB
   - Cache mapping in Redis
4. Return short URL to client

---

### URL Redirection

1. Client requests short URL
2. Server queries Redis
   - **Cache hit** → return long URL
   - **Cache miss** → query CassandraDB
3. On cache miss:
   - Populate Redis
   - Redirect client

---

## Scalability & Fault Tolerance

- **Stateless Go services** allow horizontal scaling behind a load balancer
- Redis reduces DB read pressure
- Cassandra provides:
  - Replication
  - High availability
  - No single point of failure
- System remains operational under partial cache failures

---

## Performance Characteristics

- Optimized for read-heavy traffic
- Handles high concurrency with stable latency
- Suitable for:
  - URL shorteners
  - Redirect services
  - Read-dominant key-value workloads

---

## Technology Stack

- **Language:** Go
- **Cache:** Redis
- **Database:** CassandraDB
- **Load Testing:** Locust


### Requirements

- Go 1.25+
- Docker (optional, for containerized runs)

### Environment variables

Configuration is provided via environment variables. For local development you can use a `.env` file in the project root. Common variables include:

- `REDIS_ADDR`
- `REDIS_USERNAME`
- `REDIS_PASSWORD`
- `REDIS_DB_NO`
- `REDIS_COUNTER_INIT`
- `ASTRA_DB_ID`
- `ASTRA_DB_REGION`
- `KEYSPACE_NAME`
- `ASTRA_DB_URL`
- `ASTRA_DB_TOKEN`

### Local development

1. Create a `.env` file in the project root with the variables listed above.
2. Run the API from the project root:

```bash
go run ./cmd/api
```

### Docker

#### Build

From the project root:

```bash
docker build -t shortenit-api .
```

#### Run (with .env)

To run the container and load configuration from your local `.env`:

```bash
docker run --env-file .env -p 8080:8080 --name shortenit-api shortenit-api
```

The API will then be available at `http://localhost:8080`.

### Load testing with Locust

A Python/Locust-based load and stress test lives under the `testing` directory.

- **1. Create and activate the virtual environment**
- From the `testing` directory:

```bash
cd testing
python -m venv venv
source venv/bin/activate  # On Windows: venv\Scripts\activate
```

- **2. Install dependencies**
- With the virtual environment activated:

```bash
pip install -r requirements.txt
```

- **3. Run Locust against a local API**
- Ensure the Go API is running (e.g. `go run ./cmd/api` from the project root).
- Then, from the project root or `testing` directory:

```bash
locust -f testing/locustfile.py --host http://localhost:8080
```

Open the Locust web UI (by default at `http://localhost:8089`) to configure users, spawn rate, and start the test.

- **4. Run Locust in headless mode**
- Example headless run with 100 users for 5 minutes:

```bash
API_BASE_URL=http://localhost:8080 locust -f testing/locustfile.py --headless --users 100 --spawn-rate 10 --run-time 5m
```

The `API_BASE_URL` environment variable controls the default host for Locust, so you can easily target a different environment:

```bash
API_BASE_URL=https://your-remote-api.example.com locust -f testing/locustfile.py --headless --users 50 --spawn-rate 5 --run-time 10m
```