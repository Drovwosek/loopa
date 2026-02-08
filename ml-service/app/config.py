from pydantic_settings import BaseSettings
from typing import Optional


class Settings(BaseSettings):
    HF_TOKEN: Optional[str] = None
    DEVICE: str = "cpu"
    DIARIZATION_MODEL: str = "pyannote/speaker-diarization-3.1"
    HOST: str = "0.0.0.0"
    PORT: int = 8001
    WHISPER_MODEL: str = "large-v3"
    WHISPER_DEVICE: str = "cuda"
    WHISPER_COMPUTE_TYPE: str = "float16"

    class Config:
        env_file = ".env"


settings = Settings()
