import { useEffect, useState, useCallback, useMemo } from "react"
import { useNavigate } from "react-router-dom"
import { PageLayout } from "@/components/page-layout"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
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
  AlertCircleIcon,
  PhoneIcon,
  MailIcon,
  ActivityIcon,
  LogInIcon,
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
import { FormDialog } from "@/components/schema/DynamicForm"
import type { FormSchema } from "@/types/schema"
import { cn } from "@/lib/utils"
import { toast } from "sonner"
import { useAuthStore } from "@/stores/authStore"

type StatusFilter = "all" | "active" | "disabled"

export function TenantsPage() {
  const navigate = useNavigate()
  const isSuperAdmin = (
    useAuthStore((s) => s.user?.platform_roles) ?? []
  ).includes("super_admin")
  const startImpersonation = useAuthStore((s) => s.startImpersonation)

  const [tenants, setTenants] = useState<TenantItem[]>([])
  const [searchTerm, setSearchTerm] = useState("")
  const [statusFilter, setStatusFilter] = useState<StatusFilter>("all")
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  const [formOpen, setFormOpen] = useState(false)
  const [formMode, setFormMode] = useState<"create" | "edit">("create")
  const [current, setCurrent] = useState<TenantItem | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  const [confirmOpen, setConfirmOpen] = useState<null | {
    type: "soft-delete" | "purge" | "toggle-status" | "impersonate"
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
      setTenants(list)
    } catch (err) {
      const msg = err instanceof Error ? err.message : String(err)
      console.error("[tenants] load failed:", err)
      setError(`加载租户失败：${msg}`)
      setTenants([])
    } finally {
      setIsLoading(false)
    }
  }, [])

  useEffect(() => {
    load()
  }, [load])

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
  const stats = useMemo(
    () => ({
      total: tenants.length,
      active: tenants.filter((x) => x.status === 1).length,
      disabled: tenants.filter((x) => x.status === 0).length,
    }),
    [tenants]
  )

  // ----- Schema（与 Users/Organizations 风格一致：单列 + gap-4） -----
  const tenantFormSchema: FormSchema = useMemo(() => {
    const items: FormSchema["items"] = []
    if (formMode === "create") {
      items.push({
        field: "code",
        label: t.pages.tenants.form.codeLabel,
        type: "text",
        required: true,
        placeholder: "如 acme、beta",
        tooltip: t.pages.tenants.form.codeTooltip,
      })
    }
    items.push({
      field: "name",
      label: t.pages.tenants.form.nameLabel,
      type: "text",
      required: true,
      placeholder: t.pages.tenants.form.namePlaceholder,
    })
    items.push({
      field: "contact",
      label: t.pages.tenants.form.contactLabel,
      type: "text",
    })
    items.push({
      field: "phone",
      label: t.pages.tenants.form.phoneLabel,
      type: "text",
    })
    items.push({
      field: "email",
      label: t.pages.tenants.form.emailLabel,
      type: "email",
    })
    if (formMode === "edit") {
      items.push({
        field: "_region",
        label: t.pages.tenants.form.regionTitle,
        type: "divider",
      })
      items.push({
        field: "province",
        label: t.pages.tenants.form.provinceLabel,
        type: "text",
      })
      items.push({
        field: "city",
        label: t.pages.tenants.form.cityLabel,
        type: "text",
      })
      items.push({
        field: "area",
        label: t.pages.tenants.form.areaLabel,
        type: "text",
      })
      items.push({
        field: "address",
        label: t.pages.tenants.form.addressLabel,
        type: "text",
      })
    }
    return { items }
  }, [formMode])

  // ----- Initial values -----
  const formInitialValues = useMemo(() => {
    if (formMode === "edit" && current) {
      return {
        name: current.name,
        contact: current.contact ?? "",
        phone: current.phone ?? "",
        email: current.email ?? "",
        province: current.province ?? "",
        city: current.city ?? "",
        area: current.area ?? "",
        address: current.address ?? "",
      }
    }
    return { name: "", contact: "", phone: "", email: "" }
  }, [formMode, current])

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

  const handleSubmit = async (values: Record<string, unknown>) => {
    setIsSubmitting(true)
    try {
      // 清理空值
      const clean: Record<string, any> = {}
      for (const [k, v] of Object.entries(values)) {
        if (k.startsWith("_")) continue // skip divider markers
        if (typeof v === "string") clean[k] = v.trim()
        else clean[k] = v
      }
      if (formMode === "create") {
        const created = await tenantApi.create(clean)
        toast.success(`租户「${created.name}」已创建并完成首装`)
      } else if (current) {
        await tenantApi.update(current.id, clean)
        toast.success("更新成功")
      }
      setFormOpen(false)
      await load()
    } catch (err: any) {
      const msg = err?.response?.data?.msg ?? err?.message ?? "操作失败"
      toast.error(msg)
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleToggleStatus = (tenant: TenantItem) => {
    const next: 0 | 1 = tenant.status === 1 ? 0 : 1
    setConfirmOpen({ type: "toggle-status", tenant, nextStatus: next })
  }

  const handleSoftDelete = (tenant: TenantItem) => {
    setConfirmOpen({ type: "soft-delete", tenant })
  }

  const handleImpersonate = (tenant: TenantItem) => {
    setConfirmOpen({ type: "impersonate", tenant })
  }

  const executeConfirm = async () => {
    if (!confirmOpen) return
    const { type, tenant, nextStatus } = confirmOpen
    try {
      if (type === "toggle-status" && nextStatus !== undefined) {
        await tenantApi.updateStatus(tenant.id, nextStatus)
        toast.success(
          `租户「${tenant.name}」已${nextStatus === 1 ? "启用" : "停用"}`
        )
        await load()
      } else if (type === "soft-delete") {
        await tenantApi.delete(tenant.id)
        toast.success(`租户「${tenant.name}」已软删`)
        await load()
      } else if (type === "purge") {
        const res = await tenantApi.purge(tenant.id)
        toast.success(
          `租户「${tenant.code}」已硬删（清理 ${res.tables_purged} 张表）`
        )
        await load()
      } else if (type === "impersonate") {
        // 关闭确认框后再发起，避免 Dialog 蒙层还在时跳页
        const targetTenant = tenant
        setConfirmOpen(null)
        const ok = await startImpersonation(targetTenant.id, targetTenant.name)
        if (ok) {
          toast.success(
            t.pages.tenants.impersonation.toastSuccess.replace(
              "{name}",
              targetTenant.name
            )
          )
          navigate("/app/dashboard", { replace: true })
        } else {
          toast.error(t.pages.tenants.impersonation.toastFailed)
        }
        return
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
            <h1 className="flex items-center gap-2 text-2xl font-bold">
              <Building2Icon className="h-6 w-6" />
              {t.pages.tenants.title}
            </h1>
            <p className="mt-1 text-sm text-muted-foreground">
              {t.pages.tenants.subtitle}
            </p>
          </div>
          <div className="flex gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={load}
              disabled={isLoading}
            >
              <RefreshCw
                className={cn("mr-2 h-4 w-4", isLoading && "animate-spin")}
              />
              {t.pages.tenants.refresh}
            </Button>
            <Button size="sm" onClick={openCreate}>
              <PlusIcon className="mr-2 h-4 w-4" />
              {t.pages.tenants.create}
            </Button>
          </div>
        </div>

        {/* Stats cards */}
        <div className="grid grid-cols-3 gap-4">
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                {t.pages.tenants.stats.total}
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex items-center gap-2 text-2xl font-bold">
                <Building2Icon className="h-5 w-5" />
                {stats.total}
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                {t.pages.tenants.stats.active}
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex items-center gap-2 text-2xl font-bold text-green-600">
                <ActivityIcon className="h-5 w-5" />
                {stats.active}
              </div>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardTitle className="text-sm font-medium text-muted-foreground">
                {t.pages.tenants.stats.disabled}
              </CardTitle>
            </CardHeader>
            <CardContent>
              <div className="flex items-center gap-2 text-2xl font-bold text-gray-500">
                <PowerIcon className="h-5 w-5" />
                {stats.disabled}
              </div>
            </CardContent>
          </Card>
        </div>

        {/* Table（搜索 + 状态筛选直接放在 CardHeader，与 Users.tsx 风格一致） */}
        <Card>
          <CardHeader>
            {error && (
              <div className="mb-3 flex items-start gap-2 rounded-md border border-destructive/40 bg-destructive/5 p-3 text-sm">
                <AlertCircleIcon className="mt-0.5 h-4 w-4 shrink-0 text-destructive" />
                <div className="min-w-0 flex-1">
                  <div className="font-medium text-destructive">
                    接口调用失败
                  </div>
                  <div className="mt-0.5 text-xs break-all text-muted-foreground">
                    {error}
                  </div>
                </div>
                <Button variant="ghost" size="sm" onClick={load}>
                  重试
                </Button>
              </div>
            )}
            <CardTitle>{t.pages.tenants.list}</CardTitle>
            <CardDescription>
              {t.pages.tenants.total} {filtered.length} {t.pages.tenants.unit}
            </CardDescription>
            <div className="flex items-center gap-2 pt-2">
              {/* 搜索框：max-w-sm 与 Users.tsx / Organizations.tsx 一致 */}
              <div className="relative max-w-sm flex-1">
                <SearchIcon className="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  placeholder={t.pages.tenants.searchPlaceholder}
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                  className="pl-9"
                />
              </div>
              <Select
                value={statusFilter}
                onValueChange={(v) => setStatusFilter(v as StatusFilter)}
              >
                <SelectTrigger className="w-[140px]">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="all">
                    {t.pages.tenants.filter.all}
                  </SelectItem>
                  <SelectItem value="active">
                    {t.pages.tenants.filter.active}
                  </SelectItem>
                  <SelectItem value="disabled">
                    {t.pages.tenants.filter.disabled}
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-[60px]">ID</TableHead>
                  <TableHead>{t.pages.tenants.columns.code}</TableHead>
                  <TableHead>{t.pages.tenants.columns.name}</TableHead>
                  <TableHead>{t.pages.tenants.columns.contact}</TableHead>
                  <TableHead>{t.pages.tenants.columns.status}</TableHead>
                  <TableHead>{t.pages.tenants.columns.createdAt}</TableHead>
                  <TableHead className="text-right">
                    {t.pages.tenants.columns.actions}
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filtered.length === 0 ? (
                  <TableRow>
                    <TableCell
                      colSpan={7}
                      className="py-8 text-center text-muted-foreground"
                    >
                      {isLoading ? "加载中..." : t.pages.tenants.empty}
                    </TableCell>
                  </TableRow>
                ) : (
                  filtered.map((tenant) => (
                    <TableRow key={tenant.id}>
                      <TableCell className="font-mono text-xs text-muted-foreground">
                        {tenant.id}
                      </TableCell>
                      <TableCell>
                        <code className="rounded bg-muted px-1.5 py-0.5 font-mono text-xs">
                          {tenant.code}
                        </code>
                      </TableCell>
                      <TableCell className="font-medium">
                        {tenant.name}
                      </TableCell>
                      <TableCell>
                        <div className="text-sm">
                          {tenant.contact && <div>{tenant.contact}</div>}
                          {tenant.email && (
                            <div className="flex items-center gap-1 text-xs text-muted-foreground">
                              <MailIcon className="h-3 w-3" /> {tenant.email}
                            </div>
                          )}
                          {tenant.phone && (
                            <div className="flex items-center gap-1 text-xs text-muted-foreground">
                              <PhoneIcon className="h-3 w-3" /> {tenant.phone}
                            </div>
                          )}
                        </div>
                      </TableCell>
                      <TableCell>
                        {tenant.status === 1 ? (
                          <Badge variant="default" className="bg-green-600">
                            {t.pages.tenants.status.active}
                          </Badge>
                        ) : (
                          <Badge variant="secondary">
                            {t.pages.tenants.status.disabled}
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
                            title={t.pages.tenants.actions.edit}
                          >
                            <EditIcon className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="icon"
                            onClick={() => handleToggleStatus(tenant)}
                            title={
                              tenant.status === 1
                                ? t.pages.tenants.actions.disable
                                : t.pages.tenants.actions.enable
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
                            title={t.pages.tenants.actions.softDelete}
                          >
                            <TrashIcon className="h-4 w-4 text-amber-600" />
                          </Button>
                          {isSuperAdmin && (
                            <Button
                              variant="ghost"
                              size="icon"
                              onClick={() => handleImpersonate(tenant)}
                              title={t.pages.tenants.actions.impersonate}
                              disabled={tenant.status !== 1}
                            >
                              <LogInIcon
                                className={cn(
                                  "h-4 w-4",
                                  tenant.status === 1
                                    ? "text-blue-600"
                                    : "text-gray-300"
                                )}
                              />
                            </Button>
                          )}
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

      {/* Form Dialog（项目标准 FormDialog：单列 + gap-4，与 Users/Roles 风格一致） */}
      <FormDialog
        open={formOpen}
        onOpenChange={setFormOpen}
        title={
          formMode === "create"
            ? t.pages.tenants.form.createTitle
            : t.pages.tenants.form.editTitle
        }
        width={520}
        schema={tenantFormSchema}
        initialValues={formInitialValues}
        onSubmit={handleSubmit}
        loading={isSubmitting}
      />

      {/* Confirm Dialog */}
      <Dialog
        open={!!confirmOpen}
        onOpenChange={(o) => !o && setConfirmOpen(null)}
      >
        <DialogContent>
          <DialogHeader>
            <DialogTitle className="flex items-center gap-2">
              <AlertTriangleIcon className="h-5 w-5 text-amber-500" />
              {confirmOpen?.type === "purge"
                ? t.pages.tenants.confirm.purgeTitle
                : confirmOpen?.type === "soft-delete"
                  ? t.pages.tenants.confirm.deleteTitle
                  : confirmOpen?.type === "impersonate"
                    ? t.pages.tenants.confirm.impersonateTitle
                    : t.pages.tenants.confirm.statusTitle}
            </DialogTitle>
            <DialogDescription>
              {confirmOpen?.type === "purge" && (
                <>
                  {t.pages.tenants.confirm.purgeDesc
                    .replace("{code}", confirmOpen.tenant.code)
                    .replace("{name}", confirmOpen.tenant.name)}
                </>
              )}
              {confirmOpen?.type === "soft-delete" && (
                <>
                  {t.pages.tenants.confirm.deleteDesc.replace(
                    "{name}",
                    confirmOpen.tenant.name
                  )}
                </>
              )}
              {confirmOpen?.type === "toggle-status" && (
                <>
                  {t.pages.tenants.confirm.statusDesc
                    .replace("{name}", confirmOpen.tenant.name)
                    .replace(
                      "{action}",
                      confirmOpen.nextStatus === 1
                        ? t.pages.tenants.actions.enable
                        : t.pages.tenants.actions.disable
                    )}
                </>
              )}
              {confirmOpen?.type === "impersonate" && (
                <>
                  {t.pages.tenants.confirm.impersonateDesc.replace(
                    "{name}",
                    confirmOpen.tenant.name
                  )}
                </>
              )}
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setConfirmOpen(null)}>
              {t.pages.tenants.confirm.cancel}
            </Button>
            <Button
              variant={
                confirmOpen?.type === "purge" ? "destructive" : "default"
              }
              onClick={executeConfirm}
            >
              {confirmOpen?.type === "purge"
                ? t.pages.tenants.confirm.purgeOk
                : confirmOpen?.type === "impersonate"
                  ? t.pages.tenants.actions.impersonate
                  : t.pages.tenants.confirm.ok}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </PageLayout>
  )
}
