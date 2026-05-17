import { createFileRoute } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { useState, useRef, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card } from '@/components/ui/card'
import { Send, Bot, User, MessageSquare } from 'lucide-react'

export const Route = createFileRoute('/(auth)/chat')({
  component: ChatPage,
})

interface ChatMessage {
  role: 'user' | 'assistant'
  content: string
}

function ChatPage() {
  const { t } = useTranslation()
  const [messages, setMessages] = useState<ChatMessage[]>([
    {
      role: 'assistant',
      content: t('chat_welcome_message', 'Hello! I\'m your AI assistant. How can I help you today?'),
    },
  ])
  const [input, setInput] = useState('')
  const [loading, setLoading] = useState(false)
  const messagesEndRef = useRef<HTMLDivElement>(null)
  const inputRef = useRef<HTMLInputElement>(null)

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' })
  }, [messages])

  useEffect(() => {
    inputRef.current?.focus()
  }, [])

  const handleSend = async () => {
    const trimmed = input.trim()
    if (!trimmed || loading) return

    const userMessage: ChatMessage = { role: 'user', content: trimmed }
    setMessages((prev) => [...prev, userMessage])
    setInput('')
    setLoading(true)

    try {
      const apiMessages = [...messages, userMessage].map((m) => ({
        role: m.role,
        content: m.content,
      }))

      const res = await fetch('/v1/chat/completions', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          model: 'gpt-4o',
          messages: apiMessages,
          stream: false,
          chat_mode: 'guest',
        }),
      })

      if (!res.ok) {
        throw new Error(`HTTP ${res.status}`)
      }

      const data = await res.json()
      const reply = data.choices?.[0]?.message?.content || t('chat_no_response', 'No response received.')

      setMessages((prev) => [...prev, { role: 'assistant', content: reply }])
    } catch (err) {
      const errorMsg = err instanceof Error ? err.message : String(err)
      setMessages((prev) => [
        ...prev,
        { role: 'assistant', content: `${t('chat_error_prefix', 'Error')}: ${errorMsg}` },
      ])
    } finally {
      setLoading(false)
    }
  }

  const handleKeyDown = (e: React.KeyboardEvent) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault()
      handleSend()
    }
  }

  return (
    <div className="mx-auto flex h-screen max-w-4xl flex-col bg-background p-4">
      {/* Header */}
      <div className="mb-4 flex items-center gap-2 border-b pb-3">
        <div className="flex h-8 w-8 items-center justify-center rounded-lg bg-blue-600">
          <MessageSquare className="h-4 w-4 text-white" />
        </div>
        <h1 className="text-lg font-semibold">{t('chat_title', 'Chat')}</h1>
      </div>

      {/* Messages Area */}
      <div className="flex-1 space-y-4 overflow-y-auto px-2">
        {messages.map((msg, idx) => (
          <div key={idx} className={`flex gap-3 ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}>
            {msg.role === 'assistant' && (
              <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-blue-100 dark:bg-blue-900">
                <Bot className="h-4 w-4 text-blue-600 dark:text-blue-400" />
              </div>
            )}
            <div
              className={`max-w-[80%] rounded-2xl px-4 py-2.5 text-sm ${
                msg.role === 'user'
                  ? 'bg-blue-600 text-white'
                  : 'bg-muted text-foreground'
              }`}
            >
              <p className="whitespace-pre-wrap">{msg.content}</p>
            </div>
            {msg.role === 'user' && (
              <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-blue-600">
                <User className="h-4 w-4 text-white" />
              </div>
            )}
          </div>
        ))}
        {loading && (
          <div className="flex items-center gap-2 text-sm text-muted-foreground">
            <div className="flex h-6 w-6 items-center justify-center rounded-full bg-muted">
              <Bot className="h-3.5 w-3.5" />
            </div>
            <span className="animate-pulse">{t('chat_thinking', 'Thinking...')}</span>
          </div>
        )}
        <div ref={messagesEndRef} />
      </div>

      {/* Input Area */}
      <div className="mt-4 border-t pt-4">
        <form
          onSubmit={(e) => {
            e.preventDefault()
            handleSend()
          }}
          className="flex gap-2"
        >
          <Input
            ref={inputRef}
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder={t('chat_input_placeholder', 'Type your message...')}
            disabled={loading}
            className="flex-1"
          />
          <Button type="submit" disabled={!input.trim() || loading}>
            <Send className="h-4 w-4" />
          </Button>
        </form>
      </div>
    </div>
  )
}
