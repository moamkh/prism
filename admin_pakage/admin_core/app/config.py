import os
from pydantic_settings import BaseSettings


class Settings(BaseSettings):
    database_url: str = os.getenv(
        "DATABASE_URL",
        "postgresql://postgres:postgres@localhost:5432/reverse_proxy_manager_db",
    )
    admin_port: int = int(os.getenv("ADMIN_PORT", "8000"))
    secret_key: str = os.getenv("SECRET_KEY", "changeme")
    cors_origins: str = os.getenv("CORS_ORIGINS", "*")
    display_timezone: str = os.getenv("DISPLAY_TIMEZONE", "UTC")

    class Config:
        env_file = ".env"
        env_file_encoding = "utf-8"


settings = Settings()
