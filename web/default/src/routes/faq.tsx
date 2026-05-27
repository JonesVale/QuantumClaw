import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'

export const Route = createFileRoute('/faq')({
  component: FAQPage,
})

function FAQPage() {
  const { t, language } = useT()
  const [faqs, setFaqs] = useState<{ id: number; question: string; answer: string; category: string }[]>([])
  const [categories, setCategories] = useState<string[]>([])
  const [activeCat, setActiveCat] = useState('')
  const [search, setSearch] = useState('')
  const [openId, setOpenId] = useState<number | null>(null)
  const [loading, setLoading] = useState(true)

  useEffect(() => {
    fetch('/api/faq/categories')
      .then(r => r.json())
      .then(d => { if (d.success) setCategories(d.data) })
      .catch(() => {})
  }, [])

  useEffect(() => {
    setLoading(true)
    const params = new URLSearchParams()
    if (activeCat) params.set('category', activeCat)
    fetch(`/api/faq?${params}`)
      .then(r => r.json())
      .then(d => { if (d.success) setFaqs(d.data); setLoading(false) })
      .catch(() => setLoading(false))
  }, [activeCat])

  const filtered = faqs.filter(f =>
    !search || f.question.toLowerCase().includes(search.toLowerCase()) ||
    f.answer.toLowerCase().includes(search.toLowerCase())
  )

  return (
    <div className="max-w-4xl mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold mb-2">{t('faq_title') || 'FAQ'}</h1>
      <p className="text-muted-foreground mb-6">{t('faq_subtitle') || 'Common questions about QuantumClaw'}</p>

      <Input
        placeholder={t('faq_search') || 'Search FAQ...'}
        value={search}
        onChange={e => setSearch(e.target.value)}
        className="mb-6 max-w-md"
      />

      <div className="flex gap-2 flex-wrap mb-6">
        <button
          onClick={() => setActiveCat('')}
          className={`px-4 py-1.5 rounded-full text-sm font-medium transition-colors ${
            !activeCat ? 'bg-primary text-primary-foreground' : 'bg-secondary text-secondary-foreground hover:bg-secondary/80'
          }`}
        >
          {t('faq_all') || 'All'}
        </button>
        {categories.map(cat => (
          <button
            key={cat}
            onClick={() => setActiveCat(cat)}
            className={`px-4 py-1.5 rounded-full text-sm font-medium transition-colors ${
              activeCat === cat ? 'bg-primary text-primary-foreground' : 'bg-secondary text-secondary-foreground hover:bg-secondary/80'
            }`}
          >
            {cat}
          </button>
        ))}
      </div>

      {loading ? (
        <div className="text-center py-12 text-muted-foreground">{t('loading') || 'Loading...'}</div>
      ) : filtered.length === 0 ? (
        <div className="text-center py-12 text-muted-foreground">{t('faq_no_results') || 'No matching questions found.'}</div>
      ) : (
        <div className="space-y-3">
          {filtered.map(faq => (
            <div
              key={faq.id}
              className="border rounded-lg overflow-hidden transition-all"
            >
              <button
                onClick={() => setOpenId(openId === faq.id ? null : faq.id)}
                className="w-full flex justify-between items-center p-4 text-left hover:bg-muted/50 transition-colors"
              >
                <span className="font-medium">{faq.question}</span>
                <svg
                  className={`w-5 h-5 transition-transform ${openId === faq.id ? 'rotate-180' : ''}`}
                  fill="none" stroke="currentColor" viewBox="0 0 24 24"
                >
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                </svg>
              </button>
              {openId === faq.id && (
                <div className="px-4 pb-4 text-muted-foreground leading-relaxed whitespace-pre-wrap">
                  {faq.answer}
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
