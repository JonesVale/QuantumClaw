import * as https from 'https';
import * as http from 'http';

import type {
  ChatCompletionRequest,
  ChatCompletionResponse,
  StreamEvent,
  Balance,
  ChatMessage,
} from './types';

export interface QuantumClawConfig {
  apiKey: string;
  baseURL?: string;
  timeout?: number;
}

/**
 * QuantumClaw API Client for Node.js
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

  private async request<T>(method: string, path: string, body?: unknown): Promise<T> {
    const url = new URL(this.baseURL + path);
    const isHttps = url.protocol === 'https:';
    const lib = isHttps ? https : http;

    return new Promise<T>((resolve, reject) => {
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
          if (res.statusCode !== 200) {
            reject(new Error(`HTTP ${res.statusCode}: ${raw}`));
            return;
          }
          try {
            resolve(JSON.parse(raw) as T);
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

  /** Non-streaming chat completion */
  async chatCompletions(req: ChatCompletionRequest): Promise<ChatCompletionResponse> {
    return this.request<ChatCompletionResponse>('POST', '/v1/chat/completions', { ...req, stream: false });
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

  /** List available models */
  async listModels(): Promise<Array<Record<string, unknown>>> {
    const resp = await this.request<{ success: boolean; data: Array<Record<string, unknown>> }>('GET', '/api/models');
    return resp.data || [];
  }

  /** Check account balance */
  async getBalance(): Promise<Balance | null> {
    const resp = await this.request<{ success: boolean; data: Balance }>('GET', '/api/user/balance');
    return resp.data || null;
  }
}
