import { useEffect, useState, useCallback, useMemo } from "react"
import { PageLayout } from "@/components/page-layout"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
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
  UsersIcon,
  RefreshCw,
  Building2Icon,
  ChevronRightIcon,
  ChevronDownIcon,
  XIcon,
  FolderTreeIcon,
} from "lucide-react"
import { useTranslation } from "@/locales"
import { userApi, organizationApi, type UserItem, type OrganizationItem } from "@/api"
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

type OrgNode = OrganizationItem & { children?: OrgNode[] }

const ALL_ORG_ID = 0

const mockUsers: UserItem[] = [
  { id: 1, code: "admin", real_name: "超级管理员", email: "admin@example.com", phone: "13800000001", role: "super_admin", status: 1, org_id: 1 },
  { id: 2, code: "ceo", real_name: "张总", email: "ceo@example.com", phone: "13800000002", role: "admin", status: 1, org_id: 1 },
  { id: 3, code: "cto", real_name: "李技术", email: "cto@example.com", phone: "13800000003", role: "admin", status: 1, org_id: 2 },
  { id: 4, code: "fe01", real_name: "王前端", email: "fe01@example.com", phone: "13800000004", role: "user", status: 1, org_id: 5 },
  { id: 5, code: "fe02", real_name: "赵前端", email: "fe02@example.com", phone: "13800000005", role: "user", status: 1, org_id: 5 },
  { id: 6, code: "be01", real_name: "钱后端", email: "be01@example.com", phone: "13800000006", role: "user", status: 1, org_id: 6 },
  { id: 7, code: "be02", real_name: "孙后端", email: "be02@example.com", phone: "13800000007", role: "user", status: 2, org_id: 6 },
  { id: 8, code: "pm01", real_name: "周产品", email: "pm01@example.com", phone: "13800000008", role: "user", status: 1, org_id: 3 },
  { id: 9, code: "op01", real_name: "吴运营", email: "op01@example.com", phone: "13800000009", role: "user", status: 1, org_id: 4 },
  { id: 10, code: "guest", real_name: "访客", email: "guest@example.com", phone: "13800000010", role: "guest", status: 2 },
]

const mockOrgTree: OrgNode[] = [
  {
    id: 1, code: "HQ", name: "总部", type: "company", parent_id: 0, ancestors: "0", sort: 0, status: 1,
    children: [
      {
        id: 2, code: "TECH", name: "技术中心", type: "department", parent_id: 1, ancestors: "0.1", sort: 1, status: 1,
        children: [
          { id: 5, code: "FRONTEND", name: "前端组", type: "team", parent_id: 2, ancestors: "0.1.2", sort: 1, status: 1 },
          { id: 6, code: "BACKEND", name: "后端组", type: "team", parent_id: 2, ancestors: "0.1.2", sort: 2, status: 1 },
        ],
      },
      { id: 3, code: "PRODUCT", name: "产品中心", type: "department", parent_id: 1, ancestors: "0.1", sort: 2, status: 1 },
      { id: 4, code: "OPS", name: "运营中心", type: "department", parent_id: 1, ancestors: "0.1", sort: 3, status: 1 },
    ],
  },
]

function collectOrgSubtreeIds(nodes: OrgNode[], target: number): Set<number> | null {
  let found: OrgNode | null = null
  const walk = (arr: OrgNode[]): boolean => {
    for (const n of arr) {
      if (n.id === target) { found = n; return true }
      if (n.children && walk(n.children)) return true
    }
    return false
  }
  walk(nodes)
  if (!found) return null
  const ids = new Set<number>()
  const collect = (n: OrgNode) => {
    ids.add(n.id)
    n.children?.forEach(collect)
  }
  collect(found)
  return ids
}

function findOrgName(nodes: OrgNode[], id: number): string | null {
  for (const n of nodes) {
    if (n.id === id) return n.name
    if (n.children) {
      const r = findOrgName(n.children, id)
      if (r) return r
    }
  }
  return null
}

