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
                "postgresql://user:password@localhost:5432/analytics_db"
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


# MongoDB Connection
class MongoDBConnection:
    """Wrapper for MongoDB connection"""
    
    def __init__(self):
        self.client = None
        self.db = None
        self.initialize()
    
    def initialize(self):
        """Initialize MongoDB connection"""
        try:
            from pymongo import MongoClient
            mongo_url = os.getenv(
                "MONGO_URL",
                "mongodb://localhost:27017"
            )
            self.client = MongoClient(mongo_url)
            db_name = os.getenv("MONGO_DB_NAME", "analytics_db")
            self.db = self.client[db_name]
            logger.info("MongoDB connection initialized")
        except Exception as e:
            logger.error(f"Failed to initialize MongoDB: {e}")
            raise
    
    def get_collection(self, collection_name: str):
        """Get MongoDB collection"""
        return self.db[collection_name]
    
    def close(self):
        """Close MongoDB connection"""
        if self.client:
            self.client.close()


# Global connection instances
_postgres = None
_mongo = None

def get_postgres() -> PostgreSQLConnection:
    """Get or initialize PostgreSQL connection"""
    global _postgres
    if _postgres is None:
        _postgres = PostgreSQLConnection()
    return _postgres

def get_mongo() -> MongoDBConnection:
    """Get or initialize MongoDB connection"""
    global _mongo
    if _mongo is None:
        _mongo = MongoDBConnection()
    return _mongo
