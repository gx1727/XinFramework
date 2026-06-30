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
  ShieldCheckIcon,
  UsersIcon,
  RefreshCw,
  AlertTriangleIcon,
  CheckSquare,
  Square,
} from "lucide-react"
import { t } from "@/locales"
import {
  platformUserApi,
  platformRoleApi,
  type PlatformUserItem,
  type PlatformRoleItem,
  ApiError,
} from "@/api"
import { FormDialog } from "@/components/schema/DynamicForm"
import type { FormSchema } from "@/types/schema"
import { Checkbox } from "@/components/ui/checkbox"
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

const LS_KEY_USE_MOCK = "platformUsersPage.useMockFallback"

// ---- Mock 兜底（仅当用户主动勾选"使用 Mock"开关时使用）----
const mockUsers: PlatformUserItem[] = [
  { id: 1, account_id: 1, code: "super_admin", real_name: "超级管理员", nickname: "super", status: 1, roles: [{ id: 1, code: "super_admin", name: "超级管理员" }] },
  { id: 2, account_id: 2, code: "ops_admin", real_name: "运营管理员", nickname: "ops", status: 1, roles: [{ id: 2, code: "ops_admin", name: "运营管理员" }] },
  { id: 3, account_id: 3, code: "auditor", real_name: "审计员", status: 0, roles: [] },
]

