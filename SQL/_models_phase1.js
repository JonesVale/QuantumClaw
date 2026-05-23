const fs = require('fs');
const path = 'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src/routes/models.tsx';
let c = fs.readFileSync(path, 'utf-8');

// ===== 1. Provider computation: split into AI and Quantum =====
const oldP = `  const providers = useMemo(() => {
    if (catalog.length === 0) return []
    const map = new Map<string, number>()
    for (const m of catalog) {
      const p = m.provider || 'Unknown'
      map.set(p, (map.get(p) || 0) + 1)
    }
    return Array.from(map.entries()).sort((a, b) => b[1] - a[1])
  }, [catalog.length]) // eslint-disable-line`;

const newP = `  const aiProviders = useMemo(() => {
    if (catalog.length === 0) return []
    const map = new Map<string, number>()
    for (const m of catalog) {
      if (m.use_case === 'quantum') continue
      const p = m.provider || 'Unknown'
      if (p === 'Unknown' || p.startsWith('~')) continue
      map.set(p, (map.get(p) || 0) + 1)
    }
    return Array.from(map.entries()).sort((a, b) => b[1] - a[1])
  }, [catalog.length]) // eslint-disable-line

  const quantumProviders = useMemo(() => {
    if (catalog.length === 0) return []
    const map = new Map<string, number>()
    for (const m of catalog) {
      if (m.use_case !== 'quantum') continue
      const p = m.provider || 'Unknown'
      if (p === 'Unknown' || p.startsWith('~')) continue
      map.set(p, (map.get(p) || 0) + 1)
    }
    return Array.from(map.entries()).sort((a, b) => b[1] - a[1])
  }, [catalog.length]) // eslint-disable-line`;

c = c.replace(oldP, newP);

// ===== 2. Clean up: remove seriesNames, seriesFilter, collapsed.series =====
c = c.replace("  const [seriesFilter, setSeriesFilter] = useState('all')\n", "");
c = c.replace("  const [collapsed, setCollapsed] = useState<Record<string, boolean>>({\n    provider: false, aiProvider: false, quantumProvider: false, categories: false, modalities: false, context: false, series: false\n  })",
  "  const [collapsed, setCollapsed] = useState<Record<string, boolean>>({\n    aiProvider: false, quantumProvider: false, categories: false, modalities: false, context: false\n  })");

c = c.replace("    if (seriesFilter !== 'all') result = result.filter(m => m.series === seriesFilter)\n", "");
c = c.replace("setSeriesFilter('all');\n                ", "");

// Remove seriesNames useMemo
c = c.replace(/const seriesNames = useMemo\([\s\S]*?catalog\.length\)\) \/\/ eslint-disable-line\n\n/, "");

fs.writeFileSync(path, c, 'utf-8');
console.log('Phase 1 done: providers split, series removed');
