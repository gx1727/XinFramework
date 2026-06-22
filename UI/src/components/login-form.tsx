import { useEffect, useState } from "react"
import { useNavigate, Link } from "react-router-dom"
import { cn } from "@/lib/utils"
import { Button } from "@/components/ui/button"
import {
  Field,
  FieldGroup,
  FieldLabel,
} from "@/components/ui/field"
import { Input } from "@/components/ui/input"
import { useAuthStore } from "@/stores/authStore"
import { t } from "@/locales"
import { IdentityPickerDialog } from "@/components/identity-picker-dialog"

type LoginMode = "tenant" | "platform"

export function LoginForm({
  mode,
  className,
  ...props
}: React.ComponentProps<"form"> & { mode: LoginMode }) {
  const navigate = useNavigate()
  const tenantLogin = useAuthStore((s) => s.tenantLogin)
  const platformLogin = useAuthStore((s) => s.platformLogin)
  const loginPrecheck = useAuthStore((s) => s.loginPrecheck)
  const selectTenant = useAuthStore((s) => s.selectTenant)
  const isLoading = useAuthStore((s) => s.isLoading)
  const error = useAuthStore((s) => s.error)
  const clearError = useAuthStore((s) => s.clearError)
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)
  const scope = useAuthStore((s) => s.scope)
  const availableIdentities = useAuthStore((s) => s.availableIdentities)
  const platformAvailable = useAuthStore((s) => s.platformAvailable)
  const availablePlatformRoles = useAuthStore((s) => s.availablePlatformRoles)
  const clearIdentities = useAuthStore((s) => s.clearIdentities)

  // 多身份账号登录：弹"选择身份"对话框
  const [pickerOpen, setPickerOpen] = useState(false)
  const [pendingAccount, setPendingAccount] = useState<{ account: string; password: string } | null>(null)

  useEffect(() => {
    if (isAuthenticated) {
      // 按 scope 跳转到对应 dashboard
      navigate(scope === "platform" ? "/platform/dashboard" : "/app/dashboard")
    }
  }, [isAuthenticated, scope, navigate])

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    clearError()
    clearIdentities()

    const formData = new FormData(e.currentTarget)
    const account = formData.get("account") as string
    const password = formData.get("password") as string

    if (mode === "platform") {
      // platform 模式：直接 platform-login（不 precheck）
      await platformLogin(account, password)
      return
    }

    // tenant 模式：precheck 智能分支（路径 B 多身份支持）
    const result = await loginPrecheck(account, password)
    if (!result) return  // precheck 失败，error 已在 store

    // 防御：后端返回的 tenant_identities 理论上是 []（不是 null），
    // 但加了保护避免后端契约变更导致前端崩。
    const identities = result.tenant_identities ?? []
    const tenantCount = identities.length
    const hasPlatform = result.platform_available

    if (tenantCount === 0 && hasPlatform) {
      // 纯平台账号
      await platformLogin(account, password)
      return
    }

    if (tenantCount === 1 && !hasPlatform) {
      // 单身份账号
      await tenantLogin(account, password, identities[0].tenant_id)
      return
    }

    // 多身份账号：弹"选择身份"对话框
    setPendingAccount({ account, password })
    setPickerOpen(true)
  }

  const handlePickerSelectTenant = async (tenantId: number) => {
    if (!pendingAccount) return
    setPickerOpen(false)
    await selectTenant(pendingAccount.account, pendingAccount.password, tenantId)
    setPendingAccount(null)
  }

  const handlePickerSelectPlatform = async () => {
    if (!pendingAccount) return
    setPickerOpen(false)
    await platformLogin(pendingAccount.account, pendingAccount.password)
    setPendingAccount(null)
  }

  const handlePickerCancel = () => {
    setPickerOpen(false)
    setPendingAccount(null)
    clearIdentities()
  }

  return (
    <>
    <form
      className={cn("flex flex-col gap-6", className)}
      onSubmit={handleSubmit}
      {...props}
    >
      <div className="flex flex-col items-center gap-2 text-center">
        <h1 className="text-2xl font-bold">
          {mode === "platform" ? "平台管理员登录" : t.auth.loginTitle}
        </h1>
        <p className="text-sm text-balance text-muted-foreground">
          {mode === "platform"
            ? "super_admin 等平台管理员专用入口"
            : t.auth.loginSubtitle}
        </p>
      </div>
      {error && (
        <div className="rounded-md bg-destructive/10 p-3 text-sm text-destructive">
          {error}
        </div>
      )}
      <FieldGroup>
        <Field>
          <FieldLabel htmlFor="account">{t.auth.account}</FieldLabel>
          <Input
            id="account"
            name="account"
            type="text"
            placeholder="admin"
            required
            className="bg-background"
            disabled={isLoading}
          />
        </Field>
        <Field>
          <FieldLabel htmlFor="password">{t.auth.password}</FieldLabel>
          <Input
            id="password"
            name="password"
            type="password"
            required
            className="bg-background"
            disabled={isLoading}
          />
        </Field>
        <Field>
          <div className="flex items-center justify-between w-full">
            <a
              href="#"
              className="text-sm underline-offset-4 hover:underline"
            >
              {t.auth.forgotPassword}
            </a>
            <div className="text-sm">
              {mode === "tenant" ? (
                <>
                  {t.auth.noAccount}{" "}
                  <Link to="/signup" className="underline underline-offset-4">
                    {t.auth.signUp}
                  </Link>
                </>
              ) : (
                <Link
                  to="/login"
                  className="text-sm underline-offset-4 hover:underline"
                >
                  租户登录入口 →
                </Link>
              )}
            </div>
          </div>
        </Field>
        <Field>
          <Button type="submit" className="w-full" disabled={isLoading}>
            {isLoading
              ? t.auth.loggingIn
              : mode === "platform"
                ? "进入平台后台"
                : t.auth.login}
          </Button>
        </Field>
      </FieldGroup>
      <div className="text-balance text-center text-xs text-muted-foreground [&>a]:underline [&>a]:underline-offset-4 hover:[&>a]:text-primary">
        {t.auth.termsAgree} <a href="#">{t.auth.termsOfService}</a>{" "}
        {t.auth.and} <a href="#">{t.auth.privacyPolicy}</a>.
      </div>
    </form>

    <IdentityPickerDialog
      open={pickerOpen}
      onOpenChange={setPickerOpen}
      identities={availableIdentities}
      platformAvailable={platformAvailable}
      platformRoles={availablePlatformRoles}
      onSelectTenant={handlePickerSelectTenant}
      onSelectPlatform={handlePickerSelectPlatform}
      onCancel={handlePickerCancel}
    />
    </>
  )
}