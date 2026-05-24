/**
 * QuantumClaw - Menu Permissions Admin Page
 *
 * CRUD interface for super admins to manage menu items and their role permissions.
 * Only accessible by users with role >= 10.
 */

import { useState, useEffect } from 'react'
import { createFileRoute } from '@tanstack/react-router'
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Label } from '@/components/ui/label'
import { Switch } from '@/components/ui/switch'
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from '@/components/ui/table'
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from '@/components/ui/dialog'
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select'
import { Separator } from '@/components/ui/separator'
import { Badge } from '@/components/ui/badge'
import { Skeleton } from '@/components/ui/skeleton'
import { useT } from '@/lib/use-t'
import { useAuthStore } from '@/stores/auth-store'
import { adminGetAllMenus, adminSaveMenu, adminDeleteMenu, type MenuItemData } from '@/lib/use-menus'

// ---------------------------------------------------------------------------
// Route
// ---------------------------------------------------------------------------

export const Route = createFileRoute('/_authenticated/menu-permissions')({
  component: MenuPermissionsPage,
})

// ---------------------------------------------------------------------------
// Constants
// ---------------------------------------------------------------------------

const MENU_TYPES = [
  { value: 'nav', label: 'Navigation (Top Bar)' },
  { value: 'sidebar', label: 'Sidebar' },
]

const ROLE_OPTIONS = [
  { value: '0', label: 'Guest (0)' },
  { value: '1', label: 'User (1)' },
  { value: '2', label: 'Supplier (2)' },
  { value: '10', label: 'Admin (10)' },
  { value: '100', label: 'Root (100)' },
]

// ---------------------------------------------------------------------------
// Component
// ---------------------------------------------------------------------------

