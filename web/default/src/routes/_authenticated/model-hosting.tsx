import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useState, useEffect } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'

export const Route = createFileRoute('/_authenticated/model-hosting')({
  component: ModelHostingPage,
})

interface InferenceNode {
  id: number
  name: string
  node_type: string
  base_url: string
  api_key: string
  models: string
  model_count: number
  status: string
  last_health_check: number
}

function ModelHostingPage() {
  const { t, language } = useT()
  const [nodes, setNodes] = useState<InferenceNode[]>([])
  const [loading, setLoading] = useState(true)
  const [showAdd, setShowAdd] = useState(false)
  const [name, setName] = useState('')
  const [nodeType, setNodeType] = useState('vllm')
  const [baseURL, setBaseURL] = useState('')
  const [apiKey, setApiKey] = useState('')
  const [testing, setTesting] = useState<number | null>(null)
  const [message, setMessage] = useState('')

  useEffect(() => { loadNodes() }, [])

  const loadNodes = () => {
    setLoading(true)
    fetch('/api/user/self/inference-nodes')
      .then(r => r.json())
      .then(d => { if (d.success) setNodes(d.data); setLoading(false) })
      .catch(() => setLoading(false))
  }

  const addNode = async () => {
    if (!name.trim() || !baseURL.trim()) return
    const r = await fetch('/api/user/self/inference-nodes', {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ name: name.trim(), node_type: nodeType, base_url: baseURL.trim(), api_key: apiKey }),
    })
    const d = await r.json()
    if (d.success) {
      setShowAdd(false); setName(''); setBaseURL(''); setApiKey('')
      loadNodes()
    } else {
      setMessage(d.message || 'Error')
    }
  }

  const deleteNode = async (id: number) => {
    const r = await fetch(`/api/user/self/inference-nodes/${id}`, { method: 'DELETE' })
    const d = await r.json()
    if (d.success) loadNodes()
    else setMessage(d.message || 'Error')
  }

  const testNode = async (id: number) => {
    setTesting(id)
    const r = await fetch(`/api/user/self/inference-nodes/${id}/test`, { method: 'POST' })
    const d = await r.json()
    if (d.success) {
      setMessage(`✅ ${d.data.count} models discovered`)
    } else {
      setMessage(`❌ ${d.message}`)
    }
    setTesting(null)
    loadNodes()
  }

  const statusBadge = (status: string) => {
    const colors: Record<string, string> = {
      active: 'bg-green-100 text-green-700',
      inactive: 'bg-yellow-100 text-yellow-700',
      error: 'bg-red-100 text-red-700',
    }
    return colors[status] || 'bg-gray-100 text-gray-700'
  }

  const typeIcon: Record<string, string> = { vllm: '⚡', sglang: '🔷', ollama: '🦙' }

  return (
    <div className="max-w-4xl mx-auto px-4 py-6">
      <div className="flex justify-between items-center mb-6">
        <div>
          <h1 className="text-2xl font-bold">{t('mh_title') || 'Model Hosting'}</h1>
          <p className="text-sm text-muted-foreground">{t('mh_subtitle') || 'Host your own inference nodes'}</p>
        </div>
        <Button onClick={() => setShowAdd(true)}>{t('mh_add_node') || '+ Add Node'}</Button>
      </div>

      {message && (
        <div className="mb-4 p-3 rounded-lg bg-primary/10 text-primary text-sm">{message}</div>
      )}

      {showAdd && (
        <div className="mb-6 p-4 border rounded-lg space-y-3">
          <h3 className="font-bold">{t('mh_add_node') || 'Add Inference Node'}</h3>
          <Input value={name} onChange={e => setName(e.target.value)} placeholder={t('mh_node_name') || 'Node name'} />
          <div className="flex gap-2">
            {['vllm', 'sglang', 'ollama'].map(tp => (
              <button key={tp} onClick={() => setNodeType(tp)}
                className={`px-3 py-1.5 rounded text-sm font-medium ${
                  nodeType === tp ? 'bg-primary text-primary-foreground' : 'bg-secondary'
                }`}>{tp}</button>
            ))}
          </div>
          <Input value={baseURL} onChange={e => setBaseURL(e.target.value)} placeholder="http://192.168.1.100:8000" />
          <Input value={apiKey} onChange={e => setApiKey(e.target.value)} type="password" placeholder={t('mh_api_key') || 'API Key (optional)'} />
          <div className="flex gap-2">
            <Button onClick={addNode} disabled={!name.trim() || !baseURL.trim()}>{t('save') || 'Save'}</Button>
            <Button variant="outline" onClick={() => setShowAdd(false)}>{t('cancel') || 'Cancel'}</Button>
          </div>
        </div>
      )}

      {loading ? (
        <div className="text-center py-12 text-muted-foreground">{t('loading') || 'Loading...'}</div>
      ) : nodes.length === 0 ? (
        <div className="text-center py-12 border rounded-lg">
          <p className="text-4xl mb-2">🖥️</p>
          <p className="text-muted-foreground">{t('mh_no_nodes') || 'No inference nodes registered yet.'}</p>
          <Button className="mt-4" onClick={() => setShowAdd(true)}>{t('mh_add_first') || 'Add your first node'}</Button>
        </div>
      ) : (
        <div className="grid gap-4">
          {nodes.map(node => (
            <div key={node.id} className="border rounded-lg p-4">
              <div className="flex items-start justify-between mb-3">
                <div>
                  <div className="flex items-center gap-2">
                    <span className="text-xl">{typeIcon[node.node_type] || '🖥️'}</span>
                    <h3 className="font-bold">{node.name}</h3>
                    <span className={`text-xs px-2 py-0.5 rounded-full ${statusBadge(node.status)}`}>{node.status}</span>
                  </div>
                  <p className="text-sm text-muted-foreground mt-1">{node.base_url}</p>
                </div>
                <div className="flex gap-1">
                  <Button size="sm" variant="outline" onClick={() => testNode(node.id)} disabled={testing === node.id}>
                    {testing === node.id ? (t('testing') || 'Testing...') : (t('test') || 'Test')}
                  </Button>
                  <Button size="sm" variant="destructive" onClick={() => deleteNode(node.id)}>{t('delete') || 'Delete'}</Button>
                </div>
              </div>
              <div className="flex gap-4 text-sm text-muted-foreground">
                <span>{t('mh_type') || 'Type'}: {node.node_type}</span>
                <span>{t('mh_models') || 'Models'}: {node.model_count}</span>
                {node.last_health_check > 0 && (
                  <span>{t('mh_last_check') || 'Last check'}: {new Date(node.last_health_check * 1000).toLocaleString()}</span>
                )}
              </div>
              {node.models && node.models !== '[]' && node.model_count > 0 && (
                <div className="mt-2 flex flex-wrap gap-1">
                  {(JSON.parse(node.models) as string[]).map((m: string) => (
                    <span key={m} className="text-xs bg-muted px-1.5 py-0.5 rounded">{m}</span>
                  ))}
                </div>
              )}
            </div>
          ))}
        </div>
      )}
    </div>
  )
}
