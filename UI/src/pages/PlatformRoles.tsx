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
  PlusIcon,
  SearchIcon,
  EditIcon,
  TrashIcon,
  ShieldIcon,
  RefreshCw,
  AlertTriangleIcon,
  CheckSquare,
  Square,
  ChevronDown,
  ChevronRight,
  KeyIcon,
} from "lucide-react"
import { t } from "@/locales"
import {
  platformRoleApi,
  platformMenuApi,
  platformPermissionApi,
  type PlatformRoleItem,
  type PlatformMenuItem,
  type PlatformPermissionItem,
  ApiError,
} from "@/api"
import { FormDialog } from "@/components/schema/DynamicForm"
import type { FormSchema } from "@/types/schema"
import { Checkbox } from "@/components/ui/checkbox"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
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

const LS_KEY_USE_MOCK = "platformRolesPage.useMockFallback"

// ---- Mock 兜底 ----
const mockRoles: PlatformRoleItem[] = [
  { id: 1, code: "super_admin", name: "超级管理员", description: "拥有平台所有权限", data_scope: 1, is_default: true, sort: 1, status: 1, menus: [], permissions: [] },
  { id: 2, code: "ops_admin", name: "运营管理员", description: "负责平台日常运营", data_scope: 1, is_default: false, sort: 2, status: 1, menus: [], permissions: [] },
  { id: 3, code: "auditor", name: "审计员", description: "只读审计权限", data_scope: 1, is_default: false, sort: 3, status: 0, menus: [], permissions: [] },
]

