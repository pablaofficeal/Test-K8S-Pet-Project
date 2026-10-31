from contextvars import ContextVar
from flask import Flask, g, request
from configs.app_config import Config
from api.routers.auth_routers import register_all_blueprints
from db.init_db import db
from db.wait_db import wait_for_db, wait_for_tables
from db.import_all_models import *
import logging
import uuid

TRACE_ID_HEADER = "X-Request-ID"
trace_id_ctx: ContextVar[str] = ContextVar("trace_id", default="-")


class TraceIdFilter(logging.Filter):
    def filter(self, record):
        record.trace_id = trace_id_ctx.get()
        return True


app = Flask(__name__)
app.config.from_object(Config)
register_all_blueprints(app)
db.init_app(app)

logging.basicConfig(filename=Config.LOG_FILE, level=Config.LOG_LEVEL, format=Config.LOG_FORMAT, force=True)
for handler in logging.getLogger().handlers:
    handler.addFilter(TraceIdFilter())


@app.before_request
def attach_trace_id():
    g.trace_id = request.headers.get(TRACE_ID_HEADER) or str(uuid.uuid4())
    trace_id_ctx.set(g.trace_id)


@app.after_request
def propagate_trace_id(response):
    response.headers[TRACE_ID_HEADER] = g.get("trace_id", "-")
    return response

with app.app_context():
    wait_for_db(Config.WATCH_INTERVAL)
    wait_for_tables(Config.WATCH_INTERVAL, Users, RefreshSession)

if __name__ == "__main__":
    app.run(
        host=Config.HOST,
        port=Config.PORT,
        debug=Config.DEBUG,
    )