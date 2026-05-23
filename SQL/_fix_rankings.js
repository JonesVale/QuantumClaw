const fs = require('fs');
const path = 'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src/routes/rankings.tsx';
let c = fs.readFileSync(path, 'utf-8');

// 1. Clean up imports - remove Card, Badge, Skeleton (will use simple divs)
c = c.replace("import { Card, CardContent } from '@/components/ui/card'\n", "");
c = c.replace("import { Badge } from '@/components/ui/badge'\n", "");
c = c.replace("import { Skeleton } from '@/components/ui/skeleton'\n", "");

// 2. Remove seriesFilter state
c = c.replace("  const [seriesFilter, setSeriesFilter] = useState('All')\n", "");

// 3. Remove seriesFilter from filter logic and memo
c = c.replace(".filter(m => seriesFilter === 'All' || m.model.toLowerCase().includes(seriesFilter.toLowerCase()))\n    ", "");
c = c.replace("rankings, seriesFilter, ", "rankings, ");

// 4. Replace the entire return block
const returnIdx = c.indexOf("  return (");
const endIdx = c.lastIndexOf("}");
// Make sure we get the right closing
const compEnd = c.indexOf("}\n", endIdx - 100);
const finalIdx = compEnd > returnIdx ? compEnd + 1 : endIdx;

const beforeReturn = c.substring(0, returnIdx);
const afterReturn = c.substring(finalIdx + 1);

