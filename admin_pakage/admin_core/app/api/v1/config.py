from typing import List
from fastapi import APIRouter, Depends, HTTPException
from sqlalchemy.orm import Session
from app.api.deps import get_db
from app.crud.config import config as crud_config
from app.schemas.schemas import ConfigItem, ConfigUpdate

router = APIRouter()


@router.get("/", response_model=List[ConfigItem])
def list_config(db: Session = Depends(get_db)):
    return crud_config.get_multi(db)


@router.get("/{key}", response_model=ConfigItem)
def get_config_item(key: str, db: Session = Depends(get_db)):
    obj = crud_config.get(db, key=key)
    if not obj:
        raise HTTPException(status_code=404, detail="Config key not found")
    return obj


@router.put("/{key}", response_model=ConfigItem)
def update_config_item(key: str, obj_in: ConfigUpdate, db: Session = Depends(get_db)):
    return crud_config.set_value(db, key=key, value=obj_in.value)