function MenuPermissionsPage() {
  const { t } = useT()
  const { auth } = useAuthStore()
  const queryClient = useQueryClient()

  const [editDialogOpen, setEditDialogOpen] = useState(false)
  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [editingMenu, setEditingMenu] = useState<Partial<MenuItemData> | null>(null)
  const [deletingId, setDeletingId] = useState<number | null>(null)

  // Form state
  const [formKey, setFormKey] = useState('')
  const [formParentKey, setFormParentKey] = useState('')
  const [formType, setFormType] = useState('sidebar')
  const [formLabel, setFormLabel] = useState('')
  const [formIcon, setFormIcon] = useState('')
  const [formPath, setFormPath] = useState('')
  const [formSort, setFormSort] = useState('0')
  const [formRoles, setFormRoles] = useState('[1]')
  const [formGroup, setFormGroup] = useState('')
  const [formEnabled, setFormEnabled] = useState(true)

  // Check auth
  const isAdmin = (auth.user?.role ?? 0) >= 10

  // Fetch menus
  const { data: menus = [], isLoading, error } = useQuery({
    queryKey: ['admin-menus'],
    queryFn: adminGetAllMenus,
    enabled: isAdmin,
  })

  // Save mutation
  const saveMutation = useMutation({
    mutationFn: (menu: Parameters<typeof adminSaveMenu>[0]) => adminSaveMenu(menu),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-menus'] })
      queryClient.invalidateQueries({ queryKey: ['menus'] })
      setEditDialogOpen(false)
    },
  })

  // Delete mutation
  const deleteMutation = useMutation({
    mutationFn: (id: number) => adminDeleteMenu(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['admin-menus'] })
      queryClient.invalidateQueries({ queryKey: ['menus'] })
      setDeleteDialogOpen(false)
      setDeletingId(null)
    },
  })

  // Open edit dialog for creating
  const handleCreate = () => {
    setEditingMenu(null)
    setFormKey('')
    setFormParentKey('')
    setFormType('sidebar')
    setFormLabel('')
    setFormIcon('')
    setFormPath('')
    setFormSort('0')
    setFormRoles('[1]')
    setFormGroup('')
    setFormEnabled(true)
    setEditDialogOpen(true)
  }

  // Open edit dialog for editing
  const handleEdit = (menu: MenuItemData) => {
    setEditingMenu(menu)
    setFormKey(menu.menu_key)
    setFormParentKey(menu.parent_key || '')
    setFormType(menu.menu_type)
    setFormLabel(menu.label_key)
    setFormIcon(menu.icon || '')
    setFormPath(menu.path)
    setFormSort(String(menu.sort_order))
    setFormRoles(menu.roles || '[1]')
    setFormGroup(menu.group_name || '')
    setFormEnabled(menu.enabled)
    setEditDialogOpen(true)
  }

  // Confirm save
  const handleSave = () => {
    if (!formKey || !formLabel || !formPath) return

    saveMutation.mutate({
      id: editingMenu?.id,
      menu_key: formKey,
      parent_key: formParentKey,
      menu_type: formType,
      label_key: formLabel,
      icon: formIcon,
      path: formPath,
      sort_order: parseInt(formSort) || 0,
      roles: formRoles,
      group_name: formGroup,
      enabled: formEnabled,
    })
  }

  // Open delete dialog
  const handleDeleteClick = (id: number) => {
    setDeletingId(id)
    setDeleteDialogOpen(true)
  }

  // Confirm delete
  const handleDeleteConfirm = () => {
    if (deletingId !== null) {
      deleteMutation.mutate(deletingId)
    }
  }

  // Toggle role checkbox in the roles JSON
  const toggleRole = (roleValue: string) => {
    let roles: string[]
    try {
      roles = JSON.parse(formRoles)
    } catch {
      roles = []
    }
    const idx = roles.indexOf(roleValue)
    if (idx >= 0) {
      roles.splice(idx, 1)
    } else {
      roles.push(roleValue)
    }
    setFormRoles(JSON.stringify(roles))
  }

  const isRoleSelected = (roleValue: string): boolean => {
    try {
      const roles = JSON.parse(formRoles)
      return roles.includes(roleValue)
    } catch {
      return false
    }
  }

  // Not admin
  if (!isAdmin) {
    return (
      <div className="flex items-center justify-center h-96">
        <div className="text-center">
          <h2 className="text-xl font-semibold text-muted-foreground">{t('Access Denied')}</h2>
          <p className="text-sm text-muted-foreground/60 mt-2">{t('You do not have permission to access this page.')}</p>
        </div>
      </div>
    )
  }

  if (isLoading) {
    return (
      <div className="space-y-4">
        <Skeleton className="h-10 w-48" />
        <Skeleton className="h-64 w-full" />
      </div>
    )
  }

  if (error) {
    return (
      <div className="flex items-center justify-center h-96">
        <div className="text-center">
          <h2 className="text-xl font-semibold text-destructive">{t('Error')}</h2>
          <p className="text-sm text-muted-foreground mt-2">{String(error)}</p>
        </div>
      </div>
    )
  }

  return (
    <div className="qc-wrapper py-8 space-y-6">
      <div className="flex flex-col items-center">
        <h1 className="text-3xl font-bold mb-2">{t('Menu Permissions')}</h1>
        <p className="text-muted-foreground mb-8" style={{maxWidth: 'min(65ch, 100%)'}}>{t('Configure which menu items are visible to each user role.')}</p>
      </div>

      {/* Menu Table */}
      <div className="border rounded-xl overflow-hidden">
        <Table>
          <TableHeader>
            <TableRow>
              <TableHead className="w-12">{t('ID')}</TableHead>
              <TableHead>{t('Key')}</TableHead>
              <TableHead>{t('Type')}</TableHead>
              <TableHead>{t('Label')}</TableHead>
              <TableHead>{t('Path')}</TableHead>
              <TableHead>{t('Roles')}</TableHead>
              <TableHead>{t('Group')}</TableHead>
              <TableHead>{t('Order')}</TableHead>
              <TableHead>{t('Enabled')}</TableHead>
              <TableHead className="w-32">{t('Actions')}</TableHead>
            </TableRow>
          </TableHeader>
          <TableBody>
            {menus.length === 0 ? (
              <TableRow>
                <TableCell colSpan={10} className="text-center text-muted-foreground py-8">
                  {t('No menu items configured.')}
                </TableCell>
              </TableRow>
            ) : (
              menus.map((menu) => (
                <TableRow key={menu.id}>
                  <TableCell className="font-mono text-xs">{menu.id}</TableCell>
                  <TableCell className="font-mono text-xs max-w-[120px] truncate" title={menu.menu_key}>
                    {menu.menu_key}
                  </TableCell>
                  <TableCell>
                    <Badge variant={menu.menu_type === 'nav' ? 'default' : 'secondary'}>
                      {menu.menu_type}
                    </Badge>
                  </TableCell>
                  <TableCell className="font-medium">{menu.label_key}</TableCell>
                  <TableCell className="font-mono text-xs max-w-[150px] truncate" title={menu.path}>
                    {menu.path}
                  </TableCell>
                  <TableCell>
                    <div className="flex flex-wrap gap-1">
                      {(() => {
                        try {
                          const roles = JSON.parse(menu.roles)
                          return roles.map((r: string) => (
                            <Badge key={r} variant="outline" className="text-[10px]">
                              {r === '0' ? 'G' : r === '1' ? 'U' : r === '2' ? 'S' : r === '10' ? 'A' : r === '100' ? 'R' : r}
                            </Badge>
                          ))
                        } catch {
                          return <span className="text-xs text-muted-foreground">{menu.roles}</span>
                        }
                      })()}
                    </div>
                  </TableCell>
                  <TableCell className="text-xs text-muted-foreground">
                    {menu.group_name || '-'}
                  </TableCell>
                  <TableCell className="text-xs">{menu.sort_order}</TableCell>
                  <TableCell>
                    {menu.enabled ? (
                      <Badge variant="success" className="bg-green-100 text-green-700 hover:bg-green-200">
                        {t('Yes')}
                      </Badge>
                    ) : (
                      <Badge variant="secondary" className="text-muted-foreground">
                        {t('No')}
                      </Badge>
                    )}
                  </TableCell>
                  <TableCell>
                    <div className="flex gap-1">
                      <Button variant="ghost" size="sm" onClick={() => handleEdit(menu)}>
                        {t('Edit')}
                      </Button>
                      <Button variant="ghost" size="sm" className="text-destructive" onClick={() => handleDeleteClick(menu.id)}>
                        {t('Delete')}
                      </Button>
                    </div>
                  </TableCell>
                </TableRow>
              ))
            )}
          </TableBody>
        </Table>
      </div>

      {/* Edit/Create Dialog */}
      <Dialog open={editDialogOpen} onOpenChange={setEditDialogOpen}>
        <DialogContent className="max-w-lg">
          <DialogHeader>
            <DialogTitle>{editingMenu ? t('Edit Menu Item') : t('Create Menu Item')}</DialogTitle>
            <DialogDescription>
              {t('Configure the menu item and its role-based visibility.')}
            </DialogDescription>
          </DialogHeader>

          <div className="grid gap-4 py-4">
            {/* Menu Key */}
            <div className="grid grid-cols-4 items-center gap-4">
              <Label className="text-right text-sm">{t('Menu Key')} *</Label>
              <Input
                className="col-span-3"
                value={formKey}
                onChange={(e) => setFormKey(e.target.value)}
                placeholder="e.g. sidebar-dashboard"
              />
            </div>

            {/* Menu Type */}
            <div className="grid grid-cols-4 items-center gap-4">
              <Label className="text-right text-sm">{t('Type')}</Label>
              <Select value={formType} onValueChange={setFormType}>
                <SelectTrigger className="col-span-3">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  {MENU_TYPES.map((mt) => (
                    <SelectItem key={mt.value} value={mt.value}>
                      {mt.label}
                    </SelectItem>
                  ))}
                </SelectContent>
              </Select>
            </div>

            {/* Label Key */}
            <div className="grid grid-cols-4 items-center gap-4">
              <Label className="text-right text-sm">{t('Label Key')} *</Label>
              <Input
                className="col-span-3"
                value={formLabel}
                onChange={(e) => setFormLabel(e.target.value)}
                placeholder="e.g. Dashboard"
              />
            </div>

            {/* Path */}
            <div className="grid grid-cols-4 items-center gap-4">
              <Label className="text-right text-sm">{t('Path')} *</Label>
              <Input
                className="col-span-3"
                value={formPath}
                onChange={(e) => setFormPath(e.target.value)}
                placeholder="e.g. /dashboard"
              />
            </div>

            {/* Icon */}
            <div className="grid grid-cols-4 items-center gap-4">
              <Label className="text-right text-sm">{t('Icon')}</Label>
              <Input
                className="col-span-3"
                value={formIcon}
                onChange={(e) => setFormIcon(e.target.value)}
                placeholder="e.g. LayoutDashboard"
              />
            </div>

            {/* Sort Order */}
            <div className="grid grid-cols-4 items-center gap-4">
              <Label className="text-right text-sm">{t('Sort Order')}</Label>
              <Input
                className="col-span-3"
                type="number"
                value={formSort}
                onChange={(e) => setFormSort(e.target.value)}
                placeholder="0"
              />
            </div>

            {/* Group Name */}
            <div className="grid grid-cols-4 items-center gap-4">
              <Label className="text-right text-sm">{t('Group Name')}</Label>
              <Input
                className="col-span-3"
                value={formGroup}
                onChange={(e) => setFormGroup(e.target.value)}
                placeholder="management, account, or leave empty"
              />
            </div>

            {/* Parent Key */}
            <div className="grid grid-cols-4 items-center gap-4">
              <Label className="text-right text-sm">{t('Parent Key')}</Label>
              <Input
                className="col-span-3"
                value={formParentKey}
                onChange={(e) => setFormParentKey(e.target.value)}
                placeholder="For sub-menus (optional)"
              />
            </div>

            {/* Enabled */}
            <div className="grid grid-cols-4 items-center gap-4">
              <Label className="text-right text-sm">{t('Enabled')}</Label>
              <div className="col-span-3">
                <Switch checked={formEnabled} onCheckedChange={setFormEnabled} />
              </div>
            </div>

            {/* Role Permissions */}
            <div className="grid grid-cols-4 items-start gap-4">
              <Label className="text-right text-sm pt-1">{t('Roles')}</Label>
              <div className="col-span-3 flex flex-wrap gap-2">
                {ROLE_OPTIONS.map((role) => (
                  <Button
                    key={role.value}
                    variant={isRoleSelected(role.value) ? 'default' : 'outline'}
                    size="sm"
                    onClick={() => toggleRole(role.value)}
                    className="text-xs"
                  >
                    {role.label}
                  </Button>
                ))}
              </div>
            </div>
            <div className="grid grid-cols-4 items-center gap-4">
              <div />
              <div className="col-span-3">
                <Input
                  value={formRoles}
                  onChange={(e) => setFormRoles(e.target.value)}
                  placeholder="[1,2,10,100]"
                  className="font-mono text-xs"
                />
                <p className="text-[10px] text-muted-foreground mt-1">
                  {t('JSON array of role IDs that can see this menu item.')}
                </p>
              </div>
            </div>
          </div>

          <DialogFooter>
            <Button variant="outline" onClick={() => setEditDialogOpen(false)}>
              {t('Cancel')}
            </Button>
            <Button onClick={handleSave} disabled={saveMutation.isPending || !formKey || !formLabel || !formPath}>
              {saveMutation.isPending ? t('Saving...') : t('Save')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirmation Dialog */}
      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent className="max-w-sm">
          <DialogHeader>
            <DialogTitle>{t('Delete Menu Item')}</DialogTitle>
            <DialogDescription>
              {t('Are you sure you want to delete this menu item? This action cannot be undone.')}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteDialogOpen(false)}>
              {t('Cancel')}
            </Button>
            <Button variant="destructive" onClick={handleDeleteConfirm} disabled={deleteMutation.isPending}>
              {deleteMutation.isPending ? t('Deleting...') : t('Delete')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </div>
  )
}
