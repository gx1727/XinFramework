import { useEffect, useState, useCallback, useMemo } from "react"
import { PageLayout } from "@/components/page-layout"
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import {
  Table,
  TableBody,
  TableCell,
  TableHead,
  TableHeader,
  TableRow,
} from "@/components/ui/table"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import {
  PlusIcon,
  SearchIcon,
  EditIcon,
  TrashIcon,
  Building2Icon,
  RefreshCw,
  PowerIcon,
  AlertTriangleIcon,
  PhoneIcon,
  MailIcon,
  MapPinIcon,
  ActivityIcon,
} from "lucide-react"
import { t } from "@/locales"
import { tenantApi, type TenantItem } from "@/api"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { Label } from "@/components/ui/label"
import { cn } from "@/lib/utils"
import { toast } from "sonner"

type StatusFilter = "all" | "active" | "disabled"

// Mock fallback（API 401 或断网时使用）
const mockTenants: TenantItem[] = [
  {
    id: 1,
    code: "default",
    name: "默认租户",
    status: 1,
    contact: "管理员",
    phone: "13800138000",
    email: "admin@example.com",
    province: "北京市",
    city: "北京市",
    area: "朝阳区",
    address: "xxx 街道",
    created_at: "2026-01-01 10:00:00",
    updated_at: "2026-04-26 10:00:00",
  },
  {
    id: 100,
    code: "acme",
    name: "Acme Corp",
    status: 1,
    contact: "张总",
    phone: "13900000001",
    email: "zhang@acme.com",
    province: "上海市",
    city: "上海市",
    area: "浦东新区",
    created_at: "2026-03-15 09:00:00",
    updated_at: "2026-04-20 14:00:00",
  },
  {
    id: 101,
    code: "beta",
    name: "Beta 科技",
    status: 0,
    contact: "李工",
    phone: "13900000002",
    email: "li@beta.com",
    province: "广东省",
    city: "深圳市",
    area: "南山区",
    created_at: "2026-02-10 11:00:00",
    updated_at: "2026-04-15 16:00:00",
  },
]

