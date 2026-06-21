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

type LoginMode = "tenant" | "platform"

export function LoginForm({
  mode,
  className,
  ...props
}: React.ComponentProps<"form"> & { mode: LoginMode }) {
  const navigate = useNavigate()
  const tenantLogin = useAuthStore((s) => s.tenantLogin)
  const platformLogin = useAuthStore((s) => s.platformLogin)
  const isLoading = useAuthStore((s) => s.isLoading)
  const error = useAuthStore((s) => s.error)
  const clearError = useAuthStore((s) => s.clearError)
  const isAuthenticated = useAuthStore((s) => s.isAuthenticated)
  const scope = useAuthStore((s) => s.scope)

  // platform 模式：tenantId 默认 1（兼容旧接口，但 platform-login 不读这个字段）
  // tenant 模式：用户必须填 tenantId
  const [tenantId, setTenantId] = useState<string>("1")

  useEffect(() => {
    if (isAuthenticated) {
      // 按 scope 跳转到对应 dashboard
      navigate(scope === "platform" ? "/platform/dashboard" : "/app/dashboard")
    }
  }, [isAuthenticated, scope, navigate])

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    clearError()

    const formData = new FormData(e.currentTarget)
    const account = formData.get("account") as string
    const password = formData.get("password") as string

    let success = false
    if (mode === "platform") {
      success = await platformLogin(account, password)
    } else {
      const tid = parseInt(tenantId, 10) || 0
      if (tid <= 0) {
        clearError()
        // 这里通过 setError 复用 store 不太直观，直接返回 false 让外层处理
        return
      }
      success = await tenantLogin(account, password, tid)
    }
    if (success) {
      // useEffect 会处理跳转
    }
  }

  return (
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
        {mode === "tenant" && (
          <Field>
            <FieldLabel htmlFor="tenant_id">租户 ID</FieldLabel>
            <Input
              id="tenant_id"
              name="tenant_id"
              type="number"
              min={1}
              value={tenantId}
              onChange={(e) => setTenantId(e.target.value)}
              placeholder="1"
              required
              className="bg-background"
              disabled={isLoading}
            />
          </Field>
        )}
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
  )
}