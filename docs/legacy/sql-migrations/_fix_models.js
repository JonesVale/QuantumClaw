const fs = require('fs');
const path = 'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src/routes/models.tsx';
let c = fs.readFileSync(path, 'utf-8');

// 1. Remove seriesFilter state declaration and its line
c = c.replace(
  "  const [seriesFilter, setSeriesFilter] = useState('all')\n",
  ""
);

// 2. Remove seriesNames useMemo (lines 87-94)
c = c.replace(
  "  const seriesNames = useMemo(() => {\n    if (catalog.length === 0) return []\n    const set = new Set<string>()\n    for (const m of catalog) {\n      if (m.series) set.add(m.series)\n    }\n    return Array.from(set).sort()\n  }, [catalog.length]) // eslint-disable-line\n  \n",
  ""
);

// 3. Remove seriesFilter from filter pipeline
c = c.replace(
  "    if (seriesFilter !== 'all') result = result.filter(m => m.series === seriesFilter)\n",
  ""
);

// 4. Remove seriesFilter from filtered useMemo dependency
c = c.replace(
  "  }, [catalog, search, useCaseFilter, providerFilter, seriesFilter, modalityFilter, contextFilter, sortBy])",
  "  }, [catalog, search, useCaseFilter, providerFilter, modalityFilter, contextFilter, sortBy])"
);

// 5. Remove seriesFilter from useEffect dependency
c = c.replace(
  "  useEffect(() => setVisibleCount(PAGE_STEP), [search, useCaseFilter, providerFilter, seriesFilter, modalityFilter, contextFilter, sortBy])",
  "  useEffect(() => setVisibleCount(PAGE_STEP), [search, useCaseFilter, providerFilter, modalityFilter, contextFilter, sortBy])"
);

// 6. Remove Series section JSX (lines ~233-245)
c = c.replace(
  "          {/* Series */}\n          <section className=\"border-b\">\n            <button onClick={() => setCollapsed(c => ({...c, series: !c.series}))} className=\"flex items-center justify-between w-full px-4 py-2.5 text-sm font-semibold text-muted-foreground hover:bg-muted/30 transition-colors\">\n              <span>{t('Series')}</span>\n              {collapsed.series ? <ChevronRight className=\"h-3.5 w-3.5\" /> : <ChevronDown className=\"h-3.5 w-3.5\" />}\n            </button>\n            {!collapsed.series && <div className=\"px-4 pb-2 space-y-0.5\">\n              <button onClick={() => setSeriesFilter('all')} className={cn('w-full text-left px-3 py-2 text-sm rounded transition-colors', seriesFilter === 'all' ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>{t('All')}</button>\n              {seriesNames.map(s => (\n                <button key={s} onClick={() => setSeriesFilter(s)} className={cn('w-full text-left px-3 py-2 text-sm rounded transition-colors', seriesFilter === s ? 'bg-accent font-medium' : 'hover:bg-muted/50 text-muted-foreground')}>{s}</button>\n              ))}\n            </div>}\n          </section>\n",
  ""
);

// 7. Remove setSeriesFilter from reset button
c = c.replace(
  "                setSearch(''); setUseCaseFilter('all'); setSeriesFilter('all');\n                setModalityFilter(''); setContextFilter(''); setProviderFilter('');",
  "                setSearch(''); setUseCaseFilter('all');\n                setModalityFilter(''); setContextFilter(''); setProviderFilter('');"
);

// 8. Remove series from collapsed initial state
c = c.replace(
  "    provider: false, categories: false, modalities: false, context: false, series: false\n  })",
  "    provider: false, categories: false, modalities: false, context: false\n  })"
);

fs.writeFileSync(path, c, 'utf-8');
console.log('models.tsx updated - Series removed');