export function TenantsPage() {
  const [tenants, setTenants] = useState<TenantItem[]>([])
  const [searchTerm, setSearchTerm] = useState("")
  const [statusFilter, setStatusFilter] = useState<StatusFilter>("all")
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [dataSource, setDataSource] = useState<"api" | "mock" | null>(null)

  const [formOpen, setFormOpen] = useState(false)
  const [formMode, setFormMode] = useState<"create" | "edit">("create")
  const [current, setCurrent] = useState<TenantItem | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  const [confirmOpen, setConfirmOpen] = useState<null | {
    type: "soft-delete" | "purge" | "toggle-status"
    tenant: TenantItem
    nextStatus?: 0 | 1
  }>(null)

  // ----- Load -----
  const load = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const res = await tenantApi.list({ page: 1, size: 100 })
      const list = res?.list ?? []
      setTenants(list.length ? list : mockTenants)
      setDataSource(list.length ? "api" : "mock")
    } catch (err: any) {
      console.error("[tenants] load failed:", err)
      setTenants(mockTenants)
      setDataSource("mock")
      setError(err?.message ?? "API 不可用，已加载 mock 数据")
    } finally {
      setIsLoading(false)
    }
  }, [])

  useEffect(() => { load() }, [load])

  // ----- Filter -----
  const filtered = useMemo(() => {
    let arr = tenants
    if (statusFilter === "active") arr = arr.filter((x) => x.status === 1)
    if (statusFilter === "disabled") arr = arr.filter((x) => x.status === 0)
    if (searchTerm.trim()) {
      const kw = searchTerm.toLowerCase()
      arr = arr.filter(
        (x) =>
          x.code.toLowerCase().includes(kw) ||
          x.name.toLowerCase().includes(kw) ||
          (x.contact ?? "").toLowerCase().includes(kw) ||
          (x.email ?? "").toLowerCase().includes(kw)
      )
    }
    return arr
  }, [tenants, searchTerm, statusFilter])

  // ----- Stats -----
  const stats = useMemo(() => ({
    total: tenants.length,
    active: tenants.filter((x) => x.status === 1).length,
    disabled: tenants.filter((x) => x.status === 0).length,
  }), [tenants])

  // ----- Handlers -----
  const openCreate = () => {
    setFormMode("create")
    setCurrent(null)
    setFormOpen(true)
  }

  const openEdit = (tenant: TenantItem) => {
    setFormMode("edit")
    setCurrent(tenant)
    setFormOpen(true)
  }

  const handleSubmit = async (formData: Record<string, any>) => {
    setIsSubmitting(true)
    try {
      if (formMode === "create") {
        const created = await tenantApi.create(formData)
        toast.success(`租户「${created.name}」已创建并完成首装`)
        setFormOpen(false)
        await load()
      } else if (current) {
        await tenantApi.update(current.id, formData)
        toast.success("更新成功")
        setFormOpen(false)
        await load()
      }
    } catch (err: any) {
      const msg = err?.response?.data?.msg ?? err?.message ?? "操作失败"
      toast.error(msg)
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleToggleStatus = async (tenant: TenantItem) => {
    const next: 0 | 1 = tenant.status === 1 ? 0 : 1
    setConfirmOpen({ type: "toggle-status", tenant, nextStatus: next })
  }

  const handleSoftDelete = (tenant: TenantItem) => {
    setConfirmOpen({ type: "soft-delete", tenant })
  }

  const handlePurge = (tenant: TenantItem) => {
    setConfirmOpen({ type: "purge", tenant })
  }

  const executeConfirm = async () => {
    if (!confirmOpen) return
    const { type, tenant, nextStatus } = confirmOpen
    try {
      if (type === "toggle-status" && nextStatus !== undefined) {
        await tenantApi.updateStatus(tenant.id, nextStatus)
        toast.success(`租户「${tenant.name}」已${nextStatus === 1 ? "启用" : "停用"}`)
        await load()
      } else if (type === "soft-delete") {
        await tenantApi.delete(tenant.id)
        toast.success(`租户「${tenant.name}」已软删`)
        await load()
      } else if (type === "purge") {
        const res = await tenantApi.purge(tenant.id)
        toast.success(`租户「${tenant.code}」已硬删（清理 ${res.tables_purged} 张表）`)
        await load()
      }
    } catch (err: any) {
      const msg = err?.response?.data?.msg ?? err?.message ?? "操作失败"
      toast.error(msg)
    } finally {
      setConfirmOpen(null)
    }
  }

  return (
    <PageLayout>
      <div className="space-y-4">
        {/* Header */}
        <div className="flex items-start justify-between">
          <div>
            <h1 className="text-2xl font-bold flex items-center gap-2">
              <Building2Icon className="h-6 w-6" />
              {t("pages.tenants.title")}
            </h1>
            <p className="text-sm text-muted-foreground mt-1">
              {t("pages.tenants.subtitle")}
            </p>
          </div>
          <div className="flex gap-2">
            <Button variant="outline" size="sm" onClick={load} disabled={isLoading}>
              <RefreshCw className={cn("h-4 w-4 mr-2", isLoading && "animate-spin")} />
              {t("pages.tenants.refresh")}
            </Button>
            <Button size="sm" onClick={openCreate}>
              <PlusIcon className="h-4 w-4 mr-2" />
              {t("pages.tenants.create")}
            </Button>
          </div>
        </div>

        {/* Stats cards */}
        <div className="grid grid-cols-3 gap-4">
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                {t("pages.tenants.stats.total")}
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold flex items-center gap-2">
                <Building2Icon className="h-5 w-5" />
                {stats.total}
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                {t("pages.tenants.stats.active")}
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-green-600 flex items-center gap-2">
                <ActivityIcon className="h-5 w-5" />
                {stats.active}
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                {t("pages.tenants.stats.disabled")}
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold text-gray-500 flex items-center gap-2">
                <PowerIcon className="h-5 w-5" />
                {stats.disabled}
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Filters */}
        <Card>
          <CardContent className="pt-6">
            <div className="flex gap-2">
              <div className="relative flex-1">
                <SearchIcon className="absolute left-3 top-1/2 -translate-y-1/2 h-4 w-4 text-muted-foreground" />
                <Input
                  placeholder={t("pages.tenants.searchPlaceholder")}
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-9"
                />
              </div>
              <Select value={statusFilter} onValueChange={(v) => setStatusFilter(v as StatusFilter)}>
                <SelectTrigger className="w-[140px]">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">{t("pages.tenants.filter.all")}</SelectItem>
                  <SelectItem value="active">{t("pages.tenants.filter.active")}</SelectItem>
                  <SelectItem value="disabled">{t("pages.tenants.filter.disabled")}</SelectItem>
                </SelectContent>
              </Select>
            </div>
            {dataSource === "mock" && (
              <p className="text-xs text-amber-600 mt-2">
                ⚠ {error || "正在显示 mock 数据（API 不可用或未授权）"}
              </p>
            )}
          </CardContent>
        </Card>

        {/* Table */}
        <Card>
          <CardHeader>
            <CardTitle>{t("pages.tenants.list")}</CardTitle>
            <CardDescription>
              {t("pages.tenants.total")} {filtered.length} {t("pages.tenants.unit")}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-[60px]">ID</TableHead>
                  <TableHead>{t("pages.tenants.columns.code")}</TableHead>
                  <TableHead>{t("pages.tenants.columns.name")}</TableHead>
                  <TableHead>{t("pages.tenants.columns.contact")}</TableHead>
                  <TableHead>{t("pages.tenants.columns.status")}</TableHead>
                  <TableHead>{t("pages.tenants.columns.createdAt")}</TableHead>
                  <TableHead className="text-right">{t("pages.tenants.columns.actions")}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filtered.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={7} className="text-center text-muted-foreground py-8">
                      {isLoading ? "加载中..." : t("pages.tenants.empty")}
                    </TableCell>
                  </TableRow>
                ) : (
                  filtered.map((tenant) => (
                    <TableRow key={tenant.id}>
                      <TableCell className="font-mono text-xs text-muted-foreground">
                        {tenant.id}
                      </TableCell>
                      <TableCell>
                        <code className="px-1.5 py-0.5 rounded bg-muted text-xs font-mono">
                          {tenant.code}
                        </code>
                      </TableCell>
                      <TableCell className="font-medium">{tenant.name}</TableCell>
                      <TableCell>
                        <div className="text-sm">
                          {tenant.contact && <div>{tenant.contact}</div>}
                          {tenant.email && (
                            <div className="text-xs text-muted-foreground flex items-center gap-1">
                              <MailIcon className="h-3 w-3" /> {tenant.email}
                            </div>
                          )}
                          {tenant.phone && (
                            <div className="text-xs text-muted-foreground flex items-center gap-1">
                              <PhoneIcon className="h-3 w-3" /> {tenant.phone}
                            </div>
                          )}
                        </div>
                      </TableCell>
                      <TableCell>
                        {tenant.status === 1 ? (
                          <Badge variant="default" className="bg-green-600">
                            {t("pages.tenants.status.active")}
                          </Badge>
                        ) : (
                          <Badge variant="secondary">
                            {t("pages.tenants.status.disabled")}
                          </Badge>
                        )}
                      </TableCell>
                      <TableCell className="text-xs text-muted-foreground">
                        {tenant.created_at}
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="flex justify-end gap-1">
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => openEdit(tenant)}
                            title={t("pages.tenants.actions.edit")}
                          >
                            <EditIcon className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => handleToggleStatus(tenant)}
                            title={
                              tenant.status === 1
                                ? t("pages.tenants.actions.disable")
                                : t("pages.tenants.actions.enable")
                            }
                          >
                            <PowerIcon
                              className={cn(
                                "h-4 w-4",
                                tenant.status === 0 && "text-green-600"
                              )}
                            />
                          </Button>
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => handleSoftDelete(tenant)}
                            title={t("pages.tenants.actions.softDelete")}
                          >
                            <TrashIcon className="h-4 w-4 text-amber-600" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
          </CardContent>
        </Card>
      </div>

      {/* Form Dialog */}
      <TenantFormDialog
        open={formOpen}
        mode={formMode}
        initial={current}
        onCancel={() => setFormOpen(false)}
        onSubmit={handleSubmit}
        isSubmitting={isSubmitting}
      />

      {/* Confirm Dialog */}
      <Dialog open={!!confirmOpen} onOpenChange={(o) => !o && setConfirmOpen(null)}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <AlertTriangleIcon className="h-5 w-5 text-amber-500" />
              {confirmOpen?.type === "purge"
                ? t("pages.tenants.confirm.purgeTitle")
                : confirmOpen?.type === "soft-delete"
                ? t("pages.tenants.confirm.deleteTitle")
                : t("pages.tenants.confirm.statusTitle")}
            </DialogTitle>
            <DialogDescription>
              {confirmOpen?.type === "purge" && (
                <>
                  {t("pages.tenants.confirm.purgeDesc")
                    .replace("{code}", confirmOpen.tenant.code)
                    .replace("{name}", confirmOpen.tenant.name)}
                </>
              )}
              {confirmOpen?.type === "soft-delete" && (
                <>
                  {t("pages.tenants.confirm.deleteDesc")
                    .replace("{name}", confirmOpen.tenant.name)}
                </>
              )}
              {confirmOpen?.type === "toggle-status" && (
                <>
                  {t("pages.tenants.confirm.statusDesc")
                    .replace("{name}", confirmOpen.tenant.name)
                    .replace(
                      "{action}",
                      confirmOpen.nextStatus === 1
                        ? t("pages.tenants.actions.enable")
                        : t("pages.tenants.actions.disable")
                    )}
                </>
              )}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setConfirmOpen(null)}>
              {t("pages.tenants.confirm.cancel")}
            </Button>
            <Button
              variant={confirmOpen?.type === "purge" ? "destructive" : "default"}
              onClick={executeConfirm}
            >
              {confirmOpen?.type === "purge"
                ? t("pages.tenants.confirm.purgeOk")
                : t("pages.tenants.confirm.ok")}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </PageLayout>
  )
}

