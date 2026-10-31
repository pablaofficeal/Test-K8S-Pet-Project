import json
import logging
import pika
from configs.app_config import Config

EXCHANGE = "verify.events"
ROUTING_KEY = "verification.requested"


def publish_verification_requested(user_uuid, trace_id):
    url = Config.RABBITMQ_AUTH_URL
    if not url:
        logging.warning("RABBITMQ_AUTH_URL is empty, verification request for %s was not published", user_uuid)
        return False

    body = json.dumps({"event": ROUTING_KEY, "user_uuid": str(user_uuid), "trace_id": trace_id})

    try:
        params = pika.URLParameters(url)
        params.socket_timeout = 3
        params.blocked_connection_timeout = 3
        params.connection_attempts = 1
        connection = pika.BlockingConnection(params)
        try:
            channel = connection.channel()
            # confirm + mandatory: без них потерянное сообщение выглядит как успешная отправка
            channel.confirm_delivery()
            channel.basic_publish(
                exchange=EXCHANGE,
                routing_key=ROUTING_KEY,
                body=body,
                properties=pika.BasicProperties(content_type="application/json", delivery_mode=2),
                mandatory=True,
            )
        finally:
            connection.close()
    except Exception as err:
        logging.error("failed to publish verification request for %s: %s", user_uuid, err)
        return False

    logging.info("verification request published for %s", user_uuid)
    return True
