"""add_provider_max_concurrent_requests

Revision ID: d37a3834030b
Revises: 058343649339
Create Date: 2026-05-03 16:07:29.820492

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision: str = 'd37a3834030b'
down_revision: Union[str, None] = '058343649339'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    op.add_column('providers', sa.Column('max_concurrent_requests', sa.Integer(), server_default='100', nullable=False))


def downgrade() -> None:
    op.drop_column('providers', 'max_concurrent_requests')
