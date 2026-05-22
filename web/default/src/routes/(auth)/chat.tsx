/**
 * QuantumClaw - Dedicated Chat Page
 *
 * Full-featured conversation interface:
 * - Left: Conversation history sidebar (localStorage)
 * - Top: Model selection from /api/model-catalog
 * - Center: Markdown-rendered chat bubbles with copy/timestamps
 * - Right: Parameter panel (temperature, top_p, max_tokens, system prompt)
 * - Bottom: Input + send
 * - Auth check: redirect to /sign-in if not logged in
 */

import { createFileRoute, redirect, useNavigate } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { useState, useRef, useEffect, useCallback, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import ReactMarkdown from 'react-markdown'
import remarkGfm from 'remark-gfm'
import {
  Send, Bot, User, Trash2, Copy, Check, StopCircle, Sparkles, Server,
  AlertCircle, RefreshCw, Settings2, PanelRightOpen, PanelRightClose,
  MessageSquare, Plus, Menu, X, Search as SearchIcon, ChevronDown,
  Wifi, WifiOff, ExternalLink, Zap, Archive, Edit3, GripVertical,
  Globe, Cpu, Database, BookOpen, Loader2
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Separator } from '@/components/ui/separator'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Skeleton } from '@/components/ui/skeleton'
import {
  Select, SelectContent, SelectItem, SelectTrigger, SelectValue,
} from '@/components/ui/select'
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu'
import { Tabs, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/ui/tooltip'
import { useAuthStore, type AuthUser } from '@/stores/auth-store'
import { useConversations } from '@/lib/use-conversations'
import { ConversationSidebar, type Conversation } from '@/components/conversation-sidebar'
import { MessageBubble } from '@/components/message-bubble'
import { ParameterPanel, defaultParams, type ChatParams } from '@/components/parameter-panel'
import type { Message } from '@/components/message-types'
import { getCommonHeaders } from '@/lib/api'
import { cn } from '@/lib/utils'
import { codeToType } from '@/lib/tlanguages'

export const Route = createFileRoute('/(auth)/chat')({
  beforeLoad: async ({ location }) => {
    const { auth } = useAuthStore.getState()
    if (!auth.user) {
      throw redirect({
        to: '/sign-in',
        search: { redirect: location.href },
      })
    }
  },
  component: ChatPage,
})

// ── Types ────────────────────────────────────────────────

interface CatalogItem {
  name: string
  display_name: string
  description: string
  use_case: string
  context_window: number
  input_modalities: string[]
  series: string
  provider: string
  channel_id: number
  channel_name: string
  input_price: number
  output_price: number
  status: number
  group: string
}

type ChatMode = 'remote' | 'local'

// ── API Calls ────────────────────────────────────────────

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
      } catch { /* ignore parse errors */ }
    }
  }
}

async function callRemoteChat(model: string, messages: any[], signal?: AbortSignal) {
  const res = await fetch('/v1/chat/completions', {
    method: 'POST',
    headers: { ...getCommonHeaders(), 'Content-Type': 'application/json' },
    body: JSON.stringify({ model, messages, stream: true, max_tokens: 8192 }),
    signal,
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ message: `HTTP ${res.status}` }))
    throw new Error(err.message || `HTTP ${res.status}`)
  }
  return res
}

async function callOllamaChat(model: string, messages: any[], signal?: AbortSignal) {
  const res = await fetch('/api/ollama/chat', {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ model, messages, stream: true }),
    signal,
  })
  if (!res.ok) {
    const err = await res.json().catch(() => ({ message: `Ollama HTTP ${res.status}` }))
    throw new Error(err.message || `HTTP ${res.status}`)
  }
  return res
}

// ── Ollama stream (non-SSE, NDJSON) ─────────────────────
async function* streamOllamaResponse(response: Response): AsyncGenerator<string> {
  const reader = response.body?.getReader()
  if (!reader) throw new Error('No response body')
  const decoder = new TextDecoder()
  while (true) {
    const { done, value } = await reader.read()
    if (done) break
    const chunk = decoder.decode(value, { stream: true })
    for (const line of chunk.split('\n').filter(l => l.trim())) {
      try {
        const parsed = JSON.parse(line)
        const content = parsed.message?.content || parsed.response || ''
        if (content) yield content
        if (parsed.done) return
      } catch { /* ignore */ }
    }
  }
}

