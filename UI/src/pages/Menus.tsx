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
import { Tabs, TabsContent, TabsList, TabsTrigger } from "@/components/ui/tabs"
import {
  PlusIcon,
  SearchIcon,
  EditIcon,
  TrashIcon,
  ChevronRightIcon,
  ChevronDownIcon,
  RefreshCw,
  AlertTriangleIcon,
  GlobeIcon,
  BuildingIcon,
} from "lucide-react"
import { t } from "@/locales"
import {
  menuApi,
  platformMenuApi,
  type MenuItem,
  type PlatformMenuItem,
  ApiError,
} from "@/api"
import { useAuthStore } from "@/stores/authStore"
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

type Scope = "platform" | "tenant"

interface TreeMenuItem {
  id: number
  scope: Scope
  code: string
  name: string
  path: string
  icon?: string
  sort: number
  parent_id: number
  visible?: boolean
  children?: TreeMenuItem[]
}

// ----------------- 转换函数 -----------------

function fromTenantMenu(m: MenuItem): TreeMenuItem {
  return {
    id: m.id,
    scope: "tenant",
    code: m.code,
    name: m.name,
    path: m.path,
    icon: m.icon,
    sort: m.sort,
    parent_id: m.parent_id,
    visible: m.visible,
    children: m.children?.map(fromTenantMenu),
  }
}

function fromPlatformMenu(m: PlatformMenuItem): TreeMenuItem {
  return {
    id: m.id,
    scope: "platform",
    code: m.code,
    name: m.name,
    path: m.path,
    icon: m.icon,
    sort: m.sort,
    parent_id: m.parent_id,
    visible: m.visible,
    children: m.children?.map(fromPlatformMenu),
  }
}

// ----------------- 顶层组件 -----------------

