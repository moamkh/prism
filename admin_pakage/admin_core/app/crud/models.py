from typing import List, Optional
from uuid import UUID
from sqlalchemy.orm import Session
from app.models.models import Model
from app.schemas.schemas import ModelCreate, ModelUpdate
from app.services.timeutils import now_epoch


class CRUDModel:
    def get(self, db: Session, model_id: UUID) -> Optional[Model]:
        return db.query(Model).filter(Model.id == model_id).first()

    def get_by_provider_and_model_id(
        self, db: Session, provider_id: UUID, model_id: str
    ) -> Optional[Model]:
        return (
            db.query(Model)
            .filter(Model.provider_id == provider_id, Model.model_id == model_id)
            .first()
        )

    def get_multi(
        self, db: Session, skip: int = 0, limit: int = 100, provider_id: Optional[UUID] = None
    ) -> List[Model]:
        q = db.query(Model)
        if provider_id:
            q = q.filter(Model.provider_id == provider_id)
        return q.offset(skip).limit(limit).all()

    def create(self, db: Session, obj_in: ModelCreate) -> Model:
        db_obj = Model(
            provider_id=obj_in.provider_id,
            model_id=obj_in.model_id,
            display_model_id=obj_in.display_model_id,
            max_concurrent_requests=obj_in.max_concurrent_requests,
            queue_size=obj_in.queue_size,
            is_active=obj_in.is_active,
            created_at=now_epoch(),
        )
        db.add(db_obj)
        db.commit()
        db.refresh(db_obj)
        return db_obj

    def update(self, db: Session, db_obj: Model, obj_in: ModelUpdate) -> Model:
        update_data = obj_in.model_dump(exclude_unset=True)
        for field, value in update_data.items():
            setattr(db_obj, field, value)
        db.add(db_obj)
        db.commit()
        db.refresh(db_obj)
        return db_obj

    def remove(self, db: Session, model_id: UUID) -> Optional[Model]:
        obj = db.query(Model).filter(Model.id == model_id).first()
        if obj:
            db.delete(obj)
            db.commit()
        return obj


model = CRUDModel()
