import { createFileRoute } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { useState, useEffect, useMemo } from 'react'
import { useQuery } from '@tanstack/react-query'
import { Atom, Cpu, Clock, Database, Zap, BarChart3, Play, Loader2, History, Terminal } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Badge } from '@/components/ui/badge'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs'
import { ScrollArea } from '@/components/ui/scroll-area'
import { Skeleton } from '@/components/ui/skeleton'
import { cn } from '@/lib/utils'
import dayjs from '@/lib/dayjs'

export const Route = createFileRoute('/quantum')({
  component: QuantumPage,
})

// ── 量子 Provider 定义 ────────────────────────────────────────
interface QuantumProvider {
  id: number
  name: string
  displayName: string
  icon: string
  color: string
  description: string
  backends: { name: string; qubits: number; status: 'online' | 'offline' | 'maintenance'; queueDepth: number }[]
  website: string
}

const quantumProviders: QuantumProvider[] = [
  {
    id: 100, name: 'IonQ', displayName: 'IonQ', icon: '⚛️', color: 'from-purple-500 to-purple-700',
    description: 'Trapped-ion quantum computers with high-fidelity gates and long coherence times.',
    backends: [
      { name: 'IonQ Harmony', qubits: 11, status: 'online', queueDepth: 3 },
      { name: 'IonQ Aria', qubits: 25, status: 'online', queueDepth: 7 },
      { name: 'IonQ Forte', qubits: 36, status: 'maintenance', queueDepth: 0 },
    ],
    website: 'https://ionq.com',
  },
  {
    id: 101, name: 'IBMQ', displayName: 'IBM Q', icon: '💻', color: 'from-blue-500 to-blue-700',
    description: 'IBM\'s superconducting quantum processors with the Qiskit ecosystem.',
    backends: [
      { name: 'ibm_sherbrooke', qubits: 127, status: 'online', queueDepth: 15 },
      { name: 'ibm_brisbane', qubits: 127, status: 'online', queueDepth: 22 },
      { name: 'ibm_kyiv', qubits: 127, status: 'online', queueDepth: 8 },
    ],
    website: 'https://quantum-computing.ibm.com',
  },
  {
    id: 102, name: 'Rigetti', displayName: 'Rigetti', icon: '🔬', color: 'from-green-500 to-green-700',
    description: 'Rigetti\'s superconducting processors with multi-chip architecture.',
    backends: [
      { name: 'Aspen-M-3', qubits: 80, status: 'online', queueDepth: 5 },
      { name: 'Ankaa-3', qubits: 84, status: 'online', queueDepth: 12 },
    ],
    website: 'https://rigetti.com',
  },
  {
    id: 103, name: 'AWSBraket', displayName: 'AWS Braket', icon: '☁️', color: 'from-orange-500 to-orange-700',
    description: 'Amazon\'s fully managed quantum computing service with multiple hardware providers.',
    backends: [
      { name: 'SV1 Simulator', qubits: 34, status: 'online', queueDepth: 0 },
      { name: 'TN1 Simulator', qubits: 50, status: 'online', queueDepth: 0 },
      { name: 'Rigetti (via Braket)', qubits: 80, status: 'online', queueDepth: 4 },
    ],
    website: 'https://aws.amazon.com/braket',
  },
  {
    id: 104, name: 'AzureQuantum', displayName: 'Azure Quantum', icon: '🔷', color: 'from-cyan-500 to-cyan-700',
    description: 'Microsoft\'s cloud quantum computing ecosystem with多种硬件平台接入.',
    backends: [
      { name: 'Honeywell H1', qubits: 10, status: 'online', queueDepth: 6 },
      { name: 'IonQ (via Azure)', qubits: 29, status: 'online', queueDepth: 3 },
      { name: 'Quantinuum H2', qubits: 56, status: 'online', queueDepth: 9 },
    ],
    website: 'https://azure.microsoft.com/quantum',
  },
  {
    id: 105, name: 'GoogleQuantum', displayName: 'Google Quantum', icon: '🧬', color: 'from-red-500 to-red-700',
    description: 'Google\'s superconducting quantum processors including the Sycamore and Willow chips.',
    backends: [
      { name: 'Sycamore', qubits: 53, status: 'online', queueDepth: 11 },
      { name: 'Willow', qubits: 105, status: 'online', queueDepth: 18 },
      { name: 'Rainbow', qubits: 23, status: 'maintenance', queueDepth: 0 },
    ],
    website: 'https://quantumai.google',
  },
]

