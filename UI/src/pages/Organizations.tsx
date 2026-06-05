import { useEffect, useState, useCallback, useMemo } from "react"
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
  PlusIcon,
  SearchIcon,
  EditIcon,
  TrashIcon,
  ChevronRightIcon,
  ChevronDownIcon,
  BuildingIcon,
  Building2Icon,
  UsersIcon,
  NetworkIcon,
  RefreshCw,
  GitBranchIcon,
  HashIcon,
  PowerOffIcon,
  PowerIcon,
} from "lucide-react"
import { useTranslation } from "@/locales"
import { organizationApi, type OrganizationItem } from "@/api"
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
import { ApiError } from "@/api"

type TreeOrgItem = OrganizationItem & {
  children?: TreeOrgItem[]
}

const mockOrgTree: TreeOrgItem[] = [
  {
    id: 1,
    tenant_id: 1,
    code: "HQ",
    name: "总部",
    type: "company",
    description: "集团总部",
    admin_code: "ceo",
    parent_id: 0,
    ancestors: "0",
    sort: 0,
    status: 1,
    created_at: "2026-04-26 10:00:00",
    updated_at: "2026-04-26 10:00:00",
    children: [
      {
        id: 2,
        tenant_id: 1,
        code: "TECH",
        name: "技术中心",
        type: "department",
        description: "研发中心",
        admin_code: "cto",
        parent_id: 1,
        ancestors: "0.1",
        sort: 1,
        status: 1,
        created_at: "2026-04-26 10:05:00",
        updated_at: "2026-04-26 10:05:00",
        children: [
          {
            id: 5,
            tenant_id: 1,
            code: "FRONTEND",
            name: "前端组",
            type: "team",
            description: "前端研发",
            parent_id: 2,
            ancestors: "0.1.2",
            sort: 1,
            status: 1,
            created_at: "2026-04-26 10:30:00",
            updated_at: "2026-04-26 10:30:00",
          },
          {
            id: 6,
            tenant_id: 1,
            code: "BACKEND",
            name: "后端组",
            type: "team",
            description: "服务端研发",
            parent_id: 2,
            ancestors: "0.1.2",
            sort: 2,
            status: 1,
            created_at: "2026-04-26 10:35:00",
            updated_at: "2026-04-26 10:35:00",
          },
        ],
      },
      {
        id: 3,
        tenant_id: 1,
        code: "PRODUCT",
        name: "产品中心",
        type: "department",
        description: "产品规划",
        parent_id: 1,
        ancestors: "0.1",
        sort: 2,
        status: 1,
        created_at: "2026-04-26 10:10:00",
        updated_at: "2026-04-26 10:10:00",
      },
      {
        id: 4,
        tenant_id: 1,
        code: "OPS",
        name: "运营中心",
        type: "department",
        description: "日常运营",
        parent_id: 1,
        ancestors: "0.1",
        sort: 3,
        status: 0,
        created_at: "2026-04-26 10:15:00",
        updated_at: "2026-04-26 10:15:00",
      },
    ],
  },
]