function filterOrgTree(nodes: OrgNode[], keyword: string): OrgNode[] {
  if (!keyword.trim()) return nodes
  const kw = keyword.toLowerCase()
  const match = (n: OrgNode): OrgNode | null => {
    const selfMatch = n.name.toLowerCase().includes(kw) || n.code.toLowerCase().includes(kw)
    const matchedChildren = (n.children || []).map(match).filter(Boolean) as OrgNode[]
    if (selfMatch || matchedChildren.length) {
      return { ...n, children: matchedChildren }
    }
    return null
  }
  return nodes.map(match).filter(Boolean) as OrgNode[]
}

function collectAllIds(nodes: OrgNode[]): Set<number> {
  const ids = new Set<number>()
  const walk = (arr: OrgNode[]) => arr.forEach((n) => {
    ids.add(n.id)
    n.children && walk(n.children)
  })
  walk(nodes)
  return ids
}

interface OrgTreeViewProps {
  nodes: OrgNode[]
  expandedIds: Set<number>
  selectedId: number
  countByOrgId: Map<number, number>
  onToggle: (id: number) => void
  onSelect: (id: number) => void
  showAll?: boolean
  allLabel?: string
  totalCount?: number
  level?: number
}

function OrgTreeView({
  nodes,
  expandedIds,
  selectedId,
  countByOrgId,
  onToggle,
  onSelect,
  showAll = false,
  allLabel = "",
  totalCount = 0,
  level = 0,
}: OrgTreeViewProps) {
  return (
    <div className="space-y-0.5">
      {showAll && level === 0 && (
        <button
          onClick={() => onSelect(ALL_ORG_ID)}
          className={cn(
            "w-full flex items-center gap-1.5 px-2 py-1.5 text-sm rounded hover:bg-muted/60 transition-colors",
            selectedId === ALL_ORG_ID && "bg-primary/10 text-primary font-medium"
          )}
        >
          <UsersIcon className="h-4 w-4 text-muted-foreground shrink-0" />
          <span className="flex-1 text-left">{allLabel}</span>
          <Badge variant="secondary" className="text-[10px] h-4 px-1.5 font-normal">
            {totalCount}
          </Badge>
        </button>
      )}
      {nodes.map((n) => (
        <OrgTreeNode
          key={n.id}
          node={n}
          level={level}
          expandedIds={expandedIds}
          selectedId={selectedId}
          countByOrgId={countByOrgId}
          onToggle={onToggle}
          onSelect={onSelect}
        />
      ))}
    </div>
  )
}

function OrgTreeNode({
  node,
  level,
  expandedIds,
  selectedId,
  countByOrgId,
  onToggle,
  onSelect,
}: {
  node: OrgNode
  level: number
  expandedIds: Set<number>
  selectedId: number
  countByOrgId: Map<number, number>
  onToggle: (id: number) => void
  onSelect: (id: number) => void
}) {
  const hasChildren = !!(node.children && node.children.length)
  const isExpanded = expandedIds.has(node.id)
  const isSelected = selectedId === node.id
  const count = countByOrgId.get(node.id) || 0

  return (
    <div>
      <div
        className={cn(
          "group flex items-center gap-1 px-1 py-1 text-sm rounded hover:bg-muted/60 transition-colors",
          isSelected && "bg-primary/10 text-primary"
        )}
        style={{ paddingLeft: `${level * 12 + 4}px` }}
      >
        {hasChildren ? (
          <button
            onClick={(e) => { e.stopPropagation(); onToggle(node.id) }}
            className="p-0.5 hover:bg-muted rounded shrink-0"
            aria-label={isExpanded ? "collapse" : "expand"}
          >
            {isExpanded ? (
              <ChevronDownIcon className="h-3.5 w-3.5" />
            ) : (
              <ChevronRightIcon className="h-3.5 w-3.5" />
            )}
          </button>
        ) : (
          <span className="w-4 shrink-0" />
        )}
        <button
          onClick={() => onSelect(node.id)}
          className={cn(
            "flex-1 flex items-center gap-1.5 min-w-0 text-left",
            isSelected && "font-medium"
          )}
        >
          <Building2Icon className="h-3.5 w-3.5 text-muted-foreground shrink-0" />
          <span className="truncate">{node.name}</span>
          {node.status === 0 && (
            <Badge variant="secondary" className="text-[10px] h-4 px-1 font-normal">
              停用
            </Badge>
          )}
        </button>
        <Badge variant="outline" className="text-[10px] h-4 px-1.5 font-normal shrink-0">
          {count}
        </Badge>
      </div>
      {hasChildren && isExpanded && (
        <div>
          {node.children!.map((child) => (
            <OrgTreeNode
              key={child.id}
              node={child}
              level={level + 1}
              expandedIds={expandedIds}
              selectedId={selectedId}
              countByOrgId={countByOrgId}
              onToggle={onToggle}
              onSelect={onSelect}
            />
          ))}
        </div>
      )}
    </div>
  )
}


