import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useState, useMemo } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Plus, Search, MoreHorizontal, Pencil, Trash2, RefreshCw, Users } from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { EmptyState } from '@/components/empty-state'
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table'
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/ui/dialog'
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu'
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select'
import { Skeleton } from '@/components/ui/skeleton'
import { Switch } from '@/components/ui/switch'
import apiClient from '@/lib/api'
import { cn } from '@/lib/utils'

export const Route = createFileRoute('/_authenticated/distributors')({
  component: DistributorsPage,
})

interface Distributor {
  id: number; user_id: number; name: string; contact_email: string
  markup_rate: number; profit_split: number; status: number
  api_key: string; total_revenue: number; total_payout: number
  created_at: string
}

async function getDistributors(): Promise<Distributor[]> {
  const res = await apiClient.get('/api/distributor')
  return res.data?.data || []
}

function DistributorsPage() {
  const { t } = useT()
  const queryClient = useQueryClient()
  const [search, setSearch] = useState('')
  const [editing, setEditing] = useState<Distributor | null>(null)
  const [isOpen, setIsOpen] = useState(false)

  const { data, isLoading, refetch } = useQuery({ queryKey: ['distributors'], queryFn: getDistributors })

  const distributors = useMemo(() => {
    if (!data) return []
    return data.filter((d) => !search || d.name?.toLowerCase().includes(search.toLowerCase()) || d.contact_email?.toLowerCase().includes(search.toLowerCase()))
  }, [data, search])

  return (
    <div className="p-4 sm:p-6 w-full space-y-6 min-h-screen bg-gradient-to-br from-slate-50 via-white to-blue-50/50 dark:from-slate-950 dark:via-slate-900 dark:to-blue-950/50">
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl sm:text-3xl lg:text-4xl font-bold tracking-tight bg-gradient-to-r from-amber-600 via-orange-600 to-red-600 bg-clip-text text-transparent">
            {t('Distributors')}
          </h1>
          <p className="text-muted-foreground mt-1 sm:mt-2 text-sm sm:text-base">{t('Manage reseller distributors and pricing')}</p>
        </div>
        <div className="flex gap-2">
          <Button variant="outline" onClick={() => refetch()}><RefreshCw className="mr-2 h-4 w-4" />{t('Refresh')}</Button>
          <Button onClick={() => { setEditing(null); setIsOpen(true) }}><Plus className="mr-2 h-4 w-4" />{t('Add Distributor')}</Button>
        </div>
      </div>

      <Card>
        <CardContent className="p-0">
          <div className="p-4 border-b">
            <div className="relative w-full max-w-sm"><Search className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" /><Input className="pl-9" placeholder={t('Search...')} value={search} onChange={(e) => setSearch(e.target.value)} /></div>
          </div>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead>{t('ID')}</TableHead>
                <TableHead>{t('Name')}</TableHead>
                <TableHead>{t('Contact')}</TableHead>
                <TableHead className="text-right">{t('Markup')}</TableHead>
                <TableHead className="text-right">{t('Split')}</TableHead>
                <TableHead className="text-right">{t('Revenue')}</TableHead>
                <TableHead className="text-right">{t('Payout')}</TableHead>
                <TableHead>{t('Status')}</TableHead>
                <TableHead className="text-right">{t('Actions')}</TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {isLoading ? Array.from({ length: 3 }).map((_, i) => (
                <TableRow key={i}>{Array.from({ length: 9 }).map((_, j) => <TableCell key={j}><Skeleton className="h-4 w-16" /></TableCell>)}</TableRow>
              )) : distributors.length === 0 ? (
                <TableRow><TableCell colSpan={9} className="text-center py-8 text-muted-foreground">{t('No distributors')}</TableCell></TableRow>
              ) : distributors.map((d) => (
                <TableRow key={d.id}>
                  <TableCell className="font-medium">{d.id}</TableCell>
                  <TableCell>{d.name}</TableCell>
                  <TableCell className="text-muted-foreground text-sm">{d.contact_email}</TableCell>
                  <TableCell className="text-right font-mono">{(d.markup_rate * 100).toFixed(1)}%</TableCell>
                  <TableCell className="text-right font-mono">{(d.profit_split * 100).toFixed(1)}%</TableCell>
                  <TableCell className="text-right font-mono text-green-600">{d.total_revenue}</TableCell>
                  <TableCell className="text-right font-mono text-red-600">{d.total_payout}</TableCell>
                  <TableCell><Badge variant={d.status === 1 ? 'default' : 'secondary'}>{d.status === 1 ? t('Active') : t('Disabled')}</Badge></TableCell>
                  <TableCell className="text-right">
                    <DropdownMenu>
                      <DropdownMenuTrigger asChild><Button variant="ghost" size="icon"><MoreHorizontal className="h-4 w-4" /></Button></DropdownMenuTrigger>
                      <DropdownMenuContent align="end">
                        <DropdownMenuItem onClick={() => { setEditing(d); setIsOpen(true) }}><Pencil className="mr-2 h-4 w-4" />{t('Edit')}</DropdownMenuItem>
                        <DropdownMenuItem className="text-red-600" onClick={async () => { if (confirm(t('Are you sure?'))) { await apiClient.delete(`/api/distributor/${d.id}`); queryClient.invalidateQueries({ queryKey: ['distributors'] }) } }}>
                          <Trash2 className="mr-2 h-4 w-4" />{t('Delete')}
                        </DropdownMenuItem>
                      </DropdownMenuContent>
                    </DropdownMenu>
                  </TableCell>
                </TableRow>
              ))}
            </TableBody>
          </Table>
        </CardContent>
      </Card>
    </div>
  )
}