export function OrganizationsPage() {
  const t = useTranslation()
  const [orgTree, setOrgTree] = useState<TreeOrgItem[]>([])
  const [expandedIds, setExpandedIds] = useState<Set<number>>(new Set())
  const [searchTerm, setSearchTerm] = useState("")
  const [isLoading, setIsLoading] = useState(true)
  const [isSubmitting, setIsSubmitting] = useState(false)

  const [dialogOpen, setDialogOpen] = useState(false)
  const [dialogMode, setDialogMode] = useState<"add" | "edit">("add")
  const [currentOrg, setCurrentOrg] = useState<OrganizationItem | null>(null)
  const [parentOptions, setParentOptions] = useState<
    { label: string; value: number }[]
  >([])

  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [orgToDelete, setOrgToDelete] = useState<TreeOrgItem | null>(null)

  const buildParentOptions = (list: TreeOrgItem[], prefix = ""): { label: string; value: number }[] => {
    const options: { label: string; value: number }[] = []
    list.forEach((org) => {
      options.push({ label: `${prefix}${org.name} (${org.code})`, value: org.id })
      if (org.children && org.children.length > 0) {
        options.push(
          ...buildParentOptions(org.children, prefix + "├── ")
        )
      }
    })
    return options
  }

  const fetchOrgs = useCallback(async () => {
    setIsLoading(true)
    try {
      const response = await organizationApi.tree()
      const tree = (response?.tree || []) as TreeOrgItem[]
      setOrgTree(tree)
      setParentOptions(buildParentOptions(tree))

      const ids = new Set<number>()
      const collectExpandable = (items: TreeOrgItem[]) => {
        items.forEach((o) => {
          if (o.children && o.children.length > 0) {
            ids.add(o.id)
            collectExpandable(o.children)
          }
        })
      }
      collectExpandable(tree)
      setExpandedIds(ids)
    } catch {
      setOrgTree(mockOrgTree)
      setParentOptions(buildParentOptions(mockOrgTree))
      const ids = new Set<number>()
      const collectExpandable = (items: TreeOrgItem[]) => {
        items.forEach((o) => {
          if (o.children && o.children.length > 0) {
            ids.add(o.id)
            collectExpandable(o.children)
          }
        })
      }
      collectExpandable(mockOrgTree)
      setExpandedIds(ids)
    } finally {
      setIsLoading(false)
    }
  }, [])

  useEffect(() => {
    fetchOrgs()
  }, [fetchOrgs])

  const orgFormSchema: FormSchema = useMemo(
    () => ({
      items: [
        {
          field: "parent_id",
          label: t.pages.organizations?.parentOrg || "父级组织",
          type: "select",
          tooltip: t.pages.organizations?.parentOrgTip,
          options: [
            {
              label: t.pages.organizations?.rootOrg || "根组织 (无父级)",
              value: 0,
            },
            ...parentOptions,
          ],
        },
        {
          field: "name",
          label: t.pages.organizations?.name || "组织名称",
          type: "text",
          required: true,
          placeholder: "请输入组织名称",
        },
        {
          field: "code",
          label: t.pages.organizations?.code || "组织编码",
          type: "text",
          required: true,
          placeholder: "租户内唯一，如 TECH",
          rules: [
            {
              pattern: "^[A-Za-z0-9_]+$",
              message: "仅支持字母、数字、下划线",
            },
          ],
        },
        {
          field: "type",
          label: t.pages.organizations?.type || "组织类型",
          type: "select",
          required: true,
          defaultValue: "department",
          options: [
            {
              label: t.pages.organizations?.typeCompany || "公司",
              value: "company",
            },
            {
              label: t.pages.organizations?.typeDepartment || "部门",
              value: "department",
            },
            {
              label: t.pages.organizations?.typeTeam || "团队",
              value: "team",
            },
          ],
        },
        {
          field: "admin_code",
          label: t.pages.organizations?.adminCode || "管理员账号编码",
          type: "text",
          placeholder: "可选，业务自定义",
        },
        {
          field: "description",
          label: t.pages.organizations?.description || "组织描述",
          type: "textarea",
          placeholder: "可选",
        },
        {
          field: "sort",
          label: t.pages.organizations?.sort || "排序",
          type: "number",
          defaultValue: 0,
        },
        {
          field: "status",
          label: t.pages.organizations?.status || "状态",
          type: "radio",
          defaultValue: 1,
          options: [
            { label: t.pages.organizations?.enabled || "启用", value: 1 },
            { label: t.pages.organizations?.disabled || "停用", value: 0 },
          ],
        },
      ],
    }),
    [t, parentOptions]
  )

  const handleAddRoot = () => {
    setDialogMode("add")
    setCurrentOrg(null)
    setDialogOpen(true)
  }

  const handleAddChild = (parent: TreeOrgItem) => {
    setDialogMode("add")
    setCurrentOrg({
      id: 0,
      parent_id: parent.id,
      type: "department",
      sort: 0,
      status: 1,
      name: "",
      code: "",
    } as OrganizationItem)
    setDialogOpen(true)
  }

  const handleEdit = (org: OrganizationItem) => {
    setDialogMode("edit")
    setCurrentOrg(org)
    setDialogOpen(true)
  }

  const handleDeleteConfirm = (org: TreeOrgItem) => {
    setOrgToDelete(org)
    setDeleteDialogOpen(true)
  }

  const handleDelete = async () => {
    if (!orgToDelete) return
    if (orgToDelete.parent_id === 0) {
      alert(
        t.pages.organizations?.rootDeleteWarn || "根组织不可删除"
      )
      return
    }
    if (orgToDelete.children && orgToDelete.children.length > 0) {
      alert(
        t.pages.organizations?.hasChildrenWarn ||
          "该组织下仍有子组织，请先处理子节点"
      )
      return
    }
    setIsSubmitting(true)
    try {
      await organizationApi.delete(orgToDelete.id)
      await fetchOrgs()
      setDeleteDialogOpen(false)
    } catch (err) {
      const msg =
        err instanceof ApiError
          ? err.message
          : (t.common.failed as string) || "删除失败，请重试"
      console.error("Delete organization failed:", err)
      alert(msg)
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
      const sortVal =
        typeof values.sort === "number"
          ? values.sort
          : parseInt(String(values.sort), 10) || 0
      const statusVal =
        typeof values.status === "number"
          ? values.status
          : parseInt(String(values.status), 10)

      if (dialogMode === "add") {
        const payload: Partial<OrganizationItem> = {
          name: values.name as string,
          code: values.code as string,
          type: values.type as string,
          admin_code: (values.admin_code as string) || "",
          description: (values.description as string) || "",
          parent_id: parentId,
          sort: sortVal,
          status: statusVal,
        }
        await organizationApi.create(payload)
      } else if (currentOrg) {
        // PUT 接口仅允许更新 name / type / description / admin_code / sort / status
        // 不修改 code / parent_id / ancestors
        const payload: Partial<OrganizationItem> = {
          name: values.name as string,
          type: values.type as string,
          admin_code: (values.admin_code as string) || "",
          description: (values.description as string) || "",
          sort: sortVal,
          status: statusVal,
        }
        await organizationApi.update(currentOrg.id, payload)
      }
      await fetchOrgs()
      setDialogOpen(false)
    } catch (err) {
      const msg =
        err instanceof ApiError
          ? err.message
          : (t.common.failed as string) || "保存失败，请重试"
      console.error("Save organization failed:", err)
      alert(msg)
    } finally {
      setIsSubmitting(false)
    }
  }

  const getInitialValues = () => {
    if (currentOrg) {
      if (dialogMode === "add") {
        return {
          parent_id: currentOrg.parent_id ?? 0,
          type: currentOrg.type || "department",
          sort: 0,
          status: 1,
        }
      }
      return {
        name: currentOrg.name,
        code: currentOrg.code,
        type: currentOrg.type,
        admin_code: currentOrg.admin_code || "",
        description: currentOrg.description || "",
        sort: currentOrg.sort ?? 0,
        status: currentOrg.status ?? 1,
      }
    }
    return { parent_id: 0, type: "department", sort: 0, status: 1 }
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

  const expandAll = () => {
    const ids = new Set<number>()
    const walk = (items: TreeOrgItem[]) => {
      items.forEach((o) => {
        if (o.children && o.children.length > 0) {
          ids.add(o.id)
          walk(o.children)
        }
      })
    }
    walk(orgTree)
    setExpandedIds(ids)
  }

  const collapseAll = () => setExpandedIds(new Set())

  const filterTree = (items: TreeOrgItem[], term: string): TreeOrgItem[] => {
    if (!term) return items
    const lower = term.toLowerCase()
    return items.reduce((acc, item) => {
      const matches =
        item.name.toLowerCase().includes(lower) ||
        item.code.toLowerCase().includes(lower)
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
    }, [] as TreeOrgItem[])
  }

  const filteredTree = filterTree(orgTree, searchTerm)

  // 当搜索时自动展开所有匹配的节点
  const visibleExpandedIds = useMemo(() => {
    if (!searchTerm) return expandedIds
    const ids = new Set<number>()
    const walk = (items: TreeOrgItem[], ancestors: number[]) => {
      items.forEach((o) => {
        const path = [...ancestors, o.id]
        if (o.children && o.children.length > 0) {
          ids.add(o.id)
          walk(o.children, path)
        }
      })
    }
    walk(filteredTree, [])
    return ids
  }, [searchTerm, filteredTree, expandedIds])

  const countOrgs = (items: TreeOrgItem[]): number => {
    return items.reduce(
      (acc, item) =>
        acc + 1 + (item.children ? countOrgs(item.children) : 0),
      0
    )
  }

  const orgCount = countOrgs(orgTree)

  const flatten = (items: TreeOrgItem[]): TreeOrgItem[] => {
    const out: TreeOrgItem[] = []
    items.forEach((o) => {
      out.push(o)
      if (o.children) out.push(...flatten(o.children))
    })
    return out
  }

  const allFlat = flatten(orgTree)
  const rootCount = allFlat.filter((o) => o.parent_id === 0).length
  const enabledCount = allFlat.filter((o) => o.status === 1).length
  const disabledCount = allFlat.filter((o) => o.status === 0).length

  const getTypeBadge = (type: string) => {
    switch (type) {
      case "company":
        return (
          <Badge variant="default" className="font-normal">
            <BuildingIcon className="mr-1 h-3 w-3" />
            {t.pages.organizations?.typeCompany || "公司"}
          </Badge>
        )
      case "team":
        return (
          <Badge variant="outline" className="font-normal">
            <UsersIcon className="mr-1 h-3 w-3" />
            {t.pages.organizations?.typeTeam || "团队"}
          </Badge>
        )
      case "department":
      default:
        return (
          <Badge variant="secondary" className="font-normal">
            <Building2Icon className="mr-1 h-3 w-3" />
            {t.pages.organizations?.typeDepartment || "部门"}
          </Badge>
        )
    }
  }

  return (
    <PageLayout>
      <div className="px-4 lg:px-6">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-2xl font-bold">
              {t.pages.organizations?.title || "组织管理"}
            </h1>
            <p className="text-sm text-muted-foreground">
              {t.pages.organizations?.subtitle ||
                "管理租户组织树结构与数据范围"}
            </p>
          </div>
          <div className="flex items-center gap-2">
            <Button variant="outline" onClick={fetchOrgs} disabled={isLoading}>
              <RefreshCw
                className={`mr-2 h-4 w-4 ${isLoading ? "animate-spin" : ""}`}
              />
              {t.pages.organizations?.refresh || "刷新列表"}
            </Button>
            <Button onClick={handleAddRoot}>
              <PlusIcon className="mr-2 h-4 w-4" />
              {t.pages.organizations?.addRootOrg || "添加根组织"}
            </Button>
          </div>
        </div>

        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4 mb-6">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">
                {t.pages.organizations?.totalCount || "组织总数"}
              </CardTitle>
              <NetworkIcon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{orgCount}</div>
              <p className="text-xs text-muted-foreground">
                {t.pages.organizations?.totalOrgs || "共"} {orgCount}{" "}
                {t.pages.organizations?.orgs || "个组织"}
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">
                {t.pages.organizations?.rootCount || "根组织数"}
              </CardTitle>
              <GitBranchIcon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{rootCount}</div>
              <p className="text-xs text-muted-foreground">ancestors = 0</p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">
                {t.pages.organizations?.enabledCount || "启用中"}
              </CardTitle>
              <PowerIcon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{enabledCount}</div>
              <p className="text-xs text-muted-foreground">
                status = 1
              </p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">
                {t.pages.organizations?.disabledCount || "已停用"}
              </CardTitle>
              <PowerOffIcon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{disabledCount}</div>
              <p className="text-xs text-muted-foreground">status = 0</p>
            </CardContent>
          </Card>
        </div>

        <Card>
          <CardHeader>
            <div className="flex items-center justify-between gap-4 flex-wrap">
              <div>
                <CardTitle>
                  {t.pages.organizations?.orgTree || "组织树"}
                </CardTitle>
                <CardDescription>
                  {t.pages.organizations?.treeDesc ||
                    "当前用户数据范围下可见的组织结构"}
                </CardDescription>
              </div>
              <div className="flex items-center gap-2 flex-1 max-w-md justify-end">
                <div className="relative w-full max-w-xs">
                  <SearchIcon className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                  <Input
                    placeholder={
                      t.pages.organizations?.searchPlaceholder ||
                      "搜索组织..."
                    }
                    className="pl-9"
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                  />
                </div>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={expandAll}
                  disabled={isLoading}
                >
                  展开
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={collapseAll}
                  disabled={isLoading}
                >
                  折叠
                </Button>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-[320px]">
                    {t.pages.organizations?.name || "组织名称"}
                  </TableHead>
                  <TableHead>
                    {t.pages.organizations?.code || "组织编码"}
                  </TableHead>
                  <TableHead>
                    {t.pages.organizations?.type || "组织类型"}
                  </TableHead>
                  <TableHead>
                    {t.pages.organizations?.adminCode || "管理员"}
                  </TableHead>
                  <TableHead className="w-[80px]">
                    {t.pages.organizations?.sort || "排序"}
                  </TableHead>
                  <TableHead className="w-[100px]">
                    {t.pages.organizations?.status || "状态"}
                  </TableHead>
                  <TableHead className="w-[180px] text-right">
                    {t.common.edit}
                  </TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredTree.map((item) => (
                  <OrgTreeRow
                    key={item.id}
                    item={item}
                    level={0}
                    expandedIds={visibleExpandedIds}
                    onToggle={toggleExpand}
                    onEdit={handleEdit}
                    onDelete={handleDeleteConfirm}
                    onAddChild={handleAddChild}
                    searchTerm={searchTerm}
                    renderTypeBadge={getTypeBadge}
                  />
                ))}
                {filteredTree.length === 0 && !isLoading && (
                  <TableRow>
                    <TableCell
                      colSpan={8}
                      className="text-center py-8 text-muted-foreground"
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
            {t.pages.organizations?.deleteTip && (
              <p className="text-xs text-muted-foreground mt-4">
                {t.pages.organizations.deleteTip}
              </p>
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
            ? currentOrg && currentOrg.parent_id
              ? t.pages.organizations?.addChildOrg || "添加子组织"
              : t.pages.organizations?.addRootOrg || "添加根组织"
            : t.pages.organizations?.editOrg || "编辑组织"
        }
        schema={orgFormSchema}
        initialValues={getInitialValues()}
        onSubmit={handleSubmit}
        loading={isSubmitting}
      />

      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent className="sm:max-w-[420px]">
          <DialogHeader>
            <DialogTitle>
              {t.pages.organizations?.deleteOrg || "删除组织"}
            </DialogTitle>
            <DialogDescription>
              {(t.pages.organizations?.confirmDelete || '确定要删除组织 "{name}" 吗？').replace(
                "{name}",
                orgToDelete?.name || ""
              )}
              {orgToDelete?.parent_id === 0 && (
                <span className="block mt-2 text-destructive">
                  {t.pages.organizations?.rootDeleteWarn || "根组织不可删除"}
                </span>
              )}
              {orgToDelete &&
                orgToDelete.parent_id !== 0 &&
                orgToDelete.children &&
                orgToDelete.children.length > 0 && (
                  <span className="block mt-2 text-destructive">
                    {t.pages.organizations?.hasChildrenWarn ||
                      "该组织下仍有子组织，请先处理子节点"}
                  </span>
                )}
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
              disabled={
                isSubmitting ||
                orgToDelete?.parent_id === 0 ||
                !!(orgToDelete?.children && orgToDelete.children.length > 0)
              }
            >
              {t.common.delete}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </PageLayout>
  )
}

interface OrgTreeRowProps {
  item: TreeOrgItem
  level: number
  expandedIds: Set<number>
  onToggle: (id: number) => void
  onEdit: (org: OrganizationItem) => void
  onDelete: (org: TreeOrgItem) => void
  onAddChild: (org: TreeOrgItem) => void
  searchTerm: string
  renderTypeBadge: (type: string) => React.ReactNode
}

function OrgTreeRow({
  item,
  level,
  expandedIds,
  onToggle,
  onEdit,
  onDelete,
  onAddChild,
  searchTerm,
  renderTypeBadge,
}: OrgTreeRowProps) {
  const t = useTranslation()
  const hasChildren = item.children && item.children.length > 0
  const isExpanded = expandedIds.has(item.id) || !!searchTerm

  return (
    <>
      <TableRow className="group">
        <TableCell>
          <div
            className="flex items-center"
            style={{ paddingLeft: `${level * 24}px` }}
          >
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
            <div className="ml-1 flex items-center gap-2">
              {item.parent_id === 0 ? (
                <NetworkIcon className="h-4 w-4 text-primary" />
              ) : (
                <Building2Icon className="h-4 w-4 text-muted-foreground" />
              )}
              <span className="font-medium">{item.name}</span>
              {item.parent_id === 0 && (
                <Badge variant="outline" className="text-[10px] h-4 px-1">
                  ROOT
                </Badge>
              )}
              {item.status === 0 && (
                <Badge variant="secondary" className="text-[10px] h-4 px-1">
                  {t.pages.organizations?.disabled || "停用"}
                </Badge>
              )}
            </div>
          </div>
        </TableCell>
        <TableCell className="font-mono text-sm">
          <div className="flex items-center gap-1">
            <HashIcon className="h-3 w-3 text-muted-foreground" />
            {item.code}
          </div>
        </TableCell>
        <TableCell>{renderTypeBadge(item.type)}</TableCell>
       
        <TableCell className="text-sm text-muted-foreground">
          {item.admin_code || "-"}
        </TableCell>
        <TableCell>{item.sort}</TableCell>
        <TableCell>
          <Badge variant={item.status === 1 ? "default" : "secondary"}>
            {item.status === 1
              ? t.pages.organizations?.enabled || "启用"
              : t.pages.organizations?.disabled || "停用"}
          </Badge>
        </TableCell>
        <TableCell>
          <div className="flex items-center justify-end gap-1">
            <Button
              variant="ghost"
              size="sm"
              onClick={() => onAddChild(item)}
              title={
                t.pages.organizations?.addChildOrg || "添加子组织"
              }
            >
              <PlusIcon className="h-4 w-4" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              onClick={() => onEdit(item)}
              title={t.common.edit}
            >
              <EditIcon className="h-4 w-4" />
            </Button>
            <Button
              variant="ghost"
              size="icon"
              className="h-8 w-8"
              onClick={() => onDelete(item)}
              title={t.common.delete}
            >
              <TrashIcon className="h-4 w-4 text-destructive" />
            </Button>
          </div>
        </TableCell>
      </TableRow>
      {hasChildren && isExpanded && item.children?.map((child) => (
        <OrgTreeRow
          key={child.id}
          item={child as TreeOrgItem}
          level={level + 1}
          expandedIds={expandedIds}
          onToggle={onToggle}
          onEdit={onEdit}
          onDelete={onDelete}
          onAddChild={onAddChild}
          searchTerm={searchTerm}
          renderTypeBadge={renderTypeBadge}
        />
      ))}
    </>
  )
}
