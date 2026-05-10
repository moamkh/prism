import uuid
from sqlalchemy import (
    Column,
    String,
    Integer,
    Boolean,
    Text,
    BigInteger,
    ForeignKey,
    UniqueConstraint,
)
from sqlalchemy.dialects.postgresql import UUID
from sqlalchemy.orm import relationship
from app.database import Base


class Provider(Base):
    __tablename__ = "providers"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    name = Column(String(255), nullable=False)
    base_url = Column(Text, nullable=False)
    api_token = Column(Text, nullable=False)
    http_proxy = Column(Text, nullable=True)
    socks5_proxy = Column(Text, nullable=True)
    enable_proxy = Column(Boolean, default=True, nullable=False)
    max_concurrent_requests = Column(Integer, default=100, nullable=False)
    is_active = Column(Boolean, default=True, nullable=False)
    created_at = Column(Integer, nullable=False)
    updated_at = Column(Integer, nullable=False)

    models = relationship("Model", back_populates="provider", cascade="all, delete-orphan")
    usage_logs = relationship("UsageLog", back_populates="provider")


class Model(Base):
    __tablename__ = "models"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    provider_id = Column(UUID(as_uuid=True), ForeignKey("providers.id", ondelete="CASCADE"), nullable=False)
    model_id = Column(String(255), nullable=False)
    display_model_id = Column(String(255), nullable=True)
    is_active = Column(Boolean, default=True, nullable=False)
    created_at = Column(Integer, nullable=False)

    provider = relationship("Provider", back_populates="models")
    usage_logs = relationship("UsageLog", back_populates="model")
    token_permissions = relationship("TokenModelPermission", back_populates="model", cascade="all, delete-orphan")

    __table_args__ = (UniqueConstraint("provider_id", "model_id", name="uix_provider_model"),)


class Token(Base):
    __tablename__ = "tokens"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    name = Column(String(255), nullable=False)
    key_hash = Column(String(255), nullable=False, unique=True)
    max_input_tokens = Column(Integer, nullable=True)
    max_output_tokens = Column(Integer, nullable=True)
    requests_per_minute = Column(Integer, nullable=True)
    is_active = Column(Boolean, default=True, nullable=False)
    created_at = Column(Integer, nullable=False)

    token_permissions = relationship("TokenModelPermission", back_populates="token", cascade="all, delete-orphan")
    usage_logs = relationship("UsageLog", back_populates="token")


class TokenModelPermission(Base):
    __tablename__ = "token_model_permissions"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    token_id = Column(UUID(as_uuid=True), ForeignKey("tokens.id", ondelete="CASCADE"), nullable=False)
    model_id = Column(UUID(as_uuid=True), ForeignKey("models.id", ondelete="CASCADE"), nullable=False)
    max_input_tokens = Column(Integer, nullable=True)
    max_output_tokens = Column(Integer, nullable=True)
    created_at = Column(Integer, nullable=False)

    token = relationship("Token", back_populates="token_permissions")
    model = relationship("Model", back_populates="token_permissions")

    __table_args__ = (UniqueConstraint("token_id", "model_id", name="uix_token_model"),)


class UsageLog(Base):
    __tablename__ = "usage_logs"

    id = Column(UUID(as_uuid=True), primary_key=True, default=uuid.uuid4)
    token_id = Column(UUID(as_uuid=True), ForeignKey("tokens.id", ondelete="SET NULL"), nullable=True)
    provider_id = Column(UUID(as_uuid=True), ForeignKey("providers.id", ondelete="SET NULL"), nullable=True)
    model_id = Column(UUID(as_uuid=True), ForeignKey("models.id", ondelete="SET NULL"), nullable=True)
    model_name = Column(Text, nullable=True)
    provider_name = Column(Text, nullable=True)
    request_path = Column(Text, nullable=True)
    input_tokens = Column(Integer, default=0, nullable=False)
    output_tokens = Column(Integer, default=0, nullable=False)
    total_tokens = Column(Integer, default=0, nullable=False)
    latency_ms = Column(Integer, nullable=True)
    status_code = Column(Integer, nullable=True)
    created_at = Column(Integer, nullable=False)

    token = relationship("Token", back_populates="usage_logs")
    provider = relationship("Provider", back_populates="usage_logs")
    model = relationship("Model", back_populates="usage_logs")


class Config(Base):
    __tablename__ = "config"

    key = Column(String(255), primary_key=True)
    value = Column(Text, nullable=False)
    updated_at = Column(Integer, nullable=False)
