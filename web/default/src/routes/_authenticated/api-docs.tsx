import { createFileRoute, Link } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { useState } from 'react'
import {
  MessageSquare,
  Image,
  Music,
  Braces,
  FileText,
  Wand2,
  Bot,
  ListTodo,
  BookOpen,
  Copy,
  Check,
  Server,
  ChevronDown,
  ChevronRight,
  Key,
  Terminal,
  Eye,
  Send,
} from 'lucide-react'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { Badge } from '@/components/ui/badge'
import { cn } from '@/lib/utils'

export const Route = createFileRoute('/_authenticated/api-docs')({
  component: ApiDocsPage,
})

// ── Types ──────────────────────────────────────────────────────────

interface Endpoint {
  method: 'GET' | 'POST' | 'PUT' | 'DELETE'
  path: string
  description: string
  auth?: string
  curl: string
}

interface EndpointCategory {
  id: string
  title: string
  icon: React.ElementType
  endpoints: Endpoint[]
}

// ── Code Block Component ───────────────────────────────────────────

function CodeBlock({ code, language = 'bash' }: { code: string; language?: string }) {
  const { t } = useTranslation()
  const [copied, setCopied] = useState(false)

  const copyToClipboard = async () => {
    await navigator.clipboard.writeText(code)
    setCopied(true)
    setTimeout(() => setCopied(false), 2000)
  }

  return (
    <div className="relative group">
      <div className="flex items-center justify-between bg-muted px-4 py-1.5 rounded-t-lg border-b text-xs text-muted-foreground">
        <span>{language}</span>
        <button
          onClick={copyToClipboard}
          className="flex items-center gap-1.5 hover:text-foreground transition-colors"
        >
          {copied ? <Check className="h-3.5 w-3.5" /> : <Copy className="h-3.5 w-3.5" />}
          {copied ? t('Copied!') : t('Copy')}
        </button>
      </div>
      <pre className="bg-muted/50 px-4 py-3 rounded-b-lg overflow-x-auto text-sm">
        <code>{code}</code>
      </pre>
    </div>
  )
}

// ── Endpoint Card ───────────────────────────────────────────────────

function EndpointCard({ endpoint }: { endpoint: Endpoint }) {
  const { t } = useTranslation()
  const [expanded, setExpanded] = useState(false)
  const [keyInput, setKeyInput] = useState('')
  const [result, setResult] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)

  const methodColors: Record<string, string> = {
    GET: 'text-green-600 bg-green-100 dark:text-green-400 dark:bg-green-950',
    POST: 'text-blue-600 bg-blue-100 dark:text-blue-400 dark:bg-blue-950',
    PUT: 'text-orange-600 bg-orange-100 dark:text-orange-400 dark:bg-orange-950',
    DELETE: 'text-red-600 bg-red-100 dark:text-red-400 dark:bg-red-950',
  }

  const runTryIt = async () => {
    if (!keyInput.trim()) return
    setLoading(true)
    setResult(null)
    try {
      const baseUrl = window.location.origin
      const res = await fetch(`${baseUrl}${endpoint.path}`, {
        headers: keyInput ? { Authorization: `Bearer ${keyInput}` } : undefined,
      })
      const data = await res.json()
      setResult(JSON.stringify(data, null, 2))
    } catch (err) {
      setResult(`${t('Error')}: ${err instanceof Error ? err.message : t('Request failed')}`)
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="border rounded-lg overflow-hidden">
      <button
        onClick={() => setExpanded(!expanded)}
        className="w-full flex items-center gap-3 px-4 py-3 hover:bg-muted/50 transition-colors text-left"
      >
        <span
          className={cn(
            'inline-flex items-center px-2 py-0.5 rounded text-xs font-mono font-bold shrink-0',
            methodColors[endpoint.method]
          )}
        >
          {endpoint.method}
        </span>
        <code className="text-sm font-mono flex-1">{endpoint.path}</code>
        {expanded ? <ChevronDown className="h-4 w-4 shrink-0" /> : <ChevronRight className="h-4 w-4 shrink-0" />}
      </button>

      {expanded && (
        <div className="border-t px-4 py-4 space-y-4">
          <p className="text-sm text-muted-foreground">{t(endpoint.description)}</p>
          {endpoint.auth && (
            <div className="flex items-center gap-2 text-xs">
              <Badge variant="secondary">{t('Auth')}: {endpoint.auth}</Badge>
            </div>
          )}

          <div>
            <h4 className="text-xs font-semibold mb-2 flex items-center gap-1.5">
              <Terminal className="h-3.5 w-3.5" /> {t("Example")}
            </h4>
            <CodeBlock code={endpoint.curl} />
          </div>

          <div>
            <h4 className="text-xs font-semibold mb-2 flex items-center gap-1.5">
              <Eye className="h-3.5 w-3.5" /> {t("Try it out")}
            </h4>
            <div className="flex gap-2">
              <Input
                placeholder={t("Enter your API key...")}
                value={keyInput}
                onChange={(e) => setKeyInput(e.target.value)}
                className="flex-1"
              />
              <Button size="sm" onClick={runTryIt} disabled={loading || !keyInput.trim()}>
                {loading ? (
                  <span className="animate-spin">...</span>
                ) : (
                  <>
                    <Send className="h-3.5 w-3.5 mr-1" /> {t("Send")}
                  </>
                )}
              </Button>
            </div>
            {result && (
              <pre className="mt-2 p-3 bg-muted rounded-lg text-xs overflow-x-auto max-h-48 overflow-y-auto">
                <code>{result}</code>
              </pre>
            )}
          </div>
        </div>
      )}
    </div>
  )
}

