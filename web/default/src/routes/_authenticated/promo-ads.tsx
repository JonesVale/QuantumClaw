import { createFileRoute } from '@tanstack/react-router'
import { useT } from '@/lib/use-t'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import apiClient from '@/lib/api'
import { useState } from 'react'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Dialog, DialogContent, DialogHeader, DialogTitle, DialogFooter } from '@/components/ui/dialog'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import { toast } from 'sonner'

export const Route = createFileRoute('/_authenticated/promo-ads')({
  component: PromoAdsPage,
})

interface PromoAd {
  id?: number
  page_key: string
  icon: string
  title: string
  link_url: string
  sort_order: number
  enabled: boolean
}

const defaultForm: PromoAd = {
  page_key: 'all',
  icon: '🚀',
  title: '',
  link_url: '/models',
  sort_order: 0,
  enabled: true,
}

function PromoAdsPage() {
  const { t } = useT()
  const qc = useQueryClient()
  const [open, setOpen] = useState(false)
  const [form, setForm] = useState<PromoAd>({ ...defaultForm })

  const { data, isLoading } = useQuery({
    queryKey: ['admin-promo-ads'],
    queryFn: async () => {
      const res = await apiClient.get('/api/admin/promo-ads')
      if (res.data?.success && Array.isArray(res.data.data)) return res.data.data as PromoAd[]
      return []
    },
  })

  const saveMutation = useMutation({
    mutationFn: async (ad: PromoAd) => {
      if (ad.id) {
        return apiClient.put('/api/admin/promo-ads', ad)
      }
      return apiClient.post('/api/admin/promo-ads', ad)
    },
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin-promo-ads'] })
      qc.invalidateQueries({ queryKey: ['promo-ads'] })
      setOpen(false)
      setForm({ ...defaultForm })
      toast.success('Saved')
    },
    onError: () => toast.error('Failed to save'),
  })

  const deleteMutation = useMutation({
    mutationFn: async (id: number) => apiClient.delete(`/api/admin/promo-ads/${id}`),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['admin-promo-ads'] })
      qc.invalidateQueries({ queryKey: ['promo-ads'] })
      toast.success('Deleted')
    },
    onError: () => toast.error('Failed to delete'),
  })

  const ads = data || []

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <div>
          <h2 className="text-2xl font-bold tracking-tight">{t('Promo Ads')}</h2>
          <p className="text-sm text-muted-foreground mt-1">{t('Manage scrolling advertisement cards')}</p>
        </div>
        <Button onClick={() => { setForm({ ...defaultForm }); setOpen(true) }}>
          {t('Add Ad')}
        </Button>
      </div>

      <Card>
        <CardHeader>
          <CardTitle className="text-base">{t('Advertisement Cards')} ({ads.length})</CardTitle>
        </CardHeader>
        <CardContent>
          {isLoading ? (
            <div className="text-center py-10 text-muted-foreground">{t('Loading...')}</div>
          ) : ads.length === 0 ? (
            <div className="text-center py-10 text-muted-foreground">{t('No ads configured')}</div>
          ) : (
            <div className="space-y-2">
              <div className="hidden md:grid grid-cols-12 gap-3 px-4 py-2 text-xs font-semibold text-muted-foreground uppercase tracking-wide">
                <div className="col-span-1">{t('Sort')}</div>
                <div className="col-span-1">{t('Icon')}</div>
                <div className="col-span-3">{t('Title')}</div>
                <div className="col-span-3">{t('Link URL')}</div>
                <div className="col-span-1">{t('Page')}</div>
                <div className="col-span-1">{t('Active')}</div>
                <div className="col-span-2">{t('Actions')}</div>
              </div>
              {ads.map(ad => (
                <div key={ad.id} className="grid grid-cols-2 md:grid-cols-12 gap-2 md:gap-3 items-center px-4 py-3 rounded-xl bg-white/60 hover:bg-white/90 border border-border/10 transition-all">
                  <div className="col-span-1 text-xs text-muted-foreground">{ad.sort_order}</div>
                  <div className="col-span-1 text-xl">{ad.icon}</div>
                  <div className="col-span-3 text-sm font-medium truncate">{ad.title}</div>
                  <div className="col-span-3 text-xs text-muted-foreground truncate">{ad.link_url}</div>
                  <div className="col-span-1 text-xs text-muted-foreground">{ad.page_key}</div>
                  <div className="col-span-1">
                    <div className={`w-2 h-2 rounded-full ${ad.enabled ? 'bg-green-500' : 'bg-gray-300'}`} />
                  </div>
                  <div className="col-span-2 flex items-center gap-2">
                    <Button size="sm" variant="outline" onClick={() => { setForm(ad); setOpen(true) }}>
                      {t('Edit')}
                    </Button>
                    <Button size="sm" variant="destructive" onClick={() => { if (confirm('Delete?')) deleteMutation.mutate(ad.id!) }}>
                      {t('Delete')}
                    </Button>
                  </div>
                </div>
              ))}
            </div>
          )}
        </CardContent>
      </Card>

      {/* Edit/Create Dialog */}
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="sm:max-w-lg">
          <DialogHeader>
            <DialogTitle>{form.id ? t('Edit Ad') : t('New Ad')}</DialogTitle>
          </DialogHeader>
          <div className="space-y-4 py-4">
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>{t('Icon (emoji)')}</Label>
                <Input value={form.icon} onChange={e => setForm({ ...form, icon: e.target.value })} placeholder="🤖" />
              </div>
              <div className="space-y-2">
                <Label>{t('Sort Order')}</Label>
                <Input type="number" value={form.sort_order} onChange={e => setForm({ ...form, sort_order: parseInt(e.target.value) || 0 })} />
              </div>
            </div>
            <div className="space-y-2">
              <Label>{t('Title')}</Label>
              <Input value={form.title} onChange={e => setForm({ ...form, title: e.target.value })} placeholder="GPT-4o — Multimodal Vision & Audio" />
            </div>
            <div className="space-y-2">
              <Label>{t('Link URL')}</Label>
              <Input value={form.link_url} onChange={e => setForm({ ...form, link_url: e.target.value })} placeholder="/models" />
            </div>
            <div className="grid grid-cols-2 gap-4">
              <div className="space-y-2">
                <Label>{t('Page Key')}</Label>
                <select
                  value={form.page_key}
                  onChange={e => setForm({ ...form, page_key: e.target.value })}
                  className="w-full h-10 rounded-xl border border-border/30 bg-white px-3 text-sm outline-none"
                >
                  <option value="all">{t("All Pages")}</option>
                  <option value="home">{t("Home")}</option>
                  <option value="models">{t("Models")}</option>
                  <option value="pricing">{t("Pricing")}</option>
                  <option value="rankings">{t("Rankings")}</option>
                  <option value="apps">{t("Apps")}</option>
                  <option value="enterprise">{t("Enterprise")}</option>
                  <option value="dashboard">{t("Dashboard")}</option>
                </select>
              </div>
              <div className="space-y-2 flex items-end pb-2">
                <Label className="flex items-center gap-3 cursor-pointer">
                  <Switch checked={form.enabled} onCheckedChange={v => setForm({ ...form, enabled: v })} />
                  <span className="text-sm">{t('Enabled')}</span>
                </Label>
              </div>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setOpen(false)}>{t('Cancel')}</Button>
            <Button onClick={() => saveMutation.mutate(form)} disabled={!form.title}>
              {t('Save')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
