import json
from unittest.mock import MagicMock

import pytest

from configs.app_config import Config
from utils.mq import verification_publisher
from utils.mq.verification_publisher import EXCHANGE, ROUTING_KEY, publish_verification_requested


@pytest.fixture()
def rabbit(monkeypatch):
    monkeypatch.setattr(Config, "RABBITMQ_AUTH_URL", "amqp://auth:secret@rabbit:5672/cadence")
    connection = MagicMock()
    channel = connection.channel.return_value
    monkeypatch.setattr(verification_publisher.pika, "BlockingConnection", MagicMock(return_value=connection))
    return connection, channel


def test_empty_url_skips_publishing(monkeypatch):
    monkeypatch.setattr(Config, "RABBITMQ_AUTH_URL", "")
    connect = MagicMock()
    monkeypatch.setattr(verification_publisher.pika, "BlockingConnection", connect)
    assert publish_verification_requested("user-1", "trace-1") is False
    connect.assert_not_called()


def test_publishes_event_with_confirm_and_mandatory(rabbit):
    connection, channel = rabbit
    assert publish_verification_requested("user-1", "trace-1") is True
    channel.confirm_delivery.assert_called_once()
    kwargs = channel.basic_publish.call_args.kwargs
    assert kwargs["exchange"] == EXCHANGE
    assert kwargs["routing_key"] == ROUTING_KEY
    assert kwargs["mandatory"] is True
    assert kwargs["properties"].delivery_mode == 2
    assert kwargs["properties"].content_type == "application/json"
    assert json.loads(kwargs["body"]) == {"event": ROUTING_KEY, "user_uuid": "user-1", "trace_id": "trace-1"}
    connection.close.assert_called_once()


def test_connection_error_returns_false(monkeypatch):
    monkeypatch.setattr(Config, "RABBITMQ_AUTH_URL", "amqp://auth:secret@rabbit:5672/cadence")
    monkeypatch.setattr(verification_publisher.pika, "BlockingConnection", MagicMock(side_effect=OSError("refused")))
    assert publish_verification_requested("user-1", "trace-1") is False


def test_publish_error_returns_false_and_closes_connection(rabbit):
    connection, channel = rabbit
    channel.basic_publish.side_effect = RuntimeError("nacked")
    assert publish_verification_requested("user-1", "trace-1") is False
    connection.close.assert_called_once()


def test_invalid_url_returns_false(monkeypatch):
    monkeypatch.setattr(Config, "RABBITMQ_AUTH_URL", "not a url://")
    assert publish_verification_requested("user-1", "trace-1") is False
