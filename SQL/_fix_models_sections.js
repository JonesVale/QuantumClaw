const fs = require('fs');
const path = 'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src/routes/models.tsx';
let c = fs.readFileSync(path, 'utf-8');

// Replace single Provider section with AI + Quantum sections
const oldSection = `          {/* Provider */}
          <section className="border-b">
            <button onClick={() => setCollapsed(c => ({...c, provider: !c.provider}))} className="flex items-center justify-between w-full px-4 py-2.5 text-sm font-semibold text-muted-foreground hover:bg-muted/30 transition-colors">
              <span>{t('Provider')}</span>
              {collapsed.provider ? <ChevronRight className="h-3.5 w-3.5" /> : <ChevronDown className="h-3.5 w-3.5" />}
            </button>
            {!collapsed.provider && <div className="px-4 pb-2 space-y-0.5">
              <button onClick={() => setProviderFilter('')} className={cn('w-full text-left px-3 py-2 text-sm rounded transition-colors', !providerFilter ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>
                {t('All Providers')} <span className="text-xs text-muted-foreground">({catalog.length})</span>
              </button>
              {providers.slice(0, 15).map(([p, cnt]) => (
                <button key={p} onClick={() => setProviderFilter(p)} className={cn('w-full text-left px-3 py-2 text-sm rounded transition-colors flex items-center justify-between', providerFilter === p ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>
                  <span>{p}</span>
                  <span className="text-xs text-muted-foreground">{cnt}</span>
                </button>
              ))}
              {providers.length > 15 && (
                <details className="group">
                  <summary className="w-full text-left px-3 py-1.5 text-xs text-muted-foreground hover:bg-muted/50 rounded cursor-pointer list-none">
                    <span className="group-open:hidden">{t('Show more...')} ({providers.length - 15})</span>
                    <span className="hidden group-open:inline">{t('Show less')}</span>
                  </summary>
                  <div className="space-y-0.5">
                    {providers.slice(15).map(([p, cnt]) => (
                      <button key={p} onClick={() => setProviderFilter(p)} className={cn('w-full text-left px-3 py-2 text-sm rounded transition-colors flex items-center justify-between', providerFilter === p ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>
                        <span>{p}</span>
                        <span className="text-xs text-muted-foreground">{cnt}</span>
                      </button>
                    ))}
                  </div>
                </details>
              )}
            </div>}
          </section>`;

const newSection = `          {/* AI Providers */}
          <section className="border-b">
            <button onClick={() => setCollapsed(c => ({...c, provider: !c.provider}))} className="flex items-center justify-between w-full px-4 py-2.5 text-sm font-semibold text-muted-foreground hover:bg-muted/30 transition-colors">
              <span>{t('AI Providers')}</span>
              {collapsed.provider ? <ChevronRight className="h-3.5 w-3.5" /> : <ChevronDown className="h-3.5 w-3.5" />}
            </button>
            {!collapsed.provider && <div className="px-4 pb-2 space-y-0.5">
              <button onClick={() => setProviderFilter('')} className={cn('w-full text-left px-3 py-2 text-sm rounded transition-colors', !providerFilter ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>
                {t('All Providers')} <span className="text-xs text-muted-foreground">({catalog.filter(m => m.use_case !== 'quantum').length})</span>
              </button>
              {aiProviders.slice(0, 15).map(([p, cnt]) => (
                <button key={p} onClick={() => setProviderFilter(p)} className={cn('w-full text-left px-3 py-2 text-sm rounded transition-colors flex items-center justify-between', providerFilter === p ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>
                  <span>{p}</span>
                  <span className="text-xs text-muted-foreground">{cnt}</span>
                </button>
              ))}
              {aiProviders.length > 15 && (
                <details className="group">
                  <summary className="w-full text-left px-3 py-1.5 text-xs text-muted-foreground hover:bg-muted/50 rounded cursor-pointer list-none">
                    <span className="group-open:hidden">{t('Show more...')} ({aiProviders.length - 15})</span>
                    <span className="hidden group-open:inline">{t('Show less')}</span>
                  </summary>
                  <div className="space-y-0.5">
                    {aiProviders.slice(15).map(([p, cnt]) => (
                      <button key={p} onClick={() => setProviderFilter(p)} className={cn('w-full text-left px-3 py-2 text-sm rounded transition-colors flex items-center justify-between', providerFilter === p ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>
                        <span>{p}</span>
                        <span className="text-xs text-muted-foreground">{cnt}</span>
                      </button>
                    ))}
                  </div>
                </details>
              )}
            </div>}
          </section>
          {/* Quantum Providers */}
          {quantumProviders.length > 0 && <section className="border-b">
            <button onClick={() => setCollapsed(c => ({...c, provider: !c.provider}))} className="flex items-center justify-between w-full px-4 py-2.5 text-sm font-semibold text-muted-foreground hover:bg-muted/30 transition-colors">
              <span className="flex items-center gap-2">
                <div className="w-4 h-4 rounded bg-gradient-to-br from-violet-500 to-purple-600 flex items-center justify-center shrink-0"><svg className="h-2.5 w-2.5 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><circle cx="12" cy="12" r="3"/><path d="M12 1v4M12 19v4M4.22 4.22l2.83 2.83M16.95 16.95l2.83 2.83M1 12h4M19 12h4M4.22 19.78l2.83-2.83M16.95 7.05l2.83-2.83"/></svg></div>
                {'Quantum Providers'}
              </span>
              {collapsed.provider ? <ChevronRight className="h-3.5 w-3.5" /> : <ChevronDown className="h-3.5 w-3.5" />}
            </button>
            {!collapsed.provider && <div className="px-4 pb-2 space-y-0.5">
              <button onClick={() => setProviderFilter('')} className={cn('w-full text-left px-3 py-2 text-sm rounded transition-colors', !providerFilter ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>
                {t('All Quantum')} <span className="text-xs text-muted-foreground">({catalog.filter(m => m.use_case === 'quantum').length})</span>
              </button>
              {quantumProviders.map(([p, cnt]) => (
                <button key={p} onClick={() => setProviderFilter(p)} className={cn('w-full text-left px-3 py-2 text-sm rounded transition-colors flex items-center justify-between', providerFilter === p ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>
                  <span className="flex items-center gap-2">
                    <div className="w-3 h-3 rounded bg-gradient-to-br from-violet-500 to-purple-600 flex items-center justify-center shrink-0"><svg className="h-1.5 w-1.5 text-white" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="3"><circle cx="12" cy="12" r="2"/></svg></div>
                    {p}
                  </span>
                  <span className="text-xs text-muted-foreground">{cnt}</span>
                </button>
              ))}
            </div>}
          </section>}`;

c = c.replace(oldSection, newSection);

// Update collapsed state default - remove series
c = c.replace(
  "    provider: false, categories: false, modalities: false, context: false",
  "    provider: false, categories: false, modalities: false, context: false"
);

fs.writeFileSync(path, c, 'utf-8');
console.log('models.tsx updated - AI + Quantum provider sections');
