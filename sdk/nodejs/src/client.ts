import * as https from 'https';
import * as http from 'http';

import type {
  ChatCompletionRequest,
  ChatCompletionResponse,
  StreamEvent,
  Balance,
  ChatMessage,
  ApiToken,
  Channel,
  ApiResponse,
} from './types';

export interface QuantumClawConfig {
  apiKey: string;
  baseURL?: string;
  timeout?: number;
}

/**
 * QuantumClaw API Client for Node.js
 *
 * Key format: sk-xxxxxxxx... (51 chars, OpenAI-compatible)
 */
export class QuantumClaw {
  private apiKey: string;
  private baseURL: string;
  private timeout: number;

  constructor(config: QuantumClawConfig) {
    this.apiKey = config.apiKey;
    this.baseURL = (config.baseURL || 'http://localhost:3666').replace(/\/+$/, '');
    this.timeout = config.timeout || 120000;
  }

  private async request<T>(method: string, path: string, body?: unknown): Promise<ApiResponse<T>> {
    const url = new URL(this.baseURL + path);
    const isHttps = url.protocol === 'https:';
    const lib = isHttps ? https : http;

    return new Promise<ApiResponse<T>>((resolve, reject) => {
      const payload = body ? JSON.stringify(body) : undefined;

      const options: http.RequestOptions = {
        hostname: url.hostname,
        port: url.port,
        path: url.pathname + url.search,
        method,
        timeout: this.timeout,
        headers: {
          'Authorization': `Bearer ${this.apiKey}`,
          'Content-Type': 'application/json',
          'User-Agent': 'quantumclaw-nodejs-sdk/0.1.0',
          ...(payload ? { 'Content-Length': Buffer.byteLength(payload).toString() } : {}),
        },
      };

      const req = lib.request(options, (res) => {
        const chunks: Buffer[] = [];
        res.on('data', (chunk: Buffer) => chunks.push(chunk));
        res.on('end', () => {
          const raw = Buffer.concat(chunks).toString('utf-8');
          if (res.statusCode && res.statusCode >= 400) {
            reject(new Error(`HTTP ${res.statusCode}: ${raw}`));
            return;
          }
          try {
            resolve(JSON.parse(raw));
          } catch {
            reject(new Error(`Invalid JSON: ${raw.slice(0, 200)}`));
          }
        });
      });

      req.on('error', reject);
      req.on('timeout', () => { req.destroy(); reject(new Error('Request timed out')); });

      if (payload) req.write(payload);
      req.end();
    });
  }

  // ==================== OpenAI Compatible ====================

  /** Non-streaming chat completion */
  async chatCompletions(req: ChatCompletionRequest): Promise<ChatCompletionResponse> {
    const resp = await this.request<ChatCompletionResponse>('POST', '/v1/chat/completions', { ...req, stream: false });
    return resp as unknown as ChatCompletionResponse;
  }

  /** Streaming chat completion — returns async generator */
  async *streamChat(req: ChatCompletionRequest): AsyncGenerator<StreamEvent> {
    const url = new URL(this.baseURL + '/v1/chat/completions');
    const isHttps = url.protocol === 'https:';
    const lib = isHttps ? https : http;

    const payload = JSON.stringify({ ...req, stream: true });

    const stream = await new Promise<http.IncomingMessage>((resolve, reject) => {
      const options: http.RequestOptions = {
        hostname: url.hostname,
        port: url.port,
        path: url.pathname,
        method: 'POST',
        timeout: this.timeout,
        headers: {
          'Authorization': `Bearer ${this.apiKey}`,
          'Content-Type': 'application/json',
          'Content-Length': Buffer.byteLength(payload).toString(),
        },
      };

      const httpreq = lib.request(options, (res) => resolve(res));
      httpreq.on('error', reject);
      httpreq.write(payload);
      httpreq.end();
    });

    const decoder = new TextDecoder();
    let buffer = '';

    for await (const chunk of stream as unknown as AsyncIterable<Uint8Array>) {
      buffer += decoder.decode(chunk, { stream: true });
      const lines = buffer.split('\n');
      buffer = lines.pop() || '';

      for (const line of lines) {
        const trimmed = line.trim();
        if (!trimmed.startsWith('data: ')) continue;
        const data = trimmed.slice(6);
        if (data === '[DONE]') return;
        try {
          yield JSON.parse(data) as StreamEvent;
        } catch {
          // skip malformed lines
        }
      }
    }
  }

  /** List available models (OpenAI compatible) */
  async listModels(): Promise<Array<Record<string, unknown>>> {
    const resp = await this.request<Array<Record<string, unknown>>>('GET', '/v1/models');
    return (resp as any).data || [];
  }

  // ==================== User & Balance ====================

