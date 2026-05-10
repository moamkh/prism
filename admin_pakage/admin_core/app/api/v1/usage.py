from typing import List, Optional
from uuid import UUID
from datetime import datetime
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from app.api.deps import get_db
from app.crud.usage import usage as crud_usage
from app.schemas.schemas import UsageLogOut, UsageLogFilteredOut, UsageAggregate

router = APIRouter()


@router.get("/logs", response_model=List[UsageLogOut])
def list_logs(
    skip: int = 0,
    limit: int = 100,
    token_id: Optional[UUID] = None,
    model_id: Optional[UUID] = None,
    provider_id: Optional[UUID] = None,
    start_date: Optional[datetime] = None,
    end_date: Optional[datetime] = None,
    db: Session = Depends(get_db),
):
    return crud_usage.get_multi(
        db,
        skip=skip,
        limit=limit,
        token_id=token_id,
        model_id=model_id,
        provider_id=provider_id,
        start_date=start_date,
        end_date=end_date,
    )


@router.get("/logs/filtered", response_model=UsageLogFilteredOut)
def list_logs_filtered(
    skip: int = 0,
    limit: int = 100,
    token_id: Optional[UUID] = None,
    model_id: Optional[UUID] = None,
    provider_id: Optional[UUID] = None,
    token_name: Optional[str] = None,
    provider_name: Optional[str] = None,
    start_date: Optional[datetime] = None,
    end_date: Optional[datetime] = None,
    db: Session = Depends(get_db),
):
    logs = crud_usage.get_multi(
        db,
        skip=skip,
        limit=limit,
        token_id=token_id,
        model_id=model_id,
        provider_id=provider_id,
        token_name=token_name,
        provider_name=provider_name,
        start_date=start_date,
        end_date=end_date,
    )
    totals = crud_usage.get_totals(
        db,
        token_id=token_id,
        model_id=model_id,
        provider_id=provider_id,
        token_name=token_name,
        provider_name=provider_name,
        start_date=start_date,
        end_date=end_date,
    )
    return {"logs": logs, "totals": totals}


@router.get("/aggregate", response_model=List[dict])
def aggregate_usage(days: int = 7, db: Session = Depends(get_db)):
    return crud_usage.get_aggregate_by_model(db, days=days)
