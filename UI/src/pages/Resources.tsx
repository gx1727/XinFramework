import { useEffect, useState, useCallback } from "react"
import { PageLayout } from "@/components/page-layout"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { PlusIcon, SearchIcon, EditIcon, TrashIcon, KeyIcon, RefreshCw } from "lucide-react"
import { useTranslation } from "@/locales"
import { resourceApi, menuApi, type ResourceItem, type MenuItem } from "@/api"
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

const mockResources: ResourceItem[] = [
  { id: 1, menu_id: 1, code: "user:view", name: "查看用户", action: "view", sort: 1, status: 1 },
  { id: 2, menu_id: 1, code: "user:create", name: "创建用户", action: "create", sort: 2, status: 1 },
  { id: 3, menu_id: 1, code: "user:edit", name: "编辑用户", action: "edit", sort: 3, status: 1 },
  { id: 4, menu_id: 1, code: "user:delete", name: "删除用户", action: "delete", sort: 4, status: 1 },
  { id: 5, menu_id: 2, code: "role:view", name: "查看角色", action: "view", sort: 1, status: 1 },
  { id: 6, menu_id: 2, code: "role:create", name: "创建角色", action: "create", sort: 2, status: 1 },
  { id: 7, menu_id: 2, code: "role:edit", name: "编辑角色", action: "edit", sort: 3, status: 1 },
  { id: 8, menu_id: 0, code: "common:upload", name: "上传文件", action: "create", sort: 1, status: 1 },
]

