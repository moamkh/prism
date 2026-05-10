from typing import List, Optional
from uuid import UUID
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from app.api.deps import get_db
from app.crud.models import model as crud_model
from app.crud.providers import provider as crud_provider
from app.schemas.schemas import ModelCreate, ModelUpdate, ModelOut
from app.services.provider_client import fetch_models_from_provider
from app.services.crypto import decrypt

router = APIRouter()


@router.get("/", response_model=List[ModelOut])
def list_models(
    skip: int = 0,
    limit: int = 100,
    provider_id: Optional[UUID] = None,
    db: Session = Depends(get_db),
):
    return crud_model.get_multi(db, skip=skip, limit=limit, provider_id=provider_id)


@router.post("/", response_model=ModelOut)
def create_model(obj_in: ModelCreate, db: Session = Depends(get_db)):
    existing = crud_model.get_by_provider_and_model_id(
        db, provider_id=obj_in.provider_id, model_id=obj_in.model_id
    )
    if existing:
        raise HTTPException(
            status_code=400,
            detail=f"Model '{obj_in.model_id}' already exists for this provider",
        )
    return crud_model.create(db, obj_in=obj_in)


@router.get("/{model_id}", response_model=ModelOut)
def get_model(model_id: UUID, db: Session = Depends(get_db)):
    obj = crud_model.get(db, model_id=model_id)
    if not obj:
        raise HTTPException(status_code=404, detail="Model not found")
    return obj


@router.put("/{model_id}", response_model=ModelOut)
def update_model(model_id: UUID, obj_in: ModelUpdate, db: Session = Depends(get_db)):
    obj = crud_model.get(db, model_id=model_id)
    if not obj:
        raise HTTPException(status_code=404, detail="Model not found")
    if obj_in.model_id and obj_in.model_id != obj.model_id:
        existing = crud_model.get_by_provider_and_model_id(
            db, provider_id=obj.provider_id, model_id=obj_in.model_id
        )
        if existing:
            raise HTTPException(
                status_code=400,
                detail=f"Model '{obj_in.model_id}' already exists for this provider",
            )
    return crud_model.update(db, db_obj=obj, obj_in=obj_in)


@router.delete("/{model_id}", response_model=ModelOut)
def delete_model(model_id: UUID, db: Session = Depends(get_db)):
    obj = crud_model.remove(db, model_id=model_id)
    if not obj:
        raise HTTPException(status_code=404, detail="Model not found")
    return obj


@router.get("/available/{provider_id}")
def fetch_available_models(provider_id: UUID, db: Session = Depends(get_db)):
    provider = crud_provider.get(db, provider_id=provider_id)
    if not provider:
        raise HTTPException(status_code=404, detail="Provider not found")
    try:
        api_token = provider.api_token
        try:
            api_token = decrypt(api_token)
        except Exception:
            pass
        models = fetch_models_from_provider(
            provider.base_url,
            api_token,
            enable_proxy=provider.enable_proxy,
            http_proxy=provider.http_proxy,
            socks5_proxy=provider.socks5_proxy,
        )
        return {"models": models}
    except Exception as e:
        raise HTTPException(status_code=502, detail=f"Failed to fetch models from provider: {str(e)}")
