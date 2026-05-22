import { createFileRoute } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { useState, useRef, useEffect, useCallback } from 'react'
import { Send, Loader2, Trash2, Settings2, MessageSquarePlus, PanelRightOpen, PanelRightClose, ChevronLeft, ChevronRight } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card } from '@/components/ui/card'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { getModels, type ModelInfo } from '@/lib/api-extended'
import { getCommonHeaders } from '@/lib/api'
import { useQuery } from '@tanstack/react-query'
import { cn } from '@/lib/utils'
import { MessageBubble } from '@/components/message-bubble'
import { ConversationSidebar } from '@/components/conversation-sidebar'
import { ParameterPanel, defaultParams, type ChatParams } from '@/components/parameter-panel'
import { useConversations } from '@/lib/use-conversations'
import type { Message } from '@/components/message-types'

export const Route = createFileRoute('/_authenticated/playground')({
  component: PlaygroundPage,
})

function PlaygroundPage() {
  const { t } = useTranslation()
  const scrollRef = useRef<HTMLDivElement>(null)
  const [input, setInput] = useState('')
  const [model, setModel] = useState('auto')
  const [streaming, setStreaming] = useState(false)
  const [showParams, setShowParams] = useState(false)
  const [showSidebar, setShowSidebar] = useState(true)
  const [params, setParams] = useState<ChatParams>(defaultParams)
  const abortRef = useRef<AbortController | null>(null)

  const {
    conversations,
    activeId,
    activeConversation,
    addConversation,
    deleteConversation,
    renameConversation,
    addMessage,
    updateLastAssistantMessage,
    removeLastAssistantMessage,
    setActiveId,
  } = useConversations()

  // Collection of task states for non-chat modes
  type Mode = 'chat' | 'midjourney' | 'suno' | 'video'
  const [mode, setMode] = useState<Mode>('chat')

  interface TaskState {
    taskId: string | null
    status: string
    result: any
    loading: boolean
  }
  const emptyTask = (): TaskState => ({ taskId: null, status: '', result: null, loading: false })
  const [mjTask, setMjTask] = useState<TaskState>(emptyTask)
  const [sunoTask, setSunoTask] = useState<TaskState>(emptyTask)
  const [videoTask, setVideoTask] = useState<TaskState>(emptyTask)
  const [mjPrompt, setMjPrompt] = useState('')
  const [mjAspect, setMjAspect] = useState('1:1')
  const [sunoPrompt, setSunoPrompt] = useState('')
  const [sunoNegative, setSunoNegative] = useState('')
  const [sunoDuration, setSunoDuration] = useState('30')
  const [sunoContinueAt, setSunoContinueAt] = useState('')
  const [videoPrompt, setVideoPrompt] = useState('')

  // Fetch models
  const { data: modelsData } = useQuery({
    queryKey: ['models'],
    queryFn: getModels,
    staleTime: 60 * 1000,
  })
  const models: ModelInfo[] = modelsData?.data || []
  const modelNames = [...new Set(models.map((m) => m.name))]
  const resolvedModel = model === 'auto' ? (modelNames[0] || 'gpt-3.5-turbo') : model

  // Price info for selected model
  const selectedModelInfo = models.find((m) => m.name === resolvedModel)
  const priceInfo = selectedModelInfo
    ? `$${(selectedModelInfo as any).input_price?.toFixed(6) || '?'} in / $${(selectedModelInfo as any).output_price?.toFixed(6) || '?'} out`
    : ''

  const messages: Message[] = activeConversation?.messages || []

  // Create default conversation on first mount
  useEffect(() => {
    if (conversations.length === 0) {
      addConversation('auto')
    }
  }, [])

  useEffect(() => {
    scrollRef.current?.scrollTo({ top: scrollRef.current.scrollHeight, behavior: 'smooth' })
  }, [messages])

  const handleNewConversation = () => {
    addConversation(model)
  }

  const handleSend = useCallback(async () => {
    if (!input.trim() || streaming) return

    const userMsg: Message = { role: 'user', content: input.trim(), timestamp: Date.now() }
    const assistantMsg: Message = { role: 'assistant', content: '', timestamp: Date.now() }

    addMessage(userMsg)
    setInput('')
    setStreaming(true)

    // Add empty assistant message placeholder
    addMessage(assistantMsg)

    try {
      abortRef.current = new AbortController()

      const chatMessages: { role: string; content: string }[] = []
      // Add system prompt if set
      if (params.system_prompt.trim()) {
        chatMessages.push({ role: 'system', content: params.system_prompt.trim() })
      }
      // Add existing messages (excluding the last empty assistant)
      const currentConv = activeConversation
      if (currentConv) {
        const existingMsgs = currentConv.messages.slice(0, -1) // exclude empty assistant
        chatMessages.push(...existingMsgs.map((m) => ({ role: m.role, content: m.content })))
      }
      chatMessages.push({ role: 'user', content: input.trim() })

      const headers = getCommonHeaders()
      const res = await fetch('/v1/chat/completions', {
        method: 'POST',
        headers,
        body: JSON.stringify({
          model: resolvedModel,
          messages: chatMessages,
          stream: true,
          temperature: params.temperature,
          top_p: params.top_p,
          max_tokens: params.max_tokens,
          frequency_penalty: params.frequency_penalty,
          presence_penalty: params.presence_penalty,
        }),
        signal: abortRef.current.signal,
      })

      if (!res.ok) {
        const err = await res.text()
        updateLastAssistantMessage(`Error: ${res.status} - ${err}`)
        return
      }

      const reader = res.body?.getReader()
      const decoder = new TextDecoder()
      let buffer = ''
      let fullContent = ''

      while (reader) {
        const { done, value } = await reader.read()
        if (done) break
        buffer += decoder.decode(value, { stream: true })

        const lines = buffer.split('\n')
        buffer = lines.pop() || ''

        for (const line of lines) {
          const trimmed = line.trim()
          if (!trimmed || !trimmed.startsWith('data: ')) continue
          const data = trimmed.slice(6)
          if (data === '[DONE]') break
          try {
            const parsed = JSON.parse(data)
            const content = parsed.choices?.[0]?.delta?.content
            const usage = parsed.usage
            if (content) {
              fullContent += content
              updateLastAssistantMessage(fullContent)
            }
            if (usage && parsed.choices?.[0]?.finish_reason === 'stop') {
              // Could track token usage here
            }
          } catch { /* skip malformed SSE */ }
        }
      }
    } catch (err: any) {
      if (err.name !== 'AbortError') {
        updateLastAssistantMessage(`Error: ${err.message}`)
      }
    } finally {
      setStreaming(false)
      abortRef.current = null
    }
  }, [input, streaming, resolvedModel, params, activeConversation, addMessage, updateLastAssistantMessage])

  const handleStop = () => {
    abortRef.current?.abort()
    setStreaming(false)
  }

  const handleClear = () => {
    if (activeId) {
      renameConversation(activeId, '')
      activeConversation && updateConversation(activeId, { messages: [] })
    }
  }

  const handleRegenerate = () => {
    if (activeConversation && activeConversation.messages.length >= 2) {
      removeLastAssistantMessage()
      // Re-send with the same user message
      const lastUserMsg = [...activeConversation.messages].reverse().find((m) => m.role === 'user')
      if (lastUserMsg) {
        setInput(lastUserMsg.content)
        // Small delay to allow state to settle, then send
        setTimeout(() => handleSend(), 100)
      }
    }
  }

  // Task handlers (unchanged from original)
  const submitTask = async (mode: string, payload: any, setTask: (t: TaskState) => void) => {
    setTask({ taskId: null, status: 'submitting', result: null, loading: true })
    try {
      const headers = getCommonHeaders()
      const res = await fetch(`/v1/${mode}/submit`, {
        method: 'POST', headers: { ...headers, 'Content-Type': 'application/json' }, body: JSON.stringify(payload),
      })
      const data = await res.json()
      if (data.success && data.data?.task_id) {
        setTask({ taskId: data.data.task_id, status: 'queued', result: null, loading: true })
        pollTask(data.data.task_id, setTask)
      } else {
        setTask({ taskId: null, status: 'failed', result: data, loading: false })
      }
    } catch (err: any) {
      setTask({ taskId: null, status: 'failed', result: { error: err.message }, loading: false })
    }
  }

  const pollTask = async (taskId: string, setTask: (t: TaskState) => void) => {
    const interval = setInterval(async () => {
      try {
        const headers = getCommonHeaders()
        const res = await fetch(`/v1/task/${taskId}/fetch`, { headers })
        const data = await res.json()
        if (data.data?.status === 'success') {
          clearInterval(interval)
          setTask({ taskId, status: 'completed', result: data.data, loading: false })
        } else if (data.data?.status === 'failed') {
          clearInterval(interval)
          setTask({ taskId, status: 'failed', result: data.data, loading: false })
        } else {
          setTask({ taskId, status: data.data?.status || 'running', result: null, loading: true })
        }
      } catch {
        clearInterval(interval)
        setTask({ taskId, status: 'failed', result: { error: 'Poll failed' }, loading: false })
      }
    }, 2000)
  }

  const handleSubmitMj = () => { if (!mjPrompt.trim()) return; submitTask('midjourney', { prompt: mjPrompt, aspect_ratio: mjAspect }, setMjTask) }
  const handleSubmitSuno = () => { if (!sunoPrompt.trim()) return; submitTask('suno', { prompt: sunoPrompt, negative_prompt: sunoNegative, duration: sunoDuration }, setSunoTask) }
  const handleSubmitVideo = () => { if (!videoPrompt.trim()) return; submitTask('video', { prompt: videoPrompt }, setVideoTask) }

  return (
    <div className="w-full h-full flex flex-col overflow-hidden bg-gradient-to-br from-slate-50 via-white to-blue-50/50 dark:from-slate-950 dark:via-slate-900 dark:to-blue-950/50">
      {/* Top toolbar */}
      <div className="flex items-center justify-between gap-2 px-3 py-2 border-b bg-card/50 backdrop-blur-sm shrink-0">
        <div className="flex items-center gap-2">
          <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => setShowSidebar(!showSidebar)} title={t('Toggle sidebar')}>
            {showSidebar ? <ChevronLeft className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
          </Button>
          <h1 className="text-lg font-bold tracking-tight bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-transparent">
            {t('Playground')}
          </h1>
        </div>

        <div className="flex items-center gap-2 flex-1 max-w-md ml-auto mr-auto">
          <Select value={model} onValueChange={setModel}>
            <SelectTrigger className="w-full">
              <SelectValue placeholder={t('Select model')} />
            </SelectTrigger>
            <SelectContent>
              <SelectItem value="auto" className="font-semibold text-blue-600 dark:text-blue-400">
                🤖 {t('Auto')}
              </SelectItem>
              <div className="px-2 py-1 text-xs text-muted-foreground border-t border-b">{t('Available Models')}</div>
              {modelNames.length > 0 ? (
                modelNames.map((name) => (
                  <SelectItem key={name} value={name}>
                    {name}
                  </SelectItem>
                ))
              ) : (
                <div className="px-3 py-6 text-center text-xs text-muted-foreground">
                  <p>{t('No available models')}</p>
                  <a href="/channels" className="text-blue-500 hover:underline mt-1 inline-block">{t('Configure channels')}</a>
                </div>
              )}
            </SelectContent>
          </Select>
        </div>

        <div className="flex items-center gap-1">
          {priceInfo && (
            <span className="text-[10px] text-muted-foreground hidden sm:block mr-1">{priceInfo}</span>
          )}
          <Button
            variant={showParams ? 'default' : 'ghost'}
            size="icon"
            className="h-8 w-8"
            onClick={() => setShowParams(!showParams)}
            title={t('Parameters')}
          >
            <Settings2 className="h-4 w-4" />
          </Button>
          <Button variant="ghost" size="icon" className="h-8 w-8" onClick={handleNewConversation} title={t('New conversation')}>
            <MessageSquarePlus className="h-4 w-4" />
          </Button>
        </div>
      </div>

      {/* Main content area */}
      <div className="flex flex-1 overflow-hidden">
        {/* Conversation sidebar */}
        {showSidebar && (
          <ConversationSidebar
            conversations={conversations}
            activeId={activeId}
            onSelect={setActiveId}
            onNew={handleNewConversation}
            onDelete={deleteConversation}
            onRename={renameConversation}
            onClose={() => setShowSidebar(false)}
          />
        )}

        {/* Chat area */}
        <div className="flex-1 flex flex-col overflow-hidden">
          <Tabs value={mode} onValueChange={(v) => setMode(v as Mode)} className="flex-1 flex flex-col overflow-hidden">
            <div className="px-3 pt-1 shrink-0">
              <TabsList className="h-8">
                <TabsTrigger value="chat" className="text-xs px-3">{t('Chat')}</TabsTrigger>
                <TabsTrigger value="midjourney" className="text-xs px-3">{t('Midjourney')}</TabsTrigger>
                <TabsTrigger value="suno" className="text-xs px-3">{t('Suno')}</TabsTrigger>
                <TabsTrigger value="video" className="text-xs px-3">{t('Video')}</TabsTrigger>
              </TabsList>
            </div>

            {/* === Chat Mode === */}
            <TabsContent value="chat" className="flex-1 flex flex-col mt-0 overflow-hidden">
              <div className="flex-1 flex overflow-hidden">
                {/* Messages */}
                <div className="flex-1 flex flex-col overflow-hidden">
                  <ScrollArea className="flex-1 p-4" ref={scrollRef as any}>
                    {messages.length === 0 ? (
                      <div className="flex h-full flex-col items-center justify-center text-muted-foreground gap-3 py-20">
                        <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-blue-500 to-purple-600 flex items-center justify-center shadow-lg">
                          <MessageSquarePlus className="h-8 w-8 text-white" />
                        </div>
                        <p className="text-sm font-medium">{t('Start a conversation')}</p>
                        <p className="text-xs">{t('Select a model and send a message')}</p>
                      </div>
                    ) : (
                      <div className="space-y-4 max-w-4xl mx-auto">
                        {messages.map((msg, idx) => (
                          <MessageBubble
                            key={idx}
                            message={msg}
                            isLast={idx === messages.length - 1}
                            onRegenerate={handleRegenerate}
                            streaming={streaming && idx === messages.length - 1}
                          />
                        ))}
                      </div>
                    )}
                  </ScrollArea>

                  {/* Input area */}
                  <div className="border-t bg-card/50 backdrop-blur-sm p-3 shrink-0">
                    <form
                      onSubmit={(e) => { e.preventDefault(); handleSend() }}
                      className="flex gap-2 max-w-4xl mx-auto"
                    >
                      <Input
                        value={input}
                        onChange={(e) => setInput(e.target.value)}
                        placeholder={t('Type a message...')}
                        disabled={streaming}
                        className="flex-1 h-10"
                      />
                      {streaming ? (
                        <Button type="button" variant="destructive" onClick={handleStop} className="h-10">
                          <Loader2 className="h-4 w-4 animate-spin mr-1" />
                          {t('Stop')}
                        </Button>
                      ) : (
                        <Button type="submit" disabled={!input.trim()} className="h-10">
                          <Send className="h-4 w-4" />
                        </Button>
                      )}
                    </form>
                  </div>
                </div>

                {/* Parameter panel */}
                {showParams && (
                  <ParameterPanel
                    params={params}
                    onParamsChange={setParams}
                    onClose={() => setShowParams(false)}
                  />
                )}
              </div>
            </TabsContent>

            {/* === Midjourney Mode === */}
            <TabsContent value="midjourney" className="flex-1 mt-0 overflow-y-auto p-4">
              <Card className="max-w-2xl mx-auto p-6 space-y-4">
                <h2 className="text-lg font-semibold">{t('Midjourney')}</h2>
                <div className="space-y-2">
                  <label className="text-sm font-medium">{t('Prompt')}</label>
                  <textarea
                    value={mjPrompt}
                    onChange={(e) => setMjPrompt(e.target.value)}
                    placeholder={t('Describe the image you want to generate...')}
                    className="flex min-h-[100px] w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                  />
                </div>
                <div className="flex items-center gap-4">
                  <div className="space-y-1">
                    <label className="text-sm font-medium">{t('Aspect Ratio')}</label>
                    <Select value={mjAspect} onValueChange={setMjAspect}>
                      <SelectTrigger className="w-28"><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="1:1">1:1</SelectItem>
                        <SelectItem value="16:9">16:9</SelectItem>
                        <SelectItem value="4:3">4:3</SelectItem>
                        <SelectItem value="3:2">3:2</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </div>
                {taskStatusDisplay(mjTask, t)}
                <Button onClick={handleSubmitMj} disabled={!mjPrompt.trim() || mjTask.loading}>
                  {mjTask.loading ? <><Loader2 className="mr-2 h-4 w-4 animate-spin" />{t('Generating...')}</> : t('Generate')}
                </Button>
              </Card>
            </TabsContent>

            {/* === Suno Mode === */}
            <TabsContent value="suno" className="flex-1 mt-0 overflow-y-auto p-4">
              <Card className="max-w-2xl mx-auto p-6 space-y-4">
                <h2 className="text-lg font-semibold">{t('Suno')}</h2>
                <div className="space-y-2">
                  <label className="text-sm font-medium">{t('Prompt')}</label>
                  <textarea value={sunoPrompt} onChange={(e) => setSunoPrompt(e.target.value)} placeholder={t('Describe the music...')} className="flex min-h-[80px] w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring" />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">{t('Negative Prompt (optional)')}</label>
                  <textarea value={sunoNegative} onChange={(e) => setSunoNegative(e.target.value)} placeholder={t('What to avoid...')} className="flex min-h-[60px] w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring" />
                </div>
                <div className="flex items-center gap-4">
                  <div className="space-y-1">
                    <label className="text-sm font-medium">{t('Duration')}</label>
                    <Select value={sunoDuration} onValueChange={setSunoDuration}>
                      <SelectTrigger className="w-28"><SelectValue /></SelectTrigger>
                      <SelectContent>
                        <SelectItem value="15">{t('15s')}</SelectItem>
                        <SelectItem value="30">{t('30s')}</SelectItem>
                        <SelectItem value="60">{t('60s')}</SelectItem>
                      </SelectContent>
                    </Select>
                  </div>
                </div>
                {taskStatusDisplay(sunoTask, t)}
                <Button onClick={handleSubmitSuno} disabled={!sunoPrompt.trim() || sunoTask.loading}>
                  {sunoTask.loading ? <><Loader2 className="mr-2 h-4 w-4 animate-spin" />{t('Generating...')}</> : t('Generate')}
                </Button>
              </Card>
            </TabsContent>

            {/* === Video Mode === */}
            <TabsContent value="video" className="flex-1 mt-0 overflow-y-auto p-4">
              <Card className="max-w-2xl mx-auto p-6 space-y-4">
                <h2 className="text-lg font-semibold">{t('Video Generation')}</h2>
                <div className="space-y-2">
                  <label className="text-sm font-medium">{t('Prompt')}</label>
                  <textarea value={videoPrompt} onChange={(e) => setVideoPrompt(e.target.value)} placeholder={t('Describe the video...')} className="flex min-h-[100px] w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring" />
                </div>
                {taskStatusDisplay(videoTask, t)}
                <Button onClick={handleSubmitVideo} disabled={!videoPrompt.trim() || videoTask.loading}>
                  {videoTask.loading ? <><Loader2 className="mr-2 h-4 w-4 animate-spin" />{t('Generating...')}</> : t('Generate')}
                </Button>
              </Card>
            </TabsContent>
          </Tabs>
        </div>
      </div>
    </div>
  )
}

