"""QuantumClaw SDK 自定义异常"""


class AuthenticationError(Exception):
    """认证失败（API Key 无效或过期）"""
    pass


class RateLimitError(Exception):
    """请求频率超限"""
    pass


class InsufficientQuotaError(Exception):
    """额度不足"""
    pass


class APIError(Exception):
    """通用 API 错误"""
    pass
