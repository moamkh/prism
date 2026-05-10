from typing import List, Optional
from sqlalchemy.orm import Session
from app.models.models import Config as ConfigModel
from app.services.timeutils import now_epoch


class CRUDConfig:
    def get(self, db: Session, key: str) -> Optional[ConfigModel]:
        return db.query(ConfigModel).filter(ConfigModel.key == key).first()

    def get_multi(self, db: Session) -> List[ConfigModel]:
        return db.query(ConfigModel).all()

    def get_value(self, db: Session, key: str, default: str = "") -> str:
        obj = self.get(db, key)
        return obj.value if obj else default

    def set_value(self, db: Session, key: str, value: str) -> ConfigModel:
        obj = self.get(db, key)
        now = now_epoch()
        if obj:
            obj.value = value
            obj.updated_at = now
        else:
            obj = ConfigModel(key=key, value=value, updated_at=now)
            db.add(obj)
        db.commit()
        db.refresh(obj)
        return obj


config = CRUDConfig()
