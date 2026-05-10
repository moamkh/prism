"""convert_datetime_to_epoch

Revision ID: cddc821b5571
Revises: 47efed5e55a6
Create Date: 2026-05-03 15:00:00.000000

"""
from typing import Sequence, Union

from alembic import op
import sqlalchemy as sa


# revision identifiers, used by Alembic.
revision: str = 'cddc821b5571'
down_revision: Union[str, None] = '47efed5e55a6'
branch_labels: Union[str, Sequence[str], None] = None
depends_on: Union[str, Sequence[str], None] = None


def upgrade() -> None:
    # Drop existing defaults first to avoid cast errors
    op.execute("ALTER TABLE providers ALTER COLUMN created_at DROP DEFAULT")
    op.execute("ALTER TABLE providers ALTER COLUMN updated_at DROP DEFAULT")
    op.execute("ALTER TABLE models ALTER COLUMN created_at DROP DEFAULT")
    op.execute("ALTER TABLE tokens ALTER COLUMN created_at DROP DEFAULT")
    op.execute("ALTER TABLE token_model_permissions ALTER COLUMN created_at DROP DEFAULT")
    op.execute("ALTER TABLE usage_logs ALTER COLUMN created_at DROP DEFAULT")
    op.execute("ALTER TABLE config ALTER COLUMN updated_at DROP DEFAULT")

    # Convert types
    op.alter_column('providers', 'created_at',
                    type_=sa.BigInteger(),
                    postgresql_using='EXTRACT(EPOCH FROM created_at)::bigint')
    op.alter_column('providers', 'updated_at',
                    type_=sa.BigInteger(),
                    postgresql_using='EXTRACT(EPOCH FROM updated_at)::bigint')
    op.alter_column('models', 'created_at',
                    type_=sa.BigInteger(),
                    postgresql_using='EXTRACT(EPOCH FROM created_at)::bigint')
    op.alter_column('tokens', 'created_at',
                    type_=sa.BigInteger(),
                    postgresql_using='EXTRACT(EPOCH FROM created_at)::bigint')
    op.alter_column('token_model_permissions', 'created_at',
                    type_=sa.BigInteger(),
                    postgresql_using='EXTRACT(EPOCH FROM created_at)::bigint')
    op.alter_column('usage_logs', 'created_at',
                    type_=sa.BigInteger(),
                    postgresql_using='EXTRACT(EPOCH FROM created_at)::bigint')
    op.alter_column('config', 'updated_at',
                    type_=sa.BigInteger(),
                    postgresql_using='EXTRACT(EPOCH FROM updated_at)::bigint')


def downgrade() -> None:
    op.alter_column('config', 'updated_at',
                    type_=sa.DateTime(),
                    postgresql_using='TO_TIMESTAMP(updated_at)')
    op.alter_column('usage_logs', 'created_at',
                    type_=sa.DateTime(),
                    postgresql_using='TO_TIMESTAMP(created_at)')
    op.alter_column('token_model_permissions', 'created_at',
                    type_=sa.DateTime(),
                    postgresql_using='TO_TIMESTAMP(created_at)')
    op.alter_column('tokens', 'created_at',
                    type_=sa.DateTime(),
                    postgresql_using='TO_TIMESTAMP(created_at)')
    op.alter_column('models', 'created_at',
                    type_=sa.DateTime(),
                    postgresql_using='TO_TIMESTAMP(created_at)')
    op.alter_column('providers', 'updated_at',
                    type_=sa.DateTime(),
                    postgresql_using='TO_TIMESTAMP(updated_at)')
    op.alter_column('providers', 'created_at',
                    type_=sa.DateTime(),
                    postgresql_using='TO_TIMESTAMP(created_at)')
