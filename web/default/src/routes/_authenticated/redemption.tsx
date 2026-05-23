import { createFileRoute } from '@tanstack/react-router'
import { useTranslation } from 'react-i18next'
import { useAuthStore } from '@/stores/auth-store'
import { useState } from 'react'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import {
  Plus,
  Search,
  Trash2,
  RefreshCw,
  Copy,
  Ticket,
  Shield,
} from 'lucide-react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Badge } from '@/components/ui/badge'
import { Card, CardContent } from '@/components/ui/card'
import { EmptyState } from '@/components/empty-state'
import {
  Table, TableBody, TableCell, TableHead, TableHeader, TableRow,
} from '@/components/ui/table'
import {
  Dialog, DialogContent, DialogDescription, DialogFooter, DialogHeader, DialogTitle,
} from '@/components/ui/dialog'
import { Skeleton } from '@/components/ui/skeleton'
import {
  type Redemption,
  getRedemptionCodes,
  createRedemptionCode,
  deleteRedemptionCode,
} from '@/lib/api-extended'
import { toast } from 'sonner'
import dayjs from '@/lib/dayjs'

export const Route = createFileRoute('/_authenticated/redemption')({
  component: RedemptionPage,
})

function RedemptionPage() {
  const { t } = useTranslation()
  const { auth } = useAuthStore();
  const isAdmin = auth.user?.role === 100 || auth.user?.role === 10;
  if (!isAdmin) {
    return (
      <div className="flex items-center justify-center min-h-[60vh] p-4">
        <div className="text-center max-w-md">
          <Shield className="h-16 w-16 mx-auto text-muted-foreground mb-4" />
          <h2 className="text-2xl font-bold mb-2">{t('Access Denied')}</h2>
          <p className="text-muted-foreground">{t('You do not have permission to access this page.')}</p>
        </div>
      </div>
    );
  }
  const queryClient = useQueryClient()
  const [dialogOpen, setDialogOpen] = useState(false)
  const [form, setForm] = useState({ name: '', count: 1, quota: 100 })

  const { data, isLoading, refetch } = useQuery({
    queryKey: ['redemption-codes'],
    queryFn: () => getRedemptionCodes(),
    staleTime: 30 * 1000,
  })

  const createMutation = useMutation({
    mutationFn: createRedemptionCode,
    onSuccess: () => {
      toast.success(t('Redemption codes created'))
      queryClient.invalidateQueries({ queryKey: ['redemption-codes'] })
      setDialogOpen(false)
    },
    onError: () => toast.error(t('Failed to create codes')),
  })

  const deleteMutation = useMutation({
    mutationFn: deleteRedemptionCode,
    onSuccess: () => {
      toast.success(t('Deleted'))
      queryClient.invalidateQueries({ queryKey: ['redemption-codes'] })
    },
  })

  const codes: Redemption[] = data?.data || []

  return (
    <div className=" w-full p-4 sm:p-6 space-y-6 min-h-screen bg-gradient-to-br from-slate-50 via-white to-blue-50/50 dark:from-slate-950 dark:via-slate-900 dark:to-blue-950/50">
      <div className="flex flex-col sm:flex-row items-start sm:items-center justify-between gap-3">
        <div>
          <h1 className="text-2xl sm:text-3xl lg:text-4xl font-bold tracking-tight bg-gradient-to-r from-blue-600 via-purple-600 to-pink-600 bg-clip-text text-transparent">
            {t('Redemption Codes')}
          </h1>
          <p className="text-muted-foreground mt-2 text-sm sm:text-base lg:text-lg">
            {t('Manage quota redemption codes')}
          </p>
        </div>
        <div className="flex flex-wrap gap-2">
          <Button variant="outline" onClick={() => refetch()}>
            <RefreshCw className="mr-2 h-4 w-4" />
            {t('Refresh')}
          </Button>
          <Button onClick={() => setDialogOpen(true)}>
            <Plus className="mr-2 h-4 w-4" />
            {t('Create Codes')}
          </Button>
        </div>
      </div>

      {!isLoading && codes.length === 0 ? (
        <EmptyState
          icon={Ticket}
          title={t('No redemption codes')}
          description={t('Create redemption codes for users to claim quota')}
          action={{ label: t('Create Codes'), onClick: () => setDialogOpen(true) }}
        />
      ) : (
        <Card>
          <CardContent className="p-0">
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-12">#</TableHead>
                  <TableHead>{t('Name')}</TableHead>
                  <TableHead>{t('Status')}</TableHead>
                  <TableHead>{t('Quota')}</TableHead>
                  <TableHead>{t('Count')}</TableHead>
                  <TableHead>{t('Used')}</TableHead>
                  <TableHead>{t('Created')}</TableHead>
                  <TableHead className="w-12" />
                </TableRow>
              </TableHeader>
              <TableBody>
                {isLoading ? (
                  Array.from({ length: 5 }).map((_, i) => (
                    <TableRow key={i}>
                      {Array.from({ length: 8 }).map((_, j) => (
                        <TableCell key={j}><Skeleton className="h-4 w-20" /></TableCell>
                      ))}
                    </TableRow>
                  ))
                ) : (
                  codes.map((code, idx) => (
                    <TableRow key={code.id}>
                      <TableCell className="text-muted-foreground">{idx + 1}</TableCell>
                      <TableCell className="font-medium">{code.code}</TableCell>
                      <TableCell>
                        <Badge variant={code.status === 1 ? 'default' : 'secondary'}>
                          {code.status === 1 ? t('Unused') : t('Disabled')}
                        </Badge>
                      </TableCell>
                      <TableCell>{code.quota?.toLocaleString()}</TableCell>
                      <TableCell>{code.max_count}</TableCell>
                      <TableCell>{code.used_count || 0}</TableCell>
                      <TableCell className="text-muted-foreground text-xs">
                        {dayjs(code.created_at * 1000).format('YYYY-MM-DD')}
                      </TableCell>
                      <TableCell>
                        <Button
                          variant="ghost"
                          size="icon"
                          className="h-8 w-8 text-destructive"
                          onClick={() => {
                            if (confirm(t('Are you sure?'))) deleteMutation.mutate(code.id)
                          }}
                        >
                          <Trash2 className="h-4 w-4" />
                        </Button>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      )}

      <Dialog open={dialogOpen} onOpenChange={setDialogOpen}>
        <DialogContent className="max-w-md">
          <DialogHeader>
            <DialogTitle>{t('Create Redemption Codes')}</DialogTitle>
            <DialogDescription>{t('Generate quota codes for users to redeem')}</DialogDescription>
          </DialogHeader>
          <form
            onSubmit={(e) => {
              e.preventDefault()
              createMutation.mutate(form)
            }}
            className="space-y-4"
          >
            <div className="space-y-2">
              <Label>{t('Batch Name')}</Label>
              <Input
                value={form.name}
                onChange={(e) => setForm({ ...form, name: e.target.value })}
                required
              />
            </div>
            <div className="space-y-2">
              <Label>{t('Code Count')}</Label>
              <Input
                type="number"
                min={1}
                value={form.count}
                onChange={(e) => setForm({ ...form, count: Number(e.target.value) })}
              />
            </div>
            <div className="space-y-2">
              <Label>{t('Quota Per Code')}</Label>
              <Input
                type="number"
                min={1}
                value={form.quota}
                onChange={(e) => setForm({ ...form, quota: Number(e.target.value) })}
              />
            </div>
            <DialogFooter>
              <Button type="button" variant="outline" onClick={() => setDialogOpen(false)}>
                {t('Cancel')}
              </Button>
              <Button type="submit" disabled={createMutation.isPending}>
                {createMutation.isPending ? t('Creating...') : t('Create')}
              </Button>
            </DialogFooter>
          </form>
        </DialogContent>
      </Dialog>
    </div>
  )
}

