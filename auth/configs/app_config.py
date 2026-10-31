from dotenv import load_dotenv
from pathlib import Path
import os
import logging

BASE_DIR = Path(__file__).resolve().parent.parent
load_dotenv(BASE_DIR.parent / ".env")


class Config:
    """Database configuration"""
    SQLALCHEMY_DATABASE_URI = os.getenv("DATABASE_URL")
    SQLALCHEMY_TRACK_MODIFICATIONS = os.getenv("TRACK_MODIFICATIONS")
    WATCH_INTERVAL = 10

    """Application configuration"""
    HOST = os.getenv("HOSTS")
    PORT = int(os.getenv("PORTS"))
    DEBUG = os.getenv("DEBUGS")

    """JWT configuration"""
    JWT_SECRET_KEY = os.getenv("JWT_SECRET_KEY")
    JWT_PUBLIC_KEY = os.getenv("JWT_PUBLIC_KEY")
    JWT_ALGORITHM = os.getenv("JWT_ALGORITHM")
    JWT_ACCESS_TTL = os.getenv("JWT_ACCESS_TTL")
    JWT_REFRESH_TTL = os.getenv("JWT_REFRESH_TTL")

    """Cookie configuration"""
    COOKIE_SECURE = os.getenv("COOKIE_SECURE", "false").lower() == "true"

    """Argon2 configuration"""
    ARGON2_VERSION = int(os.getenv("ARGON2_VERSION"))
    ARGON2_MEMORY = int(os.getenv("ARGON2_MEMORY"))
    ARGON2_ITERATIONS = int(os.getenv("ARGON2_ITERATIONS"))
    ARGON2_PARALLELISM = int(os.getenv("ARGON2_PARALLELISM"))
    ARGON2_KEY_LENGTH = int(os.getenv("ARGON2_KEY_LENGTH"))
    ARGON2_SALT_LENGTH = int(os.getenv("ARGON2_SALT_LENGTH"))

    """Logging configuration"""
    LOG_LEVEL = logging.DEBUG
    LOG_FORMAT = "%(asctime)s - %(name)s - %(levelname)s - [trace_id=%(trace_id)s] - %(message)s"
    LOG_FILE = "logs/auth-service.log"