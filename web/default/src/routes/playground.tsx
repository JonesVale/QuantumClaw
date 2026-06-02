import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useState, useRef, useCallback } from 'react'
import { useQuery } from '@tanstack/react-query'
import { getModels, type ModelInfo } from '@/lib/api-extended'
import { getCommonHeaders } from '@/lib/api'

export const Route = createFileRoute('/playground')({
  component: PlaygroundPage,
})

interface Message { role: 'user' | 'assistant'; content: string }

function PlaygroundPage() {
  const { t } = useT()
  const [model, setModel] = useState('')
  const [input, setInput] = useState('')
  const [msgs, setMsgs] = useState<Message[]>([])
  const [loading, setLoading] = useState(false)
  const [showSidebar, setShowSidebar] = useState(true)
  const ref = useRef<HTMLDivElement>(null)

  const { data } = useQuery({
    queryKey: ['models-chat'],
    queryFn: async () => { const r = await getModels(); return r },
  })
  const models: ModelInfo[] = data?.data || []
  const chatModels = models.filter(m => m.use_case === 'chat' || !m.use_case)

  const send = useCallback(async () => {
    if (!input.trim() || !model) return
    const userMsg: Message = { role: 'user', content: input }
    setMsgs(p => [...p, userMsg])
    setInput('')
    setLoading(true)
    try {
      const r = await fetch('/api/relay', {
        method: 'POST',
        headers: { ...getCommonHeaders(), 'Content-Type': 'application/json' },
        body: JSON.stringify({ model, messages: [{ role: 'user', content: input }] }),
      })
      const data = await r.json()
      setMsgs(p => [...p, { role: 'assistant', content: data.choices?.[0]?.message?.content || 'No response' }])
    } catch {
      setMsgs(p => [...p, { role: 'assistant', content: 'Error: Request failed' }])
    }
    setLoading(false)
  }, [input, model])

  return (
    <div className="min-h-screen bg-background flex"
      style={{ backgroundImage: 'radial-gradient(ellipse at 50% -20%, oklch(0.92 0.03 52 / 0.3), transparent 60%)' }}>
      <div className="flex-1 flex flex-col">
        <div className="qc-wrap flex-1 flex flex-col py-6">
          {/* Header */}
          <div className="qc-fade-up mb-4">
            <h1 className="qc-title-hero font-bold tracking-tight text-foreground text-2xl">{t('Playground')}</h1>
          </div>

          {/* Controls */}
          <div className="qc-fade-up flex items-center gap-3 mb-4 flex-wrap">
            <select value={model} onChange={e => setModel(e.target.value)}
              className="h-10 rounded-xl border border-border/30 bg-white/70 px-3 text-sm outline-none focus:border-[oklch(0.72_0.18_52)]/40 transition-all min-w-[180px]">
              <option value="">{t('Select model')}</option>
              {chatModels.map((m: any) => (
                <option key={m.name} value={m.name}>
                  {m.name}{m.is_premium ? ' ⚠️' : ''}
                </option>
              ))}
            </select>
            {(() => {
              const selected = chatModels.find((m: any) => m.name === model)
              if (selected?.is_premium) {
                return (
                  <div className="inline-flex items-center gap-1.5 px-2.5 py-1 rounded-full text-xs font-medium bg-amber-100 text-amber-800 border border-amber-200">
                    ⚠️ {t('Premium channel')} — {selected.channel_name} ({(selected.sell_price_rate * 100).toFixed(0)}% {t('of base price')})
                  </div>
                )
              }
              return null
            })()}
            <button onClick={() => { setMsgs([]); setInput('') }}
              className="h-10 px-3 rounded-xl border border-border/30 bg-white/70 hover:bg-muted/40 text-xs text-muted-foreground transition-all flex items-center gap-1.5">
              <svg className="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M3 6h18M8 6V4a2 2 0 012-2h4a2 2 0 012 2v2"/><line x1="10" y1="11" x2="10" y2="17"/><line x1="14" y1="11" x2="14" y2="17"/><path d="M5 6v12a2 2 0 002 2h10a2 2 0 002-2V6"/></svg>
              {t('Clear')}
            </button>
            <button onClick={() => setShowSidebar(!showSidebar)}
              className="h-10 px-3 rounded-xl border border-border/30 bg-white/70 hover:bg-muted/40 text-xs text-muted-foreground transition-all flex items-center gap-1.5">
              <svg className="w-3.5 h-3.5" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><line x1="4" y1="6" x2="20" y2="6"/><line x1="8" y1="12" x2="20" y2="12"/><line x1="12" y1="18" x2="20" y2="18"/></svg>
              {t('Settings')}
            </button>
          </div>

          {/* Messages */}
          <div ref={ref} className="flex-1 overflow-y-auto space-y-4 mb-4 min-h-[300px] rounded-2xl bg-white/40 border border-border/10 p-4">
            {msgs.length === 0 && (
              <div className="flex items-center justify-center h-full text-muted-foreground/30">
                <div className="text-center">
                  <div className="text-3xl mb-3">💬</div>
                  <p className="text-sm">{t('Select a model and start chatting')}</p>
                </div>
              </div>
            )}
            {msgs.map((m, i) => (
              <div key={i} className={`flex ${m.role === 'user' ? 'justify-end' : 'justify-start'} qc-fade-up`} style={{ animationDelay: '0.05s' }}>
                <div className={`max-w-[80%] px-4 py-3 rounded-2xl text-sm leading-relaxed ${
                  m.role === 'user'
                    ? 'bg-gradient-to-r from-amber-500 to-orange-500 text-white'
                    : 'bg-white border border-border/10 text-foreground shadow-sm'
                }`}>
                  {m.content}
                </div>
              </div>
            ))}
            {loading && (
              <div className="flex justify-start qc-fade-up">
                <div className="px-4 py-3 rounded-2xl bg-white border border-border/10 shadow-sm">
                  <div className="flex gap-1.5">
                    <div className="w-2 h-2 rounded-full bg-amber-400 animate-bounce" style={{ animationDelay: '0s' }} />
                    <div className="w-2 h-2 rounded-full bg-amber-400 animate-bounce" style={{ animationDelay: '0.15s' }} />
                    <div className="w-2 h-2 rounded-full bg-amber-400 animate-bounce" style={{ animationDelay: '0.3s' }} />
                  </div>
                </div>
              </div>
            )}
          </div>

          {/* Input */}
          <div className="qc-fade-up flex gap-3">
            <input value={input} onChange={e => setInput(e.target.value)}
              onKeyDown={e => e.key === 'Enter' && send()}
              disabled={!model || loading}
              placeholder={model ? t('Type your message...') : t('Select a model first')}
              className="flex-1 h-12 rounded-xl border border-border/30 bg-white/80 px-4 text-sm outline-none focus:border-[oklch(0.72_0.18_52)]/40 focus:bg-white transition-all placeholder:text-muted-foreground/40 disabled:opacity-50"
            />
            <button onClick={send} disabled={!model || loading || !input.trim()}
              className="h-12 px-5 rounded-xl text-sm font-semibold text-white bg-gradient-to-r from-amber-500 to-orange-500 shadow-md shadow-orange-500/20 hover:shadow-lg disabled:opacity-50 disabled:cursor-not-allowed transition-all flex items-center gap-2">
              <svg className="w-4 h-4" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M22 2L11 13M22 2l-7 20-4-9-9-4 20-7z"/></svg>
            </button>
          </div>
        </div>
      </div>

      {/* Settings sidebar */}
      {showSidebar && (
        <div className="w-64 border-l border-border/20 bg-white/50 p-5 hidden lg:block">
          <h3 className="text-xs font-bold tracking-tight text-foreground mb-4">{t('Settings')}</h3>
          <div className="space-y-4">
            <div>
              <label className="text-xs font-medium text-muted-foreground/60 block mb-1.5">{t('Temperature')}</label>
              <input type="range" min="0" max="2" step="0.1" defaultValue="0.7"
                className="w-full accent-[oklch(0.72_0.18_52)]" />
            </div>
            <div>
              <label className="text-xs font-medium text-muted-foreground/60 block mb-1.5">{t('Max Tokens')}</label>
              <input type="number" defaultValue={4096}
                className="w-full h-9 rounded-lg border border-border/30 bg-white px-3 text-xs outline-none focus:border-[oklch(0.72_0.18_52)]/40 transition-all" />
            </div>
          </div>
        </div>
      )}
    </div>
  )
}
