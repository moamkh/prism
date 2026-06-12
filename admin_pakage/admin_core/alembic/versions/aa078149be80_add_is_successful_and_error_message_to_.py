"""add is_successful and error_message to usage_logs

Revision ID: aa078149be80
Revises: e4a1b2c3d4e5
Create Date: 2026-05-10 00:00:00.000000

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision: str = 'aa078149be80'
down_revision: Union[str, None] = 'e4a1b2c3d4e5'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    op.add_column('usage_logs', sa.Column('is_successful', sa.Boolean(), nullable=False, server_default='true'))
    op.add_column('usage_logs', sa.Column('error_message', sa.Text(), nullable=True))


def downgrade() -> None:
    op.drop_column('usage_logs', 'error_message')
    op.drop_column('usage_logs', 'is_successful')
