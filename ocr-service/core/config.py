"""
Configuration settings for OCR service
"""
import os
from pydantic_settings import BaseSettings

class Settings(BaseSettings):
    # Model settings
    MODEL_PATH: str = os.getenv("MODEL_PATH", "deepseek-ai/DeepSeek-OCR")
    CUDA_VISIBLE_DEVICES: str = os.getenv("CUDA_VISIBLE_DEVICES", "0")
    
    # Processing settings
    BASE_SIZE: int = int(os.getenv("BASE_SIZE", "1024"))
    IMAGE_SIZE: int = int(os.getenv("IMAGE_SIZE", "640"))
    CROP_MODE: bool = os.getenv("CROP_MODE", "true").lower() == "true"
    MIN_CROPS: int = int(os.getenv("MIN_CROPS", "2"))
    MAX_CROPS: int = int(os.getenv("MAX_CROPS", "6"))
    MAX_WORKERS: int = int(os.getenv("MAX_WORKERS", "4"))
    
    # Storage settings
    STORAGE_PATH: str = os.getenv("STORAGE_PATH", "/app/storage")
    
    # Redis settings
    REDIS_URL: str = os.getenv("REDIS_URL", "redis://localhost:6379")
    
    class Config:
        env_file = ".env"
        case_sensitive = True

settings = Settings()