function flattenOrgOptions(nodes: OrgNode[], depth = 0): { label: string; value: number }[] {
  const out: { label: string; value: number }[] = []
  const sorted = [...nodes].sort((a, b) => a.sort - b.sort)
  for (const n of sorted) {
    out.push({ label: `${"  \u2514\u2500".repeat(depth)}${n.name}`, value: n.id })
    if (n.children && n.children.length) {
      out.push(...flattenOrgOptions(n.children, depth + 1))
    }
  }
  return out
}

export function UsersPage() {
  const t = useTranslation()
  const [allUsers, setAllUsers] = useState<UserItem[]>([])
  const [total, setTotal] = useState(0)
  const [isLoading, setIsLoading] = useState(true)
  const [searchTerm, setSearchTerm] = useState("")
  const [isSubmitting, setIsSubmitting] = useState(false)

  const [orgTree, setOrgTree] = useState<OrgNode[]>([])
  const [expandedIds, setExpandedIds] = useState<Set<number>>(new Set())
  const [selectedOrgId, setSelectedOrgId] = useState<number>(ALL_ORG_ID)
  const [orgSearch, setOrgSearch] = useState("")

  const [dialogOpen, setDialogOpen] = useState(false)
  const [dialogMode, setDialogMode] = useState<"add" | "edit">("add")
  const [currentUser, setCurrentUser] = useState<UserItem | null>(null)

  const [statusDialogOpen, setStatusDialogOpen] = useState(false)
  const [currentStatusUser, setCurrentStatusUser] = useState<UserItem | null>(null)
  const [newStatus, setNewStatus] = useState<number>(1)

  const fetchUsers = useCallback(async () => {
    setIsLoading(true)
    try {
      const response = await userApi.list({
        keyword: searchTerm || undefined,
        org_id: selectedOrgId !== ALL_ORG_ID ? selectedOrgId : undefined,
        page: 1,
        size: 100,
      })
      const list = (response.list || []).map((u) => ({
        ...u,
        status: Number(u.status),
      })) as UserItem[]
      setAllUsers(list)
      setTotal(response.total)
    } catch {
      const list = mockUsers.map((u) => ({ ...u, status: Number(u.status) }))
      const filtered = selectedOrgId === ALL_ORG_ID
        ? list
        : list.filter((u) => u.org_id === selectedOrgId)
      setAllUsers(filtered)
      setTotal(mockUsers.length)
    } finally {
      setIsLoading(false)
    }
  }, [searchTerm, selectedOrgId])

  const fetchOrgTree = useCallback(async () => {
    try {
      const res = await organizationApi.tree()
      const tree = (res?.tree || []) as OrgNode[]
      setOrgTree(tree)
      setExpandedIds(collectAllIds(tree))
    } catch {
      setOrgTree(mockOrgTree)
      setExpandedIds(collectAllIds(mockOrgTree))
    }
  }, [])

  useEffect(() => {
    fetchUsers()
  }, [fetchUsers])

  useEffect(() => {
    fetchOrgTree()
  }, [fetchOrgTree])

  const filteredOrgTree = useMemo(
    () => filterOrgTree(orgTree, orgSearch),
    [orgTree, orgSearch]
  )

  useEffect(() => {
    if (orgSearch.trim()) {
      setExpandedIds(collectAllIds(filteredOrgTree))
    }
  }, [orgSearch, filteredOrgTree])

  const countByOrgId = useMemo(() => {
    const m = new Map<number, number>()
    allUsers.forEach((u) => {
      if (u.org_id != null) m.set(u.org_id, (m.get(u.org_id) || 0) + 1)
    })
    return m
  }, [allUsers])

  const filteredUsers = useMemo(() => {
    if (selectedOrgId === ALL_ORG_ID) return allUsers
    const subtree = collectOrgSubtreeIds(orgTree, selectedOrgId)
    if (!subtree) return allUsers
    return allUsers.filter((u) => u.org_id != null && subtree.has(u.org_id))
  }, [allUsers, orgTree, selectedOrgId])

  const toggleExpand = useCallback((id: number) => {
    setExpandedIds((prev) => {
      const next = new Set(prev)
      if (next.has(id)) next.delete(id)
      else next.add(id)
      return next
    })
  }, [])

  const selectedOrgName = useMemo(() => {
    if (selectedOrgId === ALL_ORG_ID) return t.pages.users?.allUsers || "全部用户"
    return findOrgName(orgTree, selectedOrgId) || `#${selectedOrgId}`
  }, [selectedOrgId, orgTree, t])

  const userFormSchema: FormSchema = useMemo(() => {
    const items: FormSchema["items"] = [
      {
        field: "username",
        label: t.pages.users?.account || "账户",
        type: "text",
        required: true,
        placeholder: "请输入账户名（登录账号）",
        disabled: dialogMode === "edit",
      },
      {
        field: "real_name",
        label: t.pages.users?.nameLabel || "姓名",
        type: "text",
        required: true,
        placeholder: "请输入真实姓名",
      },
    ]

    if (dialogMode === "add") {
      items.push({
        field: "password",
        label: "密码",
        type: "password",
        required: true,
        placeholder: "请输入密码（至少6位）",
        rules: [{ minLength: 6, message: "密码长度至少 6 位" }],
      })
    }

    items.push(
      {
        field: "org_id",
        label: t.pages.users?.orgId || "\u6240\u5c5e\u7ec4\u7ec7",
        type: "select",
        placeholder: t.pages.users?.orgPlaceholder || "\u4e0d\u6307\u5b9a",
        options: [
          { label: t.pages.users?.orgNone || "\u4e0d\u6307\u5b9a", value: 0 },
          ...flattenOrgOptions(orgTree),
        ],
      },
      {
        field: "phone",
        label: t.pages.users?.phone || "手机",
        type: "text",
        required: true,
        placeholder: "请输入手机号",
      },
      {
        field: "email",
        label: t.pages.users?.email || "邮箱",
        type: "email",
        placeholder: "请输入邮箱地址",
      },
      {
        field: "status",
        label: t.pages.users?.status || "状态",
        type: "radio",
        defaultValue: 1,
        options: [
          { label: t.pages.users?.active || "启用", value: 1 },
          { label: t.pages.users?.inactive || "停用", value: 2 },
        ],
      }
    )

    return { items }
  }, [t, dialogMode, orgTree])

  const handleAdd = () => {
    setDialogMode("add")
    setCurrentUser(null)
    setDialogOpen(true)
  }

  const handleEdit = (user: UserItem) => {
    setDialogMode("edit")
    setCurrentUser(user)
    setDialogOpen(true)
  }

  const handleDelete = async (user: UserItem) => {
    if (!confirm(`确定要删除用户 "${user.real_name}" 吗？`)) return
    try {
      await userApi.delete(user.id)
      await fetchUsers()
    } catch (error) {
      console.error("Delete user failed:", error)
      alert("删除失败，请重试")
    }
  }

  const handleStatusChange = (user: UserItem) => {
    setCurrentStatusUser(user)
    setNewStatus(Number(user.status) === 1 ? 2 : 1)
    setStatusDialogOpen(true)
  }

  const handleStatusSubmit = async () => {
    if (!currentStatusUser) return
    setIsSubmitting(true)
    try {
      await userApi.updateStatus(currentStatusUser.id, newStatus)
      await fetchUsers()
      setStatusDialogOpen(false)
    } catch (error) {
      console.error("Update status failed:", error)
      alert("更新状态失败，请重试")
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleSubmit = async (values: Record<string, unknown>) => {
    setIsSubmitting(true)
    try {
      const statusNum =
        typeof values.status === "string"
          ? parseInt(values.status as string, 10)
          : Number(values.status)
      if (dialogMode === "add") {
        const createPayload: Partial<UserItem> & { username: string; password?: string } = {
          username: String(values.username ?? ""),
          real_name: String(values.real_name ?? ""),
          phone: (values.phone as string) || "",
          email: (values.email as string) || "",
          status: statusNum,
          code: String(values.username ?? ""),
          role: (values.role as string) || "user",
        }
        if (values.password) {
          createPayload.password = String(values.password)
        }
        if (values.org_id !== undefined && values.org_id !== null && Number(values.org_id) > 0) {
          createPayload.org_id = Number(values.org_id)
        }
        await userApi.create(createPayload)
      } else if (currentUser) {
        const updatePayload: Partial<UserItem> = {
          real_name: String(values.real_name ?? ""),
          phone: (values.phone as string) || "",
          email: (values.email as string) || "",
          status: statusNum,
        }
        updatePayload.org_id = values.org_id !== undefined && values.org_id !== null ? Number(values.org_id) : null
        await userApi.patch(currentUser.id, updatePayload)
      }
      await fetchUsers()
      setDialogOpen(false)
    } catch (error: any) {
      console.error("Save user failed:", error)
      if (error?.status === 409) {
        alert("用户名已存在")
      } else {
        alert("保存失败，请重试")
      }
    } finally {
      setIsSubmitting(false)
    }
  }

  const getInitialValues = () => {
    if (currentUser) {
      return {
        username: currentUser.username || currentUser.code,
        real_name: currentUser.real_name,
        email: currentUser.email || "",
        phone: currentUser.phone || "",
        org_id: currentUser.org_id ?? 0,
        status: currentUser.status,
      }
    }
    return { status: 1, org_id: 0 }
  }

  const getStatusBadge = (status: number | string | undefined | null) => {
    if (Number(status) === 1) {
      return <Badge variant="default">{t.pages.users?.active || "启用"}</Badge>
    }
    return <Badge variant="secondary">{t.pages.users?.inactive || "停用"}</Badge>
  }

  const activeCount = filteredUsers.filter((u) => Number(u.status) === 1).length
  const disabledCount = filteredUsers.length - activeCount

  return (
    <PageLayout>
      <div className="px-4 lg:px-6">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-2xl font-bold">{t.pages.users?.title || "用户管理"}</h1>
            <p className="text-sm text-muted-foreground">{t.pages.users?.subtitle || "管理系统用户和权限"}</p>
          </div>
          <div className="flex items-center gap-2">
            <Button variant="outline" onClick={fetchUsers} disabled={isLoading}>
              <RefreshCw className={`mr-2 h-4 w-4 ${isLoading ? "animate-spin" : ""}`} />
              {t.pages.users?.refresh || "刷新列表"}
            </Button>
            <Button onClick={handleAdd}>
              <PlusIcon className="mr-2 h-4 w-4" />
              {t.common.add}
            </Button>
          </div>
        </div>

        <div className="grid gap-4 lg:grid-cols-[280px_1fr]">
          <Card className="lg:sticky lg:top-4 lg:self-start lg:max-h-[calc(100vh-6rem)] flex flex-col">
            <CardHeader className="pb-3 space-y-2">
              <div className="flex items-center gap-2">
                <FolderTreeIcon className="h-4 w-4 text-muted-foreground" />
                <CardTitle className="text-sm font-medium">
                  {t.pages.users?.orgTree || "组织"}
                </CardTitle>
              </div>
              <div className="relative">
                <SearchIcon className="absolute left-2.5 top-1/2 h-3.5 w-3.5 -translate-y-1/2 text-muted-foreground" />
                <Input
                  placeholder={t.pages.users?.searchOrgPlaceholder || "搜索组织..."}
                  className="pl-8 h-8 text-sm"
                  value={orgSearch}
                  onChange={(e) => setOrgSearch(e.target.value)}
                />
              </div>
            </CardHeader>
            <CardContent className="flex-1 overflow-auto pt-0">
              <OrgTreeView
                nodes={filteredOrgTree}
                expandedIds={expandedIds}
                selectedId={selectedOrgId}
                countByOrgId={countByOrgId}
                onToggle={toggleExpand}
                onSelect={setSelectedOrgId}
                showAll
                allLabel={t.pages.users?.allUsers || "全部用户"}
                totalCount={allUsers.length}
              />
            </CardContent>
          </Card>

          <div className="space-y-4 min-w-0">
            {selectedOrgId !== ALL_ORG_ID && (
              <div className="flex items-center gap-2">
                <Badge variant="secondary" className="px-2 py-1 gap-1">
                  <Building2Icon className="h-3 w-3" />
                  {selectedOrgName}
                  <button
                    className="ml-1 hover:text-foreground"
                    onClick={() => setSelectedOrgId(ALL_ORG_ID)}
                    aria-label="clear filter"
                  >
                    <XIcon className="h-3 w-3" />
                  </button>
                </Badge>
                <span className="text-xs text-muted-foreground">
                  {t.pages.users?.matchedUsers || "个匹配用户"}: {filteredUsers.length}
                </span>
              </div>
            )}

            <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3">
              <Card>
                <CardHeader className="flex flex-row items-center justify-between pb-2">
                  <CardTitle className="text-sm font-medium">
                    {selectedOrgId === ALL_ORG_ID
                      ? (t.pages.users?.totalUsers || "总用户数")
                      : (t.pages.users?.currentUsers || "当前用户")}
                  </CardTitle>
                  <UsersIcon className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold">{filteredUsers.length}</div>
                  <p className="text-xs text-muted-foreground mt-1">
                    {t.pages.users?.total || "总计"} {total}
                  </p>
                </CardContent>
              </Card>
              <Card>
                <CardHeader className="flex flex-row items-center justify-between pb-2">
                  <CardTitle className="text-sm font-medium">
                    {t.pages.users?.active || "启用"}
                  </CardTitle>
                  <UsersIcon className="h-4 w-4 text-green-600" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold text-green-600">{activeCount}</div>
                </CardContent>
              </Card>
              <Card>
                <CardHeader className="flex flex-row items-center justify-between pb-2">
                  <CardTitle className="text-sm font-medium">
                    {t.pages.users?.inactive || "停用"}
                  </CardTitle>
                  <UsersIcon className="h-4 w-4 text-muted-foreground" />
                </CardHeader>
                <CardContent>
                  <div className="text-2xl font-bold text-muted-foreground">{disabledCount}</div>
                </CardContent>
              </Card>
            </div>

            <Card>
              <CardHeader>
                <div className="flex items-center gap-2">
                  <div className="relative flex-1 max-w-sm">
                    <SearchIcon className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                    <Input
                      placeholder={t.pages.users?.searchPlaceholder || "搜索用户..."}
                      className="pl-9"
                      value={searchTerm}
                      onChange={(e) => setSearchTerm(e.target.value)}
                    />
                  </div>
                  <Badge variant="secondary">
                    {t.pages.users?.total || "共"} {filteredUsers.length} {t.pages.users?.matchedUsers?.replace("个匹配用户", "个") || "个用户"}
                  </Badge>
                </div>
              </CardHeader>
              <CardContent>
                <Table>
                  <TableHeader>
                    <TableRow>
                      <TableHead>ID</TableHead>
                      <TableHead>{t.pages.users?.nameLabel || "姓名"}</TableHead>
                      <TableHead>{t.pages.users?.account || "账户"}</TableHead>
                      <TableHead>{t.pages.users?.email || "邮箱"}</TableHead>
                      <TableHead>{t.pages.users?.phone || "手机"}</TableHead>
                      <TableHead>{t.pages.users?.role || "角色"}</TableHead>
                      <TableHead>{t.pages.users?.orgName || "所属组织"}</TableHead>
                      <TableHead>{t.pages.users?.status || "状态"}</TableHead>
                      <TableHead className="text-right">{t.common.edit}</TableHead>
                    </TableRow>
                  </TableHeader>
                  <TableBody>
                    {filteredUsers.map((user) => (
                      <TableRow key={user.id}>
                        <TableCell className="font-medium">{user.id}</TableCell>
                        <TableCell>{user.real_name}</TableCell>
                        <TableCell className="font-mono text-sm">{user.code}</TableCell>
                        <TableCell>{user.email || "-"}</TableCell>
                        <TableCell>{user.phone || "-"}</TableCell>
                        <TableCell>
                          <Badge variant="outline">{user.role}</Badge>
                        </TableCell>
                        <TableCell className="text-sm">
                          {user.org_name ? (
                            <span className="text-muted-foreground">{user.org_name}</span>
                          ) : (
                            <span className="text-muted-foreground/50">-</span>
                          )}
                        </TableCell>
                        <TableCell>
                          <button
                            onClick={() => handleStatusChange(user)}
                            className="cursor-pointer"
                          >
                            {getStatusBadge(user.status)}
                          </button>
                        </TableCell>
                        <TableCell className="text-right">
                          <div className="flex items-center justify-end gap-1">
                            <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => handleEdit(user)}>
                              <EditIcon className="h-4 w-4" />
                            </Button>
                            <Button variant="ghost" size="icon" className="h-8 w-8" onClick={() => handleDelete(user)}>
                              <TrashIcon className="h-4 w-4 text-destructive" />
                            </Button>
                          </div>
                        </TableCell>
                      </TableRow>
                    ))}
                    {filteredUsers.length === 0 && !isLoading && (
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
        </div>
      </div>

      <FormDialog
        open={dialogOpen}
        onOpenChange={setDialogOpen}
        title={dialogMode === "add" ? (t.pages.users?.addUser || "添加用户") : (t.pages.users?.editUser || "编辑用户")}
        schema={userFormSchema}
        initialValues={getInitialValues()}
        onSubmit={handleSubmit}
        loading={isSubmitting}
      />

      <Dialog open={statusDialogOpen} onOpenChange={setStatusDialogOpen}>
        <DialogContent className="sm:max-w-[400px]">
          <DialogHeader>
            <DialogTitle>{t.pages.users?.changeStatus || "修改状态"}</DialogTitle>
            <DialogDescription>
              确定要{newStatus === 1 ? "启用" : "停用"}用户 "{currentStatusUser?.real_name}" 吗？
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button variant="outline" onClick={() => setStatusDialogOpen(false)}>
              {t.common.cancel}
            </Button>
            <Button onClick={handleStatusSubmit} disabled={isSubmitting}>
              {t.common.confirm}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </PageLayout>
  )
}
