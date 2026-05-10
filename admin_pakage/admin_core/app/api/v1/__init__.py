from fastapi import APIRouter
from app.api.v1 import providers, models, tokens, usage, config, dashboard, proxy_status

router = APIRouter(prefix="/api/v1")

router.include_router(providers.router, prefix="/providers", tags=["providers"])
router.include_router(models.router, prefix="/models", tags=["models"])
router.include_router(tokens.router, prefix="/tokens", tags=["tokens"])
router.include_router(usage.router, prefix="/usage", tags=["usage"])
router.include_router(config.router, prefix="/config", tags=["config"])
router.include_router(dashboard.router, prefix="/dashboard", tags=["dashboard"])
router.include_router(proxy_status.router, prefix="/proxy", tags=["proxy"])
