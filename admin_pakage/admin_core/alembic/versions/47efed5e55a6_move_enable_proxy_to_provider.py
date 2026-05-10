"""move_enable_proxy_to_provider

Revision ID: 47efed5e55a6
Revises: b03a40f7d41f
Create Date: 2026-05-03 14:00:00.000000

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision: str = '47efed5e55a6'
down_revision: Union[str, None] = 'b03a40f7d41f'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    op.drop_column('models', 'enable_proxy')
    op.add_column('providers', sa.Column('enable_proxy', sa.Boolean(), nullable=False, server_default='true'))


def downgrade() -> None:
    op.drop_column('providers', 'enable_proxy')
    op.add_column('models', sa.Column('enable_proxy', sa.Boolean(), nullable=False, server_default='true'))
