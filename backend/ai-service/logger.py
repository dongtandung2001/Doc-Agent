# logger.py
import logging
import sys
from config import Config

def setup_logger(name: str) -> logging.Logger:
    logger = logging.getLogger(name)
    logger.setLevel(getattr(logging, Config.LOG_LEVEL))

    handler = logging.StreamHandler(sys.stdout)
    handler.setLevel(getattr(logging, Config.LOG_LEVEL))

    formatter = logging.Formatter(
        '%(asctime)s - %(name)s - %(levelname)s - %(message)s'
    )
    handler.setFormatter(formatter)

    logger.addHandler(handler)
    return logger