import { useTranslation } from 'react-i18next'
import { MessageSquare, Plus, Trash2, Edit3, Check, X } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { cn } from '@/lib/utils'
import { useState } from 'react'

export interface Conversation {
  id: string
  title: string
  messages: { role: 'user' | 'assistant' | 'system'; content: string; timestamp: number }[]
  model: string
  createdAt: number
}

interface ConversationSidebarProps {
  conversations: Conversation[]
  activeId: string | null
  onSelect: (id: string) => void
  onNew: () => void
  onDelete: (id: string) => void
  onRename: (id: string, title: string) => void
  onClose: () => void
}

function ConversationItem({
  conv,
  isActive,
  onSelect,
  onDelete,
  onRename,
}: {
  conv: Conversation
  isActive: boolean
  onSelect: () => void
  onDelete: () => void
  onRename: (title: string) => void
}) {
  const { t } = useTranslation()
  const [editing, setEditing] = useState(false)
  const [editValue, setEditValue] = useState(conv.title)

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
          onSubmit={(e) => {
            e.preventDefault()
            e.stopPropagation()
            onRename(editValue || conv.title)
            setEditing(false)
          }}
          className="flex-1 flex gap-1"
          onClick={(e) => e.stopPropagation()}
        >
          <Input
            value={editValue}
            onChange={(e) => setEditValue(e.target.value)}
            className="h-6 text-xs py-0"
            autoFocus
          />
          <Button type="submit" variant="ghost" size="icon" className="h-6 w-6 shrink-0">
            <Check className="h-3 w-3" />
          </Button>
          <Button
            type="button"
            variant="ghost"
            size="icon"
            className="h-6 w-6 shrink-0"
            onClick={() => setEditing(false)}
          >
            <X className="h-3 w-3" />
          </Button>
        </form>
      ) : (
        <>
          <span className="flex-1 truncate">{conv.title || t('New conversation')}</span>
          <div className="hidden group-hover:flex items-center gap-0.5">
            <Button
              variant="ghost"
              size="icon"
              className="h-6 w-6"
              onClick={(e) => {
                e.stopPropagation()
                setEditValue(conv.title)
                setEditing(true)
              }}
            >
              <Edit3 className="h-3 w-3" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="h-6 w-6 text-red-500 hover:text-red-600"
              onClick={(e) => {
                e.stopPropagation()
                if (confirm(t('Delete this conversation?'))) onDelete()
              }}
            >
              <Trash2 className="h-3 w-3" />
            </Button>
          </div>
        </>
      )}
    </div>
  )
}

export function ConversationSidebar({
  conversations,
  activeId,
  onSelect,
  onNew,
  onDelete,
  onRename,
  onClose,
}: ConversationSidebarProps) {
  const { t } = useTranslation()

  return (
    <div className="w-64 shrink-0 border-r bg-card p-3 flex flex-col gap-2 overflow-y-auto">
      <div className="flex items-center justify-between mb-1">
        <h3 className="text-xs font-semibold text-muted-foreground uppercase tracking-wider">
          {t('Conversations')}
        </h3>
        <div className="flex gap-1">
          <Button variant="ghost" size="icon" className="h-6 w-6" onClick={onNew} title={t('New conversation')}>
            <Plus className="h-4 w-4" />
          </Button>
        </div>
      </div>
      <div className="space-y-1 flex-1 overflow-y-auto">
        {conversations.length === 0 ? (
          <p className="text-xs text-muted-foreground text-center py-8">{t('No conversations yet')}</p>
        ) : (
          conversations.map((conv) => (
            <ConversationItem
              key={conv.id}
              conv={conv}
              isActive={conv.id === activeId}
              onSelect={() => onSelect(conv.id)}
              onDelete={() => onDelete(conv.id)}
              onRename={(title) => onRename(conv.id, title)}
            />
          ))
        )}
      </div>
    </div>
  )
}
