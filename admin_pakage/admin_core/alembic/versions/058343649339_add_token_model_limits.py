"""add_token_model_limits

Revision ID: 058343649339
Revises: cddc821b5571
Create Date: 2026-05-03 16:00:00.000000

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision: str = '058343649339'
down_revision: Union[str, None] = 'cddc821b5571'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    op.add_column('token_model_permissions', sa.Column('max_input_tokens', sa.Integer(), nullable=True))
    op.add_column('token_model_permissions', sa.Column('max_output_tokens', sa.Integer(), nullable=True))


def downgrade() -> None:
    op.drop_column('token_model_permissions', 'max_output_tokens')
    op.drop_column('token_model_permissions', 'max_input_tokens')
