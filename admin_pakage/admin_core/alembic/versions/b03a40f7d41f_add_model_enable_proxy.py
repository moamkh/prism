"""add_model_enable_proxy

Revision ID: b03a40f7d41f
Revises: 0aa31f33273e
Create Date: 2026-05-03 13:33:29.581381

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision: str = 'b03a40f7d41f'
down_revision: Union[str, None] = '0aa31f33273e'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    op.add_column('models', sa.Column('enable_proxy', sa.Boolean(), nullable=False, server_default='true'))


def downgrade() -> None:
    op.drop_column('models', 'enable_proxy')