// ============================================
// 子组件：创建 / 编辑租户表单弹窗
// ============================================
interface TenantFormDialogProps {
  open: boolean
  mode: "create" | "edit"
  initial: TenantItem | null
  onCancel: () => void
  onSubmit: (data: Record<string, any>) => void
  isSubmitting: boolean
}

function TenantFormDialog({ open, mode, initial, onCancel, onSubmit, isSubmitting }: TenantFormDialogProps) {
  const [form, setForm] = useState<Record<string, any>>({})

  useEffect(() => {
    if (open) {
      if (mode === "edit" && initial) {
        setForm({
          name: initial.name,
          contact: initial.contact ?? "",
          phone: initial.phone ?? "",
          email: initial.email ?? "",
          province: initial.province ?? "",
          city: initial.city ?? "",
          area: initial.area ?? "",
          address: initial.address ?? "",
        })
      } else {
        setForm({ name: "", contact: "", phone: "", email: "", province: "", city: "", area: "", address: "" })
      }
    }
  }, [open, mode, initial])

  const update = (k: string, v: any) => setForm((f) => ({ ...f, [k]: v }))

  const handleOk = () => {
    if (mode === "create" && !form.code?.trim()) {
      toast.error("请填写租户 code")
      return
    }
    if (!form.name?.trim()) {
      toast.error("请填写租户名称")
      return
    }
    if (mode === "create") {
      onSubmit({
        code: form.code.trim(),
        name: form.name.trim(),
        contact: form.contact?.trim(),
        phone: form.phone?.trim(),
        email: form.email?.trim(),
      })
    } else {
      onSubmit({
        name: form.name.trim(),
        contact: form.contact?.trim(),
        phone: form.phone?.trim(),
        email: form.email?.trim(),
        province: form.province?.trim(),
        city: form.city?.trim(),
        area: form.area?.trim(),
        address: form.address?.trim(),
      })
    }
  }

  return (
    <Dialog open={open} onOpenChange={(o) => !o && onCancel()}>
      <DialogContent className="max-w-xl">
        <DialogHeader>
          <DialogTitle>
            {mode === "create" ? t("pages.tenants.form.createTitle") : t("pages.tenants.form.editTitle")}
          </DialogTitle>
          <DialogDescription>
            {mode === "create"
              ? "新建租户会自动触发首装（root 组织 / admin 角色 / 菜单资源 / 字典）"
              : t("pages.tenants.form.editDesc")}
          </DialogDescription>
        </DialogHeader>

        <div className="grid grid-cols-2 gap-3">
          {mode === "create" && (
            <div className="col-span-2">
              <Label htmlFor="code">租户 code（必填，租户唯一标识，创建后不可改）</Label>
              <Input
                id="code"
                value={form.code ?? ""}
                onChange={(e) => update("code", e.target.value)}
                placeholder="如 acme、beta"
              />
            </div>
          )}

          <div className="col-span-2">
            <Label htmlFor="name">租户名称（必填）</Label>
            <Input
              id="name"
              value={form.name ?? ""}
              onChange={(e) => update("name", e.target.value)}
            />
          </div>

          <div>
            <Label htmlFor="contact">联系人</Label>
            <Input
              id="contact"
              value={form.contact ?? ""}
              onChange={(e) => update("contact", e.target.value)}
            />
          </div>
          <div>
            <Label htmlFor="phone">手机</Label>
            <Input
              id="phone"
              value={form.phone ?? ""}
              onChange={(e) => update("phone", e.target.value)}
            />
          </div>

          <div className="col-span-2">
            <Label htmlFor="email">邮箱</Label>
            <Input
              id="email"
              type="email"
              value={form.email ?? ""}
              onChange={(e) => update("email", e.target.value)}
            />
          </div>

          {mode === "edit" && (
            <>
              <div className="col-span-2 pt-2 border-t">
                <div className="text-xs text-muted-foreground flex items-center gap-1">
                  <MapPinIcon className="h-3 w-3" />
                  区域信息（可选）
                </div>
              </div>
              <div>
                <Label htmlFor="province">省</Label>
                <Input id="province" value={form.province ?? ""} onChange={(e) => update("province", e.target.value)} />
              </div>
              <div>
                <Label htmlFor="city">市</Label>
                <Input id="city" value={form.city ?? ""} onChange={(e) => update("city", e.target.value)} />
              </div>
              <div>
                <Label htmlFor="area">区</Label>
                <Input id="area" value={form.area ?? ""} onChange={(e) => update("area", e.target.value)} />
              </div>
              <div>
                <Label htmlFor="address">详细地址</Label>
                <Input id="address" value={form.address ?? ""} onChange={(e) => update("address", e.target.value)} />
              </div>
            </>
          )}
        </div>

        <DialogFooter>
          <Button variant="outline" onClick={onCancel} disabled={isSubmitting}>
            {t("pages.tenants.confirm.cancel")}
          </Button>
          <Button onClick={handleOk} disabled={isSubmitting}>
            {isSubmitting ? "提交中..." : t("pages.tenants.confirm.ok")}
          </Button>
        </DialogFooter>
      </DialogContent>
    </Dialog>
  )
}