export function ResourcesPage() {
  const t = useTranslation()
  const [resources, setResources] = useState<ResourceItem[]>([])
  const [menus, setMenus] = useState<MenuItem[]>([])
  const [total, setTotal] = useState(0)
  const [isLoading, setIsLoading] = useState(true)
  const [searchTerm, setSearchTerm] = useState("")
  const [isSubmitting, setIsSubmitting] = useState(false)

  const [dialogOpen, setDialogOpen] = useState(false)
  const [dialogMode, setDialogMode] = useState<"add" | "edit">("add")
  const [currentResource, setCurrentResource] = useState<ResourceItem | null>(null)

  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [resourceToDelete, setResourceToDelete] = useState<ResourceItem | null>(null)

  const fetchResources = useCallback(async () => {
    setIsLoading(true)
    try {
      const response = await resourceApi.list({ page: 1, size: 500 })
      setResources(response.list)
      setTotal(response.total)
    } catch {
      setResources(mockResources)
      setTotal(mockResources.length)
    } finally {
      setIsLoading(false)
    }
  }, [])

  const fetchMenus = useCallback(async () => {
    try {
      const response = await menuApi.tree()
      setMenus(response)
    } catch {
      setMenus([])
    }
  }, [])

  useEffect(() => {
    fetchResources()
    fetchMenus()
  }, [fetchResources, fetchMenus])

  const buildMenuOptions = (menuList: MenuItem[], prefix = "") => {
    const options: { label: string; value: number }[] = []
    menuList.forEach(menu => {
      options.push({ label: `${prefix}${menu.name}`, value: menu.id })
      if (menu.children && menu.children.length > 0) {
        options.push(...buildMenuOptions(menu.children, prefix + "├── "))
      }
    })
    return options
  }

  const resourceFormSchema: FormSchema = {
    items: [
      {
        field: "menu_id",
        label: t.pages.resources?.menu || "所属菜单",
        type: "select",
        required: false,
        placeholder: "请选择所属菜单 (为空视为公共资源)",
        options: [
          { label: "无 (公共资源)", value: 0 },
          ...buildMenuOptions(menus),
        ],
      },
      {
        field: "name",
        label: t.pages.resources?.name || "资源名称",
        type: "text",
        required: true,
        placeholder: "请输入资源名称，如 查看用户",
      },
      {
        field: "code",
        label: t.pages.resources?.code || "资源代码",
        type: "text",
        required: true,
        placeholder: "请输入资源代码，如 user:view",
      },
      {
        field: "action",
        label: t.pages.resources?.action || "操作类型",
        type: "select",
        required: true,
        placeholder: "请选择操作类型",
        options: [
          { label: "查看 (view)", value: "view" },
          { label: "创建 (create)", value: "create" },
          { label: "编辑 (edit)", value: "edit" },
          { label: "删除 (delete)", value: "delete" },
          { label: "导出 (export)", value: "export" },
          { label: "导入 (import)", value: "import" },
        ],
      },
      {
        field: "description",
        label: t.pages.resources?.description || "描述",
        type: "textarea",
        placeholder: "请输入资源描述",
      },
      {
        field: "sort",
        label: t.pages.resources?.sortOrder || "排序",
        type: "number",
        defaultValue: 1,
      },
      {
        field: "status",
        label: t.common?.status || "状态",
        type: "radio",
        options: [
          { label: t.common?.enable || "启用", value: 1 },
          { label: t.common?.disable || "禁用", value: 0 },
        ],
        required: true,
        defaultValue: 1,
      },
    ],
  }

  const handleAdd = () => {
    setDialogMode("add")
    setCurrentResource(null)
    setDialogOpen(true)
  }

  const handleEdit = (resource: ResourceItem) => {
    setDialogMode("edit")
    setCurrentResource(resource)
    setDialogOpen(true)
  }

  const handleDeleteConfirm = (resource: ResourceItem) => {
    setResourceToDelete(resource)
    setDeleteDialogOpen(true)
  }

  const handleDelete = async () => {
    if (!resourceToDelete) return
    setIsSubmitting(true)
    try {
      await resourceApi.delete(resourceToDelete.id)
      await fetchResources()
      setDeleteDialogOpen(false)
    } catch (error) {
      console.error("Delete resource failed:", error)
      alert("删除失败，请重试")
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleSubmit = async (values: Record<string, unknown>) => {
    setIsSubmitting(true)
    try {
      const resourceData: Partial<ResourceItem> = {
        menu_id: Number(values.menu_id) || 0,
        name: values.name as string,
        code: values.code as string,
        action: values.action as string,
        description: values.description as string,
        sort: values.sort !== undefined ? Number(values.sort) : 1,
        status: values.status !== undefined ? Number(values.status) : 1,
      }
      if (dialogMode === "add") {
        await resourceApi.create(resourceData)
      } else if (currentResource) {
        await resourceApi.update(currentResource.id, resourceData)
      }
      await fetchResources()
      setDialogOpen(false)
    } catch (error) {
      console.error("Save resource failed:", error)
      alert("保存失败，请重试")
    } finally {
      setIsSubmitting(false)
    }
  }

  const getInitialValues = () => {
    if (currentResource) {
      return {
        menu_id: currentResource.menu_id || 0,
        name: currentResource.name,
        code: currentResource.code,
        action: currentResource.action,
        description: currentResource.description || "",
        sort: currentResource.sort,
        status: currentResource.status,
      }
    }
    return { status: 1, sort: 1 }
  }

  const getMenuName = (menuId: number) => {
    if (!menuId || menuId === 0) return "公共资源"
    const menu = menus.find(m => m.id === menuId)
    return menu?.name || `菜单 ${menuId}`
  }

  const getActionBadge = (action: string) => {
    const actionMap: Record<string, { label: string; variant: "default" | "secondary" | "outline" }> = {
      view: { label: "查看", variant: "default" },
      create: { label: "创建", variant: "default" },
      edit: { label: "编辑", variant: "outline" },
      delete: { label: "删除", variant: "secondary" },
      export: { label: "导出", variant: "outline" },
      import: { label: "导入", variant: "outline" },
    }
    const config = actionMap[action] || { label: action, variant: "secondary" }
    return <Badge variant={config.variant}>{config.label}</Badge>
  }

  const filteredResources = searchTerm
    ? resources.filter(r =>
        r.name.toLowerCase().includes(searchTerm.toLowerCase()) ||
        r.code.toLowerCase().includes(searchTerm.toLowerCase())
      )
    : resources

  return (
    <PageLayout>
      <div className="px-4 lg:px-6">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-2xl font-bold">{t.pages.resources?.title || "资源管理"}</h1>
            <p className="text-sm text-muted-foreground">{t.pages.resources?.subtitle || "管理系统按钮和操作权限"}</p>
          </div>
          <div className="flex items-center gap-2">
            <Button variant="outline" onClick={fetchResources} disabled={isLoading}>
              <RefreshCw className={`mr-2 h-4 w-4 ${isLoading ? "animate-spin" : ""}`} />
              {t.pages.resources?.refresh || "刷新列表"}
            </Button>
            <Button onClick={handleAdd}>
              <PlusIcon className="mr-2 h-4 w-4" />
              {t.common.add}
            </Button>
          </div>
        </div>

        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3 mb-6">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">资源总数</CardTitle>
              <KeyIcon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{total}</div>
              <p className="text-xs text-muted-foreground">按钮/操作权限数量</p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">启用资源</CardTitle>
              <KeyIcon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{resources.filter(r => r.status === 1).length}</div>
              <p className="text-xs text-muted-foreground">当前启用的资源</p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">关联菜单</CardTitle>
              <KeyIcon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{new Set(resources.map(r => r.menu_id).filter(id => id && id !== 0)).size}</div>
              <p className="text-xs text-muted-foreground">关联的菜单数量</p>
            </CardContent>
          </Card>
        </div>

        <Card>
          <CardHeader>
            <div className="flex items-center gap-4">
              <div className="relative flex-1 max-w-sm">
                <SearchIcon className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  placeholder={t.pages.resources?.searchPlaceholder || "搜索资源..."}
                  className="pl-9"
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                />
              </div>
              <Badge variant="secondary">共 {filteredResources.length} 个资源</Badge>
            </div>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead>ID</TableHead>
                  <TableHead>{t.pages.resources?.name || "资源名称"}</TableHead>
                  <TableHead>{t.pages.resources?.code || "资源代码"}</TableHead>
                  <TableHead>{t.pages.resources?.menu || "所属菜单"}</TableHead>
                  <TableHead>{t.pages.resources?.action || "操作"}</TableHead>
                  <TableHead>{t.pages.resources?.description || "描述"}</TableHead>
                  <TableHead className="w-[80px]">{t.pages.resources?.sortOrder || "排序"}</TableHead>
                  <TableHead>{t.pages.resources?.status || "状态"}</TableHead>
                  <TableHead className="w-[120px] text-right">{t.common.edit}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredResources.map((resource) => (
                  <TableRow key={resource.id}>
                    <TableCell className="font-medium">{resource.id}</TableCell>
                    <TableCell>{resource.name}</TableCell>
                    <TableCell className="font-mono text-sm">{resource.code}</TableCell>
                    <TableCell>
                      <Badge variant="outline">{getMenuName(resource.menu_id)}</Badge>
                    </TableCell>
                    <TableCell>{getActionBadge(resource.action)}</TableCell>
                    <TableCell className="text-sm text-muted-foreground">{resource.description || "-"}</TableCell>
                    <TableCell>{resource.sort}</TableCell>
                    <TableCell>
                      <Badge variant={resource.status === 1 ? "default" : "secondary"}>
                        {resource.status === 1 ? "启用" : "停用"}
                      </Badge>
                    </TableCell>
                    <TableCell className="text-right">
                      <div className="flex items-center justify-end gap-1">
                        <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => handleEdit(resource)}>
                          <EditIcon className="h-4 w-4" />
                        </Button>
                        <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => handleDeleteConfirm(resource)}>
                          <TrashIcon className="h-4 w-4 text-destructive" />
                        </Button>
                      </div>
                    </TableCell>
                  </TableRow>
                ))}
                {filteredResources.length === 0 && !isLoading && (
                  <TableRow>
                    <TableCell colSpan={9} className="text-center py-8 text-muted-foreground">
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
        title={dialogMode === "add" ? (t.pages.resources?.addResource || "添加资源") : (t.pages.resources?.editResource || "编辑资源")}
        schema={resourceFormSchema}
        initialValues={getInitialValues()}
        onSubmit={handleSubmit}
        loading={isSubmitting}
      />

      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent className="sm:max-w-[400px]">
          <DialogHeader>
            <DialogTitle>{t.pages.resources?.deleteResource || "删除资源"}</DialogTitle>
            <DialogDescription>
              确定要删除资源 "{resourceToDelete?.name}" 吗？删除后将影响已分配该权限的角色。
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