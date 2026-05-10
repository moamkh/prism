"""add_display_model_id

Revision ID: e4a1b2c3d4e5
Revises: ad48a6ac6db5
Create Date: 2026-05-09 15:55:00.000000

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision: str = 'e4a1b2c3d4e5'
down_revision: Union[str, None] = 'ad48a6ac6db5'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    op.add_column('models', sa.Column('display_model_id', sa.String(255), nullable=True))


def downgrade() -> None:
    op.drop_column('models', 'display_model_id')
