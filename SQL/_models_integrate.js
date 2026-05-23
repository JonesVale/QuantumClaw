const fs = require('fs');
const path = 'H:/AiData/openclaw/workspace/QuantumClaw/web/default/src/routes/models.tsx';
let c = fs.readFileSync(path, 'utf-8');

// 1. Add import for ModelsPageView
c = c.replace(
  "import { ModelComparisonDialog } from '@/components/model-comparison-dialog'",
  "import { ModelComparisonDialog } from '@/components/model-comparison-dialog'\nimport { ModelsPageView } from '@/components/models-view'"
);

// 2. Remove unused imports (small optimization)
c = c.replace("import { Badge } from '@/components/ui/badge'\n", "");
c = c.replace("import { Card, CardContent } from '@/components/ui/card'\n", "");
c = c.replace("import { ScrollArea } from '@/components/ui/scroll-area'\n", "");
c = c.replace("import { Input } from '@/components/ui/input'\n", "");

// 3. Clean up lucide imports - keep only what's used by useCaseLabels
c = c.replace(
  "  Search, SlidersHorizontal, X, ArrowUpDown,\n  ChevronRight, ChevronDown, MessageSquare, Code, Brain,\n  Image, Cpu, Atom, Play,\n  CheckSquare, BarChart3\n} from 'lucide-react'",
  "  MessageSquare, Code, Brain, Image, Atom\n} from 'lucide-react'"
);

// 4. Clean up seriesFilter state
c = c.replace("  const [seriesFilter, setSeriesFilter] = useState('all')\n", "");

// 5. Remove seriesFilter from filter pipeline
c = c.replace("    if (seriesFilter !== 'all') result = result.filter(m => m.series === seriesFilter)\n", "");

// 6. Remove seriesFilter from dependencies
c = c.replace(", seriesFilter", "");

// 7. Remove setSeriesFilter from reset
c = c.replace("setSeriesFilter('all');\n                ", "");

// 8. Replace the entire function return block
// Find the return statement
const retIdx = c.indexOf("\n  return (");
const funEnd = c.lastIndexOf("</div>") + 6;

// The part from return to end of function
const beforeRet = c.substring(0, retIdx);
const afterFun = c.substring(funEnd);

// New: keep all state/logic, replace only the return with view component
const newBody = `\n  return (
    <ModelsPageView
      t={t}
      catalog={catalog}
      search={search}
      setSearch={setSearch}
      providerFilter={providerFilter}
      setProviderFilter={setProviderFilter}
      useCaseFilter={useCaseFilter}
      setUseCaseFilter={setUseCaseFilter}
      contextFilter={contextFilter}
      setContextFilter={setContextFilter}
      modalityFilter={modalityFilter}
      setModalityFilter={setModalityFilter}
      sortBy={sortBy}
      setSortBy={setSortBy}
      sidebarOpen={sidebarOpen}
      setSidebarOpen={setSidebarOpen}
      collapsed={collapsed}
      setCollapsed={setCollapsed}
      aiProviders={aiProviders}
      quantumProviders={quantumProviders}
      useCaseLabels={useCaseLabels}
      filtered={filtered}
      displayed={displayed}
      selectedModels={selectedModels}
      toggleSelect={toggleSelect}
      selectedModelsList={selectedModelsList}
      comparisonOpen={comparisonOpen}
      setComparisonOpen={setComparisonOpen}
      openDetail={openDetail}
      auth={auth}
      visibleCount={visibleCount}
      setVisibleCount={setVisibleCount}
      PAGE_STEP={PAGE_STEP}
    />
  )`;
  
c = beforeRet + newBody + afterFun;

fs.writeFileSync(path, c, 'utf-8');
console.log('Done - models.tsx now uses ModelsPageView');
