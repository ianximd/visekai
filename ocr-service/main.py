"""
FastAPI server for DeepSeek-OCR service
"""
import os
from fastapi import FastAPI, File, UploadFile, HTTPException
from fastapi.middleware.cors import CORSMiddleware
from pydantic import BaseModel
from typing import Optional, List
import uvicorn

from api.routes import router as api_router
from core.config import settings
from core.logging import logger

# Create FastAPI app
app = FastAPI(
    title="DeepSeek-OCR Service",
    description="OCR service using DeepSeek-OCR model",
    version="1.0.0"
)

# CORS middleware
app.add_middleware(
    CORSMiddleware,
    allow_origins=["*"],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Include API routes
app.include_router(api_router, prefix="/api/v1")

@app.get("/")
async def root():
    return {
        "service": "DeepSeek-OCR",
        "version": "1.0.0",
        "status": "running"
    }

@app.get("/health")
async def health_check():
    return {
        "status": "healthy",
        "service": "ocr-service",
        "model": settings.MODEL_PATH
    }

@app.on_event("startup")
async def startup_event():
    """Initialize service on startup"""
    logger.info("Starting DeepSeek-OCR service...")
    logger.info(f"Model path: {settings.MODEL_PATH}")
    logger.info(f"Base size: {settings.BASE_SIZE}")
    logger.info(f"Image size: {settings.IMAGE_SIZE}")
    logger.info(f"Crop mode: {settings.CROP_MODE}")
    # TODO: Load model here
    logger.info("Service started successfully")

@app.on_event("shutdown")
async def shutdown_event():
    """Cleanup on shutdown"""
    logger.info("Shutting down DeepSeek-OCR service...")
    # TODO: Cleanup model resources

if __name__ == "__main__":
    uvicorn.run(
        "main:app",
        host="0.0.0.0",
        port=8000,
        reload=True,
        log_level="info"
    )
