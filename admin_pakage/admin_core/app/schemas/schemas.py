from datetime import datetime
from typing import Optional, List
from uuid import UUID
from pydantic import BaseModel, ConfigDict, field_serializer
from app.services.timeutils import format_epoch


class ProviderBase(BaseModel):
    name: str
    base_url: str
    api_token: str
    http_proxy: str | None = None
    socks5_proxy: str | None = None
    enable_proxy: bool = True
    max_concurrent_requests: int = 100
    is_active: bool = True


class ProviderCreate(ProviderBase):
    pass


class ProviderUpdate(BaseModel):
    name: Optional[str] = None
    base_url: Optional[str] = None
    api_token: Optional[str] = None
    http_proxy: Optional[str] = None
    socks5_proxy: Optional[str] = None
    enable_proxy: Optional[bool] = None
    max_concurrent_requests: Optional[int] = None
    is_active: Optional[bool] = None


class ProviderOut(ProviderBase):
    model_config = ConfigDict(from_attributes=True)
    id: UUID
    created_at: int
    updated_at: int

    @field_serializer("created_at", "updated_at")
    def serialize_timestamp(self, value: int) -> str:
        return format_epoch(value)


class ModelBase(BaseModel):
    model_id: str
    display_model_id: Optional[str] = None
    is_active: bool = True


class ModelCreate(ModelBase):
    provider_id: UUID


class ModelUpdate(BaseModel):
    model_id: Optional[str] = None
    display_model_id: Optional[str] = None
    is_active: Optional[bool] = None


class ModelOut(ModelBase):
    model_config = ConfigDict(from_attributes=True)
    id: UUID
    provider_id: UUID
    created_at: int

    @field_serializer("created_at")
    def serialize_timestamp(self, value: int) -> str:
        return format_epoch(value)


class TokenBase(BaseModel):
    name: str
    max_input_tokens: Optional[int] = None
    max_output_tokens: Optional[int] = None
    requests_per_minute: Optional[int] = None
    is_active: bool = True


class ModelPermissionCreate(BaseModel):
    model_id: UUID
    max_input_tokens: Optional[int] = None
    max_output_tokens: Optional[int] = None


class TokenCreate(TokenBase):
    model_permissions: List[ModelPermissionCreate] = []


class TokenUpdate(BaseModel):
    name: Optional[str] = None
    max_input_tokens: Optional[int] = None
    max_output_tokens: Optional[int] = None
    requests_per_minute: Optional[int] = None
    is_active: Optional[bool] = None
    model_permissions: Optional[List[ModelPermissionCreate]] = None


class ModelPermissionOut(BaseModel):
    model_config = ConfigDict(from_attributes=True)
    model_id: UUID
    max_input_tokens: Optional[int] = None
    max_output_tokens: Optional[int] = None


class TokenOut(TokenBase):
    model_config = ConfigDict(from_attributes=True)
    id: UUID
    key_hash: str
    created_at: int
    model_permissions: List[ModelPermissionOut] = []

    @field_serializer("created_at")
    def serialize_timestamp(self, value: int) -> str:
        return format_epoch(value)


class TokenWithPlainKey(TokenBase):
    model_config = ConfigDict(from_attributes=True)
    id: UUID
    key_hash: str
    created_at: int
    model_permissions: List[ModelPermissionOut] = []
    plain_key: str

    @field_serializer("created_at")
    def serialize_timestamp(self, value: int) -> str:
        return format_epoch(value)


class TokenModelPermissionCreate(BaseModel):
    token_id: UUID
    model_id: UUID


class TokenModelPermissionOut(BaseModel):
    model_config = ConfigDict(from_attributes=True)
    id: UUID
    token_id: UUID
    model_id: UUID
    created_at: int

    @field_serializer("created_at")
    def serialize_timestamp(self, value: int) -> str:
        return format_epoch(value)


class UsageLogOut(BaseModel):
    model_config = ConfigDict(from_attributes=True)
    id: UUID
    token_id: Optional[UUID]
    provider_id: Optional[UUID]
    model_id: Optional[UUID]
    model_name: Optional[str]
    provider_name: Optional[str]
    request_path: Optional[str]
    input_tokens: int
    output_tokens: int
    total_tokens: int
    latency_ms: Optional[int]
    status_code: Optional[int]
    created_at: int

    @field_serializer("created_at")
    def serialize_timestamp(self, value: int) -> str:
        return format_epoch(value)


class UsageLogTotals(BaseModel):
    count: int
    total_input_tokens: int
    total_output_tokens: int
    total_tokens: int


class UsageLogFilteredOut(BaseModel):
    logs: List[UsageLogOut]
    totals: UsageLogTotals


class UsageAggregate(BaseModel):
    group_key: str
    total_requests: int
    total_input_tokens: int
    total_output_tokens: int
    total_tokens: int


class ConfigItem(BaseModel):
    key: str
    value: str
    updated_at: int

    @field_serializer("updated_at")
    def serialize_timestamp(self, value: int) -> str:
        return format_epoch(value)


class ConfigUpdate(BaseModel):
    value: str


class TokenModelUsageOut(BaseModel):
    model_id: str
    model_name: str
    max_tokens: int
    current_usage: int
    percentage: float


class TopModelOut(BaseModel):
    model_id: Optional[str]
    model_name: Optional[str]
    provider_name: Optional[str]
    total_requests: int
    total_input_tokens: int
    total_output_tokens: int
    total_tokens: int


class DashboardStats(BaseModel):
    total_requests: int
    total_input_tokens: int
    total_output_tokens: int
    active_providers: int
    active_tokens: int
    top_models: List[TopModelOut]
    recent_logs: List[UsageLogOut]
