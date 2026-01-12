import os
import random
import string

from dotenv import load_dotenv
from locust import HttpUser, between, task, events


# Load environment variables from a local .env if present.
# This is primarily for convenient local testing.
load_dotenv()


def random_long_url() -> str:
    """Generate a pseudo-random long URL for testing."""
    # Keep it deterministic-ish but varied enough to avoid total cache hits.
    path = "".join(random.choices(string.ascii_letters + string.digits, k=16))
    return f"https://example.com/{path}"


class ShortenItUser(HttpUser):
    """
    Simulated user for the shorten.it API.

    Tasks:
    - create_short_url: POST /forward with a random long_url
    - resolve_short_url: POST /backward with a known short_url (if any exist)
    - health_check: GET /healthz
    """

    wait_time = between(0.1, 0.5)

    # Shared pool of short URLs that have been created by any user.
    # This is kept simple and in-memory for load testing purposes.
    short_urls = []

    def on_start(self) -> None:
        """
        Optionally prime the user with a short URL so resolve task has data.
        """
        if not self.short_urls:
            self.create_short_url()

    @task(5)
    def create_short_url(self) -> None:
        """
        Call POST /forward to create or retrieve a short URL for a long URL.
        Weighted higher to simulate more writes, or adjust as needed.
        """
        long_url = random_long_url()
        payload = {"long_url": long_url}

        with self.client.post("/forward", json=payload, name="POST /forward", catch_response=True) as response:
            if response.status_code != 200:
                response.failure(f"Unexpected status code: {response.status_code} - {response.text}")
                return

            try:
                data = response.json()
            except Exception as exc:  # noqa: BLE001
                response.failure(f"Failed to decode JSON: {exc}")
                return

            short_url = data.get("short_url")
            if not short_url:
                response.failure(f"Missing short_url in response: {data}")
                return

            # Store for subsequent resolve requests.
            self.short_urls.append(short_url)
            response.success()

    @task(10)
    def resolve_short_url(self) -> None:
        """
        Call POST /backward to resolve a known short URL back to a long URL.

        This is typically the more frequent operation in a URL shortener,
        so it has a higher weight than create_short_url.
        """
        if not self.short_urls:
            # No data to resolve yet, fall back to creating one.
            self.create_short_url()
            return

        short_url = random.choice(self.short_urls)
        payload = {"short_url": short_url}

        with self.client.post("/backward", json=payload, name="POST /backward", catch_response=True) as response:
            if response.status_code != 200:
                response.failure(f"Unexpected status code: {response.status_code} - {response.text}")
                return

            try:
                data = response.json()
            except Exception as exc:  # noqa: BLE001
                response.failure(f"Failed to decode JSON: {exc}")
                return

            # Both "not_found" and a proper long_url are acceptable outcomes,
            # but we flag completely malformed responses.
            if "long_url" not in data and data.get("status") != "not_found":
                response.failure(f"Unexpected response body: {data}")
                return

            response.success()

    @task(1)
    def health_check(self) -> None:
        """
        Lightweight health check hitting GET /healthz.
        """
        with self.client.get("/healthz", name="GET /healthz", catch_response=True) as response:
            if response.status_code != 200:
                response.failure(f"Unexpected status code: {response.status_code} - {response.text}")
                return

            response.success()


@events.init_command_line_parser.add_listener
def _(parser) -> None:  # type: ignore[no-untyped-def]
    """
    Extend Locust's CLI to accept a default host from environment variables.

    Example:
        API_BASE_URL=http://localhost:8080 locust -f testing/locustfile.py
    """
    default_host = os.getenv("API_BASE_URL", "http://localhost:8080")
    parser.set_defaults(host=default_host)