// ── Category Section ───────────────────────────────────────────────

function CategorySection({ category }: { category: EndpointCategory }) {
  const { t } = useTranslation()
  return (
    <Card>
      <CardHeader>
        <CardTitle className="flex items-center gap-2 text-lg">
          <category.icon className="h-5 w-5 text-blue-600" />
          {category.title}
        </CardTitle>
      </CardHeader>
      <CardContent className="space-y-3">
        {category.endpoints.map((ep, i) => (
          <EndpointCard key={i} endpoint={ep} />
        ))}
      </CardContent>
    </Card>
  )
}

// ── Base URL ────────────────────────────────────────────────────────

function BaseUrlBanner() {
  const { t } = useTranslation()
  return (
    <div className="flex items-center gap-2 px-4 py-3 bg-blue-50 dark:bg-blue-950 rounded-lg border border-blue-200 dark:border-blue-800">
      <Server className="h-4 w-4 text-blue-600 shrink-0" />
      <span className="text-sm text-muted-foreground">{t("Base URL")}:</span>
      <code className="text-sm font-mono bg-blue-100 dark:bg-blue-900 px-2 py-0.5 rounded">
        {window.location.origin}
      </code>
    </div>
  )
}

// ── Endpoint Data ──────────────────────────────────────────────────

const API_ENDPOINTS: EndpointCategory[] = [
  {
    id: 'chat',
    title: 'Chat Completions',
    icon: MessageSquare,
    endpoints: [
      {
        method: 'POST',
        path: '/v1/chat/completions',
        description: 'Create a chat completion. Compatible with OpenAI\'s chat completions API.',
        auth: 'Bearer Token',
        curl: `curl ${window.location.origin}/v1/chat/completions \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer sk-your-key-here" \\
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "Hello!"}]
  }'`,
      },
      {
        method: 'POST',
        path: '/v1/chat/completions (streaming)',
        description: 'Stream chat completions using Server-Sent Events.',
        auth: 'Bearer Token',
        curl: `curl ${window.location.origin}/v1/chat/completions \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer sk-your-key-here" \\
  -d '{
    "model": "gpt-4o",
    "messages": [{"role": "user", "content": "Hello!"}],
    "stream": true
  }'`,
      },
    ],
  },
  {
    id: 'images',
    title: 'Images',
    icon: Image,
    endpoints: [
      {
        method: 'POST',
        path: '/v1/images/generations',
        description: 'Generate images from text descriptions.',
        auth: 'Bearer Token',
        curl: `curl ${window.location.origin}/v1/images/generations \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer sk-your-key-here" \\
  -d '{
    "model": "dall-e-3",
    "prompt": "A cute cat",
    "n": 1,
    "size": "1024x1024"
  }'`,
      },
    ],
  },
  {
    id: 'audio',
    title: 'Audio / Speech',
    icon: Music,
    endpoints: [
      {
        method: 'POST',
        path: '/v1/audio/transcriptions',
        description: 'Transcribe audio to text using Whisper models.',
        auth: 'Bearer Token',
        curl: `curl ${window.location.origin}/v1/audio/transcriptions \\
  -H "Authorization: Bearer sk-your-key-here" \\
  -F "file=@audio.mp3" \\
  -F "model=whisper-1"`,
      },
      {
        method: 'POST',
        path: '/v1/audio/speech',
        description: 'Generate speech from text using TTS models.',
        auth: 'Bearer Token',
        curl: `curl ${window.location.origin}/v1/audio/speech \\
  -H "Authorization: Bearer sk-your-key-here" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "tts-1",
    "input": "Hello world",
    "voice": "alloy"
  }'`,
      },
    ],
  },
  {
    id: 'embeddings',
    title: 'Embeddings',
    icon: Braces,
    endpoints: [
      {
        method: 'POST',
        path: '/v1/embeddings',
        description: 'Create vector embeddings for text input.',
        auth: 'Bearer Token',
        curl: `curl ${window.location.origin}/v1/embeddings \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer sk-your-key-here" \\
  -d '{
    "model": "text-embedding-3-small",
    "input": "The quick brown fox"
  }'`,
      },
    ],
  },
  {
    id: 'files',
    title: 'Files',
    icon: FileText,
    endpoints: [
      {
        method: 'GET',
        path: '/v1/files',
        description: 'List uploaded files.',
        auth: 'Bearer Token',
        curl: `curl ${window.location.origin}/v1/files \\
  -H "Authorization: Bearer sk-your-key-here"`,
      },
      {
        method: 'POST',
        path: '/v1/files',
        description: 'Upload a file for use with models.',
        auth: 'Bearer Token',
        curl: `curl ${window.location.origin}/v1/files \\
  -H "Authorization: Bearer sk-your-key-here" \\
  -F "file=@document.pdf" \\
  -F "purpose=assistants"`,
      },
      {
        method: 'GET',
        path: '/v1/files/:id',
        description: 'Retrieve file details by ID.',
        auth: 'Bearer Token',
        curl: `curl ${window.location.origin}/v1/files/file-abc123 \\
  -H "Authorization: Bearer sk-your-key-here"`,
      },
      {
        method: 'DELETE',
        path: '/v1/files/:id',
        description: 'Delete a file by ID.',
        auth: 'Bearer Token',
        curl: `curl -X DELETE ${window.location.origin}/v1/files/file-abc123 \\
  -H "Authorization: Bearer sk-your-key-here"`,
      },
    ],
  },
  {
    id: 'finetuning',
    title: 'Fine-tuning',
    icon: Wand2,
    endpoints: [
      {
        method: 'POST',
        path: '/v1/fine_tuning/jobs',
        description: 'Create a fine-tuning job.',
        auth: 'Bearer Token',
        curl: `curl ${window.location.origin}/v1/fine_tuning/jobs \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer sk-your-key-here" \\
  -d '{
    "model": "gpt-4o-mini-2024-07-18",
    "training_file": "file-abc123"
  }'`,
      },
      {
        method: 'GET',
        path: '/v1/fine_tuning/jobs',
        description: 'List fine-tuning jobs.',
        auth: 'Bearer Token',
        curl: `curl ${window.location.origin}/v1/fine_tuning/jobs \\
  -H "Authorization: Bearer sk-your-key-here"`,
      },
    ],
  },
  {
    id: 'assistants',
    title: 'Assistants',
    icon: Bot,
    endpoints: [
      {
        method: 'POST',
        path: '/v1/assistants',
        description: 'Create an assistant.',
        auth: 'Bearer Token',
        curl: `curl ${window.location.origin}/v1/assistants \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer sk-your-key-here" \\
  -d '{
    "instructions": "You are a helpful assistant.",
    "name": "My Assistant",
    "tools": [{"type": "code_interpreter"}],
    "model": "gpt-4o"
  }'`,
      },
      {
        method: 'GET',
        path: '/v1/assistants',
        description: 'List all assistants.',
        auth: 'Bearer Token',
        curl: `curl ${window.location.origin}/v1/assistants \\
  -H "Authorization: Bearer sk-your-key-here"`,
      },
      {
        method: 'POST',
        path: '/v1/threads/:id/messages',
        description: 'Create a message in a thread.',
        auth: 'Bearer Token',
        curl: `curl ${window.location.origin}/v1/threads/thread_abc123/messages \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer sk-your-key-here" \\
  -d '{"role": "user", "content": "Hello!"}'`,
      },
    ],
  },
  {
    id: 'async-tasks',
    title: 'Async Tasks',
    icon: ListTodo,
    endpoints: [
      {
        method: 'POST',
        path: '/api/task/midjourney',
        description: 'Create a Midjourney image generation task.',
        auth: 'Bearer Token',
        curl: `curl ${window.location.origin}/api/task/midjourney \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer sk-your-key-here" \\
  -d '{"type": "imagine", "prompt": "a beautiful sunset"}'`,
      },
      {
        method: 'GET',
        path: '/api/task/:task_id',
        description: 'Get the status of an async task by ID.',
        auth: 'Bearer Token',
        curl: `curl ${window.location.origin}/api/task/task_abc123 \\
  -H "Authorization: Bearer sk-your-key-here"`,
      },
      {
        method: 'GET',
        path: '/api/task',
        description: 'List all async tasks for the current user.',
        auth: 'Bearer Token',
        curl: `curl ${window.location.origin}/api/task \\
  -H "Authorization: Bearer sk-your-key-here"`,
      },
    ],
  },
  {
    id: 'quantum',
    title: 'Quantum Computing',
    icon: Server,
    endpoints: [
      {
        method: 'POST',
        path: '/v1/quantum/run',
        description: 'Submit a quantum circuit for execution. Supports multiple quantum backends including IonQ, IBM Q, Rigetti, AWS Braket, and Azure Quantum.',
        auth: 'Bearer Token',
        curl: `curl -X POST ${window.location.origin}/v1/quantum/run \\
  -H "Content-Type: application/json" \\
  -H "Authorization: Bearer sk-your-key-here" \\
  -d '{
    "backend": "ionq_harmony",
    "shots": 1000,
    "circuit": {
      "qubits": 2,
      "gates": [
        {"name": "h", "targets": [0]},
        {"name": "cx", "control": [0], "targets": [1]}
      ]
    }
  }'`,
      },
      {
        method: 'GET',
        path: '/v1/quantum/status/:task_id',
        description: 'Query the status and results of a submitted quantum task by task ID.',
        auth: 'Bearer Token',
        curl: `curl ${window.location.origin}/v1/quantum/status/qt_abc123 \\
  -H "Authorization: Bearer sk-your-key-here"`,
      },
      {
        method: 'POST',
        path: '/v1/quantum/cancel/:task_id',
        description: 'Cancel a running quantum task before it completes.',
        auth: 'Bearer Token',
        curl: `curl -X POST ${window.location.origin}/v1/quantum/cancel/qt_abc123 \\
  -H "Authorization: Bearer sk-your-key-here"`,
      },
      {
        method: 'GET',
        path: '/v1/quantum/backends',
        description: 'List available quantum backends for the configured quantum channel.',
        auth: 'Bearer Token',
        curl: `curl ${window.location.origin}/v1/quantum/backends \\
  -H "Authorization: Bearer sk-your-key-here"`,
      },
    ],
  },
]

// ── Page Component ─────────────────────────────────────────────────

function ApiDocsPage() {
  const { t } = useTranslation()
  return (
    <div className=" w-full p-4 sm:p-6 space-y-6">
      {/* Header */}
      <div className="space-y-2">
        <h1 className="text-3xl font-bold tracking-tight flex items-center gap-3">
          <BookOpen className="h-8 w-8 text-blue-600" />
          {t('API Documentation')}
        </h1>
        <p className="text-muted-foreground">
          {t('Comprehensive API reference for QuantumClaw. Supports AI model APIs and Quantum computing endpoints.')}
        </p>
      </div>

      {/* Apifox Playground Callout */}
      <div className="bg-blue-50 dark:bg-blue-950 border border-blue-200 dark:border-blue-800 rounded-lg p-4 mb-8">
        <p className="text-sm text-blue-700 dark:text-blue-300">
          {t('apifox_debug_hint')}{' '}
          <a href="https://apifox.newapi.ai/" target="_blank" rel="noopener noreferrer" className="underline font-medium">
            Apifox Playground
          </a>
        </p>
      </div>

      <BaseUrlBanner />

      {/* Authentication */}
      <Card>
        <CardHeader>
          <CardTitle className="flex items-center gap-2">
            <Key className="h-5 w-5 text-blue-600" />
            {t('Authentication')}
          </CardTitle>
        </CardHeader>
        <CardContent className="space-y-3 text-sm">
          <p>
            {t('Most API endpoints require authentication via a')} <strong>Bearer Token</strong> {t('in the')}
            <code className="mx-1 px-1 py-0.5 bg-muted rounded text-xs font-mono">Authorization</code> {t("header")}:
          </p>
          <CodeBlock code={`Authorization: Bearer sk-your-api-key-here`} language="http" />
          <p>
            {t('You can create and manage your API keys from the')}{' '}
            <Link to="/keys" className="text-blue-600 hover:underline">
              {t("API Keys page")}
            </Link>
            .
          </p>
        </CardContent>
      </Card>

      {/* Endpoint Categories */}
      <Tabs defaultValue="chat" className="space-y-4">
        <TabsList className="flex-wrap">
          {API_ENDPOINTS.map((cat) => (
            <TabsTrigger key={cat.id} value={cat.id} className="gap-1.5">
              <cat.icon className="h-3.5 w-3.5" />
              <span className="hidden sm:inline">{t(cat.title)}</span>
            </TabsTrigger>
          ))}
        </TabsList>

        {API_ENDPOINTS.map((cat) => (
          <TabsContent key={cat.id} value={cat.id} className="space-y-4">
            <CategorySection category={cat} />
          </TabsContent>
        ))}
      </Tabs>

      {/* Footer */}
      <Card>
        <CardContent className="py-4 text-center text-sm text-muted-foreground">
          {t('This documentation is OpenAPI-compatible. Generate clients using')}{' '}
          <a
            href="https://openapi-generator.tech"
            target="_blank"
            rel="noopener noreferrer"
            className="text-blue-600 hover:underline"
          >
            OpenAPI Generator
          </a>
          .
        </CardContent>
      </Card>
    </div>
  )
}
