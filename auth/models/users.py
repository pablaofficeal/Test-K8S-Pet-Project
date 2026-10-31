from datetime import datetime
from sqlalchemy import Uuid
from db.init_db import db

class Users(db.Model):
    uuid = db.Column(Uuid, primary_key=True)
    username = db.Column(db.String, nullable=False)
    email = db.Column(db.String, nullable=False)
    password = db.Column(db.String, nullable=False)
    created_at = db.Column(db.DateTime, nullable=False, default=datetime.now)
    updated_at = db.Column(db.DateTime, nullable=False, default=datetime.now, onupdate=datetime.now)

    def __init__(self, uuid, username, email, password):
        self.uuid = uuid
        self.username = username
        self.email = email
        self.password = password