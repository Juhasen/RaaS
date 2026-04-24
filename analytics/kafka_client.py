import os
from contextlib import contextmanager
from kafka import KafkaProducer, KafkaConsumer
from kafka.errors import KafkaError
import json
import logging

logger = logging.getLogger(__name__)

class KafkaProducerWrapper:
    """Wrapper for Kafka Producer with idempotent configuration"""
    
    def __init__(self):
        self.producer = None
        self.initialize()
    
    def initialize(self):
        """Initialize Kafka producer with idempotent settings"""
        try:
            self.producer = KafkaProducer(
                bootstrap_servers=os.getenv("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092"),
                value_serializer=lambda v: json.dumps(v).encode('utf-8'),
                acks='all',
                retries=3,
                client_id=os.getenv("KAFKA_CLIENT_ID", "python-producer")
            )
            logger.info("Kafka Producer initialized successfully")
        except Exception as e:
            logger.error(f"Failed to initialize Kafka Producer: {e}")
            raise
    
    def send_event(self, topic: str, event: dict, key: str = None) -> bool:
        """
        Send event to Kafka topic
        
        Args:
            topic: Kafka topic name
            event: Event data as dictionary
            key: Optional partition key for idempotency
            
        Returns:
            bool: True if sent successfully, False otherwise
        """
        try:
            future = self.producer.send(
                topic, 
                value=event, 
                key=key.encode('utf-8') if key else None
            )
            future.get(timeout=10)
            logger.info(f"Event sent to {topic}")
            return True
        except KafkaError as e:
            logger.error(f"Failed to send event to {topic}: {e}")
            return False
    
    def close(self):
        """Close producer connection"""
        if self.producer:
            self.producer.close()


class KafkaConsumerWrapper:
    """Wrapper for Kafka Consumer with idempotent configuration"""
    
    def __init__(self, group_id: str, topics: list):
        self.group_id = group_id
        self.topics = topics
        self.consumer = None
        self.initialize()
    
    def initialize(self):
        """Initialize Kafka consumer"""
        try:
            self.consumer = KafkaConsumer(
                *self.topics,
                bootstrap_servers=os.getenv("KAFKA_BOOTSTRAP_SERVERS", "localhost:9092"),
                group_id=self.group_id,
                value_deserializer=lambda m: json.loads(m.decode('utf-8')),
                auto_offset_reset='earliest',
                enable_auto_commit=True,
                max_poll_records=100
            )
            logger.info(f"Kafka Consumer initialized for group: {self.group_id}")
        except Exception as e:
            logger.error(f"Failed to initialize Kafka Consumer: {e}")
            raise
    
    def consume_messages(self, callback, timeout_ms=1000):
        """
        Consume messages from subscribed topics
        
        Args:
            callback: Function to call for each message
            timeout_ms: Poll timeout in milliseconds
        """
        try:
            for message in self.consumer:
                try:
                    # Process message
                    result = callback(message.value)
                    logger.info(f"Message processed: {result}")
                except Exception as e:
                    logger.error(f"Error processing message: {e}")
                    # Continue processing without throwing
        except Exception as e:
            logger.error(f"Consumer error: {e}")
            raise
    
    def close(self):
        """Close consumer connection"""
        if self.consumer:
            self.consumer.close()
