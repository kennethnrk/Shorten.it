## shorten.it Go API

This is a basic Go API for URL shortening backed by Redis and Astra DB.

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