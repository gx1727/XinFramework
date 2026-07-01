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
  ChevronRightIcon,
  ChevronDownIcon,
  RefreshCw,
  AlertTriangleIcon,
} from "lucide-react"
import { t } from "@/locales"
import { sysMenuApi, type SysMenuItem, ApiError } from "@/api"
import { useAuthStore, hasPermission } from "@/stores/authStore"
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

/**
 * Sys 菜单管理（/sys/menus）。
 *
 * 0024+ 拆分原则：sys_menus 与 tenant_menus 是两张独立的表、两个独立路由、
 * 两套独立的 CRUD 入口。本页只关心 sys_menus，不引入租户菜单 tab 切换。
 * 租户菜单请去 /app/menus。
 *
 * 服务端 service.Tree 已按当前用户的 sys_role_menus 绑定收敛可见性，
 * 所以非 super_admin 的 sys 用户只能看到被分配的那部分菜单。
 */
interface TreeMenuItem {
  id: number
  code: string
  name: string
  path: string
  icon?: string
  sort: number
  parent_id: number
  visible?: boolean
  children?: TreeMenuItem[]
}

function fromSysMenu(m: SysMenuItem): TreeMenuItem {
  return {
    id: m.id,
    code: m.code,
    name: m.name,
    path: m.path,
    icon: m.icon,
    sort: m.sort,
    parent_id: m.parent_id,
    visible: m.visible,
    children: m.children?.map(fromSysMenu),
  }
}

export function SysMenusPage() {
  // 权限码：与后端 permission.P(Res, Act) 一一对应。
  // super_admin 因持有 "*:*" 通配天然通过，不需要在组件里写 isSuperAdmin 分支。
  const user = useAuthStore((s) => s.user)
  const canCreateMenu = hasPermission(user, "menu:create")
  const canUpdateMenu = hasPermission(user, "menu:update")
  const canDeleteMenu = hasPermission(user, "menu:delete")

  const [menus, setMenus] = useState<TreeMenuItem[]>([])
  const [error, setError] = useState<string | null>(null)
  const [loading, setLoading] = useState(false)
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
  const fetchMenus = useCallback(async () => {
    setLoading(true)
    setError(null)
    try {
      const res = (await sysMenuApi.tree()) as SysMenuItem[]
      const list = (res ?? []).map(fromSysMenu)
      setMenus(list)
    } catch (err) {
      const msg =
        err instanceof ApiError
          ? `${err.status} ${err.message}`
          : err instanceof Error
            ? err.message
            : "Sys 菜单加载失败"
      console.error("[SysMenus] fetch failed:", err)
      setMenus([])
      setError(msg)
    } finally {
      setLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchMenus()
  }, [fetchMenus])

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

  useEffect(() => {
    if (loading) return
    buildParentOptions(menus)
    setExpandedIds((prev) => {
      const next = new Set<number>()
      menus.forEach((m) => {
        if (m.children && m.children.length > 0) next.add(m.id)
      })
      prev.forEach((id) => {
        if (
          menus.some(
            (m) => m.id === id || m.children?.some((c) => c.id === id),
          )
        ) {
          next.add(id)
        }
      })
      return next
    })
  }, [menus, loading, buildParentOptions])

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
          placeholder: "请输入菜单代码，如 sys-users",
        },
        {
          field: "path",
          label: t.pages.menus?.path || "路由路径",
          type: "text",
          required: true,
          placeholder: "请输入路由路径，如 /sys/users",
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
      await sysMenuApi.delete(menuToDelete.id)
      setDeleteDialogOpen(false)
      await fetchMenus()
    } catch (err) {
      console.error("[SysMenus] delete failed:", err)
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
        await sysMenuApi.create(payload as Partial<SysMenuItem>)
      } else if (currentMenu) {
        await sysMenuApi.update(
          currentMenu.id,
          payload as Partial<SysMenuItem>,
        )
      }
      setDialogOpen(false)
      await fetchMenus()
    } catch (err) {
      console.error("[SysMenus] save failed:", err)
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

  const filteredTree = filterTree(menus, searchTerm)
  const menuCount = countMenus(menus)

  return (
    <PageLayout>
      <div className="px-4 lg:px-6">
        <div className="mb-6 flex items-center justify-between">
          <div>
            <h1 className="text-2xl font-bold">Sys 菜单管理</h1>
            <p className="text-sm text-muted-foreground">
              管理 Sys 域菜单树（sys_menus）
            </p>
          </div>
          <div className="flex items-center gap-4">
            <Button
              variant="outline"
              onClick={fetchMenus}
              disabled={loading}
            >
              <RefreshCw
                className={cn("mr-2 h-4 w-4", loading && "animate-spin")}
              />
              {t.pages.menus?.refresh || "刷新列表"}
            </Button>
            <Button
              onClick={() => handleAdd()}
              disabled={!canCreateMenu}
              title={!canCreateMenu ? "当前账号无菜单新建权限" : undefined}
            >
              <PlusIcon className="mr-2 h-4 w-4" />
              {t.common.add}
            </Button>
          </div>
        </div>

        {error && (
          <div className="mb-4 flex items-start gap-2 rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-sm text-destructive">
            <AlertTriangleIcon className="mt-0.5 h-4 w-4 shrink-0" />
            <div className="flex-1">
              <div className="font-medium">加载失败</div>
              <div className="text-xs opacity-80">{error}</div>
            </div>
            <Button size="sm" variant="outline" onClick={fetchMenus}>
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
                  <SysMenuTreeRow
                    key={item.id}
                    item={item}
                    level={0}
                    expandedIds={expandedIds}
                    onToggle={toggleExpand}
                    onEdit={handleEdit}
                    onDelete={handleDeleteConfirm}
                    onAddChild={handleAdd}
                    canCreate={canCreateMenu}
                    canUpdate={canUpdateMenu}
                    canDelete={canDeleteMenu}
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
      </div>

      <FormDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        title={
          dialogMode === "add" ? "添加 Sys 菜单" : "编辑 Sys 菜单"
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

// ----------------- SysMenuTreeRow -----------------

interface SysMenuTreeRowProps {
  item: TreeMenuItem
  level: number
  expandedIds: Set<number>
  onToggle: (id: number) => void
  onEdit: (menu: TreeMenuItem) => void
  onDelete: (menu: TreeMenuItem) => void
  onAddChild: (parentId: number) => void
  canCreate: boolean
  canUpdate: boolean
  canDelete: boolean
}

function SysMenuTreeRow({
  item,
  level,
  expandedIds,
  onToggle,
  onEdit,
  onDelete,
  onAddChild,
  canCreate,
  canUpdate,
  canDelete,
}: SysMenuTreeRowProps) {
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
              disabled={!canCreate}
            >
              <PlusIcon className="h-4 w-4" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              onClick={() => onEdit(item)}
              disabled={!canUpdate}
            >
              <EditIcon className="h-4 w-4" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              onClick={() => onDelete(item)}
              disabled={!canDelete}
            >
              <TrashIcon className="h-4 w-4 text-destructive" />
            </Button>
          </div>
        </TableCell>
      </TableRow>
      {hasChildren &&
        isExpanded &&
        item.children?.map((child) => (
          <SysMenuTreeRow
            key={child.id}
            item={child as TreeMenuItem}
            level={level + 1}
            expandedIds={expandedIds}
            onToggle={onToggle}
            onEdit={onEdit}
            onDelete={onDelete}
            onAddChild={onAddChild}
            canCreate={canCreate}
            canUpdate={canUpdate}
            canDelete={canDelete}
          />
        ))}
    </>
  )
}