export function MenusPage() {
  const isSuperAdmin = (
    useAuthStore((s) => s.user?.platform_roles) ?? []
  ).includes("super_admin")

  // 默认 tab：super_admin 先看平台；普通用户进租户
  const [activeTab, setActiveTab] = useState<Scope>(
    isSuperAdmin ? "platform" : "tenant"
  )

  // ---------- tenant tab 状态 ----------
  const [tenantMenus, setTenantMenus] = useState<TreeMenuItem[]>([])
  const [tenantError, setTenantError] = useState<string | null>(null)
  const [tenantLoading, setTenantLoading] = useState(false)

  // ---------- platform tab 状态 ----------
  const [platformMenus, setPlatformMenus] = useState<TreeMenuItem[]>([])
  const [platformError, setPlatformError] = useState<string | null>(null)
  const [platformLoading, setPlatformLoading] = useState(false)

  // ---------- 共享 UI 状态 ----------
  const [expandedIds, setExpandedIds] = useState<Set<number>>(new Set())
  const [searchTerm, setSearchTerm] = useState("")
  const [isSubmitting, setIsSubmitting] = useState(false)

  const [dialogOpen, setDialogOpen] = useState(false)
  const [dialogMode, setDialogMode] = useState<"add" | "edit">("add")
  const [currentMenu, setCurrentMenu] = useState<TreeMenuItem | null>(null)
  const [parentMenuOptions, setParentMenuOptions] = useState<
    { label: string; value: number }[]
  >([])

  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [menuToDelete, setMenuToDelete] = useState<TreeMenuItem | null>(null)

  // ---------- fetch ----------
  const fetchTenant = useCallback(async () => {
    setTenantLoading(true)
    setTenantError(null)
    try {
      const res = (await menuApi.tree()) as MenuItem[]
      const list = (res ?? []).map(fromTenantMenu)
      setTenantMenus(list)
    } catch (err) {
      const msg =
        err instanceof ApiError
          ? `${err.status} ${err.message}`
          : err instanceof Error
            ? err.message
            : "租户菜单加载失败"
      console.error("[Menus] tenant fetch failed:", err)
      setTenantMenus([])
      setTenantError(msg)
    } finally {
      setTenantLoading(false)
    }
  }, [])

  const fetchPlatform = useCallback(async () => {
    if (!isSuperAdmin) return
    setPlatformLoading(true)
    setPlatformError(null)
    try {
      const res = (await platformMenuApi.tree()) as PlatformMenuItem[]
      const list = (res ?? []).map(fromPlatformMenu)
      setPlatformMenus(list)
    } catch (err) {
      const msg =
        err instanceof ApiError
          ? `${err.status} ${err.message}`
          : err instanceof Error
            ? err.message
            : "平台菜单加载失败"
      console.error("[Menus] platform fetch failed:", err)
      setPlatformMenus([])
      setPlatformError(msg)
    } finally {
      setPlatformLoading(false)
    }
  }, [isSuperAdmin])

  useEffect(() => {
    fetchTenant()
  }, [fetchTenant])

  useEffect(() => {
    if (isSuperAdmin) fetchPlatform()
  }, [fetchPlatform, isSuperAdmin])

  const buildParentOptions = useCallback((menuList: TreeMenuItem[]) => {
    const opts: { label: string; value: number }[] = []
    const walk = (items: TreeMenuItem[], prefix: string) => {
      items.forEach((m) => {
        opts.push({ label: `${prefix}${m.name}`, value: m.id })
        if (m.children?.length) walk(m.children, prefix + "├── ")
      })
    }
    walk(menuList, "")
    setParentMenuOptions(opts)
  }, [])

  // 切换 tab 时构建 parent options（form 父菜单下拉）
  useEffect(() => {
    const currentList = activeTab === "platform" ? platformMenus : tenantMenus
    const currentLoading =
      activeTab === "platform" ? platformLoading : tenantLoading
    if (currentLoading) return
    buildParentOptions(currentList)
    setExpandedIds((prev) => {
      const next = new Set<number>()
      currentList.forEach((m) => {
        if (m.children && m.children.length > 0) next.add(m.id)
      })
      // 保留用户已展开的
      prev.forEach((id) => {
        if (
          currentList.some(
            (m) => m.id === id || m.children?.some((c) => c.id === id)
          )
        ) {
          next.add(id)
        }
      })
      return next
    })
  }, [
    activeTab,
    platformMenus,
    tenantMenus,
    platformLoading,
    tenantLoading,
    buildParentOptions,
  ])

  // 当前 tab 的活跃数据
  const currentMenus = activeTab === "platform" ? platformMenus : tenantMenus
  const currentLoading =
    activeTab === "platform" ? platformLoading : tenantLoading

  const handleRefresh = () => {
    if (activeTab === "platform") fetchPlatform()
    else fetchTenant()
  }

  // ---------- form schema ----------
  const menuFormSchema = useMemo<FormSchema>(
    () => ({
      items: [
        {
          field: "parent_id",
          label: t.pages.menus?.parentMenu || "父菜单",
          type: "select",
          placeholder: "请选择父菜单（根菜单请选择「无」）",
          options: [
            { label: "无（作为一级菜单）", value: 0 },
            ...parentMenuOptions,
          ],
        },
        {
          field: "name",
          label: t.pages.menus?.menuName || "菜单名称",
          type: "text",
          required: true,
          placeholder: "请输入菜单名称",
        },
        {
          field: "code",
          label: t.pages.menus?.code || "菜单代码",
          type: "text",
          required: true,
          placeholder: "请输入菜单代码，如 users",
        },
        {
          field: "path",
          label: t.pages.menus?.path || "路由路径",
          type: "text",
          required: true,
          placeholder: "请输入路由路径，如 /users",
        },
        {
          field: "icon",
          label: t.pages.menus?.icon || "图标",
          type: "icon",
          placeholder: "点击选择图标",
        },
        {
          field: "sort",
          label: t.pages.menus?.sortOrder || "排序",
          type: "number",
          defaultValue: 1,
        },
        {
          field: "visible",
          label: t.pages.menus?.visible || "是否显示",
          type: "radio",
          defaultValue: "true",
          options: [
            { label: "显示", value: "true" },
            { label: "隐藏", value: "false" },
          ],
        },
      ],
    }),
    [parentMenuOptions]
  )

  // ---------- handlers ----------
  const handleAdd = (parentId?: number) => {
    setDialogMode("add")
    setCurrentMenu({
      id: 0,
      scope: activeTab,
      parent_id: parentId ?? 0,
      name: "",
      code: "",
      path: "",
      sort: 1,
      visible: true,
    } as TreeMenuItem)
    setDialogOpen(true)
  }

  const handleEdit = (menu: TreeMenuItem) => {
    setDialogMode("edit")
    setCurrentMenu(menu)
    setDialogOpen(true)
  }

  const handleDeleteConfirm = (menu: TreeMenuItem) => {
    setMenuToDelete(menu)
    setDeleteDialogOpen(true)
  }

  const handleDelete = async () => {
    if (!menuToDelete) return
    setIsSubmitting(true)
    try {
      if (menuToDelete.scope === "platform") {
        await platformMenuApi.delete(menuToDelete.id)
      } else {
        await menuApi.delete(menuToDelete.id)
      }
      setDeleteDialogOpen(false)
      if (activeTab === "platform") await fetchPlatform()
      else await fetchTenant()
    } catch (err) {
      console.error("[Menus] delete failed:", err)
      alert("删除失败，请重试")
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleSubmit = async (values: Record<string, unknown>) => {
    setIsSubmitting(true)
    try {
      const parentId =
        typeof values.parent_id === "number"
          ? values.parent_id
          : parseInt(String(values.parent_id), 10) || 0

      const payload: Partial<TreeMenuItem> = {
        name: values.name as string,
        code: values.code as string,
        path: values.path as string,
        icon: (values.icon as string) || "",
        sort:
          typeof values.sort === "number"
            ? values.sort
            : parseInt(String(values.sort), 10) || 1,
        parent_id: parentId,
        visible: values.visible === "true" || values.visible === true,
      }
      if (dialogMode === "add") {
        if (activeTab === "platform") {
          await platformMenuApi.create(payload as Partial<PlatformMenuItem>)
        } else {
          await menuApi.create(payload as Partial<MenuItem>)
        }
      } else if (currentMenu) {
        if (currentMenu.scope === "platform") {
          await platformMenuApi.update(
            currentMenu.id,
            payload as Partial<PlatformMenuItem>
          )
        } else {
          await menuApi.update(currentMenu.id, payload as Partial<MenuItem>)
        }
      }
      setDialogOpen(false)
      if (activeTab === "platform") await fetchPlatform()
      else await fetchTenant()
    } catch (err) {
      console.error("[Menus] save failed:", err)
      alert("保存失败，请重试")
    } finally {
      setIsSubmitting(false)
    }
  }

  const getInitialValues = () => {
    if (currentMenu) {
      return {
        parent_id: currentMenu.parent_id || 0,
        name: currentMenu.name,
        code: currentMenu.code,
        path: currentMenu.path || "",
        icon: currentMenu.icon || "",
        sort: currentMenu.sort,
        visible: currentMenu.visible !== false,
      }
    }
    return { parent_id: 0, sort: 1, visible: "true" }
  }

  const toggleExpand = (id: number) => {
    setExpandedIds((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }

  const filterTree = (items: TreeMenuItem[], term: string): TreeMenuItem[] => {
    if (!term) return items
    return items.reduce((acc, item) => {
      const matches =
        item.name.toLowerCase().includes(term.toLowerCase()) ||
        (item.path || "").toLowerCase().includes(term.toLowerCase()) ||
        (item.code || "").toLowerCase().includes(term.toLowerCase())
      const filteredChildren = item.children
        ? filterTree(item.children, term)
        : []
      if (matches || filteredChildren.length > 0) {
        acc.push({
          ...item,
          children:
            filteredChildren.length > 0 ? filteredChildren : item.children,
        })
      }
      return acc
    }, [] as TreeMenuItem[])
  }

  const countMenus = (items: TreeMenuItem[]): number =>
    items.reduce((acc, item) => {
      return acc + 1 + (item.children ? countMenus(item.children) : 0)
    }, 0)

  const filteredTree = filterTree(currentMenus, searchTerm)
  const menuCount = countMenus(currentMenus)

  // ---------- 渲染 ----------
  const renderTable = (
    errorMsg: string | null,
    onRetry: () => void,
    loading: boolean
  ) => (
    <>
      {errorMsg && (
        <div className="mb-4 flex items-start gap-2 rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-sm text-destructive">
          <AlertTriangleIcon className="mt-0.5 h-4 w-4 shrink-0" />
          <div className="flex-1">
            <div className="font-medium">加载失败</div>
            <div className="text-xs opacity-80">{errorMsg}</div>
          </div>
          <Button size="sm" variant="outline" onClick={onRetry}>
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
                placeholder={t.pages.menus?.searchPlaceholder || "搜索菜单..."}
                className="pl-9"
                value={searchTerm}
                onChange={(e) => setSearchTerm(e.target.value)}
              />
            </div>
            <Badge variant="secondary">
              {t.pages.menus?.totalMenus || "共"} {menuCount}{" "}
              {t.pages.menus?.menus || "个菜单"}
            </Badge>
          </div>
        </CardHeader>
        <CardContent>
          <Table>
            <TableHeader>
              <TableRow>
                <TableHead className="w-[300px]">
                  {t.pages.menus?.menuName || "菜单名称"}
                </TableHead>
                <TableHead>{t.pages.menus?.code || "代码"}</TableHead>
                <TableHead>{t.pages.menus?.path || "路径"}</TableHead>
                <TableHead>{t.pages.menus?.icon || "图标"}</TableHead>
                <TableHead className="w-[80px]">
                  {t.pages.menus?.sortOrder || "排序"}
                </TableHead>
                <TableHead>{t.pages.menus?.visible || "显示"}</TableHead>
                <TableHead className="w-[150px] text-right">
                  {t.common.edit}
                </TableHead>
              </TableRow>
            </TableHeader>
            <TableBody>
              {filteredTree.map((item) => (
                <MenuTreeRow
                  key={`${item.scope}-${item.id}`}
                  item={item}
                  level={0}
                  expandedIds={expandedIds}
                  onToggle={toggleExpand}
                  onEdit={handleEdit}
                  onDelete={handleDeleteConfirm}
                  onAddChild={handleAdd}
                />
              ))}
              {filteredTree.length === 0 && !loading && (
                <TableRow>
                  <TableCell
                    colSpan={7}
                    className="py-8 text-center text-muted-foreground"
                  >
                    {t.common.noData}
                  </TableCell>
                </TableRow>
              )}
            </TableBody>
          </Table>
          {loading && (
            <div className="flex items-center justify-center py-8">
              <div className="text-sm text-muted-foreground">
                {t.common.loading}
              </div>
            </div>
          )}
        </CardContent>
      </Card>
    </>
  )

  return (
    <PageLayout>
      <div className="px-4 lg:px-6">
        <div className="mb-6 flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold">
              {t.pages.menus?.title || "菜单管理"}
            </h1>
            <p className="text-sm text-muted-foreground">
              {isSuperAdmin
                ? "管理平台共享菜单与本租户私有菜单"
                : "管理本租户私有菜单"}
            </p>
          </div>
          <div className="flex items-center gap-4">
            <Button
              variant="outline"
              onClick={handleRefresh}
              disabled={currentLoading}
            >
              <RefreshCw
                className={cn("mr-2 h-4 w-4", currentLoading && "animate-spin")}
              />
              {t.pages.menus?.refresh || "刷新列表"}
            </Button>
            <Button onClick={() => handleAdd()}>
              <PlusIcon className="mr-2 h-4 w-4" />
              {t.common.add}
            </Button>
          </div>
        </div>

        {isSuperAdmin ? (
          <Tabs
            value={activeTab}
            onValueChange={(v) => setActiveTab(v as Scope)}
            className="space-y-4"
          >
            <TabsList>
              <TabsTrigger value="platform" className="gap-2">
                <GlobeIcon className="h-4 w-4" />
                平台菜单
              </TabsTrigger>
              <TabsTrigger value="tenant" className="gap-2">
                <BuildingIcon className="h-4 w-4" />
                租户菜单
              </TabsTrigger>
            </TabsList>

            <TabsContent value="platform">
              {renderTable(platformError, fetchPlatform, platformLoading)}
            </TabsContent>
            <TabsContent value="tenant">
              {renderTable(tenantError, fetchTenant, tenantLoading)}
            </TabsContent>
          </Tabs>
        ) : (
          // 普通用户：直接渲染租户 tab，无平台入口
          renderTable(tenantError, fetchTenant, tenantLoading)
        )}
      </div>

      <FormDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        title={
          dialogMode === "add"
            ? `添加${activeTab === "platform" ? "平台" : "租户"}菜单`
            : `编辑${currentMenu?.scope === "platform" ? "平台" : "租户"}菜单`
        }
        schema={menuFormSchema}
        initialValues={getInitialValues()}
        onSubmit={handleSubmit}
        loading={isSubmitting}
      />

      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent className="sm:max-w-[400px]">
          <DialogHeader>
            <DialogTitle>删除菜单</DialogTitle>
            <DialogDescription>
              确定要删除菜单 "{menuToDelete?.name}" 吗？
              {menuToDelete?.children && menuToDelete.children.length > 0
                ? "该菜单下有子菜单，将一并删除。"
                : ""}
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

// ----------------- MenuTreeRow 组件（与原版一致，仅 item 类型用 TreeMenuItem） -----------------

interface MenuTreeRowProps {
  item: TreeMenuItem
  level: number
  expandedIds: Set<number>
  onToggle: (id: number) => void
  onEdit: (menu: TreeMenuItem) => void
  onDelete: (menu: TreeMenuItem) => void
  onAddChild: (parentId: number) => void
}

function MenuTreeRow({
  item,
  level,
  expandedIds,
  onToggle,
  onEdit,
  onDelete,
  onAddChild,
}: MenuTreeRowProps) {
  const hasChildren = item.children && item.children.length > 0
  const isExpanded = expandedIds.has(item.id)

  return (
    <>
      <TableRow>
        <TableCell>
          <div
            className="flex items-center"
            style={{ paddingLeft: `${level * 24}px` }}
          >
            {hasChildren ? (
              <button
                onClick={() => onToggle(item.id)}
                className="rounded p-1 hover:bg-accent"
              >
                {isExpanded ? (
                  <ChevronDownIcon className="h-4 w-4" />
                ) : (
                  <ChevronRightIcon className="h-4 w-4" />
                )}
              </button>
            ) : (
              <span className="w-6" />
            )}
            <span className="ml-1">{item.name}</span>
            {item.scope === "platform" && (
              <Badge variant="outline" className="ml-2 text-xs">
                <GlobeIcon className="mr-1 h-3 w-3" /> 平台
              </Badge>
            )}
          </div>
        </TableCell>
        <TableCell className="font-mono text-sm">{item.code}</TableCell>
        <TableCell className="text-sm">{item.path || "-"}</TableCell>
        <TableCell>{item.icon || "-"}</TableCell>
        <TableCell>{item.sort}</TableCell>
        <TableCell>
          <Badge variant={item.visible !== false ? "default" : "secondary"}>
            {item.visible !== false ? "显示" : "隐藏"}
          </Badge>
        </TableCell>
        <TableCell>
          <div className="flex items-center justify-end gap-1">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => onAddChild(item.id)}
              title="添加子菜单"
            >
              <PlusIcon className="h-4 w-4" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              onClick={() => onEdit(item)}
            >
              <EditIcon className="h-4 w-4" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              onClick={() => onDelete(item)}
            >
              <TrashIcon className="h-4 w-4 text-destructive" />
            </Button>
          </div>
        </TableCell>
      </TableRow>
      {hasChildren &&
        isExpanded &&
        item.children?.map((child) => (
          <MenuTreeRow
            key={child.id}
            item={child as TreeMenuItem}
            level={level + 1}
            expandedIds={expandedIds}
            onToggle={onToggle}
            onEdit={onEdit}
            onDelete={onDelete}
            onAddChild={onAddChild}
          />
        ))}
    </>
  )
}
