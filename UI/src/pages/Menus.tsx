import { useEffect, useState, useCallback, useMemo } from "react"
import { PageLayout } from "@/components/page-layout"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { PlusIcon, SearchIcon, EditIcon, TrashIcon, ChevronRightIcon, ChevronDownIcon, MenuIcon, RefreshCw } from "lucide-react"
import { useTranslation } from "@/locales"
import { menuApi, type MenuItem } from "@/api"
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

interface TreeMenuItem extends MenuItem {
  children?: TreeMenuItem[]
}

const mockMenuTree: TreeMenuItem[] = [
  {
    id: 1, code: "dashboard", name: "仪表盘", path: "/dashboard", icon: "LayoutDashboardIcon", sort: 1, parent_id: 0, children: [],
  },
  {
    id: 2, code: "analytics", name: "数据分析", path: "/analytics", icon: "ChartBarIcon", sort: 2, parent_id: 0, children: [],
  },
  {
    id: 3, code: "projects", name: "项目管理", path: "/projects", icon: "FolderIcon", sort: 3, parent_id: 0,
    children: [
      { id: 31, code: "project_list", name: "项目列表", path: "/projects/list", icon: "ListIcon", sort: 1, parent_id: 3 },
      { id: 32, code: "project_create", name: "新建项目", path: "/projects/create", icon: "PlusIcon", sort: 2, parent_id: 3 },
    ],
  },
  {
    id: 4, code: "team", name: "团队管理", path: "/team", icon: "UsersIcon", sort: 4, parent_id: 0,
    children: [
      { id: 41, code: "team_members", name: "成员管理", path: "/team/members", icon: "UsersIcon", sort: 1, parent_id: 4 },
    ],
  },
  {
    id: 5, code: "system", name: "系统管理", path: "/system", icon: "SettingsIcon", sort: 5, parent_id: 0,
    children: [
      { id: 51, code: "users", name: "用户管理", path: "/users", icon: "UsersIcon", sort: 1, parent_id: 5 },
      { id: 52, code: "roles", name: "角色管理", path: "/roles", icon: "ShieldIcon", sort: 2, parent_id: 5 },
      { id: 53, code: "menus", name: "菜单管理", path: "/menus", icon: "MenuIcon", sort: 3, parent_id: 5 },
    ],
  },
]

