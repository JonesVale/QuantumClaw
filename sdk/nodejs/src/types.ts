/** Chat message in a conversation */
export interface ChatMessage {
  role: 'user' | 'assistant' | 'system' | 'tool';
  content: string;
}

/** Chat completion request body */
export interface ChatCompletionRequest {
  model: string;
  messages: ChatMessage[];
  temperature?: number;
  max_tokens?: number;
  stream?: boolean;
  top_p?: number;
  frequency_penalty?: number;
  presence_penalty?: number;
}

/** Chat completion response (non-streaming) */
export interface ChatCompletionResponse {
  id: string;
  object: string;
  created: number;
  model: string;
  choices: ChatCompletionChoice[];
  usage?: UsageInfo;
}

export interface ChatCompletionChoice {
  index: number;
  message: ChatMessage;
  finish_reason: string;
}

export interface UsageInfo {
  prompt_tokens: number;
  completion_tokens: number;
  total_tokens: number;
}

/** SSE stream event */
export interface StreamEvent {
  id: string;
  object: string;
  created: number;
  model: string;
  choices: StreamChoice[];
  usage?: UsageInfo;
}

export interface StreamChoice {
  index: number;
  delta: StreamDelta;
  finish_reason?: string;
}

export interface StreamDelta {
  role?: string;
  content?: string;
}

/** Balance information */
export interface Balance {
  quota: number;
  used_quota: number;
}

/** API Token */
export interface ApiToken {
  id: number;
  user_id: number;
  key: string;
  name: string;
  status: number;
  created_time: number;
  accessed_time: number;
  expired_time: number;
  remain_quota: number;
  unlimited_quota: boolean;
  used_quota: number;
  models: string | null;
  subnet: string | null;
}

/** Channel */
export interface Channel {
  id: number;
  name: string;
  type: number;
  key: string;
  status: number;
  models: string;
}

/** Standard API Response */
export interface ApiResponse<T> {
  success: boolean;
  message: string;
  data: T;
}
