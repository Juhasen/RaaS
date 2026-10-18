import sys
from unittest.mock import MagicMock

# main.py connects to Postgres and Kafka at import time; stub them.
sys.modules["kafka_client"] = MagicMock()
sys.modules["db_utils"] = MagicMock()

from fastapi.testclient import TestClient  # noqa: E402

from main import app  # noqa: E402


def test_metrics_endpoint():
    client = TestClient(app)
    client.get("/health")
    r = client.get("/metrics")
    assert r.status_code == 200
    assert "http_requests_total" in r.text
