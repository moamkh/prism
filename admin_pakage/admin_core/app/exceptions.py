import logging
import traceback
from fastapi import Request
from fastapi.responses import JSONResponse
from fastapi.exceptions import RequestValidationError
from pydantic import BaseModel
from sqlalchemy.exc import IntegrityError, OperationalError
from starlette.exceptions import HTTPException as StarletteHTTPException

logger = logging.getLogger(__name__)


class ErrorResponse(BaseModel):
    error: str
    detail: str | None = None
    request_id: str | None = None


def _request_id(request: Request) -> str:
    return request.headers.get("x-request-id", "-")


def register_exception_handlers(app):
    """Register global exception handlers on the FastAPI app."""

    @app.exception_handler(StarletteHTTPException)
    async def http_exception_handler(request: Request, exc: StarletteHTTPException):
        rid = _request_id(request)
        if exc.status_code >= 500:
            logger.error(
                "[HTTPException] %s %s -> %s | detail=%s | request_id=%s",
                request.method, request.url.path, exc.status_code, exc.detail, rid,
                exc_info=exc.status_code >= 500,
            )
        return JSONResponse(
            status_code=exc.status_code,
            content={"error": exc.detail or "Request failed", "request_id": rid},
        )

    @app.exception_handler(RequestValidationError)
    async def validation_exception_handler(request: Request, exc: RequestValidationError):
        rid = _request_id(request)
        # Build a concise message from Pydantic errors
        errors = exc.errors()
        details = []
        for err in errors:
            loc = ".".join(str(x) for x in err.get("loc", []))
            msg = err.get("msg", "invalid value")
            details.append(f"{loc}: {msg}")
        detail_str = "; ".join(details[:3])  # cap to first 3 errors
        logger.warning(
            "[ValidationError] %s %s -> 422 | detail=%s | request_id=%s",
            request.method, request.url.path, detail_str, rid,
        )
        return JSONResponse(
            status_code=422,
            content={"error": "Validation failed", "detail": detail_str, "request_id": rid},
        )

    @app.exception_handler(IntegrityError)
    async def integrity_error_handler(request: Request, exc: IntegrityError):
        rid = _request_id(request)
        orig = getattr(exc.orig, "pgerror", str(exc.orig)) if exc.orig else str(exc)
        logger.error(
            "[IntegrityError] %s %s | detail=%s | request_id=%s\n%s",
            request.method, request.url.path, orig, rid, traceback.format_exc(),
        )
        # Try to surface a friendlier message
        detail = "Database constraint violation"
        if orig and "unique constraint" in orig.lower():
            detail = "A record with this value already exists"
        elif orig and "foreign key" in orig.lower():
            detail = "Related record does not exist"
        return JSONResponse(
            status_code=409,
            content={"error": detail, "detail": orig, "request_id": rid},
        )

    @app.exception_handler(OperationalError)
    async def operational_error_handler(request: Request, exc: OperationalError):
        rid = _request_id(request)
        logger.error(
            "[OperationalError] %s %s | request_id=%s\n%s",
            request.method, request.url.path, rid, traceback.format_exc(),
        )
        return JSONResponse(
            status_code=503,
            content={"error": "Database is temporarily unavailable", "request_id": rid},
        )

    @app.exception_handler(Exception)
    async def generic_exception_handler(request: Request, exc: Exception):
        rid = _request_id(request)
        logger.critical(
            "[UnhandledException] %s %s | request_id=%s\n%s",
            request.method, request.url.path, rid, traceback.format_exc(),
        )
        return JSONResponse(
            status_code=500,
            content={"error": "Internal server error", "request_id": rid},
        )
