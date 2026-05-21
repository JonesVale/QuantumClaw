import { createFileRoute } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { useState, useRef, useEffect } from 'react'
import { Send, Bot, User, Loader2, Trash2, Settings2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Switch } from '@/components/ui/switch'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { getModels, type ModelInfo } from '@/lib/api-extended'
import { getCommonHeaders } from '@/lib/api'
import { useQuery } from '@tanstack/react-query'
import { cn } from '@/lib/utils'

export const Route = createFileRoute('/_authenticated/playground')({
  component: PlaygroundPage,
})

interface Message {
  role: 'user' | 'assistant' | 'system'
  content: string
  timestamp: number
}

function PlaygroundPage() {
  const { t } = useTranslation()
  const scrollRef = useRef<HTMLDivElement>(null)
  const [messages, setMessages] = useState<Message[]>([])
  const [input, setInput] = useState('')
  const [model, setModel] = useState('auto')
  const [streaming, setStreaming] = useState(false)
  const abortRef = useRef<AbortController | null>(null)

  type Mode = 'chat' | 'midjourney' | 'suno' | 'video'
  const [mode, setMode] = useState<Mode>('chat')

  interface TaskState {
    taskId: string | null
    status: string
    result: any
    loading: boolean
  }

  const emptyTask = (): TaskState => ({
    taskId: null,
    status: '',
    result: null,
    loading: false,
  })

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

  const submitTask = async (mode: string, payload: any, setTask: (t: TaskState) => void) => {
    setTask({ taskId: null, status: 'submitting', result: null, loading: true })
    try {
      const headers = getCommonHeaders()
      const res = await fetch(`/v1/${mode}/submit`, {
        method: 'POST',
        headers: { ...headers, 'Content-Type': 'application/json' },
        body: JSON.stringify(payload),
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
        const res = await fetch(`/v1/task/${taskId}/fetch`, {
          headers,
        })
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

  const handleSubmitMj = () => {
    if (!mjPrompt.trim()) return
    submitTask('midjourney', { prompt: mjPrompt, aspect_ratio: mjAspect }, setMjTask)
  }

  const handleSubmitSuno = () => {
    if (!sunoPrompt.trim()) return
    submitTask('suno', { prompt: sunoPrompt, negative_prompt: sunoNegative, duration: sunoDuration }, setSunoTask)
  }

  const handleSubmitVideo = () => {
    if (!videoPrompt.trim()) return
    submitTask('video', { prompt: videoPrompt }, setVideoTask)
  }

  const { data: modelsData } = useQuery({
    queryKey: ['models'],
    queryFn: getModels,
    staleTime: 60 * 1000,
  })

  const models: ModelInfo[] = modelsData?.data || []
  const modelNames = [...new Set(models.map((m) => m.name))]

  const resolvedModel = model === 'auto' ? (modelNames[0] || 'gpt-3.5-turbo') : model

  useEffect(() => {
    if (model === 'auto' && modelNames.length > 0) {
      // auto mode selected, first model will be used
    }
  }, [modelNames, model])

  useEffect(() => {
    scrollRef.current?.scrollTo({ top: scrollRef.current.scrollHeight })
  }, [messages])

  const handleSend = async () => {
    if (!input.trim() || streaming) return

    const userMsg: Message = {
      role: 'user',
      content: input.trim(),
      timestamp: Date.now(),
    }
    setMessages((prev) => [...prev, userMsg])
    setInput('')
    setStreaming(true)

    const assistantMsg: Message = {
      role: 'assistant',
      content: '',
      timestamp: Date.now(),
    }
    setMessages((prev) => [...prev, assistantMsg])

    try {
      abortRef.current = new AbortController()
      const headers = getCommonHeaders()
      const res = await fetch('/v1/chat/completions', {
        method: 'POST',
        headers,
        body: JSON.stringify({
          model: resolvedModel,
          messages: [...messages, userMsg].map((m) => ({
            role: m.role,
            content: m.content,
          })),
          stream: true,
        }),
        signal: abortRef.current.signal,
      })

      if (!res.ok) {
        const err = await res.text()
        setMessages((prev) => [
          ...prev.slice(0, -1),
          { ...assistantMsg, content: `Error: ${res.status} - ${err}` },
        ])
        return
      }

      const reader = res.body?.getReader()
      const decoder = new TextDecoder()
      let buffer = ''

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
            if (content) {
              setMessages((prev) => {
                const last = prev[prev.length - 1]
                if (last.role === 'assistant') {
                  return [
                    ...prev.slice(0, -1),
                    { ...last, content: last.content + content },
                  ]
                }
                return prev
              })
            }
          } catch {
            /* skip malformed SSE */
          }
        }
      }
    } catch (err: any) {
      if (err.name !== 'AbortError') {
        setMessages((prev) => [
          ...prev.slice(0, -1),
          { ...assistantMsg, content: `Error: ${err.message}` },
        ])
      }
    } finally {
      setStreaming(false)
      abortRef.current = null
    }
  }

  const handleStop = () => {
    abortRef.current?.abort()
    setStreaming(false)
  }

  const handleClear = () => {
    setMessages([])
  }

  return (
    <div className=" w-full flex min-h-[600px] flex-col gap-4 p-4 sm:p-6">
      {/* Toolbar */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
        <h1 className="text-4xl font-bold tracking-tight bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-transparent">
          {t('Playground')}
        </h1>
        <Select value={model} onValueChange={setModel}>
          <SelectTrigger className="w-full sm:w-auto">
            <SelectValue placeholder={t('Select model')} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="auto" className="font-semibold text-blue-600 dark:text-blue-400">
              🤖 {t('Auto')} - {modelNames.length > 0 ? modelNames[0] : t('No models')}
            </SelectItem>
            <div className="px-2 py-1 text-xs text-muted-foreground border-t border-b">{t('Available Models')}</div>
            {modelNames.map((name) => (
              <SelectItem key={name} value={name}>
                {name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
        <div className="ml-auto flex gap-2">
          <Button variant="outline" size="sm" onClick={handleClear} disabled={messages.length === 0}>
            <Trash2 className="mr-1 h-4 w-4" />
            {t('Clear')}
          </Button>
        </div>
      </div>

      <Tabs value={mode} onValueChange={(v) => setMode(v as Mode)} className="flex-1 flex flex-col">
        <TabsList className="w-fit mb-4">
          <TabsTrigger value="chat">{t('Chat')}</TabsTrigger>
          <TabsTrigger value="midjourney">{t('Midjourney')}</TabsTrigger>
          <TabsTrigger value="suno">{t('Suno')}</TabsTrigger>
          <TabsTrigger value="video">{t('Video')}</TabsTrigger>
        </TabsList>

        {/* Chat mode */}
        <TabsContent value="chat" className="flex-1 flex flex-col mt-0">
          <Card className="flex-1 flex flex-col overflow-hidden">
            <ScrollArea className="flex-1 p-4" ref={scrollRef as any}>
              {messages.length === 0 ? (
                <div className="flex h-full flex-col items-center justify-center text-muted-foreground gap-2">
                  <img src="/logo.webp" alt="QC" className="h-12 w-12 opacity-20 rounded-xl object-cover" />
                  <p className="text-sm">{t('Start a conversation')}</p>
                </div>
              ) : (
                <div className="space-y-4">
                  {messages.map((msg, idx) => (
                    <div
                      key={idx}
                      className={cn(
                        'flex gap-3',
                        msg.role === 'user' && 'flex-row-reverse'
                      )}
                    >
                      <div
                        className={cn(
                          'flex h-8 w-8 shrink-0 items-center justify-center rounded-full',
                          msg.role === 'user'
                            ? 'bg-primary text-primary-foreground'
                            : 'bg-muted'
                        )}
                      >
                        {msg.role === 'user' ? (
                          <User className="h-4 w-4" />
                        ) : (
                          <img src="/logo.webp" alt="QC" className="h-4 w-4 rounded object-cover" />
                        )}
                      </div>
                      <div
                        className={cn(
                          'max-w-[80%] rounded-lg px-4 py-2 text-sm whitespace-pre-wrap',
                          msg.role === 'user'
                            ? 'bg-primary text-primary-foreground'
                            : 'bg-muted'
                        )}
                      >
                        {msg.content}
                        {msg.role === 'assistant' && !msg.content && streaming && (
                          <Loader2 className="inline h-4 w-4 animate-spin" />
                        )}
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </ScrollArea>

            {/* Input */}
            <div className="border-t p-4">
              <form
                onSubmit={(e) => {
                  e.preventDefault()
                  handleSend()
                }}
                className="flex flex-col sm:flex-row gap-2"
              >
                <Input
                  value={input}
                  onChange={(e) => setInput(e.target.value)}
                  placeholder={t('Type a message...')}
                  disabled={streaming}
                  className="flex-1"
                />
                {streaming ? (
                  <Button type="button" variant="destructive" onClick={handleStop}>
                    {t('Stop')}
                  </Button>
                ) : (
                  <Button type="submit" disabled={!input.trim()}>
                    <Send className="h-4 w-4" />
                  </Button>
                )}
              </form>
            </div>
          </Card>
        </TabsContent>

        {/* Midjourney mode */}
        <TabsContent value="midjourney" className="flex-1 mt-0">
          <Card className="flex-1 flex flex-col overflow-hidden">
            <CardHeader>
              <CardTitle>{t('Midjourney')}</CardTitle>
            </CardHeader>
            <CardContent className="flex-1 flex flex-col gap-4">
              <div className="space-y-2">
                <Label>{t('Prompt')}</Label>
                <textarea
                  value={mjPrompt}
                  onChange={(e) => setMjPrompt(e.target.value)}
                  placeholder={t('Describe the image you want to generate...')}
                  className="flex min-h-[120px] w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                />
              </div>
              <div className="flex items-center gap-4">
                <div className="space-y-1">
                  <Label>{t('Aspect Ratio')}</Label>
                  <Select value={mjAspect} onValueChange={setMjAspect}>
                    <SelectTrigger className="w-28">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="1:1">1:1</SelectItem>
                      <SelectItem value="16:9">16:9</SelectItem>
                      <SelectItem value="4:3">4:3</SelectItem>
                      <SelectItem value="3:2">3:2</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>

              {/* Task status display */}
              {mjTask.status && (
                <div className="rounded-lg border p-4 space-y-2">
                  <div className="flex items-center gap-2 text-sm">
                    <span className="font-medium">{t('Status')}:</span>
                    <span className={cn(
                      'capitalize',
                      mjTask.status === 'completed' && 'text-green-500',
                      mjTask.status === 'failed' && 'text-red-500',
                      mjTask.status === 'submitting' && 'text-yellow-500'
                    )}>
                      {mjTask.status === 'submitting' ? (
                        <><Loader2 className="inline h-3 w-3 animate-spin mr-1" />{t('Submitting')}</>
                      ) : mjTask.status === 'queued' || mjTask.status === 'running' ? (
                        <><Loader2 className="inline h-3 w-3 animate-spin mr-1" />{t('Running')}</>
                      ) : (
                        t(mjTask.status.charAt(0).toUpperCase() + mjTask.status.slice(1))
                      )}
                    </span>
                  </div>
                  {mjTask.result?.file_url && (
                    <div className="mt-2">
                      <img
                        src={mjTask.result.file_url}
                        alt="Generated"
                        className="max-w-full rounded-lg border"
                      />
                      <a
                        href={mjTask.result.file_url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-sm text-blue-500 hover:underline mt-1 inline-block"
                      >
                        {t('View full image')} ↗
                      </a>
                    </div>
                  )}
                  {mjTask.status === 'failed' && mjTask.result && (
                    <p className="text-sm text-red-500">
                      {mjTask.result.error || mjTask.result.reason || JSON.stringify(mjTask.result)}
                    </p>
                  )}
                </div>
              )}

              <div className="mt-auto">
                <Button onClick={handleSubmitMj} disabled={!mjPrompt.trim() || mjTask.loading}>
                  {mjTask.loading ? (
                    <><Loader2 className="mr-2 h-4 w-4 animate-spin" />{t('Generating...')}</>
                  ) : (
                    t('Generate')
                  )}
                </Button>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Suno mode */}
        <TabsContent value="suno" className="flex-1 mt-0">
          <Card className="flex-1 flex flex-col overflow-hidden">
            <CardHeader>
              <CardTitle>{t('Suno')}</CardTitle>
            </CardHeader>
            <CardContent className="flex-1 flex flex-col gap-4">
              <div className="space-y-2">
                <Label>{t('Prompt')}</Label>
                <textarea
                  value={sunoPrompt}
                  onChange={(e) => setSunoPrompt(e.target.value)}
                  placeholder={t('Describe the music you want to generate...')}
                  className="flex min-h-[100px] w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                />
              </div>
              <div className="space-y-2">
                <Label>{t('Negative Prompt (optional)')}</Label>
                <textarea
                  value={sunoNegative}
                  onChange={(e) => setSunoNegative(e.target.value)}
                  placeholder={t('What to avoid in the generated audio...')}
                  className="flex min-h-[60px] w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                />
              </div>
              <div className="flex items-center gap-4">
                <div className="space-y-1">
                  <Label>{t('Duration')}</Label>
                  <Select value={sunoDuration} onValueChange={setSunoDuration}>
                    <SelectTrigger className="w-28">
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="15">{t('15s')}</SelectItem>
                      <SelectItem value="30">{t('30s')}</SelectItem>
                      <SelectItem value="60">{t('60s')}</SelectItem>
                    </SelectContent>
                  </Select>
                </div>
              </div>

              {/* Task status display */}
              {sunoTask.status && (
                <div className="rounded-lg border p-4 space-y-2">
                  <div className="flex items-center gap-2 text-sm">
                    <span className="font-medium">{t('Status')}:</span>
                    <span className={cn(
                      'capitalize',
                      sunoTask.status === 'completed' && 'text-green-500',
                      sunoTask.status === 'failed' && 'text-red-500'
                    )}>
                      {sunoTask.status === 'submitting' ? (
                        <><Loader2 className="inline h-3 w-3 animate-spin mr-1" />{t('Submitting')}</>
                      ) : sunoTask.status === 'queued' || sunoTask.status === 'running' ? (
                        <><Loader2 className="inline h-3 w-3 animate-spin mr-1" />{t('Running')}</>
                      ) : (
                        t(sunoTask.status.charAt(0).toUpperCase() + sunoTask.status.slice(1))
                      )}
                    </span>
                  </div>
                  {sunoTask.result?.file_url && (
                    <div className="mt-2">
                      <audio controls className="w-full">
                        <source src={sunoTask.result.file_url} />
                        {t('Your browser does not support the audio element.')}
                      </audio>
                      <a
                        href={sunoTask.result.file_url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-sm text-blue-500 hover:underline mt-1 inline-block"
                      >
                        {t('Download')} ↗
                      </a>
                    </div>
                  )}
                  {sunoTask.status === 'failed' && sunoTask.result && (
                    <p className="text-sm text-red-500">
                      {sunoTask.result.error || sunoTask.result.reason || JSON.stringify(sunoTask.result)}
                    </p>
                  )}
                </div>
              )}

              <div className="mt-auto">
                <Button onClick={handleSubmitSuno} disabled={!sunoPrompt.trim() || sunoTask.loading}>
                  {sunoTask.loading ? (
                    <><Loader2 className="mr-2 h-4 w-4 animate-spin" />{t('Generating...')}</>
                  ) : (
                    t('Generate')
                  )}
                </Button>
              </div>
            </CardContent>
          </Card>
        </TabsContent>

        {/* Video mode */}
        <TabsContent value="video" className="flex-1 mt-0">
          <Card className="flex-1 flex flex-col overflow-hidden">
            <CardHeader>
              <CardTitle>{t('Video Generation')}</CardTitle>
            </CardHeader>
            <CardContent className="flex-1 flex flex-col gap-4">
              <div className="space-y-2">
                <Label>{t('Prompt')}</Label>
                <textarea
                  value={videoPrompt}
                  onChange={(e) => setVideoPrompt(e.target.value)}
                  placeholder={t('Describe the video you want to generate...')}
                  className="flex min-h-[120px] w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                />
              </div>

              {/* Task status display */}
              {videoTask.status && (
                <div className="rounded-lg border p-4 space-y-2">
                  <div className="flex items-center gap-2 text-sm">
                    <span className="font-medium">{t('Status')}:</span>
                    <span className={cn(
                      'capitalize',
                      videoTask.status === 'completed' && 'text-green-500',
                      videoTask.status === 'failed' && 'text-red-500'
                    )}>
                      {videoTask.status === 'submitting' ? (
                        <><Loader2 className="inline h-3 w-3 animate-spin mr-1" />{t('Submitting')}</>
                      ) : videoTask.status === 'queued' || videoTask.status === 'running' ? (
                        <><Loader2 className="inline h-3 w-3 animate-spin mr-1" />{t('Running')}</>
                      ) : (
                        t(videoTask.status.charAt(0).toUpperCase() + videoTask.status.slice(1))
                      )}
                    </span>
                  </div>
                  {videoTask.result?.file_url && (
                    <div className="mt-2">
                      <video controls className="max-w-full rounded-lg border">
                        <source src={videoTask.result.file_url} />
                        {t('Your browser does not support the video element.')}
                      </video>
                      <a
                        href={videoTask.result.file_url}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="text-sm text-blue-500 hover:underline mt-1 inline-block"
                      >
                        {t('Download')} ↗
                      </a>
                    </div>
                  )}
                  {videoTask.status === 'failed' && videoTask.result && (
                    <p className="text-sm text-red-500">
                      {videoTask.result.error || videoTask.result.reason || JSON.stringify(videoTask.result)}
                    </p>
                  )}
                </div>
              )}

              <div className="mt-auto">
                <Button onClick={handleSubmitVideo} disabled={!videoPrompt.trim() || videoTask.loading}>
                  {videoTask.loading ? (
                    <><Loader2 className="mr-2 h-4 w-4 animate-spin" />{t('Generating...')}</>
                  ) : (
                    t('Generate')
                  )}
                </Button>
              </div>
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}

