import os
import logging
from sqlalchemy import create_engine
from sqlalchemy.orm import sessionmaker, Session
from contextlib import contextmanager

logger = logging.getLogger(__name__)

# PostgreSQL Connection
class PostgreSQLConnection:
    """Wrapper for PostgreSQL database connection"""
    
    def __init__(self):
        self.engine = None
        self.SessionLocal = None
        self.initialize()
    
    def initialize(self):
        """Initialize PostgreSQL connection"""
        try:
            db_url = os.getenv(
                "DATABASE_URL",
                "postgresql://user:password@localhost:5432/user_db"
            )
            self.engine = create_engine(
                db_url,
                echo=os.getenv("SQL_ECHO", "false").lower() == "true",
                pool_size=10,
                max_overflow=20
            )
            self.SessionLocal = sessionmaker(bind=self.engine)
            logger.info("PostgreSQL connection initialized")
        except Exception as e:
            logger.error(f"Failed to initialize PostgreSQL: {e}")
            raise
    
    @contextmanager
    def get_session(self) -> Session:
        """Get a database session"""
        session = self.SessionLocal()
        try:
            yield session
            session.commit()
        except Exception as e:
            session.rollback()
            logger.error(f"Database session error: {e}")
            raise
        finally:
            session.close()
    
    def close(self):
        """Close database engine"""
        if self.engine:
            self.engine.dispose()


# Global connection instances
_postgres = None

def get_postgres() -> PostgreSQLConnection:
    """Get or initialize PostgreSQL connection"""
    global _postgres
    if _postgres is None:
        _postgres = PostgreSQLConnection()
    return _postgres
