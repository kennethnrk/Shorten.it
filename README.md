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


