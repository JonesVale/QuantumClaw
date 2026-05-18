import { createFileRoute } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { useState, useRef, useEffect, useCallback } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
import {
  Send, Bot, User, Trash2, Copy, Check, StopCircle, Sparkles, Server,
  AlertCircle, RefreshCw
} from 'lucide-react'
import { useQuery } from '@tanstack/react-query'
import { getModels, type ModelInfo } from '@/lib/api-extended'
import { getCommonHeaders } from '@/lib/api'

export const Route = createFileRoute('/(auth)/chat')({
  component: ChatPage,
})

interface ChatMessage {
  role: 'user' | 'assistant'
  content: string
  timestamp: number
}

// ── Free Chat（通过后端代理，API Key 不暴露）───────────────
const FREE_API = '/api/free-chat'
type FreeProvider = { name: string; label: string; ready: boolean; models: { id: string; name: string }[] }
const DEFAULT_FREE_PROVIDER = 'groq'
const STORAGE_KEY = 'quantumclaw_chat_history'
const MODE_KEY = 'quantumclaw_chat_mode'
type ChatMode = 'free' | 'user'

async function callFreeChat(provider: string, model: string, messages: any[], signal?: AbortSignal) {
  const res = await fetch(`${FREE_API}/completions`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ provider, model, messages, stream: true }),
    signal,
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ message: `HTTP ${res.status}` }))
    throw new Error(err.message || `HTTP ${res.status}`)
  }
  return res
}

async function callUserAPI(model: string, messages: any[], signal?: AbortSignal) {
  const res = await fetch('/v1/chat/completions', {
    method: 'POST',
    headers: { ...getCommonHeaders(), 'Content-Type': 'application/json' },
    body: JSON.stringify({ model, messages, stream: true, max_tokens: 4096 }),
    signal,
  })
  if (!res.ok) throw new Error(`API HTTP ${res.status}`)
  return res
}

async function* streamResponse(response: Response): AsyncGenerator<string> {
  const reader = response.body?.getReader()
  if (!reader) throw new Error('No response body')
  const decoder = new TextDecoder()
  while (true) {
    const { done, value } = await reader.read()
    if (done) break
    const chunk = decoder.decode(value, { stream: true })
    for (const line of chunk.split('\n').filter(l => l.startsWith('data: '))) {
      const data = line.slice(6)
      if (data === '[DONE]') continue
      try {
        const delta = JSON.parse(data).choices?.[0]?.delta?.content || ''
        if (delta) yield delta
      } catch {}
    }
  }
}

