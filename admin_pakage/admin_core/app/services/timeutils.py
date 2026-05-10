"""Timezone and epoch utilities."""
import time
from datetime import datetime
from zoneinfo import ZoneInfo
from app.config import settings


def now_epoch() -> int:
    """Return current UTC time as Unix epoch (seconds)."""
    return int(time.time())


def format_epoch(epoch: int) -> str:
    """Convert epoch seconds to ISO string in the configured display timezone."""
    if not epoch:
        return ""
    dt_utc = datetime.fromtimestamp(epoch, tz=ZoneInfo("UTC"))
    dt_local = dt_utc.astimezone(ZoneInfo(settings.display_timezone))
    return dt_local.isoformat()
