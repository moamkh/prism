from fastapi import APIRouter, Depends
from sqlalchemy.orm import Session
from sqlalchemy import func
from app.api.deps import get_db
from app.crud.usage import usage as crud_usage
from app.models.models import Provider, Token, UsageLog
from app.schemas.schemas import DashboardStats, UsageLogOut

router = APIRouter()


@router.get("/stats", response_model=DashboardStats)
def dashboard_stats(db: Session = Depends(get_db)):
    stats = crud_usage.get_stats(db)
    active_providers = db.query(Provider).filter(Provider.is_active == True).count()
    active_tokens = db.query(Token).filter(Token.is_active == True).count()
    top_models = crud_usage.get_aggregate_by_model(db, days=7)
    recent_logs = (
        db.query(UsageLog)
        .order_by(UsageLog.created_at.desc())
        .limit(10)
        .all()
    )
    return DashboardStats(
        total_requests=stats["total_requests"],
        total_input_tokens=stats["total_input_tokens"],
        total_output_tokens=stats["total_output_tokens"],
        active_providers=active_providers,
        active_tokens=active_tokens,
        top_models=top_models,
        recent_logs=[UsageLogOut.model_validate(r) for r in recent_logs],
    )
