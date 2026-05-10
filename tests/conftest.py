import os
import sys
import pytest
from sqlalchemy import create_engine, text
from sqlalchemy.orm import sessionmaker

# Add admin_core to path
sys.path.insert(0, os.path.join(os.path.dirname(__file__), '..', 'admin_pakage', 'admin_core'))

from app.database import Base
from app.main import app
from app.config import settings
from fastapi.testclient import TestClient

TEST_DB_URL = "postgresql://postgres:postgres@localhost:5432/reverse_proxy_manager_test_db"


def ensure_test_db():
    """Create test database if it doesn't exist."""
    engine = create_engine("postgresql://postgres:postgres@localhost:5432/postgres")
    with engine.connect() as conn:
        conn.execution_options(isolation_level="AUTOCOMMIT")
        result = conn.execute(text("SELECT 1 FROM pg_database WHERE datname = 'reverse_proxy_manager_test_db'"))
        if not result.fetchone():
            conn.execute(text("CREATE DATABASE reverse_proxy_manager_test_db"))
    engine.dispose()


@pytest.fixture(scope="session", autouse=True)
def setup_test_db():
    ensure_test_db()
    engine = create_engine(TEST_DB_URL)
    Base.metadata.create_all(bind=engine)
    engine.dispose()
    yield
    # Cleanup: drop all tables after tests
    engine = create_engine(TEST_DB_URL)
    Base.metadata.drop_all(bind=engine)
    engine.dispose()


@pytest.fixture
def db_session():
    engine = create_engine(TEST_DB_URL)
    SessionLocal = sessionmaker(autocommit=False, autoflush=False, bind=engine)
    session = SessionLocal()
    try:
        yield session
    finally:
        session.close()
        engine.dispose()


@pytest.fixture
def client(db_session):
    """FastAPI test client with test DB."""
    from app.api.deps import get_db

    original_override = app.dependency_overrides.get(get_db)

    def override_get_db():
        try:
            yield db_session
        finally:
            pass

    app.dependency_overrides[get_db] = override_get_db

    # Point config to test DB
    original_url = settings.database_url
    settings.database_url = TEST_DB_URL

    with TestClient(app) as test_client:
        yield test_client

    settings.database_url = original_url
    if original_override:
        app.dependency_overrides[get_db] = original_override
    else:
        app.dependency_overrides.pop(get_db, None)


@pytest.fixture
def clean_tables(db_session):
    """Clear all data from tables before test."""
    from app.models.models import Provider, Model, Token, UsageLog
    db_session.query(UsageLog).delete()
    db_session.query(Token).delete()
    db_session.query(Model).delete()
    db_session.query(Provider).delete()
    db_session.commit()
    yield
