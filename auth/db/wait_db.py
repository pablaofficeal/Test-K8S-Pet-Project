from sqlalchemy import inspect, text
from sqlalchemy.exc import SQLAlchemyError
from db.init_db import db
import logging
import time


def wait_for_db(interval):
    while True:
        try:
            with db.engine.connect() as conn:
                conn.execute(text("SELECT 1"))
            return
        except SQLAlchemyError as err:
            logging.warning("postgres not ready, retry in %ss: %s", interval, err)
            time.sleep(interval)


def wait_for_tables(interval, *models):
    names = [model.__table__.name for model in models]
    while True:
        try:
            missing = [name for name in names if not inspect(db.engine).has_table(name)]
        except SQLAlchemyError as err:
            missing = names
            logging.warning("postgres lost while checking tables: %s", err)
        if not missing:
            return
        logging.warning("tables not ready %s, retry in %ss", missing, interval)
        time.sleep(interval)