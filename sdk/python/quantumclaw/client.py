"""
QuantumClaw Python SDK — 核心客户端

提供 OpenAI 兼容的 API 调用接口，支持流式和非流式聊天补全。
"""

import json
import time
from typing import Iterator, Optional, Dict, Any, List, Union

import requests

from .errors import AuthenticationError, RateLimitError, InsufficientQuotaError, APIError


class QuantumClaw:
    """QuantumClaw API Client"""

    def __init__(
        self,
        api_key: str,
        base_url: str = "http://localhost:3666",
        timeout: int = 120,
        max_retries: int = 3,
    ):
        self.api_key = api_key
        self.base_url = base_url.rstrip("/")
        self.timeout = timeout
        self.max_retries = max_retries
        self._session = requests.Session()
        self._session.headers.update({
            "Authorization": f"Bearer {api_key}",
            "Content-Type": "application/json",
        })

    def _request(self, method: str, path: str, **kwargs) -> requests.Response:
        """发送请求，带重试和错误处理"""
        url = f"{self.base_url}{path}"
        kwargs.setdefault("timeout", self.timeout)

        for attempt in range(self.max_retries):
            try:
                resp = self._session.request(method, url, **kwargs)
            except requests.exceptions.Timeout:
                if attempt == self.max_retries - 1:
                    raise APIError("Request timed out")
                continue
            except requests.exceptions.ConnectionError as e:
                raise APIError(f"Connection failed: {e}")

            if resp.status_code == 200:
                body = resp.json()
                if not body.get("success", True):
                    # Check for standard QuantumClaw error format
                    msg = body.get("message", body.get("error", {}).get("message", "Unknown error"))
                    if "authentication" in msg.lower() or "token" in msg.lower() or "auth" in msg.lower():
                        raise AuthenticationError(msg)
                    if "quota" in msg.lower() or "balance" in msg.lower() or "insufficient" in msg.lower():
                        raise InsufficientQuotaError(msg)
                    if "rate" in msg.lower() or "too many" in msg.lower():
                        raise RateLimitError(msg)
                    raise APIError(msg)
                return resp
            elif resp.status_code == 401:
                raise AuthenticationError("Invalid API key")
            elif resp.status_code == 429:
                raise RateLimitError("Rate limit exceeded")
            else:
                try:
                    err_data = resp.json()
                    msg = err_data.get("error", {}).get("message", str(resp.status_code))
                except Exception:
                    msg = f"HTTP {resp.status_code}"
                raise APIError(msg)

        raise APIError("Max retries exceeded")

    # ---- Chat Completions ----

    def chat_completions(
        self,
        model: str,
        messages: List[Dict[str, str]],
        **kwargs,
    ) -> Dict[str, Any]:
        """非流式聊天补全"""
        body = {"model": model, "messages": messages, **kwargs}
        resp = self._request("POST", "/v1/chat/completions", json=body)
        return resp.json()

    def stream_chat(
        self,
        model: str,
        messages: List[Dict[str, str]],
        **kwargs,
    ) -> Iterator[Dict[str, Any]]:
        """流式聊天补全（逐行 SSE 解析）"""
        body = {"model": model, "messages": messages, "stream": True, **kwargs}
        url = f"{self.base_url}/v1/chat/completions"

        with self._session.post(url, json=body, stream=True, timeout=self.timeout) as resp:
            if resp.status_code != 200:
                raise APIError(f"Stream request failed: HTTP {resp.status_code}")
            for line in resp.iter_lines():
                if not line:
                    continue
                line = line.decode("utf-8").strip()
                if line.startswith("data: "):
                    data_str = line[6:]
                    if data_str == "[DONE]":
                        break
                    try:
                        yield json.loads(data_str)
                    except json.JSONDecodeError:
                        continue

    # ---- Models ----

    def list_models(self) -> List[Dict[str, Any]]:
        """获取可用模型列表"""
        resp = self._request("GET", "/api/models")
        data = resp.json()
        return data.get("data", [])

    # ---- Balance ----

    def check_balance(self) -> Dict[str, Any]:
        """查询账户余额"""
        resp = self._request("GET", "/api/user/balance")
        data = resp.json()
        return data.get("data", {})

    # ---- Usage ----

    def check_usage(self) -> Dict[str, Any]:
        """查询使用量"""
        try:
            resp = self._request("GET", "/api/user/usage")
            return resp.json().get("data", {})
        except APIError:
            return {}

    # ---- Token Management ----

    def list_tokens(self) -> List[Dict[str, Any]]:
        """列出所有 API Key"""
        resp = self._request("GET", "/api/token/")
        return resp.json().get("data", [])

    def create_token(self, name: str, remain_quota: int = 0) -> Dict[str, Any]:
        """创建新的 API Key"""
        resp = self._request("POST", "/api/token/", json={
            "name": name,
            "remain_quota": remain_quota,
        })
        return resp.json().get("data", {})
