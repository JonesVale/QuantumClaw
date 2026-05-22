import { useState, useCallback, useEffect } from 'react'
import { Conversation } from '@/components/conversation-sidebar'
import { Message } from '@/components/message-types'

const STORAGE_KEY = 'qc_conversations'
const MAX_CONVERSATIONS = 50

function loadConversations(): Conversation[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (raw) return JSON.parse(raw)
  } catch { /* ignore */ }
  return []
}

function saveConversations(convs: Conversation[]) {
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(convs))
  } catch { /* quota exceeded, ignore */ }
}

let idCounter = Date.now()
function genId(): string {
  return 'conv_' + (idCounter++).toString(36)
}

export function useConversations() {
  const [conversations, setConversations] = useState<Conversation[]>(loadConversations)
  const [activeId, setActiveId] = useState<string | null>(null)

  // Set initial active if none
  useEffect(() => {
    if (!activeId && conversations.length > 0) {
      setActiveId(conversations[0].id)
    }
  }, [])

  // Sync to localStorage
  useEffect(() => {
    saveConversations(conversations)
  }, [conversations])

  const activeConversation = conversations.find((c) => c.id === activeId) || null

  const addConversation = useCallback((model: string) => {
    const conv: Conversation = {
      id: genId(),
      title: '',
      messages: [],
      model,
      createdAt: Date.now(),
    }
    setConversations((prev) => {
      const next = [conv, ...prev]
      if (next.length > MAX_CONVERSATIONS) next.pop()
      return next
    })
    setActiveId(conv.id)
    return conv
  }, [])

  const deleteConversation = useCallback((id: string) => {
    setConversations((prev) => {
      const next = prev.filter((c) => c.id !== id)
      if (next.length === 0) {
        // Create a default conversation if all deleted
        const defaultConv: Conversation = {
          id: genId(),
          title: '',
          messages: [],
          model: 'auto',
          createdAt: Date.now(),
        }
        setActiveId(defaultConv.id)
        return [defaultConv]
      }
      if (activeId === id) {
        setActiveId(next[0].id)
      }
      return next
    })
  }, [activeId])

  const renameConversation = useCallback((id: string, title: string) => {
    setConversations((prev) =>
      prev.map((c) => (c.id === id ? { ...c, title } : c))
    )
  }, [])

  const updateConversation = useCallback((id: string, updates: Partial<Conversation>) => {
    setConversations((prev) =>
      prev.map((c) => (c.id === id ? { ...c, ...updates } : c))
    )
  }, [])

  const addMessage = useCallback((message: Message) => {
    setConversations((prev) =>
      prev.map((c) => {
        if (c.id !== activeId) return c
        return { ...c, messages: [...c.messages, message], title: c.title || (message.role === 'user' ? message.content.slice(0, 50) : '') }
      })
    )
  }, [activeId])

  const updateLastAssistantMessage = useCallback((content: string) => {
    setConversations((prev) =>
      prev.map((c) => {
        if (c.id !== activeId) return c
        const msgs = [...c.messages]
        for (let i = msgs.length - 1; i >= 0; i--) {
          if (msgs[i].role === 'assistant') {
            msgs[i] = { ...msgs[i], content, timestamp: Date.now() }
            break
          }
        }
        return { ...c, messages: msgs }
      })
    )
  }, [activeId])

  const removeLastAssistantMessage = useCallback(() => {
    setConversations((prev) =>
      prev.map((c) => {
        if (c.id !== activeId) return c
        const msgs = [...c.messages]
        for (let i = msgs.length - 1; i >= 0; i--) {
          if (msgs[i].role === 'assistant') {
            msgs.splice(i, 1)
            break
          }
        }
        return { ...c, messages: msgs }
      })
    )
  }, [activeId])

  return {
    conversations,
    activeId,
    activeConversation,
    setActiveId,
    addConversation,
    deleteConversation,
    renameConversation,
    updateConversation,
    addMessage,
    updateLastAssistantMessage,
    removeLastAssistantMessage,
  }
}
