import { useState } from 'react'
import { useT } from '@/lib/use-t'
import { User, Bot, Copy, Check, RefreshCw, Loader2 } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { cn } from '@/lib/utils'
import { Message } from './message-types'

interface MessageBubbleProps {
  message: Message
  isLast: boolean
  onRegenerate?: () => void
  streaming?: boolean
}

function SimpleMarkdown({ content }: { content: string }) {
  // Simple inline rendering: code blocks, bold, italic, links, inline code
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

export function MessageBubble({ message, isLast, onRegenerate, streaming }: MessageBubbleProps) {
  const { t } = useT()
  const [copied, setCopied] = useState(false)

  const copyMessage = async () => {
    await navigator.clipboard.writeText(message.content)
    setCopied(true)
    setTimeout(() => setCopied(false), 1500)
  }

  return (
    <div className={cn('flex gap-3 group', message.role === 'user' && 'flex-row-reverse')}>
      <div
        className={cn(
          'flex h-8 w-8 shrink-0 items-center justify-center rounded-full shadow-sm ring-1 ring-border/50',
          message.role === 'user'
            ? 'bg-primary text-primary-foreground'
            : 'bg-gradient-to-br from-blue-500 to-purple-600 text-white'
        )}
      >
        {message.role === 'user' ? (
          <User className="h-4 w-4" />
        ) : (
          <Bot className="h-4 w-4" />
        )}
      </div>
      <div className="flex flex-col gap-1 max-w-[80%]">
        <div
          className={cn(
            'rounded-2xl px-4 py-2.5 text-sm leading-relaxed whitespace-pre-wrap break-words',
            message.role === 'user'
              ? 'bg-primary text-primary-foreground rounded-tr-sm'
              : 'bg-muted/70 dark:bg-muted/30 rounded-tl-sm border border-border/30'
          )}
        >
          {message.content ? (
            message.role === 'assistant' ? (
              <SimpleMarkdown content={message.content} />
            ) : (
              message.content
            )
          ) : (
            streaming && <Loader2 className="inline h-4 w-4 animate-spin text-muted-foreground" />
          )}
        </div>

        {/* Hover actions */}
        <div className={cn(
          'flex items-center gap-1 opacity-0 group-hover:opacity-100 transition-opacity',
          message.role === 'user' ? 'justify-end' : 'justify-start'
        )}>
          <Button
            variant="ghost"
            size="icon"
            className="h-6 w-6"
            onClick={copyMessage}
            title={t('Copy')}
          >
            {copied ? <Check className="h-3 w-3 text-green-500" /> : <Copy className="h-3 w-3" />}
          </Button>
          {isLast && message.role === 'assistant' && !streaming && (
            <Button
              variant="ghost"
              size="icon"
              className="h-6 w-6"
              onClick={onRegenerate}
              title={t('Regenerate')}
            >
              <RefreshCw className="h-3 w-3" />
            </Button>
          )}
          <span className="text-[10px] text-muted-foreground">
            {new Date(message.timestamp).toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })}
          </span>
        </div>
      </div>
    </div>
  )
}
