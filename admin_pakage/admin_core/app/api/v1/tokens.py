from typing import List
from uuid import UUID
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from app.api.deps import get_db
from app.crud.tokens import token as crud_token
from app.crud.usage import usage as crud_usage
from app.schemas.schemas import TokenCreate, TokenUpdate, TokenOut, TokenWithPlainKey, ModelPermissionOut, TokenModelUsageOut

router = APIRouter()


def serialize_token(obj) -> TokenOut:
    model_permissions = [
        ModelPermissionOut(
            model_id=perm.model_id,
            max_input_tokens=perm.max_input_tokens,
            max_output_tokens=perm.max_output_tokens,
        )
        for perm in obj.token_permissions
    ]
    return TokenOut(
        id=obj.id,
        name=obj.name,
        key_hash=obj.key_hash,
        max_input_tokens=obj.max_input_tokens,
        max_output_tokens=obj.max_output_tokens,
        requests_per_minute=obj.requests_per_minute,
        is_active=obj.is_active,
        created_at=obj.created_at,
        model_permissions=model_permissions,
    )


@router.get("/", response_model=List[TokenOut])
def list_tokens(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    objs = crud_token.get_multi(db, skip=skip, limit=limit)
    return [serialize_token(obj) for obj in objs]


@router.post("/", response_model=TokenWithPlainKey)
def create_token(obj_in: TokenCreate, db: Session = Depends(get_db)):
    db_obj, plain_key = crud_token.create(db, obj_in=obj_in)
    model_permissions = [
        ModelPermissionOut(
            model_id=perm.model_id,
            max_input_tokens=perm.max_input_tokens,
            max_output_tokens=perm.max_output_tokens,
        )
        for perm in db_obj.token_permissions
    ]
    return TokenWithPlainKey(
        id=db_obj.id,
        name=db_obj.name,
        key_hash=db_obj.key_hash,
        max_input_tokens=db_obj.max_input_tokens,
        max_output_tokens=db_obj.max_output_tokens,
        requests_per_minute=db_obj.requests_per_minute,
        is_active=db_obj.is_active,
        created_at=db_obj.created_at,
        model_permissions=model_permissions,
        plain_key=plain_key,
    )


@router.get("/{token_id}", response_model=TokenOut)
def get_token(token_id: UUID, db: Session = Depends(get_db)):
    obj = crud_token.get(db, token_id=token_id)
    if not obj:
        raise HTTPException(status_code=404, detail="Token not found")
    return serialize_token(obj)


@router.put("/{token_id}", response_model=TokenOut)
def update_token(token_id: UUID, obj_in: TokenUpdate, db: Session = Depends(get_db)):
    obj = crud_token.get(db, token_id=token_id)
    if not obj:
        raise HTTPException(status_code=404, detail="Token not found")
    updated = crud_token.update(db, db_obj=obj, obj_in=obj_in)
    return serialize_token(updated)


@router.get("/{token_id}/usage", response_model=List[TokenModelUsageOut])
def token_usage(token_id: UUID, db: Session = Depends(get_db)):
    obj = crud_token.get(db, token_id=token_id)
    if not obj:
        raise HTTPException(status_code=404, detail="Token not found")
    return crud_usage.get_token_model_usage(db, token_id=token_id)


@router.post("/{token_id}/regenerate", response_model=TokenWithPlainKey)
def regenerate_token_key(token_id: UUID, db: Session = Depends(get_db)):
    obj = crud_token.get(db, token_id=token_id)
    if not obj:
        raise HTTPException(status_code=404, detail="Token not found")
    db_obj, plain_key = crud_token.regenerate_key(db, db_obj=obj)
    model_permissions = [
        ModelPermissionOut(
            model_id=perm.model_id,
            max_input_tokens=perm.max_input_tokens,
            max_output_tokens=perm.max_output_tokens,
        )
        for perm in db_obj.token_permissions
    ]
    return TokenWithPlainKey(
        id=db_obj.id,
        name=db_obj.name,
        key_hash=db_obj.key_hash,
        max_input_tokens=db_obj.max_input_tokens,
        max_output_tokens=db_obj.max_output_tokens,
        requests_per_minute=db_obj.requests_per_minute,
        is_active=db_obj.is_active,
        created_at=db_obj.created_at,
        model_permissions=model_permissions,
        plain_key=plain_key,
    )


@router.delete("/{token_id}", response_model=TokenOut)
def delete_token(token_id: UUID, db: Session = Depends(get_db)):
    obj = crud_token.remove(db, token_id=token_id)
    if not obj:
        raise HTTPException(status_code=404, detail="Token not found")
    return serialize_token(obj)
