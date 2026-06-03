import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

export const Route = createFileRoute('/_authenticated/feedback')({
  component: FeedbackPage,
})

const TYPES = ['bug', 'feature', 'question'] as const

function FeedbackPage() {
  const { t, language } = useT()
  const [title, setTitle] = useState('')
  const [content, setContent] = useState('')
  const [type, setType] = useState<'bug' | 'feature' | 'question'>('question')
  const [email, setEmail] = useState('')
  const [submitting, setSubmitting] = useState(false)
  const [message, setMessage] = useState('')
  const [history, setHistory] = useState<{
    id: number; title: string; type: string; status: string; content: string;
    admin_response: string; created_at: number
  }[]>([])
  const [page, setPage] = useState(1)
  const [total, setTotal] = useState(0)
  const pageSize = 10

  useEffect(() => { loadHistory() }, [page])

  const loadHistory = () => {
    fetch(`/api/user/self/feedback?page=${page}&page_size=${pageSize}`)
      .then(r => r.json())
      .then(d => { if (d.success) { setHistory(d.data); setTotal(d.total) } })
      .catch(() => {})
  }

  const handleSubmit = async () => {
    if (!title.trim() || !content.trim()) return
    setSubmitting(true)
    setMessage('')
    try {
      const r = await fetch('/api/user/self/feedback', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ title: title.trim(), content: content.trim(), type, email }),
      })
      const d = await r.json()
      if (d.success) {
        setMessage(t('feedback_submitted') || 'Submitted!')
        setTitle(''); setContent(''); setEmail('')
        loadHistory()
      } else {
        setMessage(d.message || 'Error')
      }
    } catch {
      setMessage(t('network_error') || 'Network error')
    }
    setSubmitting(false)
  }

  const totalPages = Math.ceil(total / pageSize)

  return (
    <div className="max-w-3xl mx-auto px-4 py-6">
      <h1 className="text-2xl font-bold mb-6">{t('feedback_title') || 'Feedback'}</h1>

      {message && (
        <div className="mb-4 p-3 rounded-lg bg-primary/10 text-primary text-sm">{message}</div>
      )}

      <div className="space-y-4 mb-8 p-4 border rounded-lg">
        <div>
          <label className="block text-sm font-medium mb-1">{t('feedback_type') || 'Type'}</label>
          <div className="flex gap-2">
            {TYPES.map(tp => (
              <button
                key={tp}
                onClick={() => setType(tp)}
                className={`px-3 py-1.5 rounded text-sm font-medium transition-colors ${
                  type === tp ? 'bg-primary text-primary-foreground' : 'bg-secondary text-secondary-foreground hover:bg-secondary/80'
                }`}
              >
                {tp === 'bug' ? (t('feedback_bug') || 'Bug') :
                 tp === 'feature' ? (t('feedback_feature') || 'Feature') :
                 (t('feedback_question') || 'Question')}
              </button>
            ))}
          </div>
        </div>

        <div>
          <label className="block text-sm font-medium mb-1">{t('feedback_title_label') || 'Title'} *</label>
          <Input value={title} onChange={e => setTitle(e.target.value)} placeholder={t('feedback_title_placeholder') || 'Brief title...'} maxLength={200} />
        </div>

        <div>
          <label className="block text-sm font-medium mb-1">{t('feedback_content_label') || 'Content'} *</label>
          <textarea value={content} onChange={e => setContent(e.target.value)} rows={5}
            placeholder={t(`feedback_content_placeholder`) || `Describe in detail...`}
            className="flex min-h-[80px] w-full rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring focus-visible:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50" />
        </div>

        <div>
          <label className="block text-sm font-medium mb-1">{t('feedback_email') || 'Email'} <span className="text-muted-foreground">({t('optional') || 'optional'})</span></label>
          <Input value={email} onChange={e => setEmail(e.target.value)} type="email" placeholder="your@email.com" />
        </div>

        <Button onClick={handleSubmit} disabled={submitting || !title.trim() || !content.trim()}>
          {submitting ? (t('submitting') || 'Submitting...') : (t('feedback_submit') || 'Submit')}
        </Button>
      </div>

      <h2 className="text-xl font-bold mb-4">{t('feedback_history') || 'My Feedback'}</h2>
      {history.length === 0 ? (
        <p className="text-muted-foreground">{t('feedback_no_history') || 'No feedback submitted yet.'}</p>
      ) : (
        <div className="space-y-3">
          {history.map(fb => (
            <div key={fb.id} className="p-4 border rounded-lg">
              <div className="flex items-center gap-2 mb-1">
                <span className={`text-xs px-2 py-0.5 rounded-full ${
                  fb.type === 'bug' ? 'bg-red-100 text-red-700' :
                  fb.type === 'feature' ? 'bg-blue-100 text-blue-700' :
                  'bg-green-100 text-green-700'
                }`}>{fb.type}</span>
                <span className={`text-xs px-2 py-0.5 rounded-full ${
                  fb.status === 'pending' ? 'bg-yellow-100 text-yellow-700' :
                  fb.status === 'resolved' ? 'bg-green-100 text-green-700' :
                  'bg-gray-100 text-gray-700'
                }`}>{fb.status}</span>
                <span className="text-xs text-muted-foreground ml-auto">{new Date(fb.created_at * 1000).toLocaleDateString()}</span>
              </div>
              <h3 className="font-medium">{fb.title}</h3>
              <p className="text-sm text-muted-foreground mt-1 line-clamp-2">{fb.content}</p>
              {fb.admin_response && (
                <div className="mt-2 p-2 bg-muted rounded text-sm">
                  <span className="font-medium">{t('feedback_response') || 'Response'}:</span> {fb.admin_response}
                </div>
              )}
            </div>
          ))}
        </div>
      )}

      {totalPages > 1 && (
        <div className="flex justify-center gap-2 mt-6">
          <button onClick={() => setPage(p => Math.max(1, p-1))} disabled={page===1}
            className="px-3 py-1 text-sm rounded border disabled:opacity-50">{t("Prev")}</button>
          <span className="px-3 py-1 text-sm">{page}/{totalPages}</span>
          <button onClick={() => setPage(p => Math.min(totalPages, p+1))} disabled={page===totalPages}
            className="px-3 py-1 text-sm rounded border disabled:opacity-50">Next</button>
        </div>
      )}
    </div>
  )
}
