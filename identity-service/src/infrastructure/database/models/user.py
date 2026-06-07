from sqlalchemy.orm import Mapped, mapped_column
from sqlalchemy import String
from uuid import UUID, uuid4
from src.infrastructure.database.base import Base

class UserModel(Base):
    __tablename__ = "users"

    id: Mapped[UUID] = mapped_column(primary_key=True, default=uuid4)
    email: Mapped[str] = mapped_column(String(255), unique=True)
    hashed_password: Mapped[str] = mapped_column(String)
    is_active: Mapped[bool] = mapped_column(default=True)