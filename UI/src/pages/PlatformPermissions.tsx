import { useEffect, useState, useCallback, useMemo } from "react"
import { PageLayout } from "@/components/page-layout"
import { Card, CardContent, CardHeader } from "@/components/ui/card"
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
  PlusIcon,
  SearchIcon,
  EditIcon,
  TrashIcon,
  KeyIcon,
  RefreshCw,
  AlertTriangleIcon,
} from "lucide-react"
import { t } from "@/locales"
import {
  platformPermissionApi,
  platformMenuApi,
  type PlatformPermissionItem,
  type PlatformMenuItem,
  ApiError,
} from "@/api"
import { FormDialog } from "@/components/schema/DynamicForm"
import type { FormSchema } from "@/types/schema"
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"
import { cn } from "@/lib/utils"
import { toast } from "sonner"

export function PlatformPermissionsPage() {
  const [permissions, setPermissions] = useState<PlatformPermissionItem[]>([])
  const [menuTree, setMenuTree] = useState<PlatformMenuItem[]>([])
  const [searchTerm, setSearchTerm] = useState("")
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  const [dialogOpen, setDialogOpen] = useState(false)
  const [dialogMode, setDialogMode] = useState<"add" | "edit">("add")
  const [currentPerm, setCurrentPerm] = useState<PlatformPermissionItem | null>(
    null
  )

  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [permToDelete, setPermToDelete] =
    useState<PlatformPermissionItem | null>(null)

  // ---- Fetch ----
  const fetchPermissions = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const res = await platformPermissionApi.list({ page: 1, size: 500 })
      setPermissions(res?.list ?? [])
    } catch (err) {
      const msg =
        err instanceof ApiError
          ? `${err.status} ${err.message}`
          : err instanceof Error
            ? err.message
            : "加载平台权限码失败"
      console.error("[PlatformPermissions] load failed:", err)
      setPermissions([])
      setError(msg)
    } finally {
      setIsLoading(false)
    }
  }, [])

  const fetchMenuTree = useCallback(async () => {
    try {
      const tree = await platformMenuApi.tree()
      setMenuTree(tree || [])
    } catch (err) {
      console.warn("[PlatformPermissions] load menu tree failed:", err)
      setMenuTree([])
    }
  }, [])

  useEffect(() => {
    fetchPermissions()
  }, [fetchPermissions])

  useEffect(() => {
    fetchMenuTree()
  }, [fetchMenuTree])

  // ---- Filter ----
  const filteredPerms = useMemo(() => {
    if (!searchTerm.trim()) return permissions
    const kw = searchTerm.toLowerCase()
    return permissions.filter(
      (p) =>
        p.code.toLowerCase().includes(kw) ||
        p.name.toLowerCase().includes(kw) ||
        p.action.toLowerCase().includes(kw)
    )
  }, [permissions, searchTerm])

  const stats = useMemo(
    () => ({
      total: permissions.length,
      active: permissions.filter((p) => p.status === 1).length,
      menuLinked: new Set(
        permissions.map((p) => p.menu_id).filter((id) => id && id !== 0)
      ).size,
    }),
    [permissions]
  )

  // 构建菜单 select 选项（递归）
  const buildMenuOptions = (
    menus: PlatformMenuItem[],
    prefix = ""
  ): { label: string; value: number }[] => {
    const opts: { label: string; value: number }[] = []
    menus.forEach((m) => {
      opts.push({ label: `${prefix}${m.name}`, value: m.id })
      if (m.children?.length)
        opts.push(...buildMenuOptions(m.children, prefix + "├── "))
    })
    return opts
  }

  const menuOptions = useMemo(
    () => [
      { label: "无（公共资源）", value: 0 },
      ...buildMenuOptions(menuTree),
    ],
    [menuTree]
  )

  // ---- Form schema ----
  const permissionFormSchema: FormSchema = useMemo(
    () => ({
      items: [
        {
          field: "menu_id",
          label: t.pages.platformPermissions?.menu || "所属菜单",
          type: "select",
          required: false,
          placeholder: "请选择所属菜单（留空视为公共资源）",
          options: menuOptions,
        },
        {
          field: "name",
          label: t.pages.platformPermissions?.name || "权限名称",
          type: "text",
          required: true,
          placeholder: "请输入权限名称，如 查看平台菜单",
        },
        {
          field: "code",
          label: t.pages.platformPermissions?.code || "权限代码",
          type: "text",
          required: true,
          placeholder: "如 platform:menu:view（必须含且仅含一个冒号）",
          disabled: dialogMode === "edit",
          tooltip: "后端强校验：resource:action 格式",
        },
        {
          field: "action",
          label: t.pages.platformPermissions?.action || "操作类型",
          type: "select",
          required: true,
          placeholder: "请选择操作类型",
          options: [
            { label: "查看 (list)", value: "list" },
            { label: "详情 (view)", value: "view" },
            { label: "创建 (create)", value: "create" },
            { label: "编辑 (edit)", value: "edit" },
            { label: "删除 (delete)", value: "delete" },
            { label: "导出 (export)", value: "export" },
            { label: "导入 (import)", value: "import" },
            { label: "授权 (assign)", value: "assign" },
          ],
        },
        {
          field: "description",
          label: t.pages.platformPermissions?.description || "描述",
          type: "textarea",
          placeholder: "请输入权限描述",
        },
        {
          field: "sort",
          label: t.pages.platformPermissions?.sortOrder || "排序",
          type: "number",
          defaultValue: 1,
        },
        {
          field: "status",
          label: t.pages.platformPermissions?.status || "状态",
          type: "radio",
          defaultValue: 1,
          options: [
            { label: t.common.enable || "启用", value: 1 },
            { label: t.common.disable || "停用", value: 0 },
          ],
        },
      ],
    }),
    [menuOptions, dialogMode]
  )

  // ---- Handlers ----
  const handleAdd = () => {
    setDialogMode("add")
    setCurrentPerm(null)
    setDialogOpen(true)
  }

  const handleEdit = (perm: PlatformPermissionItem) => {
    setDialogMode("edit")
    setCurrentPerm(perm)
    setDialogOpen(true)
  }

  const handleDeleteConfirm = (perm: PlatformPermissionItem) => {
    setPermToDelete(perm)
    setDeleteDialogOpen(true)
  }

  const handleDelete = async () => {
    if (!permToDelete) return
    setIsSubmitting(true)
    try {
      await platformPermissionApi.delete(permToDelete.id)
      toast.success("删除成功")
      setDeleteDialogOpen(false)
      await fetchPermissions()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "删除失败")
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleSubmit = async (values: Record<string, unknown>) => {
    setIsSubmitting(true)
    try {
      const menuIdNum = Number(values.menu_id) || 0
      const payload: Partial<PlatformPermissionItem> & {
        name: string
        code?: string
      } = {
        name: String(values.name ?? ""),
        menu_id: menuIdNum > 0 ? menuIdNum : null,
        action: String(values.action ?? "view"),
        description: (values.description as string) || "",
        sort: Number(values.sort) || 1,
        status: Number(values.status) || 1,
      }
      if (dialogMode === "add") {
        payload.code = String(values.code ?? "")
        await platformPermissionApi.create(payload)
        toast.success("创建成功")
      } else if (currentPerm) {
        await platformPermissionApi.update(currentPerm.id, payload)
        toast.success("更新成功")
      }
      setDialogOpen(false)
      await fetchPermissions()
    } catch (err) {
      const msg =
        err instanceof ApiError
          ? err.status === 409
            ? "权限代码已存在"
            : err.status === 400
              ? "权限代码格式错误（必须为 resource:action）"
              : err.message
          : err instanceof Error
            ? err.message
            : "保存失败"
      toast.error(msg)
    } finally {
      setIsSubmitting(false)
    }
  }

  const getInitialValues = () => {
    if (currentPerm) {
      return {
        menu_id: currentPerm.menu_id ?? 0,
        name: currentPerm.name,
        code: currentPerm.code,
        action: currentPerm.action,
        description: currentPerm.description || "",
        sort: currentPerm.sort,
        status: currentPerm.status,
      }
    }
    return { menu_id: 0, sort: 1, status: 1, action: "view" }
  }

  // 菜单 ID → 菜单名
  const menuNameById = useMemo(() => {
    const map = new Map<number, string>()
    const walk = (ms: PlatformMenuItem[]) => {
      ms.forEach((m) => {
        map.set(m.id, m.name)
        m.children && walk(m.children)
      })
    }
    walk(menuTree)
    return map
  }, [menuTree])

  const getMenuName = (menuId?: number | null) => {
    if (!menuId || menuId === 0) return "公共资源"
    return menuNameById.get(menuId) || `菜单 ${menuId}`
  }

  const actionBadge: Record<
    string,
    { label: string; variant: "default" | "secondary" | "outline" }
  > = {
    list: { label: "列表", variant: "default" },
    view: { label: "查看", variant: "default" },
    create: { label: "创建", variant: "default" },
    edit: { label: "编辑", variant: "outline" },
    delete: { label: "删除", variant: "secondary" },
    export: { label: "导出", variant: "outline" },
    import: { label: "导入", variant: "outline" },
    assign: { label: "授权", variant: "outline" },
  }

  return (
    <PageLayout>
      <div className="px-4 lg:px-6">
        <div className="mb-6 flex items-center justify-between">
          <div>
            <h1 className="flex items-center gap-2 text-2xl font-bold">
              <KeyIcon className="h-6 w-6" />
              {t.pages.platformPermissions?.title || "平台权限码"}
            </h1>
            <p className="mt-1 text-sm text-muted-foreground">
              {t.pages.platformPermissions?.subtitle ||
                "管理平台域权限码（sys_permission），格式：resource:action"}
            </p>
          </div>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={fetchPermissions}
              disabled={isLoading}
            >
              <RefreshCw
                className={cn("mr-2 h-4 w-4", isLoading && "animate-spin")}
              />
              {t.pages.platformPermissions?.refresh || "刷新"}
            </Button>
            <Button size="sm" onClick={handleAdd}>
              <PlusIcon className="mr-2 h-4 w-4" />
              {t.pages.platformPermissions?.create || "新建权限码"}
            </Button>
          </div>
        </div>

        <div className="mb-4 grid grid-cols-3 gap-4">
          <Card>
            <CardHeader className="pb-2">
              <div className="text-sm text-muted-foreground">权限码总数</div>
              <div className="text-2xl font-bold">{stats.total}</div>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <div className="text-sm text-muted-foreground">启用中</div>
              <div className="text-2xl font-bold text-green-600">
                {stats.active}
              </div>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <div className="text-sm text-muted-foreground">关联菜单</div>
              <div className="text-2xl font-bold text-blue-600">
                {stats.menuLinked}
              </div>
            </CardHeader>
          </Card>
        </div>

        {error && (
          <div className="mb-4 flex items-start gap-2 rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-sm text-destructive">
            <AlertTriangleIcon className="mt-0.5 h-4 w-4 shrink-0" />
            <div className="flex-1">
              <div className="font-medium">加载失败</div>
              <div className="text-xs opacity-80">{error}</div>
            </div>
            <Button size="sm" variant="outline" onClick={fetchPermissions}>
              重试
            </Button>
          </div>
        )}

        <Card>
          <CardHeader>
            <div className="flex items-center gap-4">
              <div className="relative max-w-sm flex-1">
                <SearchIcon className="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  placeholder={
                    t.pages.platformPermissions?.searchPlaceholder ||
                    "搜索 code / 名称 / action..."
                  }
                  className="pl-9"
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                />
              </div>
              <Badge variant="secondary">共 {filteredPerms.length} 条</Badge>
            </div>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-[60px]">ID</TableHead>
                  <TableHead>
                    {t.pages.platformPermissions?.name || "名称"}
                  </TableHead>
                  <TableHead>
                    {t.pages.platformPermissions?.code || "代码"}
                  </TableHead>
                  <TableHead>
                    {t.pages.platformPermissions?.menu || "所属菜单"}
                  </TableHead>
                  <TableHead>
                    {t.pages.platformPermissions?.action || "操作"}
                  </TableHead>
                  <TableHead>描述</TableHead>
                  <TableHead className="w-[80px]">
                    {t.pages.platformPermissions?.sortOrder || "排序"}
                  </TableHead>
                  <TableHead>
                    {t.pages.platformPermissions?.status || "状态"}
                  </TableHead>
                  <TableHead className="w-[120px] text-right">
                    {t.common.edit}
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredPerms.length === 0 ? (
                  <TableRow>
                    <TableCell
                      colSpan={9}
                      className="py-8 text-center text-muted-foreground"
                    >
                      {isLoading ? t.common.loading : t.common.noData}
                    </TableCell>
                  </TableRow>
                ) : (
                  filteredPerms.map((perm) => {
                    const ab = actionBadge[perm.action] || {
                      label: perm.action,
                      variant: "secondary" as const,
                    }
                    return (
                      <TableRow key={perm.id}>
                        <TableCell className="font-mono text-xs text-muted-foreground">
                          {perm.id}
                        </TableCell>
                        <TableCell className="font-medium">
                          {perm.name}
                        </TableCell>
                        <TableCell>
                          <code className="rounded bg-muted px-1.5 py-0.5 font-mono text-xs">
                            {perm.code}
                          </code>
                        </TableCell>
                        <TableCell>
                          <Badge variant="outline">
                            {getMenuName(perm.menu_id)}
                          </Badge>
                        </TableCell>
                        <TableCell>
                          <Badge variant={ab.variant}>{ab.label}</Badge>
                        </TableCell>
                        <TableCell className="max-w-[260px] truncate text-sm text-muted-foreground">
                          {perm.description || "-"}
                        </TableCell>
                        <TableCell>{perm.sort}</TableCell>
                        <TableCell>
                          <Badge
                            variant={
                              perm.status === 1 ? "default" : "secondary"
                            }
                          >
                            {perm.status === 1 ? "启用" : "停用"}
                          </Badge>
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="flex items-center justify-end gap-1">
                            <Button
                              variant="ghost"
                              size="icon"
                              className="h-8 w-8"
                              onClick={() => handleEdit(perm)}
                            >
                              <EditIcon className="h-4 w-4" />
                            </Button>
                            <Button
                              variant="ghost"
                              size="icon"
                              className="h-8 w-8"
                              onClick={() => handleDeleteConfirm(perm)}
                            >
                              <TrashIcon className="h-4 w-4 text-destructive" />
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    )
                  })
                )}
              </TableBody>
            </Table>
            {isLoading && (
              <div className="flex items-center justify-center py-8">
                <div className="text-sm text-muted-foreground">
                  {t.common.loading}
                </div>
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      <FormDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        width={560}
        title={
          dialogMode === "add"
            ? t.pages.platformPermissions?.create || "新建平台权限码"
            : t.pages.platformPermissions?.edit || "编辑平台权限码"
        }
        schema={permissionFormSchema}
        initialValues={getInitialValues()}
        onSubmit={handleSubmit}
        loading={isSubmitting}
      />

      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent className="sm:max-w-[420px]">
          <DialogHeader>
            <DialogTitle>删除平台权限码</DialogTitle>
            <DialogDescription>
              确定要删除权限码 "{permToDelete?.name}" ({permToDelete?.code})
              吗？删除后将影响已分配该权限的角色。
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setDeleteDialogOpen(false)}
            >
              {t.common.cancel}
            </Button>
            <Button
              variant="destructive"
              onClick={handleDelete}
              disabled={isSubmitting}
            >
              {t.common.delete}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </PageLayout>
  )
}