const newReturn = `  return (
    <div className="min-h-screen bg-background">
      <div className="max-w-7xl mx-auto px-6 sm:px-8 lg:px-10 py-12">
        {/* Hero */}
        <div className="mb-12">
          <div className="flex items-start justify-between">
            <div>
              <h1 className="text-3xl sm:text-4xl lg:text-5xl font-bold tracking-tight">
                <span className="bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-transparent">
                  {t('Model Rankings')}
                </span>
              </h1>
              <p className="text-lg text-muted-foreground mt-4 max-w-2xl leading-relaxed">
                {t('Model performance rankings across providers')}
              </p>
            </div>
            <button className="inline-flex items-center justify-center rounded-lg border border-input bg-background h-10 px-4 text-sm font-medium hover:bg-accent gap-2 shrink-0" onClick={() => refetch()} disabled={isFetching}>
              <svg className={\`h-4 w-4 \${isFetching ? 'animate-spin' : ''}\`} xmlns="http://www.w3.org/2000/svg" width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><path d="M21 12a9 9 0 0 0-9-9 9.75 9.75 0 0 0-6.74 2.74L3 8"/><path d="M3 3v5h5"/><path d="M3 12a9 9 0 0 0 9 9 9.75 9.75 0 0 0 6.74-2.74L21 16"/><path d="M16 21h5v-5"/></svg>
              {t('Refresh')}
            </button>
          </div>
        </div>

        {/* Tab pills */}
        <div className="flex flex-wrap gap-3 mb-10">
          {TABS.map((tab) => {
            const Icon = tab.icon;
            const active = activeTab === tab.key;
            return (
              <button
                key={tab.key}
                onClick={() => setActiveTab(tab.key)}
                className={\`inline-flex items-center gap-2 px-5 py-2.5 rounded-lg text-sm font-medium transition-all \${
                  active
                    ? 'bg-primary text-primary-foreground shadow-sm'
                    : 'bg-muted text-muted-foreground hover:bg-accent hover:text-accent-foreground'
                }\`}
              >
                <Icon className="h-4 w-4" />
                {t(tab.labelKey)}
              </button>
            );
          })}
        </div>

        {/* Rankings List */}
        {isLoading ? (
          <div className="space-y-3">
            {Array.from({ length: 6 }).map((_, i) => (
              <div key={i} className="rounded-xl p-5 bg-card animate-pulse">
                <div className="flex items-center gap-4">
                  <div className="h-10 w-10 rounded-xl bg-muted" />
                  <div className="flex-1 space-y-2">
                    <div className="h-5 w-40 rounded bg-muted" />
                    <div className="h-4 w-24 rounded bg-muted" />
                  </div>
                  <div className="h-6 w-20 rounded bg-muted" />
                </div>
              </div>
            ))}
          </div>
        ) : filteredRankings.length === 0 ? (
          <div className="text-center py-20 text-muted-foreground">
            <svg xmlns="http://www.w3.org/2000/svg" width="48" height="48" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className="mx-auto mb-4 opacity-20"><path d="M6 9H4.5a2.5 2.5 0 0 1 0-5C7 4 9 6 9 9.5v3c0 1.1-.9 2-2 2H6"/><path d="M18 9h1.5a2.5 2.5 0 0 0 0-5C17 4 15 6 15 9.5v3c0 1.1.9 2 2 2h1"/></svg>
            {t('No data available')}
          </div>
        ) : (
          <>
          <div className="space-y-3">
            {displayed.map((item, index) => {
              const rank = index + 1;
              const isTop3 = rank <= 3;
              return (
                <div key={\`\${item.provider}-\${item.model}\`} className={\`rounded-xl p-5 transition-all hover:shadow-md bg-card border-0 \${isTop3 ? 'ring-1 ring-yellow-500/20' : ''}\`}>
                  <div className="flex items-center gap-4">
                    {/* Rank */}
                    <div className={\`flex items-center justify-center w-10 h-10 rounded-xl font-bold text-white shrink-0 \${
                      isTop3 ? RANK_STYLES[rank - 1] : 'bg-muted text-muted-foreground'
                    }\`}>
                      {isTop3 ? <svg xmlns="http://www.w3.org/2000/svg" width="20" height="20" viewBox="0 0 24 24" fill="currentColor"><path d="M12 2l3.09 6.26L22 9.27l-5 4.87 1.18 6.88L12 17.77l-6.18 3.25L7 14.14 2 9.27l6.91-1.01L12 2z"/></svg> : rank}
                    </div>

                    {/* Model info */}
                    <div className="flex-1 min-w-0">
                      <div className="flex items-center gap-2 flex-wrap">
                        <span className="font-semibold text-base">{item.model}</span>
                        <span className="text-xs px-2 py-0.5 rounded bg-muted text-muted-foreground">{item.provider}</span>
                      </div>
                      <div className="text-xs text-muted-foreground mt-0.5">{item.channel_name}</div>
                    </div>

                    {/* Stats */}
                    <div className="flex items-center gap-4 sm:gap-6 text-right shrink-0">
                      {activeTab === 'request_count_7d' && (
                        <div>
                          <div className="text-sm font-semibold">{formatRequests(item.request_count_7d)}</div>
                          <div className="text-[10px] text-muted-foreground uppercase tracking-wider">{t('Requests')}</div>
                        </div>
                      )}
                      {activeTab === 'tokens_7d' && (
                        <div>
                          <div className="text-sm font-semibold">{formatTokens(item.tokens_7d)}</div>
                          <div className="text-[10px] text-muted-foreground uppercase tracking-wider">{t('Tokens')}</div>
                        </div>
                      )}
                      {activeTab === 'avg_speed_ms' && (
                        <div>
                          <div className="text-sm font-semibold">{item.avg_speed_ms}ms</div>
                          <div className="text-[10px] text-muted-foreground uppercase tracking-wider">{t('Avg Speed')}</div>
                        </div>
                      )}
                      {activeTab === 'price_per_1k' && (
                        <div>
                          <div className="text-sm font-semibold">\${item.price_per_1k.toFixed(4)}</div>
                          <div className="text-[10px] text-muted-foreground uppercase tracking-wider">/1K {t('Tokens')}</div>
                        </div>
                      )}

                      {/* Trend */}
                      <span className={\`inline-flex items-center gap-0.5 text-xs font-medium px-2 py-1 rounded-full \${
                        item.trend_percent >= 0
                          ? 'bg-green-100 text-green-700 dark:bg-green-900/30 dark:text-green-400'
                          : 'bg-red-100 text-red-700 dark:bg-red-900/30 dark:text-red-400'
                      }\`}>
                        {item.trend_percent >= 0
                          ? <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><polyline points="23 6 13.5 15.5 8.5 10.5 1 18"/><polyline points="17 6 23 6 23 12"/></svg>
                          : <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2"><polyline points="23 18 13.5 8.5 8.5 13.5 1 6"/><polyline points="17 18 23 18 23 12"/></svg>
                        }
                        {Math.abs(item.trend_percent)}%
                      </span>
                    </div>
                  </div>
                </div>
              );
            })}
          </div>
          {visibleCount < filteredRankings.length && (
            <div className="flex justify-center mt-10">
              <button className="inline-flex items-center justify-center rounded-xl border border-input bg-background w-full max-w-sm py-3 text-sm font-medium hover:bg-accent"
                onClick={() => setVisibleCount(c => c + PAGE_STEP)}>
                {t('Load More')} ({filteredRankings.length - visibleCount} {t('remaining')})
              </button>
            </div>
          )}
          </>
        )}
      </div>
    </div>
  );`;

c = beforeReturn + newReturn + afterReturn;

// Clean up duplicate newlines
c = c.replace(/\n{4,}/g, '\n\n\n');

fs.writeFileSync(path, c, 'utf-8');
console.log('Rankings page redesigned');
