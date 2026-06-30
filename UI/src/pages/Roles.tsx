import { useEffect, useState, useCallback } from "react"
import { PageLayout } from "@/components/page-layout"
import {
  Card,
  CardContent,
  CardDescription,
  CardHeader,
  CardTitle,
} from "@/components/ui/card"
import { Button } from "@/components/ui/button"
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
  ShieldIcon,
  UsersIcon,
  LayoutDashboardIcon,
  EditIcon,
  TrashIcon,
  RefreshCw,
  CheckSquare,
  Square,
  ChevronDown,
  ChevronRight,
  AlertCircleIcon,
} from "lucide-react"
import { toast } from "sonner"
import { t } from "@/locales"
import {
  roleApi,
  menuApi,
  resourceApi,
  organizationApi,
  type RoleItem,
  type MenuItem,
  type ResourceItem,
  type OrganizationItem,
  ApiError,
} from "@/api"
import { FormDialog } from "@/components/schema/DynamicForm"
import type { FormSchema } from "@/types/schema"
import { Checkbox } from "@/components/ui/checkbox"
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from "@/components/ui/select"
import { Label } from "@/components/ui/label"
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "@/components/ui/dialog"

export function RolesPage() {
  const [roles, setRoles] = useState<RoleItem[]>([])
  const [stats, setStats] = useState({
    roleCount: 0,
    userCount: 0,
    activeUserCount: 0,
    permissionCount: 0,
  })
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  const [dialogOpen, setDialogOpen] = useState(false)
  const [dialogMode, setDialogMode] = useState<"add" | "edit">("add")
  const [currentRole, setCurrentRole] = useState<RoleItem | null>(null)

  const [permDialogOpen, setPermDialogOpen] = useState(false)
  const [currentPermRole, setCurrentPermRole] = useState<RoleItem | null>(null)
  const [menuOptions, setMenuOptions] = useState<MenuItem[]>([])
  const [selectedMenus, setSelectedMenus] = useState<number[]>([])
  const [resources, setResources] = useState<ResourceItem[]>([])
  const [selectedResources, setSelectedResources] = useState<number[]>([])

  const [orgTree, setOrgTree] = useState<OrganizationItem[]>([])
  const [selectedOrgIds, setSelectedOrgIds] = useState<number[]>([])
  const [permDataScope, setPermDataScope] = useState<number>(1)

  const fetchRoles = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const response = await roleApi.list({ page: 1, size: 100 })
      setRoles(response?.list || [])
      setStats((prev) => ({ ...prev, roleCount: response?.total || 0 }))
    } catch (err) {
      const msg =
        err instanceof ApiError
          ? err.message
          : err instanceof Error
            ? err.message
            : String(err)
      setError(`加载角色失败：${msg}`)
      setRoles([])
    } finally {
      setIsLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchRoles()
  }, [fetchRoles])

  const roleFormSchema: FormSchema = {
    items: [
      {
        field: "name",
        label: t.pages.roles?.name || "角色名称",
        type: "text",
        required: true,
        placeholder: "请输入角色名称",
      },
      {
        field: "code",
        label: t.pages.roles?.code || "角色代码",
        type: "text",
        required: true,
        placeholder: "请输入角色代码，如 admin",
        props: { pattern: "^[a-z_]+$" },
      },
      {
        field: "description",
        label: t.pages.roles?.description || "描述",
        type: "textarea",
        placeholder: "请输入角色描述",
      },
      {
        field: "data_scope",
        label: t.pages.roles?.dataScope || "数据范围",
        type: "select",
        defaultValue: 1,
        tooltip: t.pages.roles?.dataScopeTip,
        options: [
          { label: t.pages.roles?.dataScopeAll || "全部数据", value: 1 },
          { label: t.pages.roles?.dataScopeCustom || "自定义", value: 2 },
          { label: t.pages.roles?.dataScopeDept || "本部门", value: 3 },
          {
            label: t.pages.roles?.dataScopeDeptAndBelow || "本部门及以下",
            value: 4,
          },
          { label: t.pages.roles?.dataScopeSelf || "本人数据", value: 5 },
        ],
      },
      {
        field: "sort",
        label: t.pages.roles?.sortOrder || "排序",
        type: "number",
        defaultValue: 1,
      },
      {
        field: "status",
        label: t.pages.roles?.status || "状态",
        type: "radio",
        defaultValue: 1,
        options: [
          { label: "启用", value: 1 },
          { label: "停用", value: 0 },
        ],
      },
    ],
  }

  const handleAdd = () => {
    setDialogMode("add")
    setCurrentRole(null)
    setDialogOpen(true)
  }

  const handleEdit = (role: RoleItem) => {
    setDialogMode("edit")
    setCurrentRole(role)
    setDialogOpen(true)
  }

  const handleDelete = async (role: RoleItem) => {
    if (!confirm(`确定要删除角色 "${role.name}" 吗？`)) return
    try {
      await roleApi.delete(role.id)
      await fetchRoles()
      toast.success("删除成功")
    } catch (error) {
      const msg =
        error instanceof ApiError
          ? error.message
          : error instanceof Error
            ? error.message
            : "删除失败，请重试"
      toast.error(msg)
    }
  }

  const handleSubmit = async (values: Record<string, unknown>) => {
    setIsSubmitting(true)
    try {
      const payload: Record<string, unknown> = { ...values }
      if (payload.data_scope !== undefined && payload.data_scope !== null) {
        const n = Number(payload.data_scope)
        payload.data_scope = Number.isFinite(n) ? n : payload.data_scope
      }
      if (dialogMode === "add") {
        await roleApi.create(payload as Partial<RoleItem>)
      } else if (currentRole) {
        await roleApi.update(currentRole.id, payload as Partial<RoleItem>)
      }
      await fetchRoles()
      setDialogOpen(false)
      toast.success(dialogMode === "add" ? "创建成功" : "更新成功")
    } catch (error) {
      const msg =
        error instanceof ApiError
          ? error.message
          : error instanceof Error
            ? error.message
            : "保存失败，请重试"
      toast.error(msg)
    } finally {
      setIsSubmitting(false)
    }
  }

  const handlePermission = async (role: RoleItem) => {
    setCurrentPermRole(role)
    setPermDialogOpen(true)
    setSelectedMenus([])
    setMenuOptions([])
    setSelectedResources([])
    setResources([])
    setOrgTree([])
    setSelectedOrgIds([])
    setPermDataScope(role.data_scope ?? 1)
    try {
      const [
        menus,
        permData,
        allResources,
        resPermData,
        orgTreeData,
        scopeData,
      ] = await Promise.all([
        menuApi.tree().catch(() => []),
        roleApi.getMenus(role.id).catch(() => ({ menu_ids: [] })),
        resourceApi
          .list({ page: 1, size: 1000 })
          .catch(() => ({ list: [], total: 0 })),
        roleApi.getPermissions(role.id).catch(() => ({ list: [] })),
        organizationApi.tree().catch(() => ({ tree: [] })),
        roleApi.getDataScopes(role.id).catch(() => ({ org_ids: [] })),
      ])
      setMenuOptions(menus || [])
      setSelectedMenus(permData?.menu_ids || [])
      setResources(allResources?.list || [])
      setSelectedResources(resPermData?.list?.map((r) => r.id) || [])
      setOrgTree(orgTreeData?.tree || [])
      setSelectedOrgIds(scopeData?.org_ids || [])
    } catch (error) {
      console.error("Load permissions failed:", error)
      setMenuOptions([])
      setSelectedMenus([])
      setResources([])
      setSelectedResources([])
      setOrgTree([])
      setSelectedOrgIds([])
    }
  }

  const handlePermissionSubmit = async () => {
    if (!currentPermRole) return
    setIsSubmitting(true)
    try {
      const scopeNum = Number(permDataScope)
      const tasks: Promise<unknown>[] = [
        roleApi.setMenus(currentPermRole.id, selectedMenus),
        roleApi.setPermissions(currentPermRole.id, selectedResources),
        // 仅 data_scope=2（自定义）时同步选中的组织 ID 列表
        ...(Number.isFinite(scopeNum) && scopeNum === 2
          ? [roleApi.setDataScopes(currentPermRole.id, selectedOrgIds)]
          : []),
        // data_scope 整型（1~5）通过 PATCH /roles/:id 写入 roles.data_scope 列
        roleApi.patch(currentPermRole.id, {
          data_scope: scopeNum,
        } as Partial<RoleItem>),
      ]
      await Promise.all(tasks)
      await fetchRoles()
      setPermDialogOpen(false)
    } catch (error) {
      const msg =
        error instanceof ApiError
          ? error.message
          : error instanceof Error
            ? error.message
            : "保存权限失败，请重试"
      toast.error(msg)
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
        sort: currentRole.sort,
        status: currentRole.status,
      }
    }
    return { data_scope: 1 }
  }

  const getDataScopeLabel = (scope?: number) => {
    switch (scope) {
      case 1:
        return t.pages.roles?.dataScopeAll || "全部数据"
      case 2:
        return t.pages.roles?.dataScopeCustom || "自定义"
      case 3:
        return t.pages.roles?.dataScopeDept || "本部门"
      case 4:
        return t.pages.roles?.dataScopeDeptAndBelow || "本部门及以下"
      case 5:
        return t.pages.roles?.dataScopeSelf || "本人数据"
      default:
        return "-"
    }
  }

  const getAllMenuIds = (): number[] => {
    const ids: number[] = []
    const collect = (menus: MenuItem[]) => {
      menus.forEach((menu) => {
        ids.push(menu.id)
        if (menu.children && menu.children.length > 0) {
          collect(menu.children)
        }
      })
    }
    collect(menuOptions)
    return ids
  }

  const handleSelectAllMenus = () => {
    setSelectedMenus(getAllMenuIds())
  }

  const handleDeselectAllMenus = () => {
    setSelectedMenus([])
  }

  const handleMenuToggle = (menuId: number, checked: boolean) => {
    if (checked) {
      setSelectedMenus((prev) => [...prev, menuId])
    } else {
      setSelectedMenus((prev) => prev.filter((id) => id !== menuId))
    }
  }

  const handleSelectChildren = (menuId: number, childIds: number[]) => {
    setSelectedMenus((prev) => {
      const newSet = new Set(prev)
      newSet.add(menuId)
      childIds.forEach((id) => newSet.add(id))
      return Array.from(newSet)
    })
  }

  const handleDeselectChildren = (menuId: number, childIds: number[]) => {
    setSelectedMenus((prev) =>
      prev.filter((id) => id !== menuId && !childIds.includes(id))
    )
  }

  const handleSelectAllResources = () => {
    setSelectedResources(resources.map((r) => r.id))
  }

  const handleDeselectAllResources = () => {
    setSelectedResources([])
  }

  const handleResourceToggle = (resourceId: number, checked: boolean) => {
    if (checked) {
      setSelectedResources((prev) => [...prev, resourceId])
    } else {
      setSelectedResources((prev) => prev.filter((id) => id !== resourceId))
    }
  }

  const handleSelectMenuResources = (menuId: number, checked: boolean) => {
    const menuResourceIds = resources
      .filter((r) => (r.menu_id || 0) === menuId)
      .map((r) => r.id)
    if (checked) {
      setSelectedResources((prev) => {
        const newSet = new Set(prev)
        menuResourceIds.forEach((id) => newSet.add(id))
        return Array.from(newSet)
      })
    } else {
      setSelectedResources((prev) =>
        prev.filter((id) => !menuResourceIds.includes(id))
      )
    }
  }

  const handleOrgToggle = (orgId: number, checked: boolean) => {
    if (checked) {
      setSelectedOrgIds((prev) =>
        prev.includes(orgId) ? prev : [...prev, orgId]
      )
    } else {
      setSelectedOrgIds((prev) => prev.filter((id) => id !== orgId))
    }
  }

  const handleSelectAllOrgs = () => {
    const ids: number[] = []
    const collect = (list: OrganizationItem[]) => {
      list.forEach((o) => {
        ids.push(o.id)
        if (o.children) collect(o.children)
      })
    }
    collect(orgTree)
    setSelectedOrgIds(Array.from(new Set(ids)))
  }

  const handleDeselectAllOrgs = () => {
    setSelectedOrgIds([])
  }

  return (
    <PageLayout>
      <div className="px-4 lg:px-6">
        <div className="mb-6 flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold">
              {t.pages.roles?.title || "角色管理"}
            </h1>
            <p className="text-sm text-muted-foreground">
              {t.pages.roles?.subtitle || "管理系统角色和权限分配"}
            </p>
          </div>
          <div className="flex items-center gap-2">
            <Button variant="outline" onClick={fetchRoles} disabled={isLoading}>
              <RefreshCw
                className={`mr-2 h-4 w-4 ${isLoading ? "animate-spin" : ""}`}
              />
              {t.pages.roles?.refresh || "刷新列表"}
            </Button>
            <Button onClick={handleAdd}>
              <PlusIcon className="mr-2 h-4 w-4" />
              {t.common.add}
            </Button>
          </div>
        </div>

        <div className="mb-6 grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          {error && (
            <div className="flex items-start gap-2 rounded-md border border-destructive/40 bg-destructive/5 p-3 text-sm md:col-span-2 lg:col-span-4">
              <AlertCircleIcon className="mt-0.5 h-4 w-4 shrink-0 text-destructive" />
              <div className="min-w-0 flex-1">
                <div className="font-medium text-destructive">接口调用失败</div>
                <div className="mt-0.5 text-xs break-all text-muted-foreground">
                  {error}
                </div>
              </div>
              <Button variant="ghost" size="sm" onClick={fetchRoles}>
                重试
              </Button>
            </div>
          )}
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">总角色数</CardTitle>
              <ShieldIcon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.roleCount}</div>
              <p className="text-xs text-muted-foreground">
                {t.pages.roles?.userCount || "角色数量"}
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">总用户数</CardTitle>
              <UsersIcon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.userCount}</div>
              <p className="text-xs text-muted-foreground">+12 本月</p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">活跃用户</CardTitle>
              <UsersIcon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.activeUserCount}</div>
              <p className="text-xs text-muted-foreground">
                {stats.userCount > 0
                  ? `${((stats.activeUserCount / stats.userCount) * 100).toFixed(1)}% 活跃率`
                  : "—"}
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">权限总数</CardTitle>
              <LayoutDashboardIcon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{stats.permissionCount}</div>
              <p className="text-xs text-muted-foreground">分布在各角色</p>
            </CardContent>
          </Card>
        </div>

        <Card>
          <CardHeader>
            <CardTitle>{t.pages.roles?.roleList || "角色列表"}</CardTitle>
            <CardDescription>
              {t.pages.roles?.roleDesc || "管理系统中的所有角色"}
            </CardDescription>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>ID</TableHead>
                  <TableHead>{t.pages.roles?.name || "角色名称"}</TableHead>
                  <TableHead>{t.pages.roles?.code || "角色代码"}</TableHead>
                  <TableHead>{t.pages.roles?.description || "描述"}</TableHead>
                  <TableHead>
                    {t.pages.roles?.dataScope || "数据范围"}
                  </TableHead>
                  <TableHead>{t.pages.roles?.sortOrder || "排序"}</TableHead>
                  <TableHead>{t.pages.roles?.status || "状态"}</TableHead>
                  <TableHead className="text-right">{t.common.edit}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {roles.map((role) => (
                  <TableRow key={role.id}>
                    <TableCell className="font-medium">{role.id}</TableCell>
                    <TableCell>
                      <Badge variant="outline">{role.name}</Badge>
                    </TableCell>
                    <TableCell className="font-mono text-sm">
                      {role.code}
                    </TableCell>
                    <TableCell>{role.description || "-"}</TableCell>
                    <TableCell>
                      <Badge
                        variant={
                          role.data_scope === 2 ? "default" : "secondary"
                        }
                      >
                        {getDataScopeLabel(role.data_scope)}
                      </Badge>
                    </TableCell>
                    <TableCell>{role.sort}</TableCell>
                    <TableCell>
                      <Badge
                        variant={role.status === 1 ? "default" : "secondary"}
                      >
                        {role.status === 1 ? "启用" : "停用"}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex items-center justify-end gap-1">
                        <Button
                          variant="ghost"
                          size="sm"
                          onClick={() => handlePermission(role)}
                        >
                          权限
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
                          onClick={() => handleDelete(role)}
                        >
                          <TrashIcon className="h-4 w-4 text-destructive" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
                {roles.length === 0 && !isLoading && (
                  <TableRow>
                    <TableCell
                      colSpan={8}
                      className="py-8 text-center text-muted-foreground"
                    >
                      {t.common.noData}
                    </TableCell>
                  </TableRow>
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
        title={
          dialogMode === "add"
            ? t.pages.roles?.addRole || "添加角色"
            : t.pages.roles?.editRole || "编辑角色"
        }
        schema={roleFormSchema}
        initialValues={getInitialValues()}
        onSubmit={handleSubmit}
        loading={isSubmitting}
      />

      <Dialog open={permDialogOpen} onOpenChange={setPermDialogOpen}>
        <DialogContent className="sm:max-w-[700px]">
          <DialogHeader>
            <DialogTitle>
              {t.pages.roles?.assignPermissions || "分配权限"} -{" "}
              {currentPermRole?.name}
            </DialogTitle>
          </DialogHeader>
          <div className="max-h-[60vh] overflow-y-auto py-4">
            <Tabs defaultValue="menus" className="w-full">
              <TabsList className="grid w-full grid-cols-3">
                <TabsTrigger value="menus">
                  {t.pages.roles?.menuPermissions || "菜单权限"}
                </TabsTrigger>
                <TabsTrigger value="resources">资源权限</TabsTrigger>
                <TabsTrigger value="dataScope">
                  {t.pages.roles?.dataPermissions || "数据范围"}
                </TabsTrigger>
              </TabsList>

              <TabsContent value="menus" className="space-y-4 pt-4">
                <div className="mb-4 flex items-center justify-between rounded-md border bg-muted/50 p-2">
                  <div className="flex items-center gap-2">
                    <div className="h-4 w-1 rounded-full bg-primary"></div>
                    <h4 className="text-sm font-semibold">
                      {t.pages.roles?.menuPermissions || "菜单权限"}
                    </h4>
                  </div>
                  <div className="flex items-center gap-1">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={handleSelectAllMenus}
                      className="h-7 text-xs"
                    >
                      <CheckSquare className="mr-1 h-3.5 w-3.5" />
                      {t.pages.roles?.selectAll || "全选"}
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={handleDeselectAllMenus}
                      className="h-7 text-xs"
                    >
                      <Square className="mr-1 h-3.5 w-3.5" />
                      {t.pages.roles?.deselectAll || "全不选"}
                    </Button>
                  </div>
                </div>
                {menuOptions.length === 0 ? (
                  <p className="py-4 text-sm text-muted-foreground">
                    {t.common.loading}
                  </p>
                ) : (
                  <div className="space-y-4">
                    <MenuPermissionTree
                      menus={menuOptions}
                      selectedMenus={selectedMenus}
                      onToggle={handleMenuToggle}
                      onSelectChildren={handleSelectChildren}
                      onDeselectChildren={handleDeselectChildren}
                    />
                  </div>
                )}
              </TabsContent>

              <TabsContent value="resources" className="space-y-4 pt-4">
                <div className="mb-4 flex items-center justify-between rounded-md border bg-muted/50 p-2">
                  <div className="flex items-center gap-2">
                    <div className="h-4 w-1 rounded-full bg-primary"></div>
                    <h4 className="text-sm font-semibold">操作资源权限</h4>
                  </div>
                  <div className="flex items-center gap-1">
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={handleSelectAllResources}
                      className="h-7 text-xs"
                    >
                      <CheckSquare className="mr-1 h-3.5 w-3.5" />
                      {t.pages.roles?.selectAll || "全选"}
                    </Button>
                    <Button
                      variant="outline"
                      size="sm"
                      onClick={handleDeselectAllResources}
                      className="h-7 text-xs"
                    >
                      <Square className="mr-1 h-3.5 w-3.5" />
                      {t.pages.roles?.deselectAll || "全不选"}
                    </Button>
                  </div>
                </div>
                {resources.length === 0 ? (
                  <p className="py-4 text-sm text-muted-foreground">
                    {t.common.loading}
                  </p>
                ) : (
                  <div className="space-y-6">
                    <ResourceGroupList
                      menus={menuOptions}
                      resources={resources}
                      selectedResources={selectedResources}
                      onToggle={handleResourceToggle}
                      onSelectMenuResources={handleSelectMenuResources}
                    />
                  </div>
                )}
              </TabsContent>

              <TabsContent value="dataScope" className="space-y-4 pt-4">
                <div className="mb-4 flex items-center gap-2 rounded-md border bg-muted/50 p-2">
                  <div className="h-4 w-1 rounded-full bg-primary"></div>
                  <h4 className="text-sm font-semibold">
                    {t.pages.roles?.dataScope || "数据范围"}
                  </h4>
                </div>

                <div className="grid gap-2">
                  <Label htmlFor="perm-data-scope" className="text-sm">
                    {t.pages.roles?.dataScope || "数据范围"}
                    <span className="ml-1 text-destructive">*</span>
                  </Label>
                  <Select
                    value={String(permDataScope)}
                    onValueChange={(v) => setPermDataScope(Number(v))}
                  >
                    <SelectTrigger
                      id="perm-data-scope"
                      className="w-full sm:w-[280px]"
                    >
                      <SelectValue
                        placeholder={t.pages.roles?.dataScope || "数据范围"}
                      />
                    </SelectTrigger>
                    <SelectContent>
                      <SelectItem value="1">
                        {t.pages.roles?.dataScopeAll || "全部数据"}
                      </SelectItem>
                      <SelectItem value="2">
                        {t.pages.roles?.dataScopeCustom || "自定义"}
                      </SelectItem>
                      <SelectItem value="3">
                        {t.pages.roles?.dataScopeDept || "本部门"}
                      </SelectItem>
                      <SelectItem value="4">
                        {t.pages.roles?.dataScopeDeptAndBelow || "本部门及以下"}
                      </SelectItem>
                      <SelectItem value="5">
                        {t.pages.roles?.dataScopeSelf || "本人数据"}
                      </SelectItem>
                    </SelectContent>
                  </Select>
                  <p className="text-xs text-muted-foreground">
                    {t.pages.roles?.dataScopeTip ||
                      "当数据范围为「自定义」时，指定可访问的组织列表"}
                  </p>
                </div>

                {permDataScope === 2 && (
                  <>
                    <div className="mt-4 flex items-center justify-between rounded-md border bg-muted/50 p-2">
                      <div className="flex items-center gap-2">
                        <div className="h-4 w-1 rounded-full bg-primary"></div>
                        <h4 className="text-sm font-semibold">
                          {t.pages.roles?.selectOrgs || "选择组织"}
                        </h4>
                      </div>
                      <div className="flex items-center gap-1">
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={handleSelectAllOrgs}
                          className="h-7 text-xs"
                        >
                          <CheckSquare className="mr-1 h-3.5 w-3.5" />
                          {t.pages.roles?.selectAll || "全选"}
                        </Button>
                        <Button
                          variant="outline"
                          size="sm"
                          onClick={handleDeselectAllOrgs}
                          className="h-7 text-xs"
                        >
                          <Square className="mr-1 h-3.5 w-3.5" />
                          {t.pages.roles?.deselectAll || "全不选"}
                        </Button>
                      </div>
                    </div>
                    {orgTree.length === 0 ? (
                      <p className="py-4 text-sm text-muted-foreground">
                        {t.common.loading}
                      </p>
                    ) : (
                      <div className="max-h-[40vh] space-y-1 overflow-y-auto rounded-md border bg-card p-3">
                        <OrgPermissionTree
                          orgs={orgTree}
                          selectedOrgIds={selectedOrgIds}
                          onToggle={handleOrgToggle}
                        />
                      </div>
                    )}
                    {selectedOrgIds.length > 0 && (
                      <p className="text-xs text-muted-foreground">
                        {t.pages.roles?.selectAll}: {selectedOrgIds.length}
                      </p>
                    )}
                  </>
                )}

                {permDataScope !== 2 && (
                  <div className="rounded-md border border-dashed bg-muted/30 p-4 text-sm text-muted-foreground">
                    {t.pages.roles?.dataScopeNonCustomHint ||
                      "当前数据范围非「自定义」，不需要选择组织。"}
                  </div>
                )}
              </TabsContent>
            </Tabs>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setPermDialogOpen(false)}>
              {t.common.cancel}
            </Button>
            <Button onClick={handlePermissionSubmit} disabled={isSubmitting}>
              {t.common.save}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </PageLayout>
  )
}

interface MenuPermissionTreeProps {
  menus: MenuItem[]
  level?: number
  selectedMenus: number[]
  onToggle: (menuId: number, checked: boolean) => void
  onSelectChildren: (menuId: number, childIds: number[]) => void
  onDeselectChildren: (menuId: number, childIds: number[]) => void
}

function MenuPermissionTree({
  menus,
  level = 0,
  selectedMenus,
  onToggle,
  onSelectChildren,
  onDeselectChildren,
}: MenuPermissionTreeProps) {
  return (
    <div className="space-y-1">
      {menus.map((menu) => (
        <MenuPermissionNode
          key={menu.id}
          menu={menu}
          level={level}
          selectedMenus={selectedMenus}
          onToggle={onToggle}
          onSelectChildren={onSelectChildren}
          onDeselectChildren={onDeselectChildren}
        />
      ))}
    </div>
  )
}

interface MenuPermissionNodeProps extends Omit<
  MenuPermissionTreeProps,
  "menus"
> {
  menu: MenuItem
}

function MenuPermissionNode({
  menu,
  level = 0,
  selectedMenus,
  onToggle,
  onSelectChildren,
  onDeselectChildren,
}: MenuPermissionNodeProps) {
  const [isExpanded, setIsExpanded] = useState(true)

  const hasChildren = menu.children && menu.children.length > 0
  const isSelected = selectedMenus.includes(menu.id)

  const childIds: number[] = []
  const collectChildIds = (items: MenuItem[]) => {
    items.forEach((item) => {
      childIds.push(item.id)
      if (item.children) collectChildIds(item.children)
    })
  }
  if (hasChildren) {
    collectChildIds(menu.children!)
  }

  const someChildrenSelected =
    childIds.length > 0 && childIds.some((id) => selectedMenus.includes(id))

  const handleCheckedChange = (checked: boolean | "indeterminate") => {
    onToggle(menu.id, checked === true)
    if (checked === true) {
      onSelectChildren(menu.id, childIds)
    } else if (checked === false) {
      onDeselectChildren(menu.id, childIds)
    }
  }

  return (
    <div>
      <div
        className="group flex items-center gap-2 rounded-sm px-2 py-1.5 hover:bg-muted/50"
        style={{ paddingLeft: `${level * 1.5 + 0.5}rem` }}
      >
        <div className="flex flex-1 items-center gap-2">
          {hasChildren ? (
            <button
              onClick={() => setIsExpanded(!isExpanded)}
              className="rounded p-0.5 text-muted-foreground hover:bg-muted"
            >
              {isExpanded ? (
                <ChevronDown className="h-4 w-4" />
              ) : (
                <ChevronRight className="h-4 w-4" />
              )}
            </button>
          ) : (
            <div className="w-5" />
          )}
          <Checkbox
            checked={
              isSelected ? true : someChildrenSelected ? "indeterminate" : false
            }
            onCheckedChange={handleCheckedChange}
            id={`menu-${menu.id}`}
          />
          <label
            htmlFor={`menu-${menu.id}`}
            className="flex flex-1 cursor-pointer items-center gap-2 text-sm select-none"
          >
            {menu.name}
            {(menu as MenuItem & { type?: number })?.type === 2 && (
              <Badge variant="secondary" className="h-4 px-1 text-[10px]">
                按钮
              </Badge>
            )}
            {(menu as MenuItem & { type?: number })?.type === 1 && (
              <Badge variant="outline" className="h-4 px-1 text-[10px]">
                菜单
              </Badge>
            )}
          </label>
        </div>
      </div>

      {hasChildren && isExpanded && (
        <div className="mt-1">
          {menu.children!.map((child) => (
            <MenuPermissionNode
              key={child.id}
              menu={child}
              level={level + 1}
              selectedMenus={selectedMenus}
              onToggle={onToggle}
              onSelectChildren={onSelectChildren}
              onDeselectChildren={onDeselectChildren}
            />
          ))}
        </div>
      )}
    </div>
  )
}

interface ResourceGroupListProps {
  menus: MenuItem[]
  resources: ResourceItem[]
  selectedResources: number[]
  onToggle: (resourceId: number, checked: boolean) => void
  onSelectMenuResources: (menuId: number, checked: boolean) => void
}

function ResourceGroupList({
  menus,
  resources,
  selectedResources,
  onToggle,
  onSelectMenuResources,
}: ResourceGroupListProps) {
  const menuMap = new Map<number, MenuItem>()
  const flatMenus = (list: MenuItem[]) => {
    list.forEach((m) => {
      menuMap.set(m.id, m)
      if (m.children) flatMenus(m.children)
    })
  }
  flatMenus(menus)

  const grouped = new Map<number, ResourceItem[]>()
  resources.forEach((r) => {
    const menuId = r.menu_id || 0
    if (!grouped.has(menuId)) grouped.set(menuId, [])
    grouped.get(menuId)!.push(r)
  })

  const groupKeys = Array.from(grouped.keys()).sort((a, b) => {
    if (a === 0) return -1
    if (b === 0) return 1
    const menuA = menuMap.get(a)
    const menuB = menuMap.get(b)
    if (menuA && menuB) return menuA.sort - menuB.sort
    return a - b
  })

  return (
    <div className="space-y-4">
      {groupKeys.map((menuId) => {
        const groupResources = grouped.get(menuId)!
        const menuName =
          menuId === 0
            ? "公共资源"
            : menuMap.get(menuId)?.name || `未知菜单 ${menuId}`
        const allSelected = groupResources.every((r) =>
          selectedResources.includes(r.id)
        )
        const someSelected = groupResources.some((r) =>
          selectedResources.includes(r.id)
        )

        return (
          <div key={menuId} className="rounded-md border bg-card p-3">
            <div className="mb-3 flex items-center gap-2 border-b pb-2">
              <Checkbox
                checked={
                  allSelected ? true : someSelected ? "indeterminate" : false
                }
                onCheckedChange={(checked) =>
                  onSelectMenuResources(menuId, checked === true)
                }
                id={`menu-res-${menuId}`}
              />
              <label
                htmlFor={`menu-res-${menuId}`}
                className="cursor-pointer text-sm font-semibold"
              >
                {menuName}
              </label>
            </div>
            <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 lg:grid-cols-3">
              {groupResources.map((res) => (
                <div key={res.id} className="flex items-center gap-2">
                  <Checkbox
                    checked={selectedResources.includes(res.id)}
                    onCheckedChange={(checked) =>
                      onToggle(res.id, checked === true)
                    }
                    id={`res-${res.id}`}
                  />
                  <label
                    htmlFor={`res-${res.id}`}
                    className="flex min-w-0 flex-1 cursor-pointer items-center gap-1.5 text-sm text-muted-foreground"
                  >
                    <span className="truncate">{res.name}</span>
                    <Badge
                      variant="outline"
                      className="h-4 shrink-0 px-1 py-0 text-[10px] font-normal"
                    >
                      {res.action}
                    </Badge>
                  </label>
                </div>
              ))}
            </div>
          </div>
        )
      })}
    </div>
  )
}

interface OrgPermissionTreeProps {
  orgs: OrganizationItem[]
  level?: number
  selectedOrgIds: number[]
  onToggle: (orgId: number, checked: boolean) => void
}

function OrgPermissionTree({
  orgs,
  level = 0,
  selectedOrgIds,
  onToggle,
}: OrgPermissionTreeProps) {
  return (
    <div className="space-y-1">
      {orgs.map((org) => (
        <OrgPermissionNode
          key={org.id}
          org={org}
          level={level}
          selectedOrgIds={selectedOrgIds}
          onToggle={onToggle}
        />
      ))}
    </div>
  )
}

interface OrgPermissionNodeProps extends Omit<OrgPermissionTreeProps, "orgs"> {
  org: OrganizationItem
}

function OrgPermissionNode({
  org,
  level = 0,
  selectedOrgIds,
  onToggle,
}: OrgPermissionNodeProps) {
  const [isExpanded, setIsExpanded] = useState(true)
  const hasChildren = org.children && org.children.length > 0
  const isSelected = selectedOrgIds.includes(org.id)

  return (
    <div>
      <div
        className="group flex items-center gap-2 rounded-sm px-2 py-1.5 hover:bg-muted/50"
        style={{ paddingLeft: `${level * 1.5 + 0.5}rem` }}
      >
        <div className="flex flex-1 items-center gap-2">
          {hasChildren ? (
            <button
              onClick={() => setIsExpanded(!isExpanded)}
              className="rounded p-0.5 text-muted-foreground hover:bg-muted"
            >
              {isExpanded ? (
                <ChevronDown className="h-4 w-4" />
              ) : (
                <ChevronRight className="h-4 w-4" />
              )}
            </button>
          ) : (
            <div className="w-5" />
          )}
          <Checkbox
            checked={isSelected}
            onCheckedChange={(checked) => onToggle(org.id, checked === true)}
            id={`org-${org.id}`}
          />
          <label
            htmlFor={`org-${org.id}`}
            className="flex flex-1 cursor-pointer items-center gap-2 text-sm select-none"
          >
            {org.name}
            {org.status === 0 && (
              <Badge variant="secondary" className="h-4 px-1 text-[10px]">
                停用
              </Badge>
            )}
          </label>
        </div>
      </div>

      {hasChildren && isExpanded && (
        <div className="mt-1">
          {org.children!.map((child) => (
            <OrgPermissionNode
              key={child.id}
              org={child}
              level={level + 1}
              selectedOrgIds={selectedOrgIds}
              onToggle={onToggle}
            />
          ))}
        </div>
      )}
    </div>
  )
}