function ChatPage() {
  const { t } = useTranslation()
  const [mode, setMode] = useState<ChatMode>(() => {
    return (localStorage.getItem(MODE_KEY) as ChatMode) || 'free'
  })
  const [messages, setMessages] = useState<ChatMessage[]>(() => {
    try {
      const saved = localStorage.getItem(STORAGE_KEY)
      return saved ? JSON.parse(saved) : [{
        role: 'assistant' as const,
        content: t('chat_welcome_message', '👋 你好！我是 AI 助手。\n\n💡 **免费模式**：使用 Groq 免费 API（无需登录）\n🔑 **我的模型**：使用你配置的 API Key'),
        timestamp: Date.now(),
      }]
    } catch { return [] }
  })
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const [freeProvider, setFreeProvider] = useState(DEFAULT_FREE_PROVIDER)
  const [freeModel, setFreeModel] = useState('llama-3.3-70b-versatile')
  const [freeProviders, setFreeProviders] = useState<FreeProvider[]>([])
  const [freeModels, setFreeModels] = useState<{id:string;name:string}[]>([{id:'llama-3.3-70b-versatile',name:'Llama 3.3 70B'}])
  const [userModel, setUserModel] = useState('')
  const [streamedContent, setStreamedContent] = useState('')
  const [copiedIdx, setCopiedIdx] = useState(-1)
  const [apiStatus, setApiStatus] = useState<'idle' | 'testing' | 'ok' | 'fail'>('idle')
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)
  const abortRef = useRef<AbortController | null>(null)

  // 从 API 加载免费提供商和模型
  useEffect(() => {
    fetch(`${FREE_API}/models`).then(r=>r.json()).then(res => {
      if (res.success && Array.isArray(res.data)) {
        setFreeProviders(res.data)
        const ready = res.data.find((p:FreeProvider) => p.ready)
        if (ready) {
          setFreeProvider(ready.name)
          setFreeModels(ready.models)
          setFreeModel(ready.models[0]?.id || '')
        }
      }
    }).catch(() => {})
  }, [])

  useEffect(() => {
    const p = freeProviders.find(fp => fp.name === freeProvider)
    if (p) { setFreeModels(p.models); setFreeModel(p.models[0]?.id || '') }
  }, [freeProvider, freeProviders])

  const { data: models } = useQuery({
    queryKey: ['models'],
    queryFn: getModels,
    staleTime: 60000,
    enabled: mode === 'user',
  })

  // Auto scroll
  useEffect(() => { messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' }) }, [messages, streamedContent])
  useEffect(() => { inputRef.current?.focus() }, [mode])

  // Persist
  useEffect(() => {
    try { localStorage.setItem(STORAGE_KEY, JSON.stringify(messages)) } catch {}
  }, [messages])
  useEffect(() => { localStorage.setItem(MODE_KEY, mode) }, [mode])

  const copyMessage = useCallback((content: string, idx: number) => {
    navigator.clipboard.writeText(content)
    setCopiedIdx(idx)
    setTimeout(() => setCopiedIdx(-1), 2000)
  }, [])

  const clearChat = () => {
    if (abortRef.current) abortRef.current.abort()
    setMessages([{
      role: 'assistant',
      content: mode === 'free'
        ? '💡 **免费模式**：使用 Groq 免费 API（无需配置）\n🔑 切换到"我的模型"可使用你购买的 API Key'
        : '🔑 **我的模型模式**：使用你配置的 API Key\n💡 切换到"免费模式"可免费体验',
      timestamp: Date.now(),
    }])
    setStreamedContent('')
    localStorage.removeItem(STORAGE_KEY)
  }

  const stopGeneration = () => {
    abortRef.current?.abort()
    setLoading(false)
    setStreamedContent('')
  }

  const testConnection = async () => {
    setApiStatus('testing')
    try {
      const res = await fetch(`${FREE_API}/status?provider=${freeProvider}`, { signal: AbortSignal.timeout(5000) })
      const data = await res.json()
      setApiStatus(data.success ? 'ok' : 'fail')
    } catch { setApiStatus('fail') }
  }

  const handleSend = async () => {
    const trimmed = input.trim()
    if (!trimmed || loading) return

    const userMsg: ChatMessage = { role: 'user', content: trimmed, timestamp: Date.now() }
    const updated = [...messages, userMsg]
    setMessages(updated)
    setInput('')
    setLoading(true)
    setStreamedContent('')

    const controller = new AbortController()
    abortRef.current = controller

    try {
      const apiMessages = [
        { role: 'system', content: 'You are a helpful AI assistant.' },
        ...updated.map(m => ({ role: m.role, content: m.content })),
      ]

      const currentModel = mode === 'free' ? freeModel : userModel
      const caller = mode === 'free'
        ? (m: string, msgs: any[], sig?: AbortSignal) => callFreeChat(freeProvider, m, msgs, sig)
        : callUserAPI
      const response = await caller(currentModel, apiMessages, controller.signal)

      let fullContent = ''
      for await (const delta of streamResponse(response)) {
        fullContent += delta
        setStreamedContent(fullContent)
      }

      setMessages(prev => [...prev, { role: 'assistant', content: fullContent, timestamp: Date.now() }])
    } catch (err: any) {
      if (err.name === 'AbortError') return
      const errorMsg = err instanceof Error ? err.message : String(err)
      // 免费模式失败时自动切换到用户的 API（如果有配置）
      if (mode === 'free' && models && models.length > 0) {
        setMessages(prev => [...prev, {
          role: 'assistant',
          content: `⚠️ 免费 Groq API 暂时不可用，已自动切换到你配置的模型。\n\n错误: ${errorMsg}`,
          timestamp: Date.now(),
        }])
        setMode('user')
      } else {
        setMessages(prev => [...prev, {
          role: 'assistant',
          content: `${t('chat_error_prefix', 'Error')}: ${errorMsg}`,
          timestamp: Date.now(),
        }])
      }
    } finally {
      setLoading(false)
      setStreamedContent('')
      abortRef.current = null
    }
  }

  return (
    <div className="mx-auto flex h-[calc(100vh-4rem)] max-w-4xl flex-col bg-background p-4">
      {/* Header */}
      <div className="mb-3 flex items-center justify-between border-b pb-3">
        <div className="flex items-center gap-3">
          <div className={`flex h-8 w-8 items-center justify-center rounded-lg ${mode === 'free' ? 'bg-green-600' : 'bg-blue-600'}`}>
            {mode === 'free' ? <Sparkles className="h-4 w-4 text-white" /> : <Server className="h-4 w-4 text-white" />}
          </div>
          <div>
            <h1 className="text-lg font-semibold">{t('chat_title', 'AI Chat')}</h1>
            <Badge variant={mode === 'free' ? 'default' : 'secondary'} className="h-5 text-[10px]">
              {mode === 'free' ? 'FREE' : '我的模型'}
            </Badge>
          </div>
        </div>
      </div>

      {/* Mode + Model Selection */}
      <div className="mb-3 flex flex-wrap items-center gap-2">
        <Tabs value={mode} onValueChange={v => setMode(v as ChatMode)} className="w-auto">
          <TabsList className="h-8">
            <TabsTrigger value="free" className="text-xs px-3 gap-1">
              <Sparkles className="h-3 w-3" /> 免费体验
            </TabsTrigger>
            <TabsTrigger value="user" className="text-xs px-3 gap-1">
              <Server className="h-3 w-3" /> 我的模型
            </TabsTrigger>
          </TabsList>
        </Tabs>

        {mode === 'free' ? (
          <>
            {/* 提供商选择 */}
            <Select value={freeProvider} onValueChange={setFreeProvider}>
              <SelectTrigger className="h-8 w-[90px] text-xs">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {freeProviders.map(p => (
                  <SelectItem key={p.name} value={p.name} className="text-xs" disabled={!p.ready}>
                    {p.ready ? '🟢 ' : '🔴 '}{p.name}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
            {/* 模型选择 */}
            <Select value={freeModel} onValueChange={setFreeModel}>
              <SelectTrigger className="h-8 w-[200px] text-xs">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {freeModels.map(m => (
                  <SelectItem key={m.id} value={m.id} className="text-xs">{m.name}</SelectItem>
                ))}
              </SelectContent>
            </Select>
            <Button variant="ghost" size="sm" className="h-8 text-xs gap-1" onClick={testConnection}>
              <RefreshCw className={`h-3 w-3 ${apiStatus === 'testing' ? 'animate-spin' : ''}`} />
              {apiStatus === 'idle' && '测试连接'}
              {apiStatus === 'testing' && '测试中...'}
              {apiStatus === 'ok' && '✅ 正常'}
              {apiStatus === 'fail' && '❌ 失败'}
            </Button>
          </>
        ) : (
          <Select value={userModel} onValueChange={setUserModel}>
            <SelectTrigger className="h-8 w-[220px] text-xs">
              <SelectValue placeholder="选择模型..." />
            </SelectTrigger>
            <SelectContent>
              {models?.map((m: any) => (
                <SelectItem key={m.id || m.name} value={m.name} className="text-xs">{m.name}</SelectItem>
              ))}
            </SelectContent>
          </Select>
        )}

        <div className="flex-1" />
        <Button variant="ghost" size="icon" onClick={clearChat} title="清除对话">
          <Trash2 className="h-4 w-4" />
        </Button>
      </div>

      {/* 免费模式提示 */}
      {mode === 'free' && (
        <div className="mb-3 rounded-lg bg-green-50 dark:bg-green-950/30 border border-green-200 dark:border-green-800 p-3 text-xs text-green-700 dark:text-green-400">
          <AlertCircle className="inline h-3 w-3 mr-1" />
          当前使用 <strong>{freeProviders.find(p=>p.name===freeProvider)?.label || freeProvider}</strong> 免费 API。
          可切换 Groq / DeepSeek 提供商。如需更高额度，切换到「我的模型」。
        </div>
      )}

      {/* Messages */}
      <div className="flex-1 space-y-4 overflow-y-auto px-2 pb-4">
        {messages.map((msg, idx) => (
          <div key={idx} className={`group flex gap-3 ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}>
            {msg.role === 'assistant' && (
              <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-blue-100 dark:bg-blue-900">
                <Bot className="h-4 w-4 text-blue-600 dark:text-blue-400" />
              </div>
            )}
            <div className={`relative max-w-[80%] rounded-2xl px-4 py-2.5 text-sm ${
              msg.role === 'user'
                ? 'bg-blue-600 text-white'
                : 'bg-muted text-foreground'
            }`}>
              <p className="whitespace-pre-wrap">{msg.content}</p>
              <span className="mt-1 block text-[10px] opacity-50">
                {new Date(msg.timestamp).toLocaleTimeString()}
              </span>
              <button
                onClick={() => copyMessage(msg.content, idx)}
                className="absolute -right-8 top-2 hidden rounded p-1 hover:bg-muted group-hover:block"
                title="复制"
              >
                {copiedIdx === idx ? <Check className="h-3.5 w-3.5 text-green-500" /> : <Copy className="h-3.5 w-3.5" />}
              </button>
            </div>
            {msg.role === 'user' && (
              <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-blue-600">
                <User className="h-4 w-4 text-white" />
              </div>
            )}
          </div>
        ))}

        {streamedContent && (
          <div className="flex gap-3">
            <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-blue-100 dark:bg-blue-900">
              <Bot className="h-4 w-4 text-blue-600 dark:text-blue-400" />
            </div>
            <div className="max-w-[80%] rounded-2xl bg-muted px-4 py-2.5 text-sm text-foreground">
              <p className="whitespace-pre-wrap">{streamedContent}</p>
              <span className="mt-1 inline-block h-3 w-2 animate-pulse rounded-full bg-blue-500" />
            </div>
          </div>
        )}

        {loading && !streamedContent && (
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <div className="flex h-6 w-6 items-center justify-center rounded-full bg-muted">
              <Bot className="h-3.5 w-3.5" />
            </div>
            <span className="animate-pulse">思考中...</span>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* Input */}
      <div className="border-t pt-3">
        <form onSubmit={e => { e.preventDefault(); handleSend() }} className="flex gap-2">
          <Input
            ref={inputRef}
            value={input}
            onChange={e => setInput(e.target.value)}
            placeholder={mode === 'free' ? '免费模式 · 输入问题...' : '输入问题...'}
            disabled={loading}
            className="flex-1"
          />
          {loading ? (
            <Button type="button" variant="destructive" onClick={stopGeneration}>
              <StopCircle className="h-4 w-4" />
            </Button>
          ) : (
            <Button type="submit" disabled={!input.trim()}>
              <Send className="h-4 w-4" />
            </Button>
          )}
        </form>
        <p className="mt-1 text-[10px] text-muted-foreground">
          {mode === 'free' ? freeProvider : '我的模型'} · {mode === 'free' ? freeModel : (userModel || '未选择')} · Enter 发送
        </p>
      </div>
    </div>
  )
}
