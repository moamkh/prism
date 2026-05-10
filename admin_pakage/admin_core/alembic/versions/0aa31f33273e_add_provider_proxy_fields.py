"""add_provider_proxy_fields

Revision ID: 0aa31f33273e
Revises: 001
Create Date: 2026-05-03 13:24:49.345143

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision: str = '0aa31f33273e'
down_revision: Union[str, None] = '001'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    op.add_column('providers', sa.Column('http_proxy', sa.Text(), nullable=True))
    op.add_column('providers', sa.Column('socks5_proxy', sa.Text(), nullable=True))


def downgrade() -> None:
    op.drop_column('providers', 'socks5_proxy')
    op.drop_column('providers', 'http_proxy')
