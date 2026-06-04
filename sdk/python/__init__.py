"""
QuantumClaw Python SDK

Usage:
    from quantumclaw import QuantumClaw

    client = QuantumClaw(api_key="sk-xxx", base_url="https://your-instance.com")
    response = client.chat_completions(model="gpt-4o", messages=[{"role":"user","content":"Hello"}])
"""

from .client import QuantumClaw
from .errors import (
    AuthenticationError,
    RateLimitError,
    InsufficientQuotaError,
    APIError,
)

__all__ = [
    "QuantumClaw",
    "AuthenticationError",
    "RateLimitError",
    "InsufficientQuotaError",
    "APIError",
]
__version__ = "0.1.0"
