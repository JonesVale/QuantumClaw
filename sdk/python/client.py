"""
QuantumClaw Python SDK — 管理工具包

定位：给 QuantumClaw 网关的管理员使用，用于在代码中自动化管理
      API Key、查用量、配渠道、管用户。

不是 AI 调用库。调 AI 请用 OpenAI 官方 SDK。

所有 API Key 格式：sk-<48位字母数字>  (51字符)
"""
import json
from typing import Optional, Dict, Any, List, Union, Generator

import requests

from .errors import AuthenticationError, RateLimitError, InsufficientQuotaError, APIError


class QuantumClaw:
    """QuantumClaw 网关管理客户端"""

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
                    msg = body.get("message", body.get("error", {}).get("message", "Unknown error"))
                    if "auth" in msg.lower() or "token" in msg.lower():
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

    # ════════════════════════════════════════════════
    # 管理端 API — 这是 SDK 的核心价值
    # ════════════════════════════════════════════════

    # ──── 1. API Key（令牌）管理 ────

    def list_tokens(self, page: int = 0, order: str = "") -> List[Dict[str, Any]]:
        """列出当前用户的所有 API Key

        >>> tokens = client.list_tokens()
        """
        params = {"p": page}
        if order:
            params["order"] = order
        resp = self._request("GET", "/api/token/", params=params)
        return resp.json().get("data", [])

    def create_token(
        self,
        name: str,
        remain_quota: int = 0,
        unlimited_quota: bool = False,
        expired_time: int = -1,
        models: Optional[str] = None,
        subnet: Optional[str] = None,
    ) -> Dict[str, Any]:
        """创建新的 API Key

        返回的 key 包含 sk- 前缀。新 Key 只在创建时返回一次，请保存。

        Args:
            name: Key 名称（便于管理）
            remain_quota: 剩余额度（分），0=耗尽
            unlimited_quota: 是否无限额度
            expired_time: 过期时间戳，-1=永不过期
            models: 可用模型列表，逗号分隔，如 "gpt-4,gpt-3.5-turbo"
            subnet: IP 白名单 CIDR，如 "192.168.1.0/24"

        >>> token = client.create_token(name="dev-key", remain_quota=100000)
        >>> print(token["key"])  # sk-xxxx...
        """
        body = {
            "name": name,
            "remain_quota": remain_quota,
            "unlimited_quota": unlimited_quota,
            "expired_time": expired_time,
        }
        if models:
            body["models"] = models
        if subnet:
            body["subnet"] = subnet
        resp = self._request("POST", "/api/token/", json=body)
        return resp.json().get("data", {})

    def batch_create_tokens(self, tokens: List[Dict[str, Any]]) -> List[Dict[str, Any]]:
        """批量创建 API Key

        Args:
            tokens: 每个元素的字段与 create_token 一致

        >>> keys = client.batch_create_tokens([
        ...     {"name": "alice-dev", "remain_quota": 100000, "models": "gpt-4o"},
        ...     {"name": "bob-dev", "remain_quota": 50000, "models": "gpt-3.5-turbo"},
        ... ])
        """
        results = []
        for t in tokens:
            results.append(self.create_token(**t))
        return results

    def update_token(self, token_id: int, **kwargs) -> Dict[str, Any]:
        """更新 API Key 设置

        可更新：name, status, remain_quota, unlimited_quota,
               expired_time, models, subnet

        >>> client.update_token(token_id=1, name="new-name", unlimited_quota=True)
        """
        body = {"id": token_id, **kwargs}
        resp = self._request("PUT", "/api/token/", json=body)
        return resp.json().get("data", {})

    def delete_token(self, token_id: int) -> bool:
        """删除 API Key"""
        resp = self._request("DELETE", f"/api/token/{token_id}")
        return resp.json().get("success", False)

    def get_token(self, token_id: int) -> Dict[str, Any]:
        """获取单个 API Key 详情"""
        resp = self._request("GET", f"/api/token/{token_id}")
        return resp.json().get("data", {})

    def search_tokens(self, keyword: str) -> List[Dict[str, Any]]:
        """按名称搜索 API Key"""
        resp = self._request("GET", "/api/token/search", params={"keyword": keyword})
        return resp.json().get("data", [])

    def toggle_token(self, token_id: int, enable: bool) -> Dict[str, Any]:
        """启用或禁用 API Key"""
        status = 1 if enable else 2
        return self.update_token(token_id, status=status)

    # ──── 2. 用户信息与余额 ────

    def get_self_info(self) -> Dict[str, Any]:
        """获取当前登录用户信息"""
        resp = self._request("GET", "/api/user/self")
        return resp.json().get("data", {})

    def check_balance(self) -> Dict[str, Any]:
        """查询账户余额"""
        resp = self._request("GET", "/api/user/self/balance")
        return resp.json().get("data", {})

    def get_dashboard(self) -> Dict[str, Any]:
        """获取用户仪表盘数据"""
        resp = self._request("GET", "/api/user/self/dashboard")
        return resp.json().get("data", {})

    def get_available_models(self) -> List[str]:
        """获取当前用户可用的模型列表"""
        resp = self._request("GET", "/api/user/self/available_models")
        return resp.json().get("data", [])

    def get_transaction_logs(self, page: int = 0) -> List[Dict[str, Any]]:
        """查询交易明细"""
        resp = self._request("GET", "/api/user/self/transaction_logs", params={"p": page})
        return resp.json().get("data", [])

    def get_topup_list(self) -> List[Dict[str, Any]]:
        """查询充值记录"""
        resp = self._request("GET", "/api/user/self/topup/list")
        return resp.json().get("data", [])

    def get_subscription_plans(self) -> List[Dict[str, Any]]:
        """获取订阅套餐列表"""
        resp = self._request("GET", "/api/user/self/subscription/plans")
        return resp.json().get("data", [])

    def get_subscription_self(self) -> Dict[str, Any]:
        """查询当前用户的订阅状态"""
        resp = self._request("GET", "/api/user/self/subscription/self")
        return resp.json().get("data", {})

    def get_notifications(self) -> List[Dict[str, Any]]:
        """获取通知列表"""
        resp = self._request("GET", "/api/user/self/notifications")
        return resp.json().get("data", [])

    def get_checkin_status(self) -> Dict[str, Any]:
        """获取签到状态"""
        resp = self._request("GET", "/api/user/self/checkin")
        return resp.json().get("data", {})

    def get_team(self) -> List[Dict[str, Any]]:
        """获取我的团队"""
        resp = self._request("GET", "/api/user/self/team")
        return resp.json().get("data", [])

    # ──── 3. 计费与账单 ────

    def get_billing_stats(self) -> Dict[str, Any]:
        """查询计费统计"""
        resp = self._request("GET", "/api/user/self/billing/stats")
        return resp.json().get("data", {})

    def get_billing_records(self) -> List[Dict[str, Any]]:
        """查询计费记录"""
        resp = self._request("GET", "/api/user/self/billing/records")
        return resp.json().get("data", [])

    def get_transactions(self, page: int = 0) -> List[Dict[str, Any]]:
        """查询结算列表（管理员）"""
        resp = self._request("GET", "/api/transactions", params={"p": page})
        return resp.json().get("data", [])

    # ──── 4. 日志 ────

    def get_logs(self, start_idx: int = 0, num: int = 10) -> List[Dict[str, Any]]:
        """查询自己的请求日志"""
        resp = self._request("GET", "/api/log/self", params={"startIdx": start_idx, "num": num})
        return resp.json().get("data", [])

    def get_logs_stat(self) -> Dict[str, Any]:
        """查询自己的日志统计"""
        resp = self._request("GET", "/api/log/self/stat")
        return resp.json().get("data", {})

    # ──── 5. 渠道管理（管理员）────

    def list_channels(self) -> List[Dict[str, Any]]:
        """列出所有上游渠道"""
        resp = self._request("GET", "/api/channel/")
        return resp.json().get("data", [])

    def search_channels(self, keyword: str) -> List[Dict[str, Any]]:
        """搜索渠道"""
        resp = self._request("GET", "/api/channel/search", params={"keyword": keyword})
        return resp.json().get("data", [])

    def get_channel(self, channel_id: int) -> Dict[str, Any]:
        """获取渠道详情"""
        resp = self._request("GET", f"/api/channel/{channel_id}")
        return resp.json().get("data", {})

    def add_channel(self, **kwargs) -> Dict[str, Any]:
        """添加上游 AI 厂商渠道

        >>> client.add_channel(
        ...     name="OpenAI Official",
        ...     type=1,
        ...     key="sk-xxx...",
        ...     models="gpt-4,gpt-3.5-turbo"
        ... )
        """
        resp = self._request("POST", "/api/channel/", json=kwargs)
        return resp.json().get("data", {})

    def update_channel(self, **kwargs) -> Dict[str, Any]:
        """更新渠道配置"""
        resp = self._request("PUT", "/api/channel/", json=kwargs)
        return resp.json().get("data", {})

    def delete_channel(self, channel_id: int) -> bool:
        """删除渠道"""
        resp = self._request("DELETE", f"/api/channel/{channel_id}")
        return resp.json().get("success", False)

    def get_channel_types(self) -> List[Dict[str, Any]]:
        """获取渠道类型列表"""
        resp = self._request("GET", "/api/channel/types")
        return resp.json().get("data", [])

    # ──── 6. 用户管理（管理员）────

    def admin_list_users(self, page: int = 0) -> List[Dict[str, Any]]:
        """列出所有用户"""
        resp = self._request("GET", "/api/user/", params={"p": page})
        return resp.json().get("data", [])

    def admin_search_users(self, keyword: str) -> List[Dict[str, Any]]:
        """搜索用户"""
        resp = self._request("GET", "/api/user/search", params={"keyword": keyword})
        return resp.json().get("data", [])

    def admin_get_user(self, user_id: int) -> Dict[str, Any]:
        """获取用户详情"""
        resp = self._request("GET", f"/api/user/{user_id}")
        return resp.json().get("data", {})

    def admin_create_user(self, username: str, password: str, **kwargs) -> Dict[str, Any]:
        """创建用户"""
        body = {"username": username, "password": password, **kwargs}
        resp = self._request("POST", "/api/user/", json=body)
        return resp.json().get("data", {})

    def admin_update_user(self, **kwargs) -> Dict[str, Any]:
        """更新用户属性"""
        resp = self._request("PUT", "/api/user/", json=kwargs)
        return resp.json().get("data", {})

    def admin_delete_user(self, user_id: int) -> bool:
        """删除用户"""
        resp = self._request("DELETE", f"/api/user/{user_id}")
        return resp.json().get("success", False)

    def admin_add_balance(self, user_id: int, amount: int, remark: str = "") -> Dict[str, Any]:
        """为用户充值额度

        Args:
            user_id: 用户 ID
            amount: 充值额度（分）
            remark: 备注
        """
        body = {"user_id": user_id, "amount": amount, "remark": remark}
        resp = self._request("POST", "/api/user/add_balance", json=body)
        return resp.json().get("data", {})

    # ──── 7. 管理员日志 ────

    def admin_get_logs(self, page: int = 0, model: str = "") -> List[Dict[str, Any]]:
        """查询所有请求日志"""
        params = {"startIdx": page * 10, "num": 10}
        if model:
            params["model"] = model
        resp = self._request("GET", "/api/log/", params=params)
        return resp.json().get("data", [])

    def admin_get_logs_stat(self) -> Dict[str, Any]:
        """日志统计"""
        resp = self._request("GET", "/api/log/stat")
        return resp.json().get("data", {})

    def admin_search_logs(self, keyword: str) -> List[Dict[str, Any]]:
        """搜索日志"""
        resp = self._request("GET", "/api/log/search", params={"keyword": keyword})
        return resp.json().get("data", [])

    # ──── 8. 兑换码管理 ────

    def list_redemptions(self) -> List[Dict[str, Any]]:
        """列出所有兑换码"""
        resp = self._request("GET", "/api/redemption/")
        return resp.json().get("data", [])

    def create_redemption(self, name: str, quota: int, count: int = 1) -> Dict[str, Any]:
        """创建兑换码"""
        resp = self._request("POST", "/api/redemption/", json={
            "name": name, "quota": quota, "count": count
        })
        return resp.json().get("data", {})

    # ──── 9. 量子计算资源 ────

    def list_quantum_backends(self) -> List[Dict[str, Any]]:
        """获取所有可用量子后端（IonQ、IBM Q、Rigetti、AWS Braket 等）

        >>> backends = client.list_quantum_backends()
        >>> for b in backends:
        ...     print(b["provider"], b["backend_name"], b["status"])
        """
        resp = self._request("GET", "/api/quantum/backends")
        return resp.json().get("data", [])

    def list_quantum_providers(self) -> List[Dict[str, Any]]:
        """获取量子供应商统计

        返回每家供应商的配置状态和后端数量。
        """
        resp = self._request("GET", "/api/quantum/providers")
        return resp.json().get("data", [])

    def submit_quantum_task(
        self,
        provider: str,
        qasm: str,
        backend: str = "",
        shots: int = 1024,
        wait: bool = False,
    ) -> Dict[str, Any]:
        """提交量子计算任务（OpenQASM 电路）

        Args:
            provider: 供应商名称（IonQ / IBMQ / Rigetti / AWSBraket / AzureQuantum / GoogleQuantum）
            qasm: OpenQASM 电路代码
            backend: 后端名称（留空使用默认）
            shots: 测量次数（默认 1024）
            wait: 是否同步等待完成（False=异步返回 task_id）

        >>> result = client.submit_quantum_task(
        ...     provider="IonQ",
        ...     qasm="OPENQASM 2.0; qreg q[2]; creg c[2]; h q[0]; cx q[0],q[1]; measure q -> c;",
        ...     shots=1024
        ... )
        """
        body = {"provider": provider, "qasm": qasm, "shots": shots, "wait": wait}
        if backend:
            body["backend"] = backend
        resp = self._request("POST", "/api/quantum/submit", json=body)
        return resp.json().get("data", {})

    def seed_quantum_models(self) -> Dict[str, Any]:
        """（管理员）重新加载量子模型到模型目录"""
        resp = self._request("GET", "/api/models/seed-quantum")
        return resp.json().get("data", {})

    # ──── 10. 异步任务管理（Midjourney/视频/音乐/量子等）────

    def list_tasks(self) -> List[Dict[str, Any]]:
        """列出当前用户的所有异步任务"""
        resp = self._request("GET", "/api/task/")
        return resp.json().get("data", [])

    def get_task(self, task_id: str) -> Dict[str, Any]:
        """获取异步任务状态"""
        resp = self._request("GET", f"/api/task/{task_id}")
        return resp.json().get("data", {})

    def cancel_task(self, task_id: str) -> Dict[str, Any]:
        """取消异步任务"""
        resp = self._request("POST", f"/api/task/{task_id}/cancel")
        return resp.json().get("data", {})

    def delete_task(self, task_id: str) -> bool:
        """删除异步任务"""
        resp = self._request("DELETE", f"/api/task/{task_id}")
        return resp.json().get("success", False)

    # ──── 11. 系统 ────

    def get_status(self) -> Dict[str, Any]:
        """获取系统状态"""
        resp = self._request("GET", "/api/status")
        return resp.json()

    def get_site_content(self) -> Dict[str, Any]:
        """获取站点配置"""
        resp = self._request("GET", "/api/site-content")
        return resp.json().get("data", {})

    def get_models_catalog(self) -> List[Dict[str, Any]]:
        """获取模型目录"""
        resp = self._request("GET", "/api/model-catalog")
        return resp.json().get("data", [])

    # ════════════════════════════════════════════════
    # Relay API（连通性测试用）
    # 生产环境调 AI 请使用 OpenAI 官方 SDK：
    #   from openai import OpenAI
    #   client = OpenAI(api_key="sk-...", base_url="https://你的实例/v1")
    # ════════════════════════════════════════════════

    def chat_completions(self, model: str, messages: List[Dict[str, str]], **kwargs) -> Dict[str, Any]:
        """⚠️ 连通性测试用。生产环境请使用 OpenAI SDK。

        非流式聊天补全，验证网关转发是否正常工作。
        """
        body = {"model": model, "messages": messages, **kwargs}
        resp = self._request("POST", "/v1/chat/completions", json=body)
        return resp.json()

    def stream_chat(self, model: str, messages: List[Dict[str, str]], **kwargs) -> Generator[Dict[str, Any], None, None]:
        """⚠️ 连通性测试用。生产环境请使用 OpenAI SDK。

        流式聊天补全，验证网关 SSE 转发是否正常工作。
        """
        body = {"model": model, "messages": messages, "stream": True, **kwargs}
        url = f"{self.base_url}/v1/chat/completions"
        with self._session.post(url, json=body, stream=True, timeout=self.timeout) as resp:
            if resp.status_code != 200:
                raise APIError(f"Stream failed: HTTP {resp.status_code}")
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

    def list_models(self) -> List[Dict[str, Any]]:
        """⚠️ 连通性测试用。生产环境请使用 OpenAI SDK。

        列出网关可用的模型列表。
        """
        resp = self._request("GET", "/v1/models")
        return resp.json().get("data", [])
