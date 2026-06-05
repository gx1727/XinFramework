import { useEffect, useState, useCallback, useMemo } from "react"
import { PageLayout } from "@/components/page-layout"
import { Card, CardContent, CardHeader, CardTitle } from "@/components/ui/card"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { Badge } from "@/components/ui/badge"
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "@/components/ui/table"
import { PlusIcon, SearchIcon, EditIcon, TrashIcon, UsersIcon, RefreshCw } from "lucide-react"
import { useTranslation } from "@/locales"
import { userApi, roleApi, type UserItem, type RoleItem } from "@/api"
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

export function UsersPage() {
  const t = useTranslation()
  const [users, setUsers] = useState<UserItem[]>([])
  const [total, setTotal] = useState(0)
  const [isLoading, setIsLoading] = useState(true)
  const [searchTerm, setSearchTerm] = useState("")
  const [isSubmitting, setIsSubmitting] = useState(false)

  const [dialogOpen, setDialogOpen] = useState(false)
  const [dialogMode, setDialogMode] = useState<"add" | "edit">("add")
  const [currentUser, setCurrentUser] = useState<UserItem | null>(null)

  const [roleOptions, setRoleOptions] = useState<RoleItem[]>([])
  const [statusDialogOpen, setStatusDialogOpen] = useState(false)
  const [currentStatusUser, setCurrentStatusUser] = useState<UserItem | null>(null)
  const [newStatus, setNewStatus] = useState<number>(1)

  const fetchUsers = useCallback(async () => {
    setIsLoading(true)
    try {
      const response = await userApi.list({ keyword: searchTerm || undefined, page: 1, size: 100 })
      // 兼容后端将 status 序列化为字符串的情况，统一归一为 number
      const list = (response.list || []).map((u) => ({
        ...u,
        status: Number(u.status),
      })) as UserItem[]
      setUsers(list)
      setTotal(response.total)
    } catch {
      const list = mockUsers.map((u) => ({ ...u, status: Number(u.status) }))
      setUsers(list)
      setTotal(mockUsers.length)
    } finally {
      setIsLoading(false)
    }
  }, [searchTerm])

  const fetchRoles = useCallback(async () => {
    try {
      const response = await roleApi.list({ page: 1, size: 100 })
      setRoleOptions(response.list)
    } catch {
      setRoleOptions([])
    }
  }, [])

  useEffect(() => {
    fetchUsers()
  }, [fetchUsers])

  useEffect(() => {
    fetchRoles()
  }, [fetchRoles])

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
        rules: [
          { minLength: 6, message: "密码长度至少 6 位" },
        ],
      })
    }

    items.push(
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
          { label: "启用", value: 1 },
          { label: "停用", value: 2 },
        ],
      }
    )

    return { items }
  }, [t, dialogMode])

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
        await userApi.create(createPayload)
      } else if (currentUser) {
        // 编辑：只提交可变字段，username/code 为登录账号不可改
        const updatePayload: Partial<UserItem> = {
          real_name: String(values.real_name ?? ""),
          phone: (values.phone as string) || "",
          email: (values.email as string) || "",
          status: statusNum,
        }
        await userApi.update(currentUser.id, updatePayload)
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
        status: currentUser.status,
      }
    }
    return { status: 1 }
  }

  const getStatusBadge = (status: number | string | undefined | null) => {
    // 兼容后端将 status 序列化为数字或字符串的情况
    if (Number(status) === 1) {
      return <Badge variant="default">{t.pages.users?.active || "活跃"}</Badge>
    }
    return <Badge variant="secondary">{t.pages.users?.inactive || "停用"}</Badge>
  }

  const activeCount = users.filter(u => Number(u.status) === 1).length

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

        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-3 mb-6">
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">总用户数</CardTitle>
              <UsersIcon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{total}</div>
              <p className="text-xs text-muted-foreground">注册用户总数</p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">活跃用户</CardTitle>
              <UsersIcon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{activeCount}</div>
              <p className="text-xs text-muted-foreground">{total > 0 ? ((activeCount / total) * 100).toFixed(1) : 0}% 活跃率</p>
            </CardContent>
          </Card>
          <Card>
            <CardHeader className="flex flex-row items-center justify-between pb-2">
              <CardTitle className="text-sm font-medium">停用用户</CardTitle>
              <UsersIcon className="h-4 w-4 text-muted-foreground" />
            </CardHeader>
            <CardContent>
              <div className="text-2xl font-bold">{total - activeCount}</div>
              <p className="text-xs text-muted-foreground">已停用账户</p>
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
              <Badge variant="secondary">共 {total} 个用户</Badge>
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
                  <TableHead>{t.pages.users?.status || "状态"}</TableHead>
                  <TableHead className="text-right">{t.common.edit}</TableHead>
                </TableRow>
              </TableHeader>
              <TableBody>
                {users.map((user) => (
                  <TableRow key={user.id}>
                    <TableCell className="font-medium">{user.id}</TableCell>
                    <TableCell>{user.real_name}</TableCell>
                    <TableCell className="font-mono text-sm">{user.code}</TableCell>
                    <TableCell>{user.email || "-"}</TableCell>
                    <TableCell>{user.phone || "-"}</TableCell>
                    <TableCell>
                      <Badge variant="outline">{user.role}</Badge>
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
                {users.length === 0 && !isLoading && (
                  <TableRow>
                    <TableCell colSpan={8} className="text-center py-8 text-muted-foreground">
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