  /** Get current user info */
  async getSelfInfo(): Promise<Record<string, unknown>> {
    const resp = await this.request<Record<string, unknown>>('GET', '/api/user/self');
    return resp.data || {};
  }

  /** Check account balance */
  async getBalance(): Promise<Balance | null> {
    const resp = await this.request<Balance>('GET', '/api/user/self/balance');
    return resp.data || null;
  }

  /** Get user dashboard */
  async getDashboard(): Promise<Record<string, unknown>> {
    const resp = await this.request<Record<string, unknown>>('GET', '/api/user/self/dashboard');
    return resp.data || {};
  }

  // ==================== Token (API Key) Management ====================

  /** List all API keys */
  async listTokens(page: number = 0): Promise<ApiToken[]> {
    const resp = await this.request<ApiToken[]>('GET', `/api/token/?p=${page}`);
    return resp.data || [];
  }

  /** Create a new API key */
  async createToken(params: {
    name: string;
    remain_quota?: number;
    unlimited_quota?: boolean;
    expired_time?: number;
    models?: string;
  }): Promise<ApiToken> {
    const resp = await this.request<ApiToken>('POST', '/api/token/', params);
    return resp.data;
  }

  /** Update an API key */
  async updateToken(tokenId: number, params: Record<string, unknown>): Promise<ApiToken> {
    const resp = await this.request<ApiToken>('PUT', '/api/token/', { id: tokenId, ...params });
    return resp.data;
  }

  /** Delete an API key */
  async deleteToken(tokenId: number): Promise<boolean> {
    const resp = await this.request<null>('DELETE', `/api/token/${tokenId}`);
    return resp.success;
  }

  /** Get a single API key details */
  async getToken(tokenId: number): Promise<ApiToken> {
    const resp = await this.request<ApiToken>('GET', `/api/token/${tokenId}`);
    return resp.data;
  }

  // ==================== Channel Management ====================

  /** List all channels */
  async listChannels(): Promise<Channel[]> {
    const resp = await this.request<Channel[]>('GET', '/api/channel/');
    return resp.data || [];
  }

  /** Get channel details */
  async getChannel(channelId: number): Promise<Channel> {
    const resp = await this.request<Channel>('GET', `/api/channel/${channelId}`);
    return resp.data;
  }

  /** Create a channel */
  async createChannel(params: Record<string, unknown>): Promise<Channel> {
    const resp = await this.request<Channel>('POST', '/api/channel/', params);
    return resp.data;
  }

  /** Update a channel */
  async updateChannel(params: Record<string, unknown>): Promise<Channel> {
    const resp = await this.request<Channel>('PUT', '/api/channel/', params);
    return resp.data;
  }

  /** Delete a channel */
  async deleteChannel(channelId: number): Promise<boolean> {
    const resp = await this.request<null>('DELETE', `/api/channel/${channelId}`);
    return resp.success;
  }

  // ==================== Quantum Resources ====================

  /** List available quantum backends */
  async listQuantumBackends(): Promise<Array<Record<string, unknown>>> {
    const resp = await this.request<Array<Record<string, unknown>>>('GET', '/api/quantum/backends');
    return resp.data || [];
  }

  /** List quantum provider statistics */
  async listQuantumProviders(): Promise<Array<Record<string, unknown>>> {
    const resp = await this.request<Array<Record<string, unknown>>>('GET', '/api/quantum/providers');
    return resp.data || [];
  }

  /** Submit a quantum circuit (QASM) */
  async submitQuantumTask(params: {
    provider: string;
    backend?: string;
    qasm: string;
    shots?: number;
    wait?: boolean;
  }): Promise<Record<string, unknown>> {
    const resp = await this.request<Record<string, unknown>>('POST', '/api/quantum/submit', params);
    return resp.data || {};
  }

  // ==================== Async Tasks ====================

  /** List all user's async tasks */
  async listTasks(): Promise<Array<Record<string, unknown>>> {
    const resp = await this.request<Array<Record<string, unknown>>>('GET', '/api/task/');
    return resp.data || [];
  }

  /** Get task status by ID */
  async getTask(taskId: string): Promise<Record<string, unknown>> {
    const resp = await this.request<Record<string, unknown>>('GET', `/api/task/${taskId}`);
    return resp.data || {};
  }

  /** Cancel a task */
  async cancelTask(taskId: string): Promise<Record<string, unknown>> {
    const resp = await this.request<Record<string, unknown>>('POST', `/api/task/${taskId}/cancel`);
    return resp.data || {};
  }

  /** Delete a task */
  async deleteTask(taskId: string): Promise<boolean> {
    const resp = await this.request<null>('DELETE', `/api/task/${taskId}`);
    return resp.success;
  }

  // ==================== System ====================

  /** Get system status */
  async getStatus(): Promise<Record<string, unknown>> {
    return this.request<Record<string, unknown>>('GET', '/api/status');
  }
}