const sampleQasm = `OPENQASM 2.0;
include "qelib1.inc";
qreg q[2];
creg c[2];
h q[0];
cx q[0],q[1];
measure q -> c;`

type TabKey = 'overview' | 'submit' | 'history'

function QuantumPage() {
  const { t } = useTranslation()
  const [activeTab, setActiveTab] = useState<TabKey>('overview')
  const [qasm, setQasm] = useState(sampleQasm)
  const [selectedProvider, setSelectedProvider] = useState('ionq')
  const [selectedBackend, setSelectedBackend] = useState('')
  const [shots, setShots] = useState('1024')
  const [submitting, setSubmitting] = useState(false)

  // Fetch real channel data from API to show quantum channels
  const { data: channelsData } = useQuery({
    queryKey: ['channels'],
    queryFn: async () => {
      const res = await fetch('/api/channel?scope=all')
      return res.json()
    },
    staleTime: 60 * 1000,
  })

  const quantumChannels = useMemo(() => {
    const all = channelsData?.data || []
    return all.filter((ch: { type: number; key?: string }) => ch.type >= 100 && ch.key && !ch.key.startsWith('PUT_YOUR'))
  }, [channelsData])

  // Mock job history
  const [jobHistory] = useState([
    { id: 'Q-20260522-001', provider: 'IonQ', backend: 'IonQ Harmony', qubits: 2, shots: 1024, status: 'completed', time: '2.3s', date: '2026-05-22 16:30' },
    { id: 'Q-20260522-002', provider: 'IBM Q', backend: 'ibm_sherbrooke', qubits: 5, shots: 4096, status: 'completed', time: '45.1s', date: '2026-05-22 15:15' },
    { id: 'Q-20260522-003', provider: 'Rigetti', backend: 'Aspen-M-3', qubits: 3, shots: 2048, status: 'running', time: '-', date: '2026-05-22 17:00' },
  ])

  const handleSubmitCircuit = () => {
    setSubmitting(true)
    setTimeout(() => setSubmitting(false), 1500)
  }

  const backendOptions = quantumProviders.find(p => p.name.toLowerCase() === selectedProvider)?.backends || []

  return (
    <div className="p-4 sm:p-6 space-y-6 w-full min-h-screen bg-gradient-to-br from-slate-50 via-white to-blue-50/50 dark:from-slate-950 dark:via-slate-900 dark:to-blue-950/50">
      {/* Header */}
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl sm:text-3xl lg:text-4xl font-bold tracking-tight bg-gradient-to-r from-purple-600 via-blue-600 to-cyan-600 bg-clip-text text-transparent flex items-center gap-3">
            <Atom className="h-8 w-8 text-purple-500" />
            {t('Quantum Computing')}
          </h1>
          <p className="text-muted-foreground mt-2 text-sm sm:text-base">
            {t('Access multiple quantum hardware providers through a unified interface')}
          </p>
        </div>
        <Badge variant="outline" className="text-sm px-4 py-2 gap-2">
          <Zap className="h-4 w-4 text-purple-500" />
          {quantumChannels.length} / {quantumProviders.length} {t('providers configured')}
        </Badge>
      </div>

      {/* Provider Stats */}
      <div className="grid gap-3 sm:gap-4 grid-cols-2 sm:grid-cols-3 lg:grid-cols-6">
        {quantumProviders.map((p) => (
          <Card key={p.id} className={cn(
            'relative overflow-hidden transition-all duration-300 hover:shadow-lg',
            quantumChannels.some((c: { type: number }) => c.type === p.id) ? 'opacity-100' : 'opacity-60'
          )}>
            <div className={cn('absolute inset-0 bg-gradient-to-br opacity-5', p.color)} />
            <CardContent className="p-4 text-center">
              <div className="text-2xl mb-1">{p.icon}</div>
              <p className="text-xs font-semibold truncate">{p.displayName}</p>
              <p className="text-[10px] text-muted-foreground">
                {p.backends.filter(b => b.status === 'online').length} {t('backends')}
              </p>
            </CardContent>
          </Card>
        ))}
      </div>

      <Tabs value={activeTab} onValueChange={(v) => setActiveTab(v as TabKey)}>
        <TabsList className="w-fit">
          <TabsTrigger value="overview" className="gap-2">
            <BarChart3 className="h-4 w-4" />{t('Overview')}
          </TabsTrigger>
          <TabsTrigger value="submit" className="gap-2">
            <Play className="h-4 w-4" />{t('Submit Circuit')}
          </TabsTrigger>
          <TabsTrigger value="history" className="gap-2">
            <History className="h-4 w-4" />{t('Job History')}
          </TabsTrigger>
        </TabsList>

        {/* ── Overview Tab ── */}
        <TabsContent value="overview" className="mt-4 space-y-6">
          {quantumProviders.map((provider) => {
            const isConfigured = quantumChannels.some((c: { type: number }) => c.type === provider.id)
            return (
              <Card key={provider.id} className={cn('overflow-hidden', !isConfigured && 'opacity-60')}>
                <CardHeader className={cn(
                  'bg-gradient-to-r text-white pb-3',
                  provider.color
                )}>
                  <div className="flex items-center justify-between">
                    <div className="flex items-center gap-3">
                      <span className="text-3xl">{provider.icon}</span>
                      <div>
                        <CardTitle className="text-white">{provider.displayName}</CardTitle>
                        <p className="text-white/80 text-sm mt-0.5">{provider.description}</p>
                      </div>
                    </div>
                    <Badge variant={isConfigured ? 'secondary' : 'outline'} className={cn(
                      isConfigured ? 'bg-white/20 text-white' : 'bg-white/10 text-white/60'
                    )}>
                      {isConfigured ? t('Configured') : t('Not Configured')}
                    </Badge>
                  </div>
                </CardHeader>
                <CardContent className="p-4">
                  <div className="space-y-3">
                    {provider.backends.map((b) => (
                      <div key={b.name} className="flex items-center justify-between p-3 rounded-lg bg-muted/50">
                        <div className="flex items-center gap-3">
                          <Cpu className="h-4 w-4 text-muted-foreground" />
                          <div>
                            <p className="text-sm font-medium">{b.name}</p>
                            <p className="text-xs text-muted-foreground">{b.qubits} {t('qubits')}</p>
                          </div>
                        </div>
                        <div className="flex items-center gap-3">
                          <Badge variant={b.status === 'online' ? 'default' : b.status === 'maintenance' ? 'secondary' : 'outline'}
                            className={cn(
                              'text-[10px]',
                              b.status === 'online' && 'bg-green-500/20 text-green-600 border-green-500/30',
                              b.status === 'maintenance' && 'bg-yellow-500/20 text-yellow-600 border-yellow-500/30',
                              b.status === 'offline' && 'bg-red-500/20 text-red-600 border-red-500/30',
                            )}>
                            {b.status === 'online' ? '● Online' : b.status === 'maintenance' ? '◐ Maintenance' : '○ Offline'}
                          </Badge>
                          <span className="text-xs text-muted-foreground">
                            ~{b.queueDepth} {t('queued')}
                          </span>
                        </div>
                      </div>
                    ))}
                  </div>
                </CardContent>
              </Card>
            )
          })}
        </TabsContent>

        {/* ── Submit Circuit Tab ── */}
        <TabsContent value="submit" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <Play className="h-5 w-5 text-purple-500" />
                {t('Submit Quantum Circuit')}
              </CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div className="grid grid-cols-1 md:grid-cols-3 gap-4">
                <div className="space-y-2">
                  <Label>{t('Provider')}</Label>
                  <Select value={selectedProvider} onValueChange={(v) => { setSelectedProvider(v); setSelectedBackend('') }}>
                    <SelectTrigger>
                      <SelectValue />
                    </SelectTrigger>
                    <SelectContent>
                      {quantumProviders.map(p => (
                        <SelectItem key={p.name} value={p.name.toLowerCase()}>
                          {p.icon} {p.displayName}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label>{t('Backend')}</Label>
                  <Select value={selectedBackend} onValueChange={setSelectedBackend}>
                    <SelectTrigger>
                      <SelectValue placeholder={t('Select backend')} />
                    </SelectTrigger>
                    <SelectContent>
                      {backendOptions.map(b => (
                        <SelectItem key={b.name} value={b.name} disabled={b.status !== 'online'}>
                          {b.name} ({b.qubits}q) {b.status !== 'online' && `— ${b.status}`}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
                </div>
                <div className="space-y-2">
                  <Label>{t('Shots')}</Label>
                  <Input type="number" min={1} max={100000} value={shots} onChange={(e) => setShots(e.target.value)} />
                </div>
              </div>

              <div className="space-y-2">
                <Label className="flex items-center gap-2">
                  <Terminal className="h-4 w-4" />
                  {t('QASM Circuit')}
                </Label>
                <textarea
                  value={qasm}
                  onChange={(e) => setQasm(e.target.value)}
                  className="flex min-h-[200px] w-full rounded-md border border-input bg-black/5 dark:bg-white/5 font-mono text-sm p-3 focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring"
                  spellCheck={false}
                />
              </div>

              <Button onClick={handleSubmitCircuit} disabled={submitting || !selectedBackend}>
                {submitting ? (
                  <><Loader2 className="mr-2 h-4 w-4 animate-spin" />{t('Submitting...')}</>
                ) : (
                  <><Play className="mr-2 h-4 w-4" />{t('Run Circuit')}</>
                )}
              </Button>

              <p className="text-xs text-muted-foreground">
                {t('Circuit execution cost depends on provider pricing and number of shots')}
              </p>
            </CardContent>
          </Card>
        </TabsContent>

        {/* ── Job History Tab ── */}
        <TabsContent value="history" className="mt-4">
          <Card>
            <CardHeader>
              <CardTitle className="flex items-center gap-2">
                <History className="h-5 w-5 text-purple-500" />
                {t('Recent Jobs')}
              </CardTitle>
            </CardHeader>
            <CardContent>
              {jobHistory.length === 0 ? (
                <p className="text-center text-muted-foreground py-8">{t('No quantum jobs yet')}</p>
              ) : (
                <div className="space-y-3">
                  {jobHistory.map((job) => (
                    <div key={job.id} className="flex items-center justify-between p-3 rounded-lg bg-muted/50 hover:bg-muted/70 transition-colors">
                      <div className="flex items-center gap-4">
                        <div className={cn(
                          'w-10 h-10 rounded-xl flex items-center justify-center font-bold text-white text-sm',
                          job.status === 'completed' ? 'bg-gradient-to-br from-green-500 to-green-600' :
                          job.status === 'running' ? 'bg-gradient-to-br from-blue-500 to-blue-600' :
                          'bg-gradient-to-br from-red-500 to-red-600'
                        )}>
                          {job.qubits}q
                        </div>
                        <div>
                          <p className="text-sm font-semibold">{job.id}</p>
                          <p className="text-xs text-muted-foreground">
                            {job.provider} — {job.backend} · {job.shots.toLocaleString()} shots
                          </p>
                        </div>
                      </div>
                      <div className="text-right">
                        <Badge variant={job.status === 'completed' ? 'default' : 'secondary'}
                          className={cn(
                            'text-[10px]',
                            job.status === 'completed' && 'bg-green-500/20 text-green-600 border-green-500/30',
                            job.status === 'running' && 'bg-blue-500/20 text-blue-600 border-blue-500/30',
                          )}>
                          {job.status === 'completed' ? 'Completed' : 'Running'}
                        </Badge>
                        <p className="text-xs text-muted-foreground mt-1">{job.time} · {job.date}</p>
                      </div>
                    </div>
                  ))}
                </div>
              )}
            </CardContent>
          </Card>
        </TabsContent>
      </Tabs>
    </div>
  )
}
