"""add_model_rate_limits

Revision ID: f5a2b1c3d4e6
Revises: aa078149be80
Create Date: 2026-05-18 11:15:00.000000

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision: str = 'f5a2b1c3d4e6'
down_revision: Union[str, None] = 'aa078149be80'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    op.add_column('models', sa.Column('max_concurrent_requests', sa.Integer(), nullable=True))
    op.add_column('models', sa.Column('queue_size', sa.Integer(), nullable=True))


def downgrade() -> None:
    op.drop_column('models', 'queue_size')
    op.drop_column('models', 'max_concurrent_requests')
