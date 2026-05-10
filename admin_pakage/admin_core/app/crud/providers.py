from typing import List, Optional
from uuid import UUID
from sqlalchemy.orm import Session
from app.models.models import Provider
from app.schemas.schemas import ProviderCreate, ProviderUpdate
from app.services.crypto import encrypt, decrypt
from app.services.timeutils import now_epoch


class CRUDProvider:
    def get(self, db: Session, provider_id: UUID) -> Optional[Provider]:
        obj = db.query(Provider).filter(Provider.id == provider_id).first()
        if obj:
            try:
                obj.api_token = decrypt(obj.api_token)
            except Exception:
                pass
        return obj

    def get_by_name(self, db: Session, name: str) -> Optional[Provider]:
        return db.query(Provider).filter(Provider.name == name).first()

    def get_multi(self, db: Session, skip: int = 0, limit: int = 100) -> List[Provider]:
        objs = db.query(Provider).offset(skip).limit(limit).all()
        for obj in objs:
            try:
                obj.api_token = decrypt(obj.api_token)
            except Exception:
                pass
        return objs

    def create(self, db: Session, obj_in: ProviderCreate) -> Provider:
        now = now_epoch()
        db_obj = Provider(
            name=obj_in.name,
            base_url=obj_in.base_url,
            api_token=encrypt(obj_in.api_token),
            http_proxy=obj_in.http_proxy,
            socks5_proxy=obj_in.socks5_proxy,
            enable_proxy=obj_in.enable_proxy,
            max_concurrent_requests=obj_in.max_concurrent_requests,
            is_active=obj_in.is_active,
            created_at=now,
            updated_at=now,
        )
        db.add(db_obj)
        db.commit()
        db.refresh(db_obj)
        return db_obj

    def update(self, db: Session, db_obj: Provider, obj_in: ProviderUpdate) -> Provider:
        update_data = obj_in.model_dump(exclude_unset=True)
        if "api_token" in update_data:
            update_data["api_token"] = encrypt(update_data["api_token"])
        update_data["updated_at"] = now_epoch()
        for field, value in update_data.items():
            setattr(db_obj, field, value)
        db.add(db_obj)
        db.commit()
        db.refresh(db_obj)
        return db_obj

    def remove(self, db: Session, provider_id: UUID) -> Optional[Provider]:
        obj = db.query(Provider).filter(Provider.id == provider_id).first()
        if obj:
            db.delete(obj)
            db.commit()
        return obj


provider = CRUDProvider()
