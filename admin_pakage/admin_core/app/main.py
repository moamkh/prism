import logging
import time
import uuid

from fastapi import FastAPI, Request
from fastapi.middleware.cors import CORSMiddleware
from fastapi.responses import JSONResponse
from app.config import settings
from app.api.v1 import router as api_v1_router
from app.exceptions import register_exception_handlers

# Configure root logger
logging.basicConfig(
    level=logging.INFO,
    format="%(asctime)s | %(levelname)-8s | %(message)s",
)
logger = logging.getLogger(__name__)

app = FastAPI(
    title="Reverse Proxy Manager API",
    description="Admin API for managing OpenAI-compatible reverse proxy backends",
    version="1.0.0",
)

# CORS
app.add_middleware(
    CORSMiddleware,
    allow_origins=[o.strip() for o in settings.cors_origins.split(",") if o.strip()],
    allow_credentials=True,
    allow_methods=["*"],
    allow_headers=["*"],
)

# Request logging middleware with timing
@app.middleware("http")
async def log_requests(request: Request, call_next):
    rid = str(uuid.uuid4())[:8]
    request.state.request_id = rid
    start = time.time()

    # Inject request-id into response headers for tracing
    response = await call_next(request)
    response.headers["x-request-id"] = rid

    duration = (time.time() - start) * 1000
    status = response.status_code
    level = logging.WARNING if status >= 400 else logging.INFO
    logger.log(
        level,
        "[%s] %s %s -> %s | %.2fms",
        rid, request.method, request.url.path, status, duration,
    )
    return response

# Global exception handlers
register_exception_handlers(app)

app.include_router(api_v1_router)


@app.get("/health")
def health_check():
    return {"status": "ok"}
