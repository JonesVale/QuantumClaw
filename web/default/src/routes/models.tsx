import { createFileRoute, Link } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useQuery } from '@tanstack/react-query'
import { useState, useMemo, useEffect, useRef } from 'react'
import {
  Search, SlidersHorizontal, X, ArrowUpDown,
  ChevronRight, ChevronDown, MessageSquare, Code, Brain,
  Image, Cpu, Atom, Play,
  CheckSquare, BarChart3
} from 'lucide-react'
import { Input } from '@/components/ui/input'
import { Badge } from '@/components/ui/badge'
import { Button } from '@/components/ui/button'
import { Card, CardContent } from '@/components/ui/card'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { cn } from '@/lib/utils'
import { codeToType } from '@/lib/tlanguages'
import { useAuthStore } from '@/stores/auth-store'
import { getEnhancedModels } from '@/lib/api-extended'
import { ModelDetailDialog, type CatalogItem } from '@/components/model-detail-dialog'
import { ModelComparisonDialog } from '@/components/model-comparison-dialog'
import { ModelsPageView } from '@/components/models-view'

export const Route = createFileRoute('/models')({
  component: ModelsPage,
})

const useCaseLabels: Record<string, { label: string; icon: React.ElementType; color: string }> = {
  'chat': { label: 'Chat & Assistant', icon: MessageSquare, color: 'from-blue-500 to-blue-600' },
  'coding': { label: 'Code Generation', icon: Code, color: 'from-green-500 to-green-600' },
  'reasoning': { label: 'Reasoning', icon: Brain, color: 'from-purple-500 to-purple-600' },
  'vision': { label: 'Vision', icon: Image, color: 'from-cyan-500 to-cyan-600' },
}

type SortOption = 'name' | 'price-asc' | 'price-desc'

function ModelsPage() {
  const { t, language } = useT()
  const { auth } = useAuthStore()
  const [search, setSearch] = useState('')
  const [useCaseFilter, setUseCaseFilter] = useState('all')
  const [seriesFilter, setSeriesFilter] = useState('all')
  const [modalityFilter, setModalityFilter] = useState('')
  const [contextFilter, setContextFilter] = useState('')
  const [providerFilter, setProviderFilter] = useState('')
  const [sortBy, setSortBy] = useState<SortOption>('name')
  const [sidebarOpen, setSidebarOpen] = useState(true)
  const [collapsed, setCollapsed] = useState<Record<string, boolean>>({
    provider: false, categories: false, modalities: false, context: false, series: false
  })

  // Detail dialog state
  const [detailModel, setDetailModel] = useState<CatalogItem | null>(null)
  const [detailOpen, setDetailOpen] = useState(false)

  // Pagination
  const [visibleCount, setVisibleCount] = useState(30)
  const PAGE_STEP = 30

  // Comparison state
  const [selectedModels, setSelectedModels] = useState<Set<string>>(new Set())
  const [comparisonOpen, setComparisonOpen] = useState(false)

  const lang = language || 'English'
  const { data } = useQuery({
    queryKey: ['model-catalog', lang],
    queryFn: async () => {
      const r = await fetch(`/api/model-catalog?lang=${encodeURIComponent(lang)}`)
      if (!r.ok) throw new Error('Failed to fetch')
      return r.json()
    },
    staleTime: 60 * 1000,
  })
  const catalog: CatalogItem[] = data?.data || []

  // Derived data - 依赖 catalog.length 避免每次 re-render 重算
  const providers = useMemo(() => {
    if (catalog.length === 0) return []
    const map = new Map<string, number>()
    for (const m of catalog) {
      const p = m.provider || 'Unknown'
      map.set(p, (map.get(p) || 0) + 1)
    }
    return Array.from(map.entries()).sort((a, b) => b[1] - a[1])
  }, [catalog.length]) // eslint-disable-line
  
  const seriesNames = useMemo(() => {
    if (catalog.length === 0) return []
    const set = new Set<string>()
    for (const m of catalog) {
      if (m.series) set.add(m.series)
    }
    return Array.from(set).sort()
  }, [catalog.length]) // eslint-disable-line

  const filtered = useMemo(() => {
    let result = catalog
    if (search) { const q = search.toLowerCase(); result = result.filter(m => m.name.toLowerCase().includes(q) || m.description.toLowerCase().includes(q)) }
    if (useCaseFilter !== 'all') result = result.filter(m => m.use_case === useCaseFilter)
    if (providerFilter) result = result.filter(m => (m.provider || '') === providerFilter)
    if (seriesFilter !== 'all') result = result.filter(m => m.series === seriesFilter)
    if (modalityFilter) result = result.filter(m => m.input_modalities.some((mod: string) => mod.toLowerCase() === modalityFilter.toLowerCase()))
    if (contextFilter) {
      const ctx = (m: any) => m.context_window || 0
      switch (contextFilter) {
        case '0-8192': result = result.filter(m => ctx(m) <= 8192); break
        case '8193-32768': result = result.filter(m => ctx(m) > 8192 && ctx(m) <= 32768); break
        case '32769-131072': result = result.filter(m => ctx(m) > 32768 && ctx(m) <= 131072); break
        case '131073-999999999': result = result.filter(m => ctx(m) > 131072); break
      }
    }
    switch (sortBy) { case 'price-asc': result.sort((a, b) => (a.input_price ?? 999) - (b.input_price ?? 999)); break; case 'price-desc': result.sort((a, b) => (b.input_price ?? 0) - (a.input_price ?? 0)); break; default: result.sort((a, b) => a.name.localeCompare(b.name)) }
    return result
  }, [catalog, search, useCaseFilter, providerFilter, modalityFilter, contextFilter, sortBy])

  // Reset pagination when filters change
  useEffect(() => setVisibleCount(PAGE_STEP), [search, useCaseFilter, providerFilter, seriesFilter, modalityFilter, contextFilter, sortBy])

  const displayed = useMemo(() => filtered.slice(0, visibleCount), [filtered, visibleCount])

  const toggleSelect = (name: string) => {
    setSelectedModels(prev => {
      const next = new Set(prev)
      if (next.has(name)) {
        next.delete(name)
      } else {
        if (next.size >= 4) return prev // Max 4
        next.add(name)
      }
      return next
    })
  }

  const openDetail = (m: CatalogItem) => {
    setDetailModel(m)
    setDetailOpen(true)
  }

  const selectedModelsList = useMemo(() => {
    return catalog.filter(m => selectedModels.has(m.name))
  }, [catalog, selectedModels])

  return (
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
      clearSelection={() => setSelectedModels(new Set())}
      auth={auth}
      visibleCount={visibleCount}
      setVisibleCount={setVisibleCount}
      PAGE_STEP={PAGE_STEP}
    />
  )
}
