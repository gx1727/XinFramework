import { GalleryVerticalEnd, ShieldIcon } from "lucide-react"
import { LoginForm } from "@/components/login-form"
import { useConfigItem } from "@/stores/configStore"
import { Link } from "react-router-dom"

export function TenantLoginPage() {
  const siteName = useConfigItem("site", "site_name") as string | undefined
  const siteLogo = useConfigItem("site", "site_logo") as string | undefined
  const loginBg = useConfigItem("site", "login_background") as
    | string
    | undefined

  const title = siteName || "XinFramework"
  const defaultBg =
    "https://coresg-normal.trae.ai/api/ide/v1/text_to_image?prompt=modern%20abstract%20dashboard%20ui%20design%20with%20glassmorphism%20style&image_size=portrait_16_9"

  if (typeof document !== "undefined" && document.title !== `${title} - 登录`) {
    document.title = `${title} - 登录`
  }

  return (
    <div className="grid min-h-svh lg:grid-cols-2">
      <div className="flex flex-col gap-4 p-6 md:p-10">
        <div className="flex justify-center gap-2 md:justify-start">
          <a href="#" className="flex items-center gap-2 font-medium">
            {siteLogo ? (
              <img
                src={siteLogo}
                alt={title}
                className="size-6 object-contain"
              />
            ) : (
              <div className="flex size-6 items-center justify-center rounded-md bg-primary text-primary-foreground">
                <GalleryVerticalEnd className="size-4" />
              </div>
            )}
            {title}
          </a>
        </div>
        <div className="flex flex-1 items-center justify-center">
          <div className="w-full max-w-xs">
            <LoginForm mode="tenant" />
          </div>
        </div>
        {/* <div className="text-center text-sm text-muted-foreground">
          Sys 管理员？{" "}
          <Link to="/sys/login" className="underline underline-offset-4">
            进入 sys 后台登录
          </Link>
        </div> */}
      </div>
      <div className="relative hidden bg-muted lg:block">
        <img
          src={loginBg || defaultBg}
          alt="Image"
          className="absolute inset-0 h-full w-full object-cover dark:brightness-[0.2] dark:grayscale"
        />
      </div>
    </div>
  )
}
