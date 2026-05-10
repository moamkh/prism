"""add model_name and provider_name to usage_logs

Revision ID: ad48a6ac6db5
Revises: d37a3834030b
Create Date: 2026-05-03 20:06:31.262432

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision: str = 'ad48a6ac6db5'
down_revision: Union[str, None] = 'd37a3834030b'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    op.add_column('usage_logs', sa.Column('model_name', sa.Text(), nullable=True))
    op.add_column('usage_logs', sa.Column('provider_name', sa.Text(), nullable=True))


def downgrade() -> None:
    op.drop_column('usage_logs', 'provider_name')
    op.drop_column('usage_logs', 'model_name')
