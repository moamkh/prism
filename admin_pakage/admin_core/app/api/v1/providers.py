from typing import List
from uuid import UUID
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from app.api.deps import get_db
from app.crud.providers import provider as crud_provider
from app.schemas.schemas import ProviderCreate, ProviderUpdate, ProviderOut

router = APIRouter()


@router.get("/", response_model=List[ProviderOut])
def list_providers(skip: int = 0, limit: int = 100, db: Session = Depends(get_db)):
    return crud_provider.get_multi(db, skip=skip, limit=limit)


@router.post("/", response_model=ProviderOut)
def create_provider(obj_in: ProviderCreate, db: Session = Depends(get_db)):
    existing = crud_provider.get_by_name(db, name=obj_in.name)
    if existing:
        raise HTTPException(status_code=400, detail="Provider with this name already exists")
    return crud_provider.create(db, obj_in=obj_in)


@router.get("/{provider_id}", response_model=ProviderOut)
def get_provider(provider_id: UUID, db: Session = Depends(get_db)):
    obj = crud_provider.get(db, provider_id=provider_id)
    if not obj:
        raise HTTPException(status_code=404, detail="Provider not found")
    return obj


@router.put("/{provider_id}", response_model=ProviderOut)
def update_provider(provider_id: UUID, obj_in: ProviderUpdate, db: Session = Depends(get_db)):
    obj = crud_provider.get(db, provider_id=provider_id)
    if not obj:
        raise HTTPException(status_code=404, detail="Provider not found")
    return crud_provider.update(db, db_obj=obj, obj_in=obj_in)


@router.delete("/{provider_id}", response_model=ProviderOut)
def delete_provider(provider_id: UUID, db: Session = Depends(get_db)):
    obj = crud_provider.remove(db, provider_id=provider_id)
    if not obj:
        raise HTTPException(status_code=404, detail="Provider not found")
    return obj