export function MenusPage() {
  const t = useTranslation()
  const [menus, setMenus] = useState<TreeMenuItem[]>([])
  const [expandedIds, setExpandedIds] = useState<Set<number>>(new Set())
  const [searchTerm, setSearchTerm] = useState("")
  const [isLoading, setIsLoading] = useState(true)
  const [isSubmitting, setIsSubmitting] = useState(false)

  const [dialogOpen, setDialogOpen] = useState(false)
  const [dialogMode, setDialogMode] = useState<"add" | "edit">("add")
  const [currentMenu, setCurrentMenu] = useState<MenuItem | null>(null)
  const [parentMenuOptions, setParentMenuOptions] = useState<{ label: string; value: number }[]>([])

  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [menuToDelete, setMenuToDelete] = useState<MenuItem | null>(null)

  const fetchMenus = useCallback(async () => {
    setIsLoading(true)
    try {
      const response = await menuApi.tree()
      const menuTree = response as MenuItem[]
      setMenus(menuTree as TreeMenuItem[])
      buildParentOptions(menuTree)
      const allIds = new Set<number>()
      menuTree.forEach((menu) => {
        if (menu.children && menu.children.length > 0) {
          allIds.add(menu.id)
        }
      })
      setExpandedIds(allIds)
    } catch {
      setMenus(mockMenuTree)
      buildParentOptions(mockMenuTree)
      const mockIds = new Set<number>()
      mockMenuTree.forEach((menu) => {
        if (menu.children && menu.children.length > 0) {
          mockIds.add(menu.id)
        }
      })
      setExpandedIds(mockIds)
    } finally {
      setIsLoading(false)
    }
  }, [])

  const buildParentOptions = useCallback((menuList: MenuItem[], prefix = "") => {
    const options: { label: string; value: number }[] = []
    menuList.forEach((menu) => {
      options.push({ label: `${prefix}${menu.name}`, value: menu.id })
      if (menu.children && menu.children.length > 0) {
        options.push(...buildParentOptionsHelper(menu.children, prefix + "├── "))
      }
    })
    setParentMenuOptions(options)
    return options
  }, [])

  const buildParentOptionsHelper = (menuList: MenuItem[], prefix: string): { label: string; value: number }[] => {
    const options: { label: string; value: number }[] = []
    menuList.forEach((menu) => {
      options.push({ label: `${prefix}${menu.name}`, value: menu.id })
      if (menu.children && menu.children.length > 0) {
        options.push(...buildParentOptionsHelper(menu.children, prefix + "│   "))
      }
    })
    return options
  }

  useEffect(() => {
    fetchMenus()
  }, [fetchMenus])

  const menuFormSchema = useMemo<FormSchema>(() => ({
    items: [
      {
        field: "parent_id",
        label: t.pages.menus?.parentMenu || "父菜单",
        type: "select",
        placeholder: "请选择父菜单 (根菜单请选择「无」)",
        options: [
          { label: "无 (作为一级菜单)", value: 0 },
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
        defaultValue: true,
        options: [
          { label: "显示", value: true },
          { label: "隐藏", value: false },
        ],
      },
    ],
  }), [t, parentMenuOptions])

  const handleAdd = (parentId?: number) => {
    setDialogMode("add")
    setCurrentMenu({ id: 0, parent_id: parentId || 0, name: "", code: "", path: "", sort: 1, visible: true } as MenuItem)
    setDialogOpen(true)
  }

  const handleEdit = (menu: MenuItem) => {
    setDialogMode("edit")
    setCurrentMenu(menu)
    setDialogOpen(true)
  }

  const handleDeleteConfirm = (menu: MenuItem) => {
    setMenuToDelete(menu)
    setDeleteDialogOpen(true)
  }

  const handleDelete = async () => {
    if (!menuToDelete) return
    setIsSubmitting(true)
    try {
      await menuApi.delete(menuToDelete.id)
      await fetchMenus()
      setDeleteDialogOpen(false)
    } catch (error) {
      console.error("Delete menu failed:", error)
      alert("删除失败，请重试")
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleSubmit = async (values: Record<string, unknown>) => {
    setIsSubmitting(true)
    try {
      const parentId = typeof values.parent_id === 'number' 
        ? values.parent_id 
        : parseInt(String(values.parent_id), 10) || 0
      
      const menuData: Partial<MenuItem> = {
        name: values.name as string,
        code: values.code as string,
        path: values.path as string,
        icon: (values.icon as string) || "",
        sort: typeof values.sort === 'number' ? values.sort : parseInt(String(values.sort), 10) || 1,
        parent_id: parentId,
        visible: Boolean(values.visible),
      }
      if (dialogMode === "add") {
        await menuApi.create(menuData)
      } else if (currentMenu) {
        await menuApi.update(currentMenu.id, menuData)
      }
      await fetchMenus()
      setDialogOpen(false)
    } catch (error) {
      console.error("Save menu failed:", error)
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
    return { parent_id: 0, sort: 1, visible: true }
  }

  const toggleExpand = (id: number) => {
    setExpandedIds((prev) => {
      const next = new Set(prev)
      if (next.has(id)) {
        next.delete(id)
      } else {
        next.add(id)
      }
      return next
    })
  }

  const filterTree = (items: TreeMenuItem[], term: string): TreeMenuItem[] => {
    if (!term) return items
    return items.reduce((acc, item) => {
      const matches = item.name.toLowerCase().includes(term.toLowerCase()) ||
        (item.path || "").toLowerCase().includes(term.toLowerCase()) ||
        (item.code || "").toLowerCase().includes(term.toLowerCase())
      const filteredChildren = item.children ? filterTree(item.children, term) : []
      if (matches || filteredChildren.length > 0) {
        acc.push({
          ...item,
          children: filteredChildren.length > 0 ? filteredChildren : item.children,
        })
      }
      return acc
    }, [] as TreeMenuItem[])
  }

  const filteredTree = filterTree(menus, searchTerm)

  const countMenus = (items: TreeMenuItem[]): number => {
    return items.reduce((acc, item) => {
      return acc + 1 + (item.children ? countMenus(item.children) : 0)
    }, 0)
  }

  const menuCount = countMenus(menus)

  return (
    <PageLayout>
      <div className="px-4 lg:px-6">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-2xl font-bold">{t.pages.menus?.title || "菜单管理"}</h1>
            <p className="text-sm text-muted-foreground">{t.pages.menus?.subtitle || "管理系统菜单结构"}</p>
          </div>
          <div className="flex items-center gap-2">
            <Button variant="outline" onClick={fetchMenus} disabled={isLoading}>
              <RefreshCw className={`mr-2 h-4 w-4 ${isLoading ? "animate-spin" : ""}`} />
              {t.pages.menus?.refresh || "刷新列表"}
            </Button>
            <Button onClick={() => handleAdd()}>
              <PlusIcon className="mr-2 h-4 w-4" />
              {t.common.add}
            </Button>
          </div>
        </div>

        <Card>
          <CardHeader>
            <div className="flex items-center gap-4">
              <div className="relative flex-1 max-w-sm">
                <SearchIcon className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  placeholder={t.pages.menus?.searchPlaceholder || "搜索菜单..."}
                  className="pl-9"
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                />
              </div>
              <Badge variant="secondary">
                {t.pages.menus?.totalMenus || "共"} {menuCount} {t.pages.menus?.menus || "个菜单"}
              </Badge>
            </div>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-[300px]">{t.pages.menus?.menuName || "菜单名称"}</TableHead>
                  <TableHead>{t.pages.menus?.code || "代码"}</TableHead>
                  <TableHead>{t.pages.menus?.path || "路径"}</TableHead>
                  <TableHead>{t.pages.menus?.icon || "图标"}</TableHead>
                  <TableHead className="w-[80px]">{t.pages.menus?.sortOrder || "排序"}</TableHead>
                  <TableHead>{t.pages.menus?.visible || "显示"}</TableHead>
                  <TableHead className="w-[150px] text-right">{t.common.edit}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredTree.map((item) => (
                  <MenuTreeRow
                    key={item.id}
                    item={item}
                    level={0}
                    expandedIds={expandedIds}
                    onToggle={toggleExpand}
                    onEdit={handleEdit}
                    onDelete={handleDeleteConfirm}
                    onAddChild={handleAdd}
                  />
                ))}
                {filteredTree.length === 0 && !isLoading && (
                  <TableRow>
                    <TableCell colSpan={7} className="text-center py-8 text-muted-foreground">
                      {t.common.noData}
                    </TableCell>
                  </TableRow>
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
        title={dialogMode === "add" ? (t.pages.menus?.addMenu || "添加菜单") : (t.pages.menus?.editMenu || "编辑菜单")}
        schema={menuFormSchema}
        initialValues={getInitialValues()}
        onSubmit={handleSubmit}
        loading={isSubmitting}
      />

      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent className="sm:max-w-[400px]">
          <DialogHeader>
            <DialogTitle>{t.pages.menus?.deleteMenu || "删除菜单"}</DialogTitle>
            <DialogDescription>
              确定要删除菜单 "{menuToDelete?.name}" 吗？{menuToDelete?.children && menuToDelete.children.length > 0 ? "该菜单下有子菜单，将一并删除。" : ""}
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
    </PageLayout>
  )
}

interface MenuTreeRowProps {
  item: TreeMenuItem
  level: number
  expandedIds: Set<number>
  onToggle: (id: number) => void
  onEdit: (menu: MenuItem) => void
  onDelete: (menu: MenuItem) => void
  onAddChild: (parentId: number) => void
}

function MenuTreeRow({ item, level, expandedIds, onToggle, onEdit, onDelete, onAddChild }: MenuTreeRowProps) {
  const hasChildren = item.children && item.children.length > 0
  const isExpanded = expandedIds.has(item.id)

  return (
    <>
      <TableRow>
        <TableCell>
          <div className="flex items-center" style={{ paddingLeft: `${level * 24}px` }}>
            {hasChildren ? (
              <button
                onClick={() => onToggle(item.id)}
                className="p-1 hover:bg-accent rounded"
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
            <Button variant="ghost" size="sm" onClick={() => onAddChild(item.id)} title="添加子菜单">
              <PlusIcon className="h-4 w-4" />
            </Button>
            <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => onEdit(item)}>
              <EditIcon className="h-4 w-4" />
            </Button>
            <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => onDelete(item)}>
              <TrashIcon className="h-4 w-4 text-destructive" />
            </Button>
          </div>
        </TableCell>
      </TableRow>
      {hasChildren && isExpanded && item.children?.map((child) => (
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