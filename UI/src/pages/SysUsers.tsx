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
  ShieldCheckIcon,
  UsersIcon,
  RefreshCw,
  AlertTriangleIcon,
  CheckSquare,
  Square,
} from "lucide-react"
import { t } from "@/locales"
import {
  sysUserApi,
  sysRoleApi,
  type SysUserItem,
  type SysRoleItem,
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

export function SysUsersPage() {
  const [users, setUsers] = useState<SysUserItem[]>([])
  const [searchTerm, setSearchTerm] = useState("")
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [isSubmitting, setIsSubmitting] = useState(false)

  const [dialogOpen, setDialogOpen] = useState(false)
  const [dialogMode, setDialogMode] = useState<"add" | "edit">("add")
  const [currentUser, setCurrentUser] = useState<SysUserItem | null>(null)

  const [statusDialogOpen, setStatusDialogOpen] = useState(false)
  const [pendingStatus, setPendingStatus] = useState<{
    user: SysUserItem
    nextStatus: 0 | 1
  } | null>(null)

  const [deleteDialogOpen, setDeleteDialogOpen] = useState(false)
  const [userToDelete, setUserToDelete] = useState<SysUserItem | null>(
    null
  )

  const [roleDialogOpen, setRoleDialogOpen] = useState(false)
  const [currentRoleUser, setCurrentRoleUser] =
    useState<SysUserItem | null>(null)
  const [allRoles, setAllRoles] = useState<SysRoleItem[]>([])
  const [selectedRoleIds, setSelectedRoleIds] = useState<number[]>([])
  const [roleLoading, setRoleLoading] = useState(false)

  // ---- Fetch ----
  const fetchUsers = useCallback(async () => {
    setIsLoading(true)
    setError(null)
    try {
      const res = await sysUserApi.list({ page: 1, size: 200 })
      setUsers(res?.list ?? [])
    } catch (err) {
      const msg =
        err instanceof ApiError
          ? `${err.status} ${err.message}`
          : err instanceof Error
            ? err.message
            : "加载 Sys 用户失败"
      console.error("[SysUsers] load failed:", err)
      setUsers([])
      setError(msg)
    } finally {
      setIsLoading(false)
    }
  }, [])

  useEffect(() => {
    // eslint-disable-next-line react-hooks/set-state-in-effect -- 首次加载触发请求是约定写法，cascading render 可接受
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
        String(u.account_id).includes(kw)
    )
  }, [users, searchTerm])

  const activeCount = useMemo(
    () => users.filter((u) => u.status === 1).length,
    [users]
  )
  const disabledCount = users.length - activeCount

  // ---- Form schema ----
  const userFormSchema: FormSchema = useMemo(() => {
    const items: FormSchema["items"] = []

    if (dialogMode === "add") {
      // ---- 模式 2：一并新建可登录账号 ----
      // 物理上是一个表单，但由分隔条划成两块：账号信息 + Sys 身份。
      items.push(
        {
          field: "section_account",
          label:
            t.pages.sysUsers?.sectionAccount || "账号信息（首次登录使用）",
          type: "divider",
        },
        {
          field: "phone",
          label: t.pages.sysUsers?.phone || "手机号",
          type: "text",
          required: true,
          placeholder:
            t.pages.sysUsers?.phonePlaceholder ||
            "请输入手机号（作为登录账号）",
        },
        {
          field: "username",
          label: t.pages.sysUsers?.username || "用户名",
          type: "text",
          placeholder:
            t.pages.sysUsers?.usernamePlaceholder ||
            "可选，留空默认同手机号",
        },
        {
          field: "email",
          label: t.pages.sysUsers?.email || "邮箱",
          type: "email",
          placeholder:
            t.pages.sysUsers?.emailPlaceholder || "可选，账号找回用",
        },
        {
          field: "password",
          label: t.pages.sysUsers?.password || "初始密码",
          type: "password",
          required: true,
          placeholder:
            t.pages.sysUsers?.passwordPlaceholder || "6-32 位，区分大小写",
          rules: [
            { minLength: 6, maxLength: 32, message: "密码长度需在 6-32 之间" },
          ],
        }
      )
    }

    // ---- Sys 身份区（两种模式都需要） ----
    // account_id 和 code 是系统字段：新建时后端自动生成（phone → accounts.id → "u<id>"），
    // 编辑时不可改（只读展示在表单上方），所以这里不再列为 form field，避免被误输入。
    items.push(
      {
        field: "section_identity",
        label: t.pages.sysUsers?.sectionIdentity || "Sys 身份信息",
        type: "divider",
      },
      {
        field: "real_name",
        label: t.pages.sysUsers?.realName || "真实姓名",
        type: "text",
        required: true,
        placeholder: "请输入真实姓名",
      },
      {
        field: "nickname",
        label: t.pages.sysUsers?.nickname || "昵称",
        type: "text",
        placeholder: "请输入昵称（可选）",
      },
      {
        field: "status",
        label: t.pages.sysUsers?.status || "状态",
        type: "radio",
        defaultValue: 1,
        options: [
          { label: t.common.enable || "启用", value: 1 },
          { label: t.common.disable || "停用", value: 0 },
        ],
      }
    )

    return { items }
  }, [dialogMode])

  // ---- Handlers ----
  const handleAdd = () => {
    setDialogMode("add")
    setCurrentUser(null)
    setDialogOpen(true)
  }

  const handleEdit = (user: SysUserItem) => {
    setDialogMode("edit")
    setCurrentUser(user)
    setDialogOpen(true)
  }

  const handleDeleteConfirm = (user: SysUserItem) => {
    setUserToDelete(user)
    setDeleteDialogOpen(true)
  }

  const handleDelete = async () => {
    if (!userToDelete) return
    setIsSubmitting(true)
    try {
      await sysUserApi.delete(userToDelete.id)
      toast.success("删除成功")
      setDeleteDialogOpen(false)
      await fetchUsers()
    } catch (err) {
      const msg = err instanceof Error ? err.message : "删除失败"
      toast.error(msg)
      console.error("[SysUsers] delete failed:", err)
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleStatusChange = (user: SysUserItem) => {
    const next: 0 | 1 = user.status === 1 ? 0 : 1
    setPendingStatus({ user, nextStatus: next })
    setStatusDialogOpen(true)
  }

  const handleStatusSubmit = async () => {
    if (!pendingStatus) return
    setIsSubmitting(true)
    try {
      await sysUserApi.updateStatus(
        pendingStatus.user.id,
        pendingStatus.nextStatus
      )
      toast.success("状态更新成功")
      setStatusDialogOpen(false)
      await fetchUsers()
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
        typeof values.status === "string"
          ? parseInt(values.status as string, 10)
          : Number(values.status)
      if (dialogMode === "add") {
        // 模式 2：AccountID 不传，由后端在同事务内创建 accounts 行 + sys_user 行
        // code/account_id 是系统字段，不由前端提交。
        await sysUserApi.create({
          phone: String(values.phone ?? ""),
          username: (values.username as string) || undefined,
          email: (values.email as string) || undefined,
          password: String(values.password ?? ""),
          real_name: String(values.real_name ?? ""),
          nickname: (values.nickname as string) || undefined,
          status: statusNum || 1,
        })
        toast.success("创建成功")
      } else if (currentUser) {
        await sysUserApi.update(currentUser.id, {
          real_name: String(values.real_name ?? ""),
          nickname: (values.nickname as string) || undefined,
          status: statusNum || 1,
        })
        toast.success("更新成功")
      }
      setDialogOpen(false)
      await fetchUsers()
    } catch (err) {
      const msg =
        err instanceof ApiError
          ? err.status === 409
            ? "该 account 已绑定其他 Sys 用户"
            : err.message
          : err instanceof Error
            ? err.message
            : "保存失败"
      toast.error(msg)
      console.error("[SysUsers] save failed:", err)
    } finally {
      setIsSubmitting(false)
    }
  }

  const handleRoleAssign = async (user: SysUserItem) => {
    setCurrentRoleUser(user)
    setSelectedRoleIds((user.roles ?? []).map((r) => r.id))
    setRoleDialogOpen(true)
    setRoleLoading(true)
    try {
      const res = await sysRoleApi.list({ page: 1, size: 200 })
      setAllRoles(res?.list ?? [])
    } catch (err) {
      console.error("[SysUsers] load roles failed:", err)
      toast.error("加载 Sys 角色失败")
      setAllRoles([])
    } finally {
      setRoleLoading(false)
    }
  }

  const handleRoleSubmit = async () => {
    if (!currentRoleUser) return
    setIsSubmitting(true)
    try {
      await sysUserApi.assignRoles(currentRoleUser.id, selectedRoleIds)
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
      // code / account_id 不进表单（只读展示在 dialog 顶部）
      return {
        real_name: currentUser.real_name,
        nickname: currentUser.nickname || "",
        status: currentUser.status,
      }
    }
    return { status: 1 }
  }

  const handleSelectAllRoles = () =>
    setSelectedRoleIds(allRoles.map((r) => r.id))
  const handleDeselectAllRoles = () => setSelectedRoleIds([])
  const handleRoleToggle = (roleId: number, checked: boolean) => {
    setSelectedRoleIds((prev) =>
      checked
        ? prev.includes(roleId)
          ? prev
          : [...prev, roleId]
        : prev.filter((id) => id !== roleId)
    )
  }

  return (
    <PageLayout>
      <div className="px-4 lg:px-6">
        <div className="mb-6 flex items-center justify-between">
          <div>
            <h1 className="flex items-center gap-2 text-2xl font-bold">
              <ShieldCheckIcon className="h-6 w-6" />
              {t.pages.sysUsers?.title || "Sys 用户"}
            </h1>
            <p className="mt-1 text-sm text-muted-foreground">
              {t.pages.sysUsers?.subtitle ||
                "管理 sys 域用户"}
            </p>
          </div>
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={fetchUsers}
              disabled={isLoading}
            >
              <RefreshCw
                className={cn("mr-2 h-4 w-4", isLoading && "animate-spin")}
              />
              {t.pages.sysUsers?.refresh || "刷新"}
            </Button>
            <Button size="sm" onClick={handleAdd}>
              <PlusIcon className="mr-2 h-4 w-4" />
              {t.pages.sysUsers?.create || "新建 Sys 用户"}
            </Button>
          </div>
        </div>

        {/* Stats cards */}
        <div className="mb-4 grid grid-cols-3 gap-4">
          <Card>
            <CardHeader className="pb-2">
              <CardDescription>
                {t.pages.sysUsers?.statsTotal || "Sys 用户总数"}
              </CardDescription>
              <CardTitle className="text-2xl">{users.length}</CardTitle>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardDescription>
                {t.pages.sysUsers?.statsActive || "启用中"}
              </CardDescription>
              <CardTitle className="text-2xl text-green-600">
                {activeCount}
              </CardTitle>
            </CardHeader>
          </Card>
          <Card>
            <CardHeader className="pb-2">
              <CardDescription>
                {t.pages.sysUsers?.statsDisabled || "已停用"}
              </CardDescription>
              <CardTitle className="text-2xl text-gray-500">
                {disabledCount}
              </CardTitle>
            </CardHeader>
          </Card>
        </div>

        {/* Error banner */}
        {error && (
          <div className="mb-4 flex items-start gap-2 rounded-md border border-destructive/40 bg-destructive/10 px-3 py-2 text-sm text-destructive">
            <AlertTriangleIcon className="mt-0.5 h-4 w-4 shrink-0" />
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
              <div className="relative max-w-sm flex-1">
                <SearchIcon className="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <Input
                  placeholder={
                    t.pages.sysUsers?.searchPlaceholder ||
                    "搜索 code / 姓名 / 账户 ID..."
                  }
                  className="pl-9"
                  value={searchTerm}
                  onChange={(e) => setSearchTerm(e.target.value)}
                />
              </div>
              <Badge variant="secondary">
                {t.pages.sysUsers?.matchedCount?.replace(
                  "{n}",
                  String(filteredUsers.length)
                ) || `共 ${filteredUsers.length} 条`}
              </Badge>
            </div>
          </CardHeader>
          <CardContent>
            <Table>
              <TableHeader>
                <TableRow>
                  <TableHead className="w-[60px]">ID</TableHead>
                  <TableHead>
                    {t.pages.sysUsers?.accountId || "账户 ID"}
                  </TableHead>
                  <TableHead>{t.pages.sysUsers?.code || "代码"}</TableHead>
                  <TableHead>
                    {t.pages.sysUsers?.realName || "姓名"}
                  </TableHead>
                  <TableHead>
                    {t.pages.sysUsers?.nickname || "昵称"}
                  </TableHead>
                  <TableHead>角色</TableHead>
                  <TableHead>
                    {t.pages.sysUsers?.status || "状态"}
                  </TableHead>
                  <TableHead className="text-right">{t.common.edit}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {filteredUsers.length === 0 ? (
                  <TableRow>
                    <TableCell
                      colSpan={8}
                      className="py-8 text-center text-muted-foreground"
                    >
                      {isLoading ? t.common.loading : t.common.noData}
                    </TableCell>
                  </TableRow>
                ) : (
                  filteredUsers.map((user) => (
                    <TableRow key={user.id}>
                      <TableCell className="font-mono text-xs text-muted-foreground">
                        {user.id}
                      </TableCell>
                      <TableCell className="font-mono text-sm">
                        {user.account_id}
                      </TableCell>
                      <TableCell>
                        <code className="rounded bg-muted px-1.5 py-0.5 font-mono text-xs">
                          {user.code}
                        </code>
                      </TableCell>
                      <TableCell className="font-medium">
                        {user.real_name}
                      </TableCell>
                      <TableCell className="text-sm text-muted-foreground">
                        {user.nickname || "-"}
                      </TableCell>
                      <TableCell>
                        <div className="flex max-w-[240px] flex-wrap gap-1">
                          {(user.roles ?? []).length === 0 ? (
                            <span className="text-xs text-muted-foreground/60">
                              -
                            </span>
                          ) : (
                            (user.roles ?? []).map((r) => (
                              <Badge
                                key={r.id}
                                variant="outline"
                                className="text-[10px]"
                              >
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
                          <Badge
                            variant={
                              user.status === 1 ? "default" : "secondary"
                            }
                          >
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
            ? t.pages.sysUsers?.create || "新建 Sys 用户"
            : t.pages.sysUsers?.edit || "编辑 Sys 用户"
        }
        // 只读展示系统字段（account_id / code），不在表单里让用户输
        headerExtra={
          <div className="mt-2 space-y-1 rounded-md border bg-muted/40 p-2 text-xs text-muted-foreground">
            {dialogMode === "edit" && currentUser ? (
              <>
                <div className="flex items-center gap-2">
                  <span className="font-semibold">
                    {t.pages.sysUsers?.accountId || "账户 ID"}:
                  </span>
                  <code className="rounded bg-background px-1 py-0.5 font-mono">
                    {currentUser.account_id}
                  </code>
                  <span className="font-semibold">
                    {t.pages.sysUsers?.code || "用户代码"}:
                  </span>
                  <code className="rounded bg-background px-1 py-0.5 font-mono">
                    {currentUser.code}
                  </code>
                </div>
                <p>
                  系统自动生成，保存后不可修改。如需更换，请删除后重新创建。
                </p>
              </>
            ) : (
              <p>
                账户 ID 与用户代码将在创建后由系统自动生成（格式： code ={" "}
                <code className="rounded bg-background px-1 font-mono">
                  u&lt;account_id&gt;
                </code>
                ）。
              </p>
            )}
          </div>
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
              确定要{pendingStatus?.nextStatus === 1 ? "启用" : "停用"}用户 "
              {pendingStatus?.user.real_name}" 吗？
            </DialogDescription>
          </DialogHeader>
          <DialogFooter>
            <Button
              variant="outline"
              onClick={() => setStatusDialogOpen(false)}
            >
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
            <DialogTitle>删除 Sys 用户</DialogTitle>
            <DialogDescription>
              确定要删除 Sys 用户 "{userToDelete?.real_name}"
              吗？删除不影响对应的登录账号。
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

      <Dialog open={roleDialogOpen} onOpenChange={setRoleDialogOpen}>
        <DialogContent className="sm:max-w-[560px]">
          <DialogHeader>
            <DialogTitle>
              分配 Sys 角色 - {currentRoleUser?.real_name}
            </DialogTitle>
            <DialogDescription>
              为该 Sys 用户分配 sys_role（可多选；将覆盖当前分配）。
            </DialogDescription>
          </DialogHeader>
          <div className="max-h-[50vh] overflow-y-auto py-2">
            <div className="mb-3 flex items-center justify-between rounded-md border bg-muted/50 p-2">
              <span className="text-sm font-semibold">
                可选 Sys 角色（共 {allRoles.length} 个，已选{" "}
                {selectedRoleIds.length}）
              </span>
              <div className="flex items-center gap-1">
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleSelectAllRoles}
                  className="h-7 text-xs"
                >
                  <CheckSquare className="mr-1 h-3.5 w-3.5" />
                  全选
                </Button>
                <Button
                  variant="outline"
                  size="sm"
                  onClick={handleDeselectAllRoles}
                  className="h-7 text-xs"
                >
                  <Square className="mr-1 h-3.5 w-3.5" />
                  全不选
                </Button>
              </div>
            </div>
            {roleLoading ? (
              <p className="py-4 text-sm text-muted-foreground">
                {t.common.loading}
              </p>
            ) : allRoles.length === 0 ? (
              <p className="py-4 text-sm text-muted-foreground">
                暂无 Sys 角色，请先在「Sys 角色」页创建。
              </p>
            ) : (
              <div className="space-y-1 rounded-md border bg-card p-3">
                {allRoles.map((role) => (
                  <div key={role.id} className="flex items-center gap-2 py-1">
                    <Checkbox
                      id={`role-${role.id}`}
                      checked={selectedRoleIds.includes(role.id)}
                      onCheckedChange={(c) =>
                        handleRoleToggle(role.id, c === true)
                      }
                    />
                    <label
                      htmlFor={`role-${role.id}`}
                      className="flex flex-1 cursor-pointer items-center gap-2 text-sm"
                    >
                      <span>{role.name}</span>
                      <code className="rounded bg-muted px-1 font-mono text-[10px]">
                        {role.code}
                      </code>
                      {role.is_default && (
                        <Badge
                          variant="outline"
                          className="h-4 px-1 py-0 text-[10px]"
                        >
                          默认
                        </Badge>
                      )}
                      {role.status === 0 && (
                        <Badge
                          variant="secondary"
                          className="h-4 px-1 py-0 text-[10px]"
                        >
                          停用
                        </Badge>
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
