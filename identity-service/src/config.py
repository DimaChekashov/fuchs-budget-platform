from pydantic_settings import BaseSettings

class Settings(BaseSettings):
    postgres_dsn: str = "postgresql+asyncpg://identity:secret@localhost/identity_db"
    jwt_secret_key: str = "my-secret"
    
    class Config:
        env_file = ".env"

settings = Settings()