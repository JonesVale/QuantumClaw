import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'

export const Route = createFileRoute('/developer')({
  component: DeveloperPage,
})

const SDK_LANGUAGES = [
  {
    name: 'Python',
    icon: '🐍',
    install: 'pip install quantumclaw',
    usage: `from quantumclaw import QuantumClaw

client = QuantumClaw(api_key="sk-xxx")
response = client.chat_completions(
    model="gpt-4o",
    messages=[{"role":"user","content":"Hello"}]
)
print(response.choices[0].message.content)`,
    link: 'https://github.com/quantumclaw/quantumclaw',
  },
  {
    name: 'Node.js',
    icon: '🟢',
    install: 'npm install quantumclaw',
    usage: `import { QuantumClaw } from 'quantumclaw'

const client = new QuantumClaw({ apiKey: 'sk-xxx' })
const response = await client.chatCompletions({
  model: 'gpt-4o',
  messages: [{ role: 'user', content: 'Hello' }],
})`,
    link: 'https://github.com/quantumclaw/quantumclaw',
  },
  {
    name: 'Go',
    icon: '🔵',
    install: 'go get github.com/quantumclaw/quantumclaw-sdk-go',
    usage: `import quantumclaw "github.com/quantumclaw/quantumclaw-sdk-go"

client := quantumclaw.NewClient("sk-xxx", "")
resp, err := client.ChatCompletion(ctx, &quantumclaw.ChatCompletionRequest{
    Model:    "gpt-4o",
    Messages: []quantumclaw.ChatMessage{{Role: "user", Content: "Hello"}},
})`,
    link: 'https://github.com/quantumclaw/quantumclaw',
  },
  {
    name: 'cURL',
    icon: '🔧',
    install: '',
    usage: `curl https://your-instance.com/v1/chat/completions \\
  -H "Authorization: Bearer sk-xxx" \\
  -H "Content-Type: application/json" \\
  -d '{
    "model": "gpt-4o",
    "messages": [{"role":"user","content":"Hello"}]
  }'`,
    link: '',
  },
]

