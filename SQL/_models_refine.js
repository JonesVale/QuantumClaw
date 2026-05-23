const fs = require('fs');
const path = 'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src/components/models-view.tsx';
let c = fs.readFileSync(path, 'utf-8');

// 1. More spacious layout
c = c.replace('px-4 sm:px-6 lg:px-8 py-8', 'px-6 sm:px-8 lg:px-10 py-12');

// 2. Better hero spacing
c = c.replace('className="mb-10"', 'className="mb-16"');
c = c.replace('text-base text-muted-foreground mt-3 max-w-2xl leading-relaxed', 'text-lg text-muted-foreground mt-4 max-w-2xl leading-relaxed');

// 3. Widen action bar
c = c.replace('className="flex items-center gap-3 mb-8 flex-wrap"', 'className="flex items-center gap-4 mb-10 flex-wrap"');

// 4. Card grid bigger gaps
c = c.replace('grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-4', 'grid-cols-1 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6');
c = c.replace('rounded-xl p-5', 'rounded-xl p-6');
c = c.replace('mb-3', 'mb-4');

// 5. Bigger typography in cards
c = c.replace('text-[11px] font-medium text-muted-foreground uppercase tracking-wider', 'text-xs font-medium text-muted-foreground uppercase tracking-wider');
c = c.replace('text-base font-semibold', 'text-lg font-semibold');
// Tag sizes
c = c.replace('px-2 py-0.5 rounded text-[10px] font-medium', 'px-2 py-0.5 rounded text-xs font-medium');
c = c.replace('rounded text-[10px] bg-muted/50', 'rounded text-xs bg-muted/50');

// 6. Replace hover overlay with persistent action links at bottom
const oldOverlay = `                  {/* Hover overlay */}
                  <div className="absolute inset-0 rounded-xl bg-background/80 backdrop-blur-[1px] opacity-0 group-hover:opacity-100 transition-opacity flex items-center justify-center gap-2">
                    <button className="inline-flex items-center justify-center rounded-lg border border-input bg-background h-8 px-3 text-xs font-medium hover:bg-accent" onClick={() => openDetail(m)}>
                      {t('Details')}
                    </button>
                    <button className="inline-flex items-center justify-center rounded-lg bg-primary text-primary-foreground h-8 px-3 text-xs font-medium gap-1" onClick={() => window.location.href = auth?.user ? '/chat' : '/sign-in'}>
                      <svg xmlns="http://www.w3.org/2000/svg" width="12" height="12" viewBox="0 0 24 24" fill="currentColor"><polygon points="5 3 19 12 5 21 5 3"/></svg>
                      {t('Call')}
                    </button>
                  </div>`;

const newActions = `                  {/* Actions at bottom */}
                  <div className="flex items-center gap-2 mt-4 pt-3 border-t border-border/40">
                    <button className="text-xs text-muted-foreground hover:text-foreground font-medium transition-colors" onClick={() => openDetail(m)}>
                      {t('Details')} →
                    </button>
                    <span className="text-muted-foreground/30">·</span>
                    <button className="text-xs text-primary hover:text-primary/80 font-medium transition-colors" onClick={() => window.location.href = auth?.user ? '/chat' : '/sign-in'}>
                      {t('Call')}
                    </button>
                  </div>`;

c = c.replace(oldOverlay, newActions);

// 7. Remove 'group' class
c = c.replace('className={\'group relative rounded-xl p-6 transition-all duration-200 bg-card hover:shadow-md border-0 \' +', 'className={\'relative rounded-xl p-6 transition-all duration-200 bg-card hover:shadow-md border-0 \' +');

// 8. Better active filter tags
c = c.replace('gap-2 mb-6 flex-wrap', 'gap-3 mb-8 flex-wrap');
c = c.replace('px-3 py-1 rounded-full text-xs font-medium', 'px-3 py-1.5 rounded-full text-xs font-medium');

// 9. Sidebar spacing
c = c.replace('p-5 space-y-5', 'p-6 space-y-6');

fs.writeFileSync(path, c, 'utf-8');
console.log('Refinements applied');