export function PlatformUsersPage() {
  const [users, setUsers] = useState<PlatformUserItem[]>([])
  const [searchTerm, setSearchTerm] = useState("")
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [dataSource, setDataSource] = useState<"api" | "mock" | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  const [dialogOpen, setDialogOpen] = useState(false)
  const [dialogMode, setDialogMode] = useState<"add" | "edit">("add")
  const [currentUser, setCurrentUser] = useState<PlatformUserItem | null>(null)

  const [statusDialogOpen, setStatusDialogOpen] = useState(false)
  const [pendingStatus, setPendingStatus] = useState<{ user: PlatformUserItem; nextStatus: 0 | 1 } | null>(null)

  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [userToDelete, setUserToDelete] = useState<PlatformUserItem | null>(null)

  const [roleDialogOpen, setRoleDialogOpen] = useState(false)
  const [currentRoleUser, setCurrentRoleUser] = useState<PlatformUserItem | null>(null)
  const [allRoles, setAllRoles] = useState<PlatformRoleItem[]>([])
  const [selectedRoleIds, setSelectedRoleIds] = useState<number[]>([])
  const [roleLoading, setRoleLoading] = useState(false)

  // ---- Mock 兜底开关 ----
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
  const fetchUsers = useCallback(async () => {
    if (useMockFallback) {
      setUsers(mockUsers)
      setDataSource("mock")
      setError(null)
      return
    }
    setIsLoading(true)
    setError(null)
    try {
      const res = await platformUserApi.list({ page: 1, size: 200 })
      setUsers(res?.list ?? [])
      setDataSource("api")
    } catch (err) {
      const msg =
        err instanceof ApiError
          ? `${err.status} ${err.message}`
          : err instanceof Error
            ? err.message
            : "加载平台用户失败"
      console.error("[PlatformUsers] load failed:", err)
      setUsers([])
      setDataSource(null)
      setError(msg)
    } finally {
      setIsLoading(false)
    }
  }, [useMockFallback])

  useEffect(() => {
    fetchUsers()
  }, [fetchUsers])

  // ---- Filter ----
  const filteredUsers = useMemo(() => {
    if (!searchTerm.trim()) return users
    const kw = searchTerm.toLowerCase()
    return users.filter(
      (u) =>
        u.code.toLowerCase().includes(kw) ||
        u.real_name.toLowerCase().includes(kw) ||
        String(u.account_id).includes(kw),
    )
  }, [users, searchTerm])

  const activeCount = useMemo(() => users.filter((u) => u.status === 1).length, [users])
  const disabledCount = users.length - activeCount

  // ---- Form schema ----
  const userFormSchema: FormSchema = useMemo(() => {
    const items: FormSchema["items"] = [
      {
        field: "account_id",
        label: t.pages.platformUsers?.accountId || "账户 ID",
        type: "number",
        required: true,
        placeholder: "请输入登录账号的 account_id",
        disabled: dialogMode === "edit",
        tooltip: "对应 accounts.id；同一 account 已绑定则后端会报错",
      },
      {
        field: "code",
        label: t.pages.platformUsers?.code || "用户代码",
        type: "text",
        required: true,
        placeholder: "请输入用户代码（平台域内唯一）",
        disabled: dialogMode === "edit",
      },
      {
        field: "real_name",
        label: t.pages.platformUsers?.realName || "真实姓名",
        type: "text",
        required: true,
        placeholder: "请输入真实姓名",
      },
      {
        field: "nickname",
        label: t.pages.platformUsers?.nickname || "昵称",
        type: "text",
        placeholder: "请输入昵称（可选）",
      },
      {
        field: "status",
        label: t.pages.platformUsers?.status || "状态",
        type: "radio",
        defaultValue: 1,
        options: [
          { label: t.common.enable || "启用", value: 1 },
          { label: t.common.disable || "停用", value: 0 },
        ],
      },
    ]
    return { items }
  }, [dialogMode])

  // ---- Handlers ----
  const handleAdd = () => {
    setDialogMode("add")
    setCurrentUser(null)
    setDialogOpen(true)
  }

  const handleEdit = (user: PlatformUserItem) => {
    setDialogMode("edit")
    setCurrentUser(user)
    setDialogOpen(true)
  }

  const handleDeleteConfirm = (user: PlatformUserItem) => {
    setUserToDelete(user)
    setDeleteDialogOpen(true)
  }

  const handleDelete = async () => {
    if (!userToDelete) return
    setIsSubmitting(true)
    try {
      if (useMockFallback) {
        setUsers((prev) => prev.filter((u) => u.id !== userToDelete.id))
        toast.success("已删除（Mock）")
      } else {
        await platformUserApi.delete(userToDelete.id)
        toast.success("删除成功")
      }
      setDeleteDialogOpen(false)
      if (!useMockFallback) await fetchUsers()
    } catch (err) {
      const msg = err instanceof Error ? err.message : "删除失败"
      toast.error(msg)
      console.error("[PlatformUsers] delete failed:", err)
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleStatusChange = (user: PlatformUserItem) => {
    const next: 0 | 1 = user.status === 1 ? 0 : 1
    setPendingStatus({ user, nextStatus: next })
    setStatusDialogOpen(true)
  }

  const handleStatusSubmit = async () => {
    if (!pendingStatus) return
    setIsSubmitting(true)
    try {
      if (useMockFallback) {
        setUsers((prev) =>
          prev.map((u) => (u.id === pendingStatus.user.id ? { ...u, status: pendingStatus.nextStatus } : u)),
        )
        toast.success("已更新状态（Mock）")
      } else {
        await platformUserApi.updateStatus(pendingStatus.user.id, pendingStatus.nextStatus)
        toast.success("状态更新成功")
      }
      setStatusDialogOpen(false)
      if (!useMockFallback) await fetchUsers()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "状态更新失败")
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleSubmit = async (values: Record<string, unknown>) => {
    setIsSubmitting(true)
    try {
      const statusNum =
        typeof values.status === "string" ? parseInt(values.status as string, 10) : Number(values.status)
      if (dialogMode === "add") {
        if (useMockFallback) {
          const newId = Math.max(0, ...users.map((u) => u.id)) + 1
          setUsers((prev) => [
            ...prev,
            {
              id: newId,
              account_id: Number(values.account_id),
              code: String(values.code ?? ""),
              real_name: String(values.real_name ?? ""),
              nickname: (values.nickname as string) || "",
              status: statusNum || 1,
              roles: [],
            },
          ])
          toast.success("已新增（Mock）")
        } else {
          await platformUserApi.create({
            account_id: Number(values.account_id),
            code: String(values.code ?? ""),
            real_name: String(values.real_name ?? ""),
            nickname: (values.nickname as string) || undefined,
            status: statusNum || 1,
          })
          toast.success("创建成功")
        }
      } else if (currentUser) {
        if (useMockFallback) {
          setUsers((prev) =>
            prev.map((u) =>
              u.id === currentUser.id
                ? {
                    ...u,
                    real_name: String(values.real_name ?? u.real_name),
                    nickname: (values.nickname as string) || u.nickname,
                    status: statusNum || u.status,
                  }
                : u,
            ),
          )
          toast.success("已更新（Mock）")
        } else {
          await platformUserApi.update(currentUser.id, {
            real_name: String(values.real_name ?? ""),
            nickname: (values.nickname as string) || undefined,
            status: statusNum || 1,
          })
          toast.success("更新成功")
        }
      }
      setDialogOpen(false)
      if (!useMockFallback) await fetchUsers()
    } catch (err) {
      const msg =
        err instanceof ApiError
          ? err.status === 409
            ? "该 account 已绑定其他平台用户"
            : err.message
          : err instanceof Error
            ? err.message
            : "保存失败"
      toast.error(msg)
      console.error("[PlatformUsers] save failed:", err)
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleRoleAssign = async (user: PlatformUserItem) => {
    setCurrentRoleUser(user)
    setSelectedRoleIds((user.roles ?? []).map((r) => r.id))
    setRoleDialogOpen(true)
    setRoleLoading(true)
    try {
      const res = await platformRoleApi.list({ page: 1, size: 200 })
      setAllRoles(res?.list ?? [])
    } catch (err) {
      console.error("[PlatformUsers] load roles failed:", err)
      toast.error("加载平台角色失败")
      setAllRoles([])
    } finally {
      setRoleLoading(false)
    }
  }

  const handleRoleSubmit = async () => {
    if (!currentRoleUser) return
    setIsSubmitting(true)
    try {
      await platformUserApi.assignRoles(currentRoleUser.id, selectedRoleIds)
      toast.success("角色分配成功")
      setRoleDialogOpen(false)
      await fetchUsers()
    } catch (err) {
      toast.error(err instanceof Error ? err.message : "角色分配失败")
    } finally {
      setIsSubmitting(false)
    }
  }

  const getInitialValues = () => {
    if (currentUser) {
      return {
        account_id: currentUser.account_id,
        code: currentUser.code,
        real_name: currentUser.real_name,
        nickname: currentUser.nickname || "",
        status: currentUser.status,
      }
    }
    return { status: 1 }
  }

  const handleSelectAllRoles = () => setSelectedRoleIds(allRoles.map((r) => r.id))
  const handleDeselectAllRoles = () => setSelectedRoleIds([])
  const handleRoleToggle = (roleId: number, checked: boolean) => {
    setSelectedRoleIds((prev) =>
      checked ? (prev.includes(roleId) ? prev : [...prev, roleId]) : prev.filter((id) => id !== roleId),
    )
  }

  return (
    <PageLayout>
      <div className="px-4 lg:px-6">
        <div className="flex items-center justify-between mb-6">
          <div>
            <h1 className="text-2xl font-bold flex items-center gap-2">
              <ShieldCheckIcon className="h-6 w-6" />
              {t.pages.platformUsers?.title || "平台用户"}
            </h1>
            <p className="text-sm text-muted-foreground mt-1">
              {t.pages.platformUsers?.subtitle || "管理平台域用户（仅 super_admin 可访问）"}
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
            <Button variant="outline" size="sm" onClick={fetchUsers} disabled={isLoading}>
              <RefreshCw className={cn("h-4 w-4 mr-2", isLoading && "animate-spin")} />
              {t.pages.platformUsers?.refresh || "刷新"}
            </Button>
            <Button size="sm" onClick={handleAdd}>
              <PlusIcon className="h-4 w-4 mr-2" />
              {t.pages.platformUsers?.create || "新建平台用户"}
            </Button>
          </div>
        </div>

        {/* Stats cards */}
        <div className="grid grid-cols-3 gap-4 mb-4">
          <Card>
            <CardHeader className="pb-2">
              <CardDescription>{t.pages.platformUsers?.statsTotal || "平台用户总数"}</CardDescription>
              <CardTitle className="text-2xl">{users.length}</CardTitle>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardDescription>{t.pages.platformUsers?.statsActive || "启用中"}</CardDescription>
              <CardTitle className="text-2xl text-green-600">{activeCount}</CardTitle>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardDescription>{t.pages.platformUsers?.statsDisabled || "已停用"}</CardDescription>
              <CardTitle className="text-2xl text-gray-500">{disabledCount}</CardTitle>
            </CardHeader>
          </Card>
        </div>

        {/* Error banner */}
        {error && (
          <div className="mb-4 flex items-start gap-2 rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-sm text-destructive">
            <AlertTriangleIcon className="h-4 w-4 mt-0.5 shrink-0" />
            <div className="flex-1">
              <div className="font-medium">加载失败</div>
              <div className="text-xs opacity-80">{error}</div>
            </div>
            <Button size="sm" variant="outline" onClick={fetchUsers}>
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
                  placeholder={t.pages.platformUsers?.searchPlaceholder || "搜索 code / 姓名 / 账户 ID..."}
                  className="pl-9"
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                />
              </div>
              <Badge variant="secondary">
                {t.pages.platformUsers?.matchedCount?.replace("{n}", String(filteredUsers.length)) ||
                  `共 ${filteredUsers.length} 条`}
              </Badge>
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
                  <TableHead>{t.pages.platformUsers?.accountId || "账户 ID"}</TableHead>
                  <TableHead>{t.pages.platformUsers?.code || "代码"}</TableHead>
                  <TableHead>{t.pages.platformUsers?.realName || "姓名"}</TableHead>
                  <TableHead>{t.pages.platformUsers?.nickname || "昵称"}</TableHead>
                  <TableHead>角色</TableHead>
                  <TableHead>{t.pages.platformUsers?.status || "状态"}</TableHead>
                  <TableHead className="text-right">{t.common.edit}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredUsers.length === 0 ? (
                  <TableRow>
                    <TableCell colSpan={8} className="text-center py-8 text-muted-foreground">
                      {isLoading ? t.common.loading : t.common.noData}
                    </TableCell>
                  </TableRow>
                ) : (
                  filteredUsers.map((user) => (
                    <TableRow key={user.id}>
                      <TableCell className="font-mono text-xs text-muted-foreground">{user.id}</TableCell>
                      <TableCell className="font-mono text-sm">{user.account_id}</TableCell>
                      <TableCell>
                        <code className="px-1.5 py-0.5 rounded bg-muted text-xs font-mono">{user.code}</code>
                      </TableCell>
                      <TableCell className="font-medium">{user.real_name}</TableCell>
                      <TableCell className="text-sm text-muted-foreground">{user.nickname || "-"}</TableCell>
                      <TableCell>
                        <div className="flex flex-wrap gap-1 max-w-[240px]">
                          {(user.roles ?? []).length === 0 ? (
                            <span className="text-xs text-muted-foreground/60">-</span>
                          ) : (
                            (user.roles ?? []).map((r) => (
                              <Badge key={r.id} variant="outline" className="text-[10px]">
                                {r.name}
                              </Badge>
                            ))
                          )}
                        </div>
                      </TableCell>
                      <TableCell>
                        <button
                          onClick={() => handleStatusChange(user)}
                          className="cursor-pointer"
                          title="点击切换状态"
                        >
                          <Badge variant={user.status === 1 ? "default" : "secondary"}>
                            {user.status === 1 ? "启用" : "停用"}
                          </Badge>
                        </button>
                      </TableCell>
                      <TableCell className="text-right">
                        <div className="flex items-center justify-end gap-1">
                          <Button
                            variant="ghost"
                            size="sm"
                            onClick={() => handleRoleAssign(user)}
                            title="分配角色"
                          >
                            <UsersIcon className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-8 w-8"
                            onClick={() => handleEdit(user)}
                          >
                            <EditIcon className="h-4 w-4" />
                          </Button>
                          <Button
                            variant="ghost"
                            size="icon"
                            className="h-8 w-8"
                            onClick={() => handleDeleteConfirm(user)}
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
            ? t.pages.platformUsers?.create || "新建平台用户"
            : t.pages.platformUsers?.edit || "编辑平台用户"
        }
        width={520}
        schema={userFormSchema}
        initialValues={getInitialValues()}
        onSubmit={handleSubmit}
        loading={isSubmitting}
      />

      <Dialog open={statusDialogOpen} onOpenChange={setStatusDialogOpen}>
        <DialogContent className="sm:max-w-[400px]">
          <DialogHeader>
            <DialogTitle>修改状态</DialogTitle>
            <DialogDescription>
              确定要{pendingStatus?.nextStatus === 1 ? "启用" : "停用"}用户 "{pendingStatus?.user.real_name}" 吗？
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

      <Dialog open={deleteDialogOpen} onOpenChange={setDeleteDialogOpen}>
        <DialogContent className="sm:max-w-[400px]">
          <DialogHeader>
            <DialogTitle>删除平台用户</DialogTitle>
            <DialogDescription>
              确定要删除平台用户 "{userToDelete?.real_name}" 吗？删除不影响对应的登录账号。
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

      <Dialog open={roleDialogOpen} onOpenChange={setRoleDialogOpen}>
        <DialogContent className="sm:max-w-[560px]">
          <DialogHeader>
            <DialogTitle>
              分配平台角色 - {currentRoleUser?.real_name}
            </DialogTitle>
            <DialogDescription>
              为该平台用户分配 sys_role（可多选；将覆盖当前分配）。
            </DialogDescription>
          </DialogHeader>
          <div className="max-h-[50vh] overflow-y-auto py-2">
            <div className="flex items-center justify-between mb-3 bg-muted/50 p-2 rounded-md border">
              <span className="text-sm font-semibold">
                可选平台角色（共 {allRoles.length} 个，已选 {selectedRoleIds.length}）
              </span>
              <div className="flex items-center gap-1">
                <Button variant="outline" size="sm" onClick={handleSelectAllRoles} className="h-7 text-xs">
                  <CheckSquare className="w-3.5 h-3.5 mr-1" />
                  全选
                </Button>
                <Button variant="outline" size="sm" onClick={handleDeselectAllRoles} className="h-7 text-xs">
                  <Square className="w-3.5 h-3.5 mr-1" />
                  全不选
                </Button>
              </div>
            </div>
            {roleLoading ? (
              <p className="text-sm text-muted-foreground py-4">{t.common.loading}</p>
            ) : allRoles.length === 0 ? (
              <p className="text-sm text-muted-foreground py-4">暂无平台角色，请先在「平台角色」页创建。</p>
            ) : (
              <div className="space-y-1 border rounded-md p-3 bg-card">
                {allRoles.map((role) => (
                  <div key={role.id} className="flex items-center gap-2 py-1">
                    <Checkbox
                      id={`role-${role.id}`}
                      checked={selectedRoleIds.includes(role.id)}
                      onCheckedChange={(c) => handleRoleToggle(role.id, c === true)}
                    />
                    <label
                      htmlFor={`role-${role.id}`}
                      className="text-sm cursor-pointer flex items-center gap-2 flex-1"
                    >
                      <span>{role.name}</span>
                      <code className="px-1 rounded bg-muted text-[10px] font-mono">{role.code}</code>
                      {role.is_default && (
                        <Badge variant="outline" className="text-[10px] h-4 px-1 py-0">默认</Badge>
                      )}
                      {role.status === 0 && (
                        <Badge variant="secondary" className="text-[10px] h-4 px-1 py-0">停用</Badge>
                      )}
                    </label>
                  </div>
                ))}
              </div>
            )}
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setRoleDialogOpen(false)}>
              {t.common.cancel}
            </Button>
            <Button onClick={handleRoleSubmit} disabled={isSubmitting}>
              {t.common.save}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    </PageLayout>
  )
}