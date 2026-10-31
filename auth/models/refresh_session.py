from db.init_db import db


class RefreshSession(db.Model):
    __tablename__ = "refresh_tokens"

    sid = db.Column("s_id", db.String, primary_key=True)
    user_uuid = db.Column(db.String, nullable=False, index=True)
    token_hash = db.Column(db.String, nullable=False)
    expires_at = db.Column(db.DateTime(timezone=True), nullable=False)
    revoked = db.Column(db.Boolean, nullable=False, default=False)
    revoked_at = db.Column(db.DateTime(timezone=True), nullable=True)
    created_at = db.Column(db.DateTime(timezone=True), nullable=False)
    updated_at = db.Column(db.DateTime(timezone=True), nullable=False)