export function DeveloperPage() {
  const { t, language } = useT()

  return (
    <div className="max-w-5xl mx-auto px-4 py-8">
      <h1 className="text-3xl font-bold mb-2">{t('dev_title') || 'Developer Portal'}</h1>
      <p className="text-muted-foreground mb-8">
        {t('dev_subtitle') || 'Build with QuantumClaw — the universal AI API gateway'}
      </p>

      {/* Quick Start */}
      <div className="mb-12">
        <h2 className="text-2xl font-bold mb-4">{t('dev_quickstart') || 'Quick Start'}</h2>
        <div className="p-4 bg-muted rounded-lg mb-4">
          <h3 className="font-medium mb-2">{t('dev_auth') || '1. Authentication'}</h3>
          <p className="text-sm text-muted-foreground mb-2">
            {t('dev_auth_desc') || 'All API requests require an API Key. Add it to the Authorization header:'}
          </p>
          <pre className="bg-card p-3 rounded text-sm overflow-x-auto">Authorization: Bearer sk-your-api-key</pre>
        </div>

        <div className="p-4 bg-muted rounded-lg">
          <h3 className="font-medium mb-2">{t('dev_base_url') || '2. Base URL'}</h3>
          <p className="text-sm text-muted-foreground mb-2">
            {t('dev_base_url_desc') || 'Use your instance URL. All endpoints follow the OpenAI API format:'}
          </p>
          <pre className="bg-card p-3 rounded text-sm overflow-x-auto">POST https://your-instance.com/v1/chat/completions</pre>
        </div>
      </div>

      {/* SDKs */}
      <div className="mb-12">
        <h2 className="text-2xl font-bold mb-4">{t('dev_sdks') || 'SDKs & Client Libraries'}</h2>
        <div className="grid gap-6 md:grid-cols-2">
          {SDK_LANGUAGES.filter(s => s.install).map(sdk => (
            <div key={sdk.name} className="border rounded-lg p-4">
              <div className="flex items-center gap-2 mb-3">
                <span className="text-2xl">{sdk.icon}</span>
                <h3 className="font-bold text-lg">{sdk.name}</h3>
              </div>
              <div className="mb-3">
                <code className="text-sm bg-muted px-2 py-1 rounded">{sdk.install}</code>
              </div>
              <pre className="bg-card p-3 rounded text-xs overflow-x-auto mb-2">{sdk.usage}</pre>
              <a href={sdk.link} target="_blank" rel="noreferrer"
                className="text-sm text-primary hover:underline">
                {t('dev_view_on_github') || 'View on GitHub'} →
              </a>
            </div>
          ))}
        </div>

        {/* cURL section */}
        <div className="mt-6 border rounded-lg p-4">
          <div className="flex items-center gap-2 mb-3">
            <span className="text-2xl">🔧</span>
            <h3 className="font-bold text-lg">cURL</h3>
          </div>
          <pre className="bg-card p-3 rounded text-xs overflow-x-auto">{SDK_LANGUAGES[3].usage}</pre>
        </div>
      </div>

      {/* API Reference */}
      <div className="mb-12">
        <h2 className="text-2xl font-bold mb-4">{t('dev_api_ref') || 'API Reference'}</h2>
        <div className="border rounded-lg overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="bg-muted">
                <th className="px-4 py-2 text-left font-medium">{t('dev_endpoint') || 'Endpoint'}</th>
                <th className="px-4 py-2 text-left font-medium">{t('dev_method') || 'Method'}</th>
                <th className="px-4 py-2 text-left font-medium">{t('dev_description') || 'Description'}</th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {[
                { path: '/v1/chat/completions', method: 'POST', desc: t('dev_api_chat') || 'Chat completions (streaming & non-streaming)' },
                { path: '/v1/models', method: 'GET', desc: t('dev_api_models') || 'List available models' },
                { path: '/api/user/balance', method: 'GET', desc: t('dev_api_balance') || 'Query account balance' },
                { path: '/api/token/', method: 'GET', desc: t('dev_api_tokens') || 'List API keys' },
                { path: '/api/token/', method: 'POST', desc: t('dev_api_create_token') || 'Create API key' },
              ].map((api, i) => (
                <tr key={i} className="hover:bg-muted/50">
                  <td className="px-4 py-2 font-mono text-xs">{api.path}</td>
                  <td className="px-4 py-2">
                    <span className={`text-xs px-1.5 py-0.5 rounded ${
                      api.method === 'GET' ? 'bg-green-100 text-green-700' :
                      api.method === 'POST' ? 'bg-blue-100 text-blue-700' :
                      'bg-orange-100 text-orange-700'
                    }`}>{api.method}</span>
                  </td>
                  <td className="px-4 py-2 text-muted-foreground">{api.desc}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
        <div className="mt-4">
          <a href="/api/swagger/index.html" target="_blank"
            className="text-primary hover:underline text-sm">
            {t('dev_swagger') || 'View full API documentation (Swagger UI)'} →
          </a>
        </div>
      </div>

      {/* Rate Limits */}
      <div className="mb-12">
        <h2 className="text-2xl font-bold mb-4">{t('dev_limits') || 'Rate Limits'}</h2>
        <div className="p-4 border rounded-lg text-sm">
          <p className="text-muted-foreground mb-2">
            {t('dev_limits_desc') || 'Rate limits vary by subscription plan:'}
          </p>
          <ul className="list-disc pl-5 space-y-1 text-muted-foreground">
            <li>{t('dev_limit_free') || 'Free: 10 requests/min'}</li>
            <li>{t('dev_limit_pro') || 'Pro: 100 requests/min'}</li>
            <li>{t('dev_limit_enterprise') || 'Enterprise: Custom'}</li>
          </ul>
        </div>
      </div>

      {/* Error Codes */}
      <div>
        <h2 className="text-2xl font-bold mb-4">{t('dev_errors') || 'Error Codes'}</h2>
        <div className="border rounded-lg overflow-hidden">
          <table className="w-full text-sm">
            <thead>
              <tr className="bg-muted">
                <th className="px-4 py-2 text-left font-medium">{t('dev_code') || 'Code'}</th>
                <th className="px-4 py-2 text-left font-medium">{t('dev_meaning') || 'Meaning'}</th>
              </tr>
            </thead>
            <tbody className="divide-y">
              {[
                { code: '401', meaning: t('dev_err_401') || 'Invalid or expired API key' },
                { code: '429', meaning: t('dev_err_429') || 'Rate limit exceeded' },
                { code: '402', meaning: t('dev_err_402') || 'Insufficient balance' },
                { code: '400', meaning: t('dev_err_400') || 'Invalid request parameters' },
                { code: '500', meaning: t('dev_err_500') || 'Internal server error' },
              ].map((err, i) => (
                <tr key={i} className="hover:bg-muted/50">
                  <td className="px-4 py-2 font-mono">{err.code}</td>
                  <td className="px-4 py-2 text-muted-foreground">{err.meaning}</td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      </div>
    </div>
  )
}
