from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(
        env_file=".env", env_ignore_empty=True, extra="ignore"
    )
    PROJECT_NAME: str = "Faust-Processor"
    KAFKA_BROKER: str
    KAFKA_PORT: int
    SCHEMA_REGISTRY_URL: str
    TOPIC_MATCH_DATA: str
    TOPIC_MATCH_TO_STORE: str
    TOPIC_MATCH_PLAYER_DATA: str
    TOPIC_MATCH_TEAM_DATA: str
    TOPIC_MATCH_GAME_DATA: str


settings = Settings()
