import secrets
from typing import Literal
from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(
        env_file=".env", env_ignore_empty=True, extra="ignore"
    )
    API_V1_STR: str = "/api/v1"
    SECRET_KEY: str = secrets.token_urlsafe(32)
    ACCESS_TOKEN_EXPIRE_MINUTES: int = 60 * 24 * 8
    DOMAIN: str = "localhost"
    ENVIRONMENT: Literal["local", "staging", "production"] = "local"
    PROJECT_NAME: str = "Backend"

    REDIS_PORT: int
    REDIS_HOST: str
    REDIS_DB: str

    KAFKA_BROKER: str
    KAFKA_PORT: int
    MIN_COMMIT_COUNT: int

    TOPIC_MATCH_DATA: str


settings = Settings()