// ── Simple Markdown component (for inline rendering if react-markdown fails) ──
function SimpleMarkdown({ content }: { content: string }) {
  const parts: React.ReactNode[] = []
  let remaining = content
  let idx = 0

  while (remaining.length > 0) {
    // Code block ```...```
    const codeBlockMatch = remaining.match(/^```(\w*)\n([\s\S]*?)```\n?/)
    if (codeBlockMatch) {
      parts.push(
        <pre key={idx++} className="my-2 rounded-lg bg-muted p-3 overflow-x-auto text-xs">
          <code>{codeBlockMatch[2]}</code>
        </pre>
      )
      remaining = remaining.slice(codeBlockMatch[0].length)
      continue
    }

    // Inline code `...`
    const inlineMatch = remaining.match(/^`([^`]+)`/)
    if (inlineMatch) {
      parts.push(
        <code key={idx++} className="rounded bg-muted px-1 py-0.5 text-xs font-mono">
          {inlineMatch[1]}
        </code>
      )
      remaining = remaining.slice(inlineMatch[0].length)
      continue
    }

    // Bold **...**
    const boldMatch = remaining.match(/^\*\*([^*]+)\*\*/)
    if (boldMatch) {
      parts.push(<strong key={idx++}>{boldMatch[1]}</strong>)
      remaining = remaining.slice(boldMatch[0].length)
      continue
    }

    // Link [...](...)
    const linkMatch = remaining.match(/^\[([^\]]+)\]\(([^)]+)\)/)
    if (linkMatch) {
      parts.push(
        <a key={idx++} href={linkMatch[2]} target="_blank" rel="noopener noreferrer" className="text-blue-500 hover:underline">
          {linkMatch[1]}
        </a>
      )
      remaining = remaining.slice(linkMatch[0].length)
      continue
    }

    // Newline
    if (remaining.startsWith('\n')) {
      parts.push(<br key={idx++} />)
      remaining = remaining.slice(1)
      continue
    }

    // Regular character
    parts.push(remaining[0])
    remaining = remaining.slice(1)
  }

  return <>{parts}</>
}

// ── Main Chat Page Component ────────────────────────────

function ChatPage() {
  const { t, i18n } = useTranslation()
  const navigate = useNavigate()
  const { auth } = useAuthStore()
  const isAuthed = !!auth.user

  // State
  const [mode, setMode] = useState<ChatMode>('remote')
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const [streamedContent, setStreamedContent] = useState('')
  const [showParams, setShowParams] = useState(false)
  const [showSidebar, setShowSidebar] = useState(true)
  const [params, setParams] = useState<ChatParams>(defaultParams)
  const [selectedModel, setSelectedModel] = useState('')
  const [copiedIdx, setCopiedIdx] = useState(-1)
  const [ollamaModels, setOllamaModels] = useState<{ name: string; label: string }[]>([])
  const [ollamaConnected, setOllamaConnected] = useState(false)
  const [ollamaChecking, setOllamaChecking] = useState(true)

  // Refs
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLTextAreaElement>(null)
  const abortRef = useRef<AbortController | null>(null)

  // Conversations hook (localStorage)
  const {
    conversations, activeId, activeConversation,
    setActiveId, addConversation, deleteConversation,
    renameConversation, addMessage, updateLastAssistantMessage,
    removeLastAssistantMessage,
  } = useConversations()

  // ── Load model catalog ────────────────────────────────
  const lang = codeToType[i18n.language] || 'English'
  const { data: catalogData, isLoading: catalogLoading } = useQuery({
    queryKey: ['model-catalog', lang],
    queryFn: async () => {
      const r = await fetch(`/api/model-catalog?lang=${encodeURIComponent(lang)}`)
      if (!r.ok) throw new Error('Failed to fetch')
      return r.json()
    },
    staleTime: 60 * 1000,
  })
  const catalog: CatalogItem[] = catalogData?.data || []

  // Available models from catalog for remote mode
  const remoteModelOptions = useMemo(() => {
    return catalog
      .filter(m => m.status === 1)
      .sort((a, b) => a.name.localeCompare(b.name))
  }, [catalog])

  // ── Ollama connection check ───────────────────────────
  useEffect(() => {
    const checkOllama = async () => {
      setOllamaChecking(true)
      try {
        const res = await fetch('/api/ollama/tags', { signal: AbortSignal.timeout(3000) })
        if (res.ok) {
          const data = await res.json()
          const models = (data.models || []).map((m: any) => ({
            name: m.name,
            label: m.name,
          }))
          setOllamaModels(models)
          setOllamaConnected(models.length > 0)
        } else {
          setOllamaConnected(false)
          setOllamaModels([])
        }
      } catch {
        setOllamaConnected(false)
        setOllamaModels([])
      } finally {
        setOllamaChecking(false)
      }
    }
    checkOllama()
  }, [])

  // ── Auto-scroll ───────────────────────────────────────
  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [activeConversation?.messages, streamedContent])

  // ── Auto focus input ──────────────────────────────────
  useEffect(() => {
    inputRef.current?.focus()
  }, [mode, selectedModel])

  // ── Ensure a conversation exists ──────────────────────
  useEffect(() => {
    if (!activeConversation && conversations.length === 0) {
      addConversation(selectedModel || 'auto')
    }
  }, [])

  // ── Copy message ──────────────────────────────────────
  const copyMessage = useCallback((content: string, idx: number) => {
    navigator.clipboard.writeText(content)
    setCopiedIdx(idx)
    setTimeout(() => setCopiedIdx(-1), 2000)
  }, [])

  // ── Clear / new conversation ──────────────────────────
  const handleNewConversation = useCallback(() => {
    const conv = addConversation(selectedModel || 'auto')
    setInput('')
    setStreamedContent('')
  }, [addConversation, selectedModel])

  // ── Stop generation ───────────────────────────────────
  const stopGeneration = useCallback(() => {
    abortRef.current?.abort()
    setLoading(false)
    setStreamedContent('')
  }, [])

  // ── Send message ──────────────────────────────────────
  const handleSend = async () => {
    const trimmed = input.trim()
    if (!trimmed || loading) return

    let conv = activeConversation
    if (!conv) {
      conv = addConversation(selectedModel || 'auto')
    }

    const userMsg: Message = { role: 'user', content: trimmed, timestamp: Date.now() }
    addMessage(userMsg)
    setInput('')
    setLoading(true)
    setStreamedContent('')

    const controller = new AbortController()
    abortRef.current = controller

    try {
      // Build messages array with system prompt
      const systemMsg = params.system_prompt
        ? { role: 'system' as const, content: params.system_prompt }
        : undefined
      const historyMessages = (conv?.messages || []).map(m => ({
        role: m.role,
        content: m.content,
      }))
      const apiMessages = systemMsg
        ? [systemMsg, ...historyMessages, { role: 'user', content: trimmed }]
        : [...historyMessages, { role: 'user', content: trimmed }]

      const currentModel = selectedModel || 'auto'

      let response: Response
      let generator: AsyncGenerator<string>

      if (mode === 'local') {
        response = await callOllamaChat(currentModel, apiMessages, controller.signal)
        generator = streamOllamaResponse(response)
      } else {
        response = await callRemoteChat(currentModel, apiMessages, controller.signal)
        generator = streamResponse(response)
      }

      let fullContent = ''
      for await (const delta of generator) {
        fullContent += delta
        setStreamedContent(fullContent)
      }

      if (fullContent) {
        addMessage({ role: 'assistant', content: fullContent, timestamp: Date.now() })
      }
    } catch (err: any) {
      if (err.name === 'AbortError') return
      const errorMsg = err instanceof Error ? err.message : String(err)
      addMessage({
        role: 'assistant',
        content: `⚠️ **Error**: ${errorMsg}\n\nPlease check your connection or try again.`,
        timestamp: Date.now(),
      })
    } finally {
      setLoading(false)
      setStreamedContent('')
      abortRef.current = null
    }
  }

  // ── Handle Enter key ──────────────────────────────────
  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  // ── Messages for current conversation ─────────────────
  const messages = activeConversation?.messages || []

  return (
    <div className="flex h-screen bg-background overflow-hidden">
      {/* ── Left: Conversation Sidebar ──────────────────── */}
      {showSidebar && (
        <div className="w-64 shrink-0 border-r bg-card flex flex-col">
          {/* Sidebar Header */}
          <div className="flex items-center justify-between p-3 border-b">
            <h3 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
              {t('Conversations')}
            </h3>
            <div className="flex items-center gap-1">
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button variant="ghost" size="icon" className="h-7 w-7" onClick={handleNewConversation}>
                    <Plus className="h-4 w-4" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>{t('New conversation')}</TooltipContent>
              </Tooltip>
              <Tooltip>
                <TooltipTrigger asChild>
                  <Button variant="ghost" size="icon" className="h-7 w-7" onClick={() => setShowSidebar(false)}>
                    <PanelRightClose className="h-4 w-4" />
                  </Button>
                </TooltipTrigger>
                <TooltipContent>{t('Close sidebar')}</TooltipContent>
              </Tooltip>
            </div>
          </div>

          {/* Conversations List */}
          <ScrollArea className="flex-1 p-2">
            {conversations.length === 0 ? (
              <div className="text-center py-8 text-xs text-muted-foreground">
                <MessageSquare className="h-6 w-6 mx-auto mb-2 opacity-30" />
                <p>{t('No conversations yet')}</p>
              </div>
            ) : (
              <div className="space-y-1">
                {conversations.map((conv) => (
                  <ConversationItem
                    key={conv.id}
                    conv={conv}
                    isActive={conv.id === activeId}
                    onSelect={() => setActiveId(conv.id)}
                    onDelete={() => deleteConversation(conv.id)}
                    onRename={(title) => renameConversation(conv.id, title)}
                    t={t}
                  />
                ))}
              </div>
            )}
          </ScrollArea>
        </div>
      )}

      {/* ── Right: Main Chat Area ───────────────────────── */}
      <div className="flex-1 flex flex-col min-w-0">
        {/* ── Top Bar ─────────────────────────────────────── */}
        <div className="flex items-center gap-2 px-4 py-2.5 border-b bg-card/50 shrink-0">
          {/* Toggle sidebar */}
          {!showSidebar && (
            <Tooltip>
              <TooltipTrigger asChild>
                <Button variant="ghost" size="icon" className="h-8 w-8 shrink-0" onClick={() => setShowSidebar(true)}>
                  <PanelRightOpen className="h-4 w-4" />
                </Button>
              </TooltipTrigger>
              <TooltipContent>{t('Show sidebar')}</TooltipContent>
            </Tooltip>
          )}

          {/* Mode Toggle */}
          <Tabs value={mode} onValueChange={v => setMode(v as ChatMode)} className="shrink-0">
            <TabsList className="h-8">
              <TabsTrigger value="remote" className="text-xs px-3 gap-1">
                <Server className="h-3 w-3" /> {t('Remote')}
              </TabsTrigger>
              <TabsTrigger
                value="local"
                className="text-xs px-3 gap-1"
                disabled={!ollamaConnected && !ollamaChecking}
              >
                {ollamaConnected ? <Wifi className="h-3 w-3 text-green-500" /> : <WifiOff className="h-3 w-3 text-muted-foreground" />}
                {t('Local')}
              </TabsTrigger>
            </TabsList>
          </Tabs>

          {/* Model Selector */}
          {mode === 'remote' ? (
            <Select value={selectedModel} onValueChange={setSelectedModel}>
              <SelectTrigger className="h-8 w-[260px] text-xs">
                <SelectValue placeholder={t('Select a model...')} />
              </SelectTrigger>
              <SelectContent className="max-h-[300px]">
                {catalogLoading ? (
                  <div className="p-2 text-xs text-muted-foreground">{t('Loading...')}</div>
                ) : remoteModelOptions.length === 0 ? (
                  <div className="p-2 text-xs text-muted-foreground">{t('No models available')}</div>
                ) : (
                  remoteModelOptions.map((m) => (
                    <SelectItem key={m.name} value={m.name} className="text-xs">
                      <div className="flex items-center gap-2">
                        <span>{m.name}</span>
                        <span className="text-[10px] text-muted-foreground ml-auto">
                          ${(m.input_price || 0).toFixed(6)}/tok
                        </span>
                      </div>
                    </SelectItem>
                  ))
                )}
              </SelectContent>
            </Select>
          ) : (
            <Select value={selectedModel} onValueChange={setSelectedModel}>
              <SelectTrigger className="h-8 w-[200px] text-xs">
                <SelectValue placeholder={t('Select local model...')} />
              </SelectTrigger>
              <SelectContent>
                {ollamaChecking ? (
                  <div className="p-2 text-xs text-muted-foreground">{t('Checking...')}</div>
                ) : ollamaModels.length === 0 ? (
                  <div className="p-2 text-xs text-muted-foreground">{t('No local models found')}</div>
                ) : (
                  ollamaModels.map((m) => (
                    <SelectItem key={m.name} value={m.name} className="text-xs">{m.label}</SelectItem>
                  ))
                )}
              </SelectContent>
            </Select>
          )}

          {/* Spacer */}
          <div className="flex-1" />

          {/* Action buttons */}
          <Tooltip>
            <TooltipTrigger asChild>
              <Button
                variant="ghost"
                size="icon"
                className="h-8 w-8"
                onClick={() => setShowParams(!showParams)}
                data-active={showParams}
              >
                <Settings2 className={cn('h-4 w-4', showParams && 'text-blue-500')} />
              </Button>
            </TooltipTrigger>
            <TooltipContent>{t('Parameters')}</TooltipContent>
          </Tooltip>

          <Tooltip>
            <TooltipTrigger asChild>
              <Button variant="ghost" size="icon" className="h-8 w-8" onClick={handleNewConversation}>
                <Trash2 className="h-4 w-4" />
              </Button>
            </TooltipTrigger>
            <TooltipContent>{t('New conversation')}</TooltipContent>
          </Tooltip>
        </div>

        {/* ── Messages Area ────────────────────────────────── */}
        <div className="flex-1 flex overflow-hidden">
          {/* Messages */}
          <div className="flex-1 flex flex-col min-w-0 overflow-hidden">
            <ScrollArea className="flex-1 px-4 py-4">
              {messages.length === 0 ? (
                <div className="flex flex-col items-center justify-center h-full text-center py-16">
                  <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center mb-4 shadow-lg">
                    <Bot className="h-8 w-8 text-white" />
                  </div>
                  <h2 className="text-xl font-semibold mb-2">{t('QuantumClaw AI Chat')}</h2>
                  <p className="text-sm text-muted-foreground max-w-md mb-6">
                    {mode === 'local'
                      ? t('Chat with your local Ollama models. Select a model to begin.')
                      : t('Select a model from the catalog and start a conversation.')}
                  </p>
                  {mode === 'remote' && remoteModelOptions.length > 0 && (
                    <div className="grid grid-cols-2 sm:grid-cols-3 gap-2 max-w-lg">
                      {remoteModelOptions.slice(0, 6).map((m) => (
                        <Button
                          key={m.name}
                          variant="outline"
                          size="sm"
                          className="text-xs h-auto py-2 px-3"
                          onClick={() => {
                            setSelectedModel(m.name)
                            // Send a welcome message
                            const conv = activeConversation || addConversation(m.name)
                            addMessage({
                              role: 'assistant',
                              content: `👋 Using **${m.name}**. How can I help you today?`,
                              timestamp: Date.now(),
                            })
                          }}
                        >
                          <Zap className="h-3 w-3 mr-1 shrink-0" />
                          <span className="truncate">{m.display_name || m.name}</span>
                        </Button>
                      ))}
                    </div>
                  )}
                </div>
              ) : (
                <div className="max-w-4xl mx-auto space-y-4 pb-4">
                  {messages.map((msg, idx) => (
                    <MessageBubble
                      key={idx}
                      message={msg}
                      isLast={idx === messages.length - 1}
                      streaming={false}
                    />
                  ))}

                  {/* Streaming message */}
                  {streamedContent && (
                    <div className="flex gap-3">
                      <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-gradient-to-br from-blue-500 to-purple-600 text-white shadow-sm">
                        <Bot className="h-4 w-4" />
                      </div>
                      <div className="max-w-[80%] rounded-2xl bg-muted/70 dark:bg-muted/30 px-4 py-2.5 text-sm leading-relaxed border border-border/30">
                        <ReactMarkdown
                          remarkPlugins={[remarkGfm]}
                          components={{
                            pre: ({ children }) => (
                              <pre className="my-2 rounded-lg bg-muted p-3 overflow-x-auto text-xs">{children}</pre>
                            ),
                            code: ({ children, ...props }) => (
                              <code className="rounded bg-muted px-1 py-0.5 text-xs font-mono" {...props}>
                                {children}
                              </code>
                            ),
                            a: ({ href, children }) => (
                              <a href={href} target="_blank" rel="noopener noreferrer" className="text-blue-500 hover:underline">
                                {children}
                              </a>
                            ),
                          }}
                        >
                          {streamedContent}
                        </ReactMarkdown>
                        <span className="mt-1 inline-block h-3 w-2 animate-pulse rounded-full bg-blue-500" />
                      </div>
                    </div>
                  )}

                  {/* Loading indicator */}
                  {loading && !streamedContent && (
                    <div className="flex items-center gap-2 text-sm text-muted-foreground">
                      <div className="flex h-6 w-6 items-center justify-center rounded-full bg-muted">
                        <Bot className="h-3.5 w-3.5" />
                      </div>
                      <span className="animate-pulse">{t('Thinking...')}</span>
                      <Loader2 className="h-3.5 w-3.5 animate-spin" />
                    </div>
                  )}

                  <div ref={messagesEndRef} />
                </div>
              )}
            </ScrollArea>

            {/* ── Input Area ──────────────────────────────────── */}
            <div className="border-t bg-card/30 px-4 py-3 shrink-0">
              <div className="max-w-4xl mx-auto">
                {/* Selected model indicator */}
                {selectedModel && (
                  <div className="flex items-center gap-1.5 mb-2">
                    <Badge variant="outline" className="text-[10px] h-5 gap-1">
                      {mode === 'local' ? <Cpu className="h-3 w-3" /> : <Globe className="h-3 w-3" />}
                      {selectedModel}
                    </Badge>
                    {params.system_prompt && (
                      <Badge variant="secondary" className="text-[10px] h-5">
                        <BookOpen className="h-3 w-3 mr-1" />
                        Sys Prompt
                      </Badge>
                    )}
                  </div>
                )}

                <div className="flex gap-2 items-end">
                  <div className="relative flex-1">
                    <textarea
                      ref={inputRef as any}
                      value={input}
                      onChange={e => setInput(e.target.value)}
                      onKeyDown={handleKeyDown}
                      placeholder={
                        mode === 'remote'
                          ? t('Type a message... (Enter to send, Shift+Enter for new line)')
                          : t('Chat with local model...')
                      }
                      disabled={loading}
                      rows={1}
                      className="w-full resize-none rounded-xl border border-input bg-background px-4 py-3 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50 min-h-[44px] max-h-[120px]"
                      style={{ height: 'auto', minHeight: '44px' }}
                      onInput={e => {
                        const el = e.currentTarget
                        el.style.height = 'auto'
                        el.style.height = Math.min(el.scrollHeight, 120) + 'px'
                      }}
                    />
                  </div>

                  {loading ? (
                    <Button
                      type="button"
                      variant="destructive"
                      size="icon"
                      className="h-11 w-11 shrink-0 rounded-xl"
                      onClick={stopGeneration}
                    >
                      <StopCircle className="h-5 w-5" />
                    </Button>
                  ) : (
                    <Button
                      type="submit"
                      size="icon"
                      className="h-11 w-11 shrink-0 rounded-xl bg-gradient-to-r from-blue-600 to-purple-600 hover:from-blue-700 hover:to-purple-700"
                      disabled={!input.trim() || (!selectedModel && mode === 'remote')}
                      onClick={handleSend}
                    >
                      <Send className="h-5 w-5" />
                    </Button>
                  )}
                </div>

                <div className="flex items-center justify-between mt-1.5">
                  <p className="text-[10px] text-muted-foreground">
                    {mode === 'remote' ? 'Remote API' : 'Local Ollama'} · {selectedModel || t('No model selected')}
                  </p>
                  <p className="text-[10px] text-muted-foreground">
                    Temp: {params.temperature.toFixed(1)} · Top P: {params.top_p.toFixed(1)} · Max: {params.max_tokens}
                  </p>
                </div>
              </div>
            </div>
          </div>

          {/* ── Right: Parameter Panel ────────────────────────── */}
          {showParams && (
            <div className="w-72 shrink-0 border-l bg-card overflow-y-auto">
              <ParameterPanel
                params={params}
                onParamsChange={setParams}
                onClose={() => setShowParams(false)}
              />
            </div>
          )}
        </div>
      </div>
    </div>
  )
}

// ── Conversation Item Component (inline) ─────────────────

function ConversationItem({
  conv, isActive, onSelect, onDelete, onRename, t,
}: {
  conv: Conversation
  isActive: boolean
  onSelect: () => void
  onDelete: () => void
  onRename: (title: string) => void
  t: (key: string) => string
}) {
  const [editing, setEditing] = useState(false)
  const [editValue, setEditValue] = useState('')
  const inputRef = useRef<HTMLInputElement>(null)

  const startEditing = useCallback((e: React.MouseEvent) => {
    e.stopPropagation()
    setEditValue(conv.title)
    setEditing(true)
    setTimeout(() => inputRef.current?.focus(), 50)
  }, [conv.title])

  const saveRename = useCallback(() => {
    onRename(editValue || conv.title)
    setEditing(false)
  }, [editValue, conv.title, onRename])

  return (
    <div
      className={cn(
        'group flex items-center gap-2 rounded-lg px-3 py-2 cursor-pointer text-sm transition-colors',
        isActive
          ? 'bg-accent text-accent-foreground'
          : 'hover:bg-accent/50 text-muted-foreground hover:text-foreground'
      )}
      onClick={onSelect}
    >
      <MessageSquare className="h-4 w-4 shrink-0" />

      {editing ? (
        <form
          onSubmit={e => { e.preventDefault(); e.stopPropagation(); saveRename() }}
          className="flex-1 flex gap-1"
          onClick={e => e.stopPropagation()}
        >
          <Input
            ref={inputRef}
            value={editValue}
            onChange={e => setEditValue(e.target.value)}
            className="h-6 text-xs py-0"
            onBlur={saveRename}
          />
        </form>
      ) : (
        <>
          <span className="flex-1 truncate text-xs">
            {conv.title || t('New conversation')}
          </span>
          <div className="hidden group-hover:flex items-center gap-0.5">
            <Button
              variant="ghost"
              size="icon"
              className="h-6 w-6"
              onClick={startEditing}
            >
              <Edit3 className="h-3 w-3" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="h-6 w-6 text-red-500 hover:text-red-600"
              onClick={e => { e.stopPropagation(); if (confirm(t('Delete this conversation?'))) onDelete() }}
            >
              <Trash2 className="h-3 w-3" />
            </Button>
          </div>
        </>
      )}
    </div>
  )
}
