/** Chat message in a conversation */
export interface ChatMessage {
  role: 'user' | 'assistant' | 'system';
  content: string;
}

/** Chat completion request body */
export interface ChatCompletionRequest {
  model: string;
  messages: ChatMessage[];
  temperature?: number;
  max_tokens?: number;
  stream?: boolean;
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