function taskStatusDisplay(task: { status: string; result: any; loading: boolean }, t: (key: string) => string) {
  if (!task.status) return null
  return (
    <div className="rounded-lg border p-4 space-y-2">
      <div className="flex items-center gap-2 text-sm">
        <span className="font-medium">{t('Status')}:</span>
        <span className={cn(
          'capitalize',
          task.status === 'completed' && 'text-green-500',
          task.status === 'failed' && 'text-red-500',
          task.status === 'submitting' && 'text-yellow-500'
        )}>
          {task.status === 'submitting' ? <><Loader2 className="inline h-3 w-3 animate-spin mr-1" />{t('Submitting')}</>
          : task.status === 'queued' || task.status === 'running' ? <><Loader2 className="inline h-3 w-3 animate-spin mr-1" />{t('Running')}</>
          : t(task.status.charAt(0).toUpperCase() + task.status.slice(1))}
        </span>
      </div>
      {task.result?.file_url && (
        <div className="mt-2">
          <img src={task.result.file_url} alt="" className="max-w-full rounded-lg border" />
          <a href={task.result.file_url} target="_blank" rel="noopener noreferrer" className="text-sm text-blue-500 hover:underline mt-1 inline-block">{t('View full image')} ↗</a>
        </div>
      )}
      {task.status === 'failed' && task.result && (
        <p className="text-sm text-red-500">{task.result.error || task.result.reason || JSON.stringify(task.result)}</p>
      )}
    </div>
  )
}