export function PlatformRolesPage() {
  const [roles, setRoles] = useState<PlatformRoleItem[]>([])
  const [searchTerm, setSearchTerm] = useState("")
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [dataSource, setDataSource] = useState<"api" | "mock" | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  const [dialogOpen, setDialogOpen] = useState(false)
  const [dialogMode, setDialogMode] = useState<"add" | "edit">("add")
  const [currentRole, setCurrentRole] = useState<PlatformRoleItem | null>(null)

  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [roleToDelete, setRoleToDelete] = useState<PlatformRoleItem | null>(null)

  // 权限分配对话框
  const [permDialogOpen, setPermDialogOpen] = useState(false)
  const [currentPermRole, setCurrentPermRole] = useState<PlatformRoleItem | null>(null)
  const [menuTree, setMenuTree] = useState<PlatformMenuItem[]>([])
  const [selectedMenuIds, setSelectedMenuIds] = useState<number[]>([])
  const [allPermissions, setAllPermissions] = useState<PlatformPermissionItem[]>([])
  const [selectedPermIds, setSelectedPermIds] = useState<number[]>([])
  const [permLoading, setPermLoading] = useState(false)

  // Mock 兜底开关
  const [useMockFallback, setUseMockFallback] = useState<boolean>(() => {
    if (typeof window === "undefined") return false
    return window.localStorage.getItem(LS_KEY_USE_MOCK) === "1"
  })

  useEffect(() => {
    if (typeof window !== "undefined") {
      window.localStorage.setItem(LS_KEY_USE_MOCK, useMockFallback ? "1" : "0")
    }
  }, [useMockFallback])

  // ---- Fetch ----
  const fetchRoles = useCallback(async () => {
    if (useMockFallback) {
      setRoles(mockRoles)
      setDataSource("mock")
      setError(null)
      return
    }
    setIsLoading(true)
    setError(null)
    try {
      const res = await platformRoleApi.list({ page: 1, size: 200 })
      setRoles(res?.list ?? [])
      setDataSource("api")
    } catch (err) {
      const msg =
        err instanceof ApiError
          ? `${err.status} ${err.message}`
          : err instanceof Error
            ? err.message
            : "加载平台角色失败"
      console.error("[PlatformRoles] load failed:", err)
      setRoles([])
      setDataSource(null)
      setError(msg)
    } finally {
      setIsLoading(false)
    }
  }, [useMockFallback])

  useEffect(() => {
    fetchRoles()
  }, [fetchRoles])

  // ---- Filter ----
  const filteredRoles = useMemo(() => {
    if (!searchTerm.trim()) return roles
    const kw = searchTerm.toLowerCase()
    return roles.filter(
      (r) =>
        r.code.toLowerCase().includes(kw) ||
        r.name.toLowerCase().includes(kw) ||
        (r.description ?? "").toLowerCase().includes(kw),
    )
  }, [roles, searchTerm])

  const stats = useMemo(
    () => ({
      total: roles.length,
      active: roles.filter((r) => r.status === 1).length,
      default: roles.filter((r) => r.is_default).length,
    }),
    [roles],
  )

  // ---- Form schema ----
  const roleFormSchema: FormSchema = useMemo(
    () => ({
      items: [
        {
          field: "name",
          label: t.pages.platformRoles?.name || "角色名称",
          type: "text",
          required: true,
          placeholder: "请输入角色名称",
        },
        {
          field: "code",
          label: t.pages.platformRoles?.code || "角色代码",
          type: "text",
          required: true,
          placeholder: "请输入角色代码，如 super_admin",
          disabled: dialogMode === "edit",
          tooltip: "创建后不可改",
        },
        {
          field: "description",
          label: t.pages.platformRoles?.description || "描述",
          type: "textarea",
          placeholder: "请输入角色描述",
        },
        {
          field: "data_scope",
          label: t.pages.platformRoles?.dataScope || "数据范围",
          type: "select",
          defaultValue: 1,
          tooltip: "平台域当前主要为租户维度；非 1 时通常无意义",
          options: [
            { label: t.pages.roles?.dataScopeAll || "全部数据", value: 1 },
            { label: t.pages.roles?.dataScopeSelf || "本人数据", value: 5 },
          ],
        },
        {
          field: "is_default",
          label: "默认角色",
          type: "switch",
          defaultValue: false,
          tooltip: "新平台用户创建时是否自动绑定",
        },
        {
          field: "sort",
          label: t.pages.platformRoles?.sortOrder || "排序",
          type: "number",
          defaultValue: 1,
        },
        {
          field: "status",
          label: t.pages.platformRoles?.status || "状态",
          type: "radio",
          defaultValue: 1,
          options: [
            { label: t.common.enable || "启用", value: 1 },
            { label: t.common.disable || "停用", value: 0 },
          ],
        },
      ],
    }),
    [dialogMode],
  )

  // ---- Handlers ----
  const handleAdd = () => {
    setDialogMode("add")
    setCurrentRole(null)
    setDialogOpen(true)
  }

  const handleEdit = (role: PlatformRoleItem) => {
    setDialogMode("edit")
    setCurrentRole(role)
    setDialogOpen(true)
  }

  const handleDeleteConfirm = (role: PlatformRoleItem) => {
    setRoleToDelete(role)
    setDeleteDialogOpen(true)
  }

  const handleDelete = async () => {
    if (!roleToDelete) return
    setIsSubmitting(true)
    try {
      if (useMockFallback) {
        setRoles((prev) => prev.filter((r) => r.id !== roleToDelete.id))
        toast.success("已删除（Mock）")
      } else {
        await platformRoleApi.delete(roleToDelete.id)
        toast.success("删除成功")
      }
      setDeleteDialogOpen(false)
      if (!useMockFallback) await fetchRoles()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "删除失败")
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleSubmit = async (values: Record<string, unknown>) => {
    setIsSubmitting(true)
    try {
      const payload = {
        name: String(values.name ?? ""),
        description: (values.description as string) || "",
        data_scope: Number(values.data_scope) || 1,
        is_default: values.is_default === true || values.is_default === "true",
        sort: Number(values.sort) || 1,
        status: Number(values.status) || 1,
      }
      if (dialogMode === "add") {
        if (useMockFallback) {
          const newId = Math.max(0, ...roles.map((r) => r.id)) + 1
          setRoles((prev) => [
            ...prev,
            {
              id: newId,
              code: String(values.code ?? ""),
              ...payload,
              menus: [],
              permissions: [],
            },
          ])
          toast.success("已新建（Mock）")
        } else {
          await platformRoleApi.create({
            code: String(values.code ?? ""),
            ...payload,
          })
          toast.success("创建成功")
        }
      } else if (currentRole) {
        if (useMockFallback) {
          setRoles((prev) =>
            prev.map((r) => (r.id === currentRole.id ? { ...r, ...payload } : r)),
          )
          toast.success("已更新（Mock）")
        } else {
          await platformRoleApi.update(currentRole.id, payload)
          toast.success("更新成功")
        }
      }
      setDialogOpen(false)
      if (!useMockFallback) await fetchRoles()
    } catch (err) {
      const msg = err instanceof Error ? err.message : "保存失败"
      toast.error(msg)
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleAssignPermission = async (role: PlatformRoleItem) => {
    setCurrentPermRole(role)
    setSelectedMenuIds((role.menus ?? []).map((m) => m.id))
    setSelectedPermIds((role.permissions ?? []).map((p) => p.id))
    setMenuTree([])
    setAllPermissions([])
    setPermDialogOpen(true)
    setPermLoading(true)
    try {
      const [menus, perms] = await Promise.all([
        platformMenuApi.tree().catch(() => []),
        platformPermissionApi.list({ page: 1, size: 500 }).catch(() => ({ list: [], total: 0 })),
      ])
      setMenuTree(menus || [])
      setAllPermissions(perms?.list ?? [])
    } catch (err) {
      console.error("[PlatformRoles] load permission resources failed:", err)
      toast.error("加载权限资源失败")
    } finally {
      setPermLoading(false)
    }
  }

  const handleAssignPermissionSubmit = async () => {
    if (!currentPermRole) return
    setIsSubmitting(true)
    try {
      await Promise.all([
        platformRoleApi.assignMenus(currentPermRole.id, selectedMenuIds),
        platformRoleApi.assignPermissions(currentPermRole.id, selectedPermIds),
      ])
      toast.success("权限分配成功")
      setPermDialogOpen(false)
      await fetchRoles()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "权限分配失败")
    } finally {
      setIsSubmitting(false)
    }
  }

  const getInitialValues = () => {
    if (currentRole) {
      return {
        name: currentRole.name,
        code: currentRole.code,
        description: currentRole.description || "",
        data_scope: currentRole.data_scope ?? 1,
        is_default: currentRole.is_default,
        sort: currentRole.sort,
        status: currentRole.status,
      }
    }
    return { data_scope: 1, is_default: false, sort: 1, status: 1 }
  }

  // 菜单树辅助
  const getAllMenuIds = (): number[] => {
    const ids: number[] = []
    const walk = (ms: PlatformMenuItem[]) => {
      ms.forEach((m) => {
        ids.push(m.id)
        if (m.children?.length) walk(m.children)
      })
    }
    walk(menuTree)
    return ids
  }

  const handleSelectAllMenus = () => setSelectedMenuIds(getAllMenuIds())
  const handleDeselectAllMenus = () => setSelectedMenuIds([])

  const handleSelectAllPermissions = () => setSelectedPermIds(allPermissions.map((p) => p.id))
  const handleDeselectAllPermissions = () => setSelectedPermIds([])

  return (
    <PageLayout>
      <div className="px-4 lg:px-6">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-2xl font-bold flex items-center gap-2">
              <ShieldIcon className="h-6 w-6" />
              {t.pages.platformRoles?.title || "平台角色"}
            </h1>
            <p className="text-sm text-muted-foreground mt-1">
              {t.pages.platformRoles?.subtitle || "管理平台域角色及其菜单/权限码授权（仅 super_admin）"}
            </p>
          </div>
          <div className="flex items-center gap-2">
            <Button
              variant={useMockFallback ? "default" : "outline"}
              size="sm"
              onClick={() => setUseMockFallback((v) => !v)}
            >
              {useMockFallback ? "Mock 已开启" : "使用 Mock 兜底"}
            </Button>
            <Button variant="outline" size="sm" onClick={fetchRoles} disabled={isLoading}>
              <RefreshCw className={cn("h-4 w-4 mr-2", isLoading && "animate-spin")} />
              {t.pages.platformRoles?.refresh || "刷新"}
            </Button>
            <Button size="sm" onClick={handleAdd}>
              <PlusIcon className="h-4 w-4 mr-2" />
              {t.pages.platformRoles?.create || "新建平台角色"}
            </Button>
          </div>
        </div>

        <div className="grid grid-cols-3 gap-4 mb-4">
          <Card>
            <CardHeader className="pb-2">
              <CardDescription>{t.pages.platformRoles?.statsTotal || "角色总数"}</CardDescription>
              <CardTitle className="text-2xl">{stats.total}</CardTitle>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardDescription>{t.pages.platformRoles?.statsActive || "启用中"}</CardDescription>
              <CardTitle className="text-2xl text-green-600">{stats.active}</CardTitle>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardDescription>{t.pages.platformRoles?.statsDefault || "默认角色"}</CardDescription>
              <CardTitle className="text-2xl text-blue-600">{stats.default}</CardTitle>
            </CardHeader>
          </Card>
        </div>

        {error && (
          <div className="mb-4 flex items-start gap-2 rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-sm text-destructive">
            <AlertTriangleIcon className="h-4 w-4 mt-0.5 shrink-0" />
            <div className="flex-1">
              <div className="font-medium">加载失败</div>
              <div className="text-xs opacity-80">{error}</div>
            </div>
            <Button size="sm" variant="outline" onClick={fetchRoles}>
              重试
            </Button>
          </div>
        )}

        <Card>
          <CardHeader>
            <div className="flex items-center gap-4">
              <div className="relative flex-1 max-w-sm">
                <SearchIcon className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  placeholder={t.pages.platformRoles?.searchPlaceholder || "搜索 code / 名称 / 描述..."}
                  className="pl-9"
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                />
              </div>
              <Badge variant="secondary">共 {filteredRoles.length} 个角色</Badge>
              <div className="ml-auto flex items-center gap-2 text-sm text-muted-foreground">
                <span>数据源</span>
                <Badge variant={dataSource === "api" ? "default" : "outline"}>
                  {dataSource === "api" ? "实时" : dataSource === "mock" ? "Mock" : "—"}
                </Badge>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-[60px]">ID</TableHead>
                  <TableHead>{t.pages.platformRoles?.name || "角色名称"}</TableHead>
                  <TableHead>{t.pages.platformRoles?.code || "角色代码"}</TableHead>
                  <TableHead>描述</TableHead>
                  <TableHead>数据范围</TableHead>
                  <TableHead className="w-[80px]">默认</TableHead>
                  <TableHead className="w-[80px]">{t.pages.platformRoles?.sortOrder || "排序"}</TableHead>
                  <TableHead>{t.pages.platformRoles?.status || "状态"}</TableHead>
                  <TableHead className="text-right">{t.common.edit}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredRoles.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={9} className="text-center py-8 text-muted-foreground">
                      {isLoading ? t.common.loading : t.common.noData}
                    </TableCell>
                  </TableRow>
                ) : (
                  filteredRoles.map((role) => (
                    <TableRow key={role.id}>
                      <TableCell className="font-mono text-xs text-muted-foreground">{role.id}</TableCell>
                      <TableCell>
                        <div className="flex items-center gap-2">
                          <Badge variant="outline">{role.name}</Badge>
                          {role.code === "super_admin" && (
                            <Badge className="bg-amber-500 text-white text-[10px]">系统内置</Badge>
                          )}
                        </div>
                      </TableCell>
                      <TableCell>
                        <code className="px-1.5 py-0.5 rounded bg-muted text-xs font-mono">{role.code}</code>
                      </TableCell>
                      <TableCell className="text-sm text-muted-foreground max-w-[260px] truncate">
                        {role.description || "-"}
                      </TableCell>
                      <TableCell>
                        <Badge variant={role.data_scope === 1 ? "default" : "secondary"}>
                          {role.data_scope === 1 ? "全部" : role.data_scope === 5 ? "本人" : `#${role.data_scope}`}
                        </Badge>
                      </TableCell>
                      <TableCell>
                        {role.is_default ? (
                          <Badge variant="default" className="bg-blue-600">默认</Badge>
                        ) : (
                          <span className="text-xs text-muted-foreground/60">-</span>
                        )}
                      </TableCell>
                      <TableCell>{role.sort}</TableCell>
                      <TableCell>
                        <Badge variant={role.status === 1 ? "default" : "secondary"}>
                          {role.status === 1 ? "启用" : "停用"}
                        </Badge>
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="flex items-center justify-end gap-1">
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleAssignPermission(role)}
                            title="分配权限"
                          >
                            <KeyIcon className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-8 w-8"
                            onClick={() => handleEdit(role)}
                          >
                            <EditIcon className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-8 w-8"
                            onClick={() => handleDeleteConfirm(role)}
                            disabled={role.code === "super_admin"}
                            title={role.code === "super_admin" ? "系统内置角色不可删除" : "删除"}
                          >
                            <TrashIcon className="h-4 w-4 text-destructive" />
                          </Button>
                        </div>
                      </TableCell>
                    </TableRow>
                  ))
                )}
              </TableBody>
            </Table>
            {isLoading && (
              <div className="flex items-center justify-center py-8">
                <div className="text-sm text-muted-foreground">{t.common.loading}</div>
              </div>
            )}
          </CardContent>
        </Card>
      </div>

      <FormDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        title={
          dialogMode === "add"
            ? t.pages.platformRoles?.create || "新建平台角色"
            : t.pages.platformRoles?.edit || "编辑平台角色"
        }
        width={520}
        schema={roleFormSchema}
        initialValues={getInitialValues()}
        onSubmit={handleSubmit}
        loading={isSubmitting}
      />

      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent className="sm:max-w-[400px]">
          <DialogHeader>
            <DialogTitle>删除平台角色</DialogTitle>
            <DialogDescription>
              确定要删除平台角色 "{roleToDelete?.name}" 吗？已有用户绑定的角色将一并解绑。
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setDeleteDialogOpen(false)}>
              {t.common.cancel}
            </Button>
            <Button variant="destructive" onClick={handleDelete} disabled={isSubmitting}>
              {t.common.delete}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      <Dialog open={permDialogOpen} onOpenChange={setPermDialogOpen}>
        <DialogContent className="sm:max-w-[760px]">
          <DialogHeader>
            <DialogTitle>分配权限 - {currentPermRole?.name}</DialogTitle>
            <DialogDescription>
              为该平台角色授予 sys_menu 与 sys_permission。两次 PUT 全量覆盖现有授权。
            </DialogDescription>
          </DialogHeader>
          <div className="max-h-[60vh] overflow-y-auto py-2">
            <Tabs defaultValue="menus" className="w-full">
              <TabsList className="grid w-full grid-cols-2">
                <TabsTrigger value="menus">菜单权限</TabsTrigger>
                <TabsTrigger value="permissions">权限码</TabsTrigger>
              </TabsList>

              <TabsContent value="menus" className="space-y-3 pt-3">
                <div className="flex items-center justify-between bg-muted/50 p-2 rounded-md border">
                  <span className="text-sm font-semibold">
                    已选 {selectedMenuIds.length} / 共 {getAllMenuIds().length} 个菜单
                  </span>
                  <div className="flex items-center gap-1">
                    <Button variant="outline" size="sm" onClick={handleSelectAllMenus} className="h-7 text-xs">
                      <CheckSquare className="w-3.5 h-3.5 mr-1" />全选
                    </Button>
                    <Button variant="outline" size="sm" onClick={handleDeselectAllMenus} className="h-7 text-xs">
                      <Square className="w-3.5 h-3.5 mr-1" />全不选
                    </Button>
                  </div>
                </div>
                {permLoading ? (
                  <p className="text-sm text-muted-foreground py-4">{t.common.loading}</p>
                ) : menuTree.length === 0 ? (
                  <p className="text-sm text-muted-foreground py-4">暂无平台菜单</p>
                ) : (
                  <div className="border rounded-md p-3 bg-card max-h-[50vh] overflow-y-auto">
                    <PlatformMenuPermTree
                      menus={menuTree}
                      selectedIds={selectedMenuIds}
                      onChange={setSelectedMenuIds}
                    />
                  </div>
                )}
              </TabsContent>

              <TabsContent value="permissions" className="space-y-3 pt-3">
                <div className="flex items-center justify-between bg-muted/50 p-2 rounded-md border">
                  <span className="text-sm font-semibold">
                    已选 {selectedPermIds.length} / 共 {allPermissions.length} 个权限码
                  </span>
                  <div className="flex items-center gap-1">
                    <Button variant="outline" size="sm" onClick={handleSelectAllPermissions} className="h-7 text-xs">
                      <CheckSquare className="w-3.5 h-3.5 mr-1" />全选
                    </Button>
                    <Button variant="outline" size="sm" onClick={handleDeselectAllPermissions} className="h-7 text-xs">
                      <Square className="w-3.5 h-3.5 mr-1" />全不选
                    </Button>
                  </div>
                </div>
                {permLoading ? (
                  <p className="text-sm text-muted-foreground py-4">{t.common.loading}</p>
                ) : allPermissions.length === 0 ? (
                  <p className="text-sm text-muted-foreground py-4">暂无权限码，请先在「平台权限码」页创建</p>
                ) : (
                  <div className="border rounded-md p-3 bg-card max-h-[50vh] overflow-y-auto space-y-1">
                    {allPermissions.map((p) => (
                      <div key={p.id} className="flex items-center gap-2 py-1">
                        <Checkbox
                          id={`sysperm-${p.id}`}
                          checked={selectedPermIds.includes(p.id)}
                          onCheckedChange={(c) => {
                            setSelectedPermIds((prev) =>
                              c === true
                                ? prev.includes(p.id) ? prev : [...prev, p.id]
                                : prev.filter((id) => id !== p.id),
                            )
                          }}
                        />
                        <label
                          htmlFor={`sysperm-${p.id}`}
                          className="text-sm cursor-pointer flex items-center gap-2 flex-1"
                        >
                          <span>{p.name}</span>
                          <code className="px-1 rounded bg-muted text-[10px] font-mono">{p.code}</code>
                          <Badge variant="outline" className="text-[10px] h-4 px-1 py-0">{p.action}</Badge>
                          {p.status === 0 && (
                            <Badge variant="secondary" className="text-[10px] h-4 px-1 py-0">停用</Badge>
                          )}
                        </label>
                      </div>
                    ))}
                  </div>
                )}
              </TabsContent>
            </Tabs>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setPermDialogOpen(false)}>
              {t.common.cancel}
            </Button>
            <Button onClick={handleAssignPermissionSubmit} disabled={isSubmitting}>
              {t.common.save}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </PageLayout>
  )
}

