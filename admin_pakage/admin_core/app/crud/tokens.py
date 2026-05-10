import secrets
import hashlib
from typing import List, Optional
from uuid import UUID
from sqlalchemy.orm import Session
from app.models.models import Token, TokenModelPermission
from app.schemas.schemas import TokenCreate, TokenUpdate
from app.services.timeutils import now_epoch


def generate_token() -> str:
    return "rpm_" + secrets.token_urlsafe(32)


def hash_token(plain: str) -> str:
    return hashlib.sha256(plain.encode()).hexdigest()


class CRUDToken:
    def get(self, db: Session, token_id: UUID) -> Optional[Token]:
        return db.query(Token).filter(Token.id == token_id).first()

    def get_by_hash(self, db: Session, key_hash: str) -> Optional[Token]:
        return db.query(Token).filter(Token.key_hash == key_hash).first()

    def get_multi(self, db: Session, skip: int = 0, limit: int = 100) -> List[Token]:
        return db.query(Token).offset(skip).limit(limit).all()

    def create(self, db: Session, obj_in: TokenCreate) -> tuple[Token, str]:
        plain_key = generate_token()
        key_hash = hash_token(plain_key)
        db_obj = Token(
            name=obj_in.name,
            key_hash=key_hash,
            max_input_tokens=obj_in.max_input_tokens,
            max_output_tokens=obj_in.max_output_tokens,
            requests_per_minute=obj_in.requests_per_minute,
            is_active=obj_in.is_active,
            created_at=now_epoch(),
        )
        db.add(db_obj)
        db.commit()
        db.refresh(db_obj)

        for perm in obj_in.model_permissions:
            p = TokenModelPermission(
                token_id=db_obj.id,
                model_id=perm.model_id,
                max_input_tokens=perm.max_input_tokens,
                max_output_tokens=perm.max_output_tokens,
                created_at=now_epoch(),
            )
            db.add(p)
        db.commit()
        db.refresh(db_obj)
        return db_obj, plain_key

    def update(self, db: Session, db_obj: Token, obj_in: TokenUpdate) -> Token:
        update_data = obj_in.model_dump(exclude_unset=True)
        model_permissions = update_data.pop("model_permissions", None)

        for field, value in update_data.items():
            setattr(db_obj, field, value)
        db.add(db_obj)

        if model_permissions is not None:
            db.query(TokenModelPermission).filter(
                TokenModelPermission.token_id == db_obj.id
            ).delete(synchronize_session=False)
            for perm in model_permissions:
                p = TokenModelPermission(
                    token_id=db_obj.id,
                    model_id=perm["model_id"],
                    max_input_tokens=perm.get("max_input_tokens"),
                    max_output_tokens=perm.get("max_output_tokens"),
                    created_at=now_epoch(),
                )
                db.add(p)

        db.commit()
        db.refresh(db_obj)
        return db_obj

    def remove(self, db: Session, token_id: UUID) -> Optional[Token]:
        obj = db.query(Token).filter(Token.id == token_id).first()
        if obj:
            db.delete(obj)
            db.commit()
        return obj


token = CRUDToken()
