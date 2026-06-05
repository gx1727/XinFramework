import { useEffect } from "react"
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
import { useTranslation } from "@/locales"

export function LoginForm({
  className,
  ...props
}: React.ComponentProps<"form">) {
  const navigate = useNavigate()
  const { login, isLoading, error, clearError, isAuthenticated } = useAuthStore()
  const t = useTranslation()

  useEffect(() => {
    if (isAuthenticated) {
      navigate("/dashboard")
    }
  }, [isAuthenticated, navigate])

  const handleSubmit = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault()
    clearError()

    const formData = new FormData(e.currentTarget)
    const account = formData.get("account") as string
    const password = formData.get("password") as string

    const success = await login(account, password)
    if (success) {
      navigate("/dashboard")
    }
  }

  return (
    <form 
      className={cn("flex flex-col gap-6", className)} 
      onSubmit={handleSubmit}
      {...props}
    >
      <div className="flex flex-col items-center gap-2 text-center">
        <h1 className="text-2xl font-bold">{t.auth.loginTitle}</h1>
        <p className="text-sm text-balance text-muted-foreground">
          {t.auth.loginSubtitle}
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
              {t.auth.noAccount}{" "}
              <Link to="/signup" className="underline underline-offset-4">
                {t.auth.signUp}
              </Link>
            </div>
          </div>
        </Field>
        <Field>
          <Button type="submit" className="w-full" disabled={isLoading}>
            {isLoading ? t.auth.loggingIn : t.auth.login}
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