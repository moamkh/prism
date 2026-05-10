from typing import List, Optional
from uuid import UUID
from datetime import datetime, timedelta, timezone
from sqlalchemy.orm import Session, joinedload
from sqlalchemy import func
from app.models.models import UsageLog, Token, Provider


class CRUDUsage:
    def _build_filter_query(
        self,
        db: Session,
        token_id: Optional[UUID] = None,
        model_id: Optional[UUID] = None,
        provider_id: Optional[UUID] = None,
        token_name: Optional[str] = None,
        provider_name: Optional[str] = None,
        start_date: Optional[datetime] = None,
        end_date: Optional[datetime] = None,
    ):
        q = db.query(UsageLog)
        if token_id:
            q = q.filter(UsageLog.token_id == token_id)
        if model_id:
            q = q.filter(UsageLog.model_id == model_id)
        if provider_id:
            q = q.filter(UsageLog.provider_id == provider_id)
        if token_name:
            q = q.join(Token, UsageLog.token_id == Token.id, isouter=True)
            q = q.filter(Token.name.ilike(f"%{token_name}%"))
        if provider_name:
            # provider_name is denormalized in UsageLog; prefer direct filter.
            q = q.filter(UsageLog.provider_name.ilike(f"%{provider_name}%"))
        if start_date:
            epoch = int(start_date.timestamp())
            q = q.filter(UsageLog.created_at >= epoch)
        if end_date:
            epoch = int(end_date.timestamp())
            q = q.filter(UsageLog.created_at <= epoch)
        return q

    def get_multi(
        self,
        db: Session,
        skip: int = 0,
        limit: int = 100,
        token_id: Optional[UUID] = None,
        model_id: Optional[UUID] = None,
        provider_id: Optional[UUID] = None,
        token_name: Optional[str] = None,
        provider_name: Optional[str] = None,
        start_date: Optional[datetime] = None,
        end_date: Optional[datetime] = None,
    ) -> List[UsageLog]:
        q = self._build_filter_query(
            db, token_id, model_id, provider_id, token_name, provider_name, start_date, end_date
        )
        return q.order_by(UsageLog.created_at.desc()).offset(skip).limit(limit).all()

    def get_totals(
        self,
        db: Session,
        token_id: Optional[UUID] = None,
        model_id: Optional[UUID] = None,
        provider_id: Optional[UUID] = None,
        token_name: Optional[str] = None,
        provider_name: Optional[str] = None,
        start_date: Optional[datetime] = None,
        end_date: Optional[datetime] = None,
    ) -> dict:
        q = self._build_filter_query(
            db, token_id, model_id, provider_id, token_name, provider_name, start_date, end_date
        )
        result = q.with_entities(
            func.count(UsageLog.id).label("count"),
            func.coalesce(func.sum(UsageLog.input_tokens), 0).label("total_input"),
            func.coalesce(func.sum(UsageLog.output_tokens), 0).label("total_output"),
            func.coalesce(func.sum(UsageLog.total_tokens), 0).label("total"),
        ).first()
        return {
            "count": result.count if result else 0,
            "total_input_tokens": int(result.total_input) if result else 0,
            "total_output_tokens": int(result.total_output) if result else 0,
            "total_tokens": int(result.total) if result else 0,
        }

    def get_aggregate_by_model(self, db: Session, days: int = 7) -> List[dict]:
        since = int((datetime.now(timezone.utc) - timedelta(days=days)).timestamp())
        rows = (
            db.query(
                UsageLog.model_id,
                func.max(UsageLog.model_name).label("model_name"),
                func.max(UsageLog.provider_name).label("provider_name"),
                func.count(UsageLog.id).label("total_requests"),
                func.sum(UsageLog.input_tokens).label("total_input"),
                func.sum(UsageLog.output_tokens).label("total_output"),
                func.sum(UsageLog.total_tokens).label("total"),
            )
            .filter(UsageLog.created_at >= since)
            .group_by(UsageLog.model_id)
            .all()
        )
        return [
            {
                "model_id": str(r.model_id) if r.model_id else None,
                "model_name": r.model_name or "Deleted Model",
                "provider_name": r.provider_name or "Unknown Provider",
                "total_requests": r.total_requests,
                "total_input_tokens": r.total_input or 0,
                "total_output_tokens": r.total_output or 0,
                "total_tokens": r.total or 0,
            }
            for r in rows
        ]

    def get_stats(self, db: Session) -> dict:
        total_requests = db.query(UsageLog).count()
        total_input = db.query(func.sum(UsageLog.input_tokens)).scalar() or 0
        total_output = db.query(func.sum(UsageLog.output_tokens)).scalar() or 0
        return {
            "total_requests": total_requests,
            "total_input_tokens": total_input,
            "total_output_tokens": total_output,
        }

    def get_token_model_usage(self, db: Session, token_id: UUID) -> List[dict]:
        from app.models.models import TokenModelPermission, Model
        # Get all permissions for this token with model names
        perms = (
            db.query(TokenModelPermission, Model)
            .join(Model, TokenModelPermission.model_id == Model.id)
            .filter(TokenModelPermission.token_id == token_id)
            .all()
        )
        result = []
        for perm, model in perms:
            max_tokens = 0
            if perm.max_input_tokens:
                max_tokens += perm.max_input_tokens
            if perm.max_output_tokens:
                max_tokens += perm.max_output_tokens
            # Sum total tokens from usage logs for this token+model
            total_used = (
                db.query(func.sum(UsageLog.total_tokens))
                .filter(UsageLog.token_id == token_id, UsageLog.model_id == perm.model_id)
                .scalar()
            ) or 0
            percentage = round((total_used / max_tokens) * 100, 1) if max_tokens > 0 else 0.0
            result.append({
                "model_id": str(model.id),
                "model_name": model.model_id,
                "max_tokens": max_tokens,
                "current_usage": int(total_used),
                "percentage": percentage,
            })
        return result


usage = CRUDUsage()
