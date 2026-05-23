const fs = require('fs');
const path = 'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src/routes/index.tsx';
let c = fs.readFileSync(path, 'utf-8');

// 1. Remove Card, Badge, Skeleton imports
c = c.replace("import { Badge } from '@/components/ui/badge'\n", "");
c = c.replace("import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card'\n", "");
c = c.replace("import { Skeleton } from '@/components/ui/skeleton'\n", "");
c = c.replace("import { Avatar, AvatarFallback } from '@/components/ui/avatar'\n", "");

// 2. Replace full return block
const retIdx = c.indexOf("  return (");
const endIdx = c.lastIndexOf("}");

// Find component's closing
const compEnd = c.lastIndexOf("}\n", endIdx - 1);
const afterIdx = compEnd > retIdx ? compEnd + 1 : endIdx;

const beforeRet = c.substring(0, retIdx);
const after = c.substring(afterIdx);

const newRet = `  return (
    <div className="min-h-screen bg-background">
      {/* ===== NAV (inline, simplified) ===== */}
      <header className="sticky top-0 z-50 border-b bg-background/95 backdrop-blur supports-[backdrop-filter]:bg-background/60">
        <div className="max-w-7xl mx-auto px-6 sm:px-8 lg:px-10">
          <div className="flex items-center justify-between h-16">
            <div className="flex items-center gap-8">
              <Link to="/" className="text-xl font-bold tracking-tight">
                <span className="bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-transparent">
                  {t('QuantumClaw')}
                </span>
              </Link>
              <nav className="hidden md:flex items-center gap-6">
                <Link to="/models" className="text-sm font-medium text-muted-foreground hover:text-foreground transition-colors">{t('Models')}</Link>
                <Link to="/rankings" className="text-sm font-medium text-muted-foreground hover:text-foreground transition-colors">{t('Rankings')}</Link>
                <Link to="/pricing" className="text-sm font-medium text-muted-foreground hover:text-foreground transition-colors">{t('Pricing')}</Link>
                <Link to="/playground" className="text-sm font-medium text-muted-foreground hover:text-foreground transition-colors">{t('Playground')}</Link>
              </nav>
            </div>
            <div className="flex items-center gap-3">
              <Link to="/sign-in">
                <button className="inline-flex items-center justify-center rounded-lg border border-input bg-background h-9 px-4 text-sm font-medium hover:bg-accent">
                  {t('Sign In')}
                </button>
              </Link>
              <Link to="/sign-in">
                <button className="inline-flex items-center justify-center rounded-lg bg-primary text-primary-foreground h-9 px-4 text-sm font-medium hover:bg-primary/90">
                  {t('Get Started')}
                </button>
              </Link>
            </div>
          </div>
        </div>
      </header>

      {/* ===== HERO ===== */}
      <section className="relative overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-br from-blue-50 via-white to-purple-50 dark:from-blue-950/20 dark:via-slate-950 dark:to-purple-950/20" />
        <div className="max-w-7xl mx-auto px-6 sm:px-8 lg:px-10 py-20 sm:py-28 lg:py-36 relative">
          <div className="max-w-3xl">
            <h1 className="text-4xl sm:text-5xl lg:text-6xl font-bold tracking-tight leading-tight">
              <span className="bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-transparent">
                {t('QuantumClaw')}
              </span>
              <br />
              <span className="text-foreground">
                {t('AI API Gateway & Token Distribution Platform')}
              </span>
            </h1>
            <p className="text-lg sm:text-xl text-muted-foreground mt-6 max-w-2xl leading-relaxed">
              {t('Access 400+ AI models and quantum computing resources through a single, unified interface. OpenAI SDK compatible.')}
            </p>
            <div className="flex items-center gap-4 mt-8">
              <Link to="/models">
                <button className="inline-flex items-center justify-center rounded-lg bg-primary text-primary-foreground h-12 px-6 text-sm font-semibold hover:bg-primary/90 gap-2">
                  {t('Browse Models')}
                  <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M5 12h14"/><path d="m12 5 7 7-7 7"/></svg>
                </button>
              </Link>
              <Link to="/sign-in">
                <button className="inline-flex items-center justify-center rounded-lg border border-input bg-background h-12 px-6 text-sm font-semibold hover:bg-accent">
                  {t('Get Started Free')}
                </button>
              </Link>
            </div>
          </div>
        </div>
      </section>

      {/* ===== STATS BAR ===== */}
      <section className="border-y bg-muted/30">
        <div className="max-w-7xl mx-auto px-6 sm:px-8 lg:px-10 py-12">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-8">
            {[
              { icon: 'database', value: '400+', label: t('AI Models') },
              { icon: 'globe', value: '60+', label: t('Providers') },
              { icon: 'cpu', value: '5', label: t('Quantum Backends') },
              { icon: 'zap', value: '99.9%', label: t('Uptime SLA') },
            ].map(s => (
              <div key={s.label} className="text-center">
                <div className="text-3xl sm:text-4xl font-bold text-foreground">{s.value}</div>
                <div className="text-sm text-muted-foreground mt-1">{s.label}</div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ===== FEATURES / USE CASES ===== */}
      <section className="py-20 sm:py-28">
        <div className="max-w-7xl mx-auto px-6 sm:px-8 lg:px-10">
          <div className="text-center mb-14">
            <h2 className="text-3xl sm:text-4xl font-bold tracking-tight">{t('One API, Every Model')}</h2>
            <p className="text-lg text-muted-foreground mt-4 max-w-2xl mx-auto">
              {t('Access leading AI models through a single OpenAI-compatible API endpoint')}
            </p>
          </div>
          <div className="grid grid-cols-1 sm:grid-cols-2 lg:grid-cols-4 gap-6">
            {[
              { icon: 'message-square', title: t('Chat & Assistant'), desc: t('GPT-4, Claude, Gemini and more chat models'), color: 'from-blue-500 to-blue-600' },
              { icon: 'code', title: t('Code Generation'), desc: t('DeepSeek Coder, Code Llama, Qwen Coder'), color: 'from-green-500 to-green-600' },
              { icon: 'brain', title: t('Reasoning'), desc: t('o1, o3, Claude Opus for complex tasks'), color: 'from-purple-500 to-purple-600' },
              { icon: 'atom', title: t('Quantum Computing'), desc: t('IonQ, IBM, Rigetti, AWS Braket backends'), color: 'from-violet-500 to-purple-600' },
            ].map(f => (
              <div key={f.title} className="rounded-xl p-6 bg-card border-0 hover:shadow-md transition-all duration-200">
                <div className={\`w-10 h-10 rounded-lg bg-gradient-to-br \${f.color} flex items-center justify-center mb-4\`}>
                  <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className="text-white"><circle cx="12" cy="12" r="3"/><path d="M12 1v4M12 19v4M4.22 4.22l2.83 2.83M16.95 16.95l2.83 2.83M1 12h4M19 12h4M4.22 19.78l2.83-2.83M16.95 7.05l2.83-2.83"/></svg>
                </div>
                <h3 className="text-base font-semibold mb-2">{f.title}</h3>
                <p className="text-sm text-muted-foreground leading-relaxed">{f.desc}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* ===== PROVIDERS SHOWCASE ===== */}
      <section className="py-20 sm:py-28 bg-muted/30">
        <div className="max-w-7xl mx-auto px-6 sm:px-8 lg:px-10">
          <div className="text-center mb-14">
            <h2 className="text-3xl sm:text-4xl font-bold tracking-tight">{t('Major Providers')}</h2>
            <p className="text-lg text-muted-foreground mt-4 max-w-2xl mx-auto">
              {t('Browse models from the world\'s leading AI companies')}
            </p>
          </div>
          <div className="grid grid-cols-2 sm:grid-cols-3 lg:grid-cols-6 gap-4">
            {['OpenAI', 'Anthropic', 'Google', 'DeepSeek', 'Meta', 'Mistral'].map(p => (
              <Link key={p} to="/models" className="rounded-xl p-5 bg-card border-0 hover:shadow-md transition-all text-center">
                <div className="font-semibold text-sm">{p}</div>
                <div className="text-xs text-muted-foreground mt-1">{t('View models')}</div>
              </Link>
            ))}
          </div>
        </div>
      </section>

      {/* ===== CTA ===== */}
      <section className="py-20 sm:py-28">
        <div className="max-w-7xl mx-auto px-6 sm:px-8 lg:px-10">
          <div className="rounded-2xl bg-gradient-to-br from-blue-600 via-purple-600 to-pink-600 p-10 sm:p-14 text-center">
            <h2 className="text-3xl sm:text-4xl font-bold text-white">{t('Ready to get started?')}</h2>
            <p className="text-lg text-white/80 mt-4 max-w-xl mx-auto">
              {t('Create your free account and get instant access to 400+ AI models and quantum computing resources.')}
            </p>
            <div className="flex items-center justify-center gap-4 mt-8">
              <Link to="/sign-in">
                <button className="inline-flex items-center justify-center rounded-lg bg-white text-blue-600 h-12 px-6 text-sm font-semibold hover:bg-white/90 gap-2">
                  {t('Create Free Account')}
                  <svg xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M5 12h14"/><path d="m12 5 7 7-7 7"/></svg>
                </button>
              </Link>
              <Link to="/models">
                <button className="inline-flex items-center justify-center rounded-lg border border-white/30 text-white h-12 px-6 text-sm font-semibold hover:bg-white/10">
                  {t('Browse Catalog')}
                </button>
              </Link>
            </div>
          </div>
        </div>
      </section>

      {/* ===== FOOTER ===== */}
      <footer className="border-t bg-muted/30 py-12">
        <div className="max-w-7xl mx-auto px-6 sm:px-8 lg:px-10">
          <div className="grid grid-cols-2 md:grid-cols-4 gap-8">
            <div>
              <h4 className="text-sm font-semibold mb-4">{t('Platform')}</h4>
              <div className="space-y-2">
                <Link to="/models" className="block text-sm text-muted-foreground hover:text-foreground">{t('Models')}</Link>
                <Link to="/rankings" className="block text-sm text-muted-foreground hover:text-foreground">{t('Rankings')}</Link>
                <Link to="/pricing" className="block text-sm text-muted-foreground hover:text-foreground">{t('Pricing')}</Link>
              </div>
            </div>
            <div>
              <h4 className="text-sm font-semibold mb-4">{t('Resources')}</h4>
              <div className="space-y-2">
                <Link to="/playground" className="block text-sm text-muted-foreground hover:text-foreground">{t('Playground')}</Link>
                <Link to="/about" className="block text-sm text-muted-foreground hover:text-foreground">{t('About')}</Link>
              </div>
            </div>
            <div>
              <h4 className="text-sm font-semibold mb-4">{t('Developers')}</h4>
              <div className="space-y-2">
                <Link to="/api-docs" className="block text-sm text-muted-foreground hover:text-foreground">{t('API Docs')}</Link>
              </div>
            </div>
            <div>
              <h4 className="text-sm font-semibold mb-4">{t('Company')}</h4>
              <div className="space-y-2">
                <span className="block text-sm text-muted-foreground">{t('QuantumClaw')}</span>
              </div>
            </div>
          </div>
          <div className="border-t mt-10 pt-8 text-center text-sm text-muted-foreground">
            &copy; 2026 {t('QuantumClaw. All Rights Reserved.')}
          </div>
        </div>
      </footer>
    </div>
  );`;

c = beforeRet + newRet + after;

// Clean up empty lines
c = c.replace(/\n{4,}/g, '\n\n\n');

fs.writeFileSync(path, c, 'utf-8');
console.log('Index page redesigned (AWS homepage style)');