// 平台菜单权限树（继承父级自动级联子级）
function PlatformMenuPermTree({
  menus,
  selectedIds,
  onChange,
  level = 0,
}: {
  menus: PlatformMenuItem[]
  selectedIds: number[]
  onChange: (ids: number[]) => void
  level?: number
}) {
  return (
    <div className="space-y-1">
      {menus.map((menu) => (
        <PlatformMenuPermNode
          key={menu.id}
          menu={menu}
          level={level}
          selectedIds={selectedIds}
          onChange={onChange}
        />
      ))}
    </div>
  )
}

function PlatformMenuPermNode({
  menu,
  level,
  selectedIds,
  onChange,
}: {
  menu: PlatformMenuItem
  level: number
  selectedIds: number[]
  onChange: (ids: number[]) => void
}) {
  const [isExpanded, setIsExpanded] = useState(true)
  const hasChildren = !!(menu.children && menu.children.length)
  const isSelected = selectedIds.includes(menu.id)

  const childIds: number[] = []
  const collect = (m: PlatformMenuItem) => {
    childIds.push(m.id)
    m.children?.forEach(collect)
  }
  if (hasChildren) {
    menu.children!.forEach(collect)
  }
  const someChildSelected = childIds.some((id) => selectedIds.includes(id))

  const toggle = (checked: boolean) => {
    let next: Set<number>
    if (checked) {
      next = new Set(selectedIds)
      next.add(menu.id)
      childIds.forEach((id) => next.add(id))
    } else {
      next = new Set(selectedIds)
      next.delete(menu.id)
      childIds.forEach((id) => next.delete(id))
    }
    onChange(Array.from(next))
  }

  return (
    <div>
      <div
        className="flex items-center gap-2 py-1 hover:bg-muted/50 rounded-sm px-2"
        style={{ paddingLeft: `${level * 1.2 + 0.5}rem` }}
      >
        {hasChildren ? (
          <button
            onClick={() => setIsExpanded(!isExpanded)}
            className="p-0.5 hover:bg-muted rounded text-muted-foreground"
          >
            {isExpanded ? <ChevronDown className="w-4 h-4" /> : <ChevronRight className="w-4 h-4" />}
          </button>
        ) : (
          <div className="w-5" />
        )}
        <Checkbox
          checked={isSelected ? true : someChildSelected ? "indeterminate" : false}
          onCheckedChange={(c) => toggle(c === true)}
          id={`sysmenu-${menu.id}`}
        />
        <label
          htmlFor={`sysmenu-${menu.id}`}
          className="text-sm cursor-pointer select-none flex items-center gap-2 flex-1"
        >
          {menu.name}
          <code className="px-1 rounded bg-muted text-[10px] font-mono">{menu.code}</code>
        </label>
      </div>
      {hasChildren && isExpanded && (
        <div>
          {menu.children!.map((child) => (
            <PlatformMenuPermNode
              key={child.id}
              menu={child}
              level={level + 1}
              selectedIds={selectedIds}
              onChange={onChange}
            />
          ))}
        </div>
      )}
    </div>
  )
}