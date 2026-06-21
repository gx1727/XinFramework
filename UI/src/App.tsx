import { lazy, Suspense } from "react"
import { Routes, Route, Navigate, useLocation } from "react-router-dom"
import { useAuthStore } from "@/stores/authStore"

// ===== 登录 =====
const TenantLoginPage = lazy(() => import("@/pages/TenantLogin").then((m) => ({ default: m.TenantLoginPage })))
const PlatformLoginPage = lazy(() => import("@/pages/PlatformLogin").then((m) => ({ default: m.PlatformLoginPage })))

// ===== 租户域（/app/*）=====
const DashboardPage = lazy(() => import("@/pages/Dashboard").then((m) => ({ default: m.DashboardPage })))
const UsersPage = lazy(() => import("@/pages/Users").then((m) => ({ default: m.UsersPage })))
const RolesPage = lazy(() => import("@/pages/Roles").then((m) => ({ default: m.RolesPage })))
const SettingsPage = lazy(() => import("@/pages/Settings").then((m) => ({ default: m.SettingsPage })))
const AnalyticsPage = lazy(() => import("@/pages/Analytics").then((m) => ({ default: m.AnalyticsPage })))
const ProjectsPage = lazy(() => import("@/pages/Projects").then((m) => ({ default: m.ProjectsPage })))
const TeamPage = lazy(() => import("@/pages/Team").then((m) => ({ default: m.TeamPage })))
const MenusPage = lazy(() => import("@/pages/Menus").then((m) => ({ default: m.MenusPage })))
const FramesPage = lazy(() => import("@/pages/Frames").then((m) => ({ default: m.FramesPage })))
const AvatarsPage = lazy(() => import("@/pages/Avatars").then((m) => ({ default: m.AvatarsPage })))
const FrameCategoriesPage = lazy(() => import("@/pages/FrameCategories").then((m) => ({ default: m.FrameCategoriesPage })))
const AvatarCategoriesPage = lazy(() => import("@/pages/AvatarCategories").then((m) => ({ default: m.AvatarCategoriesPage })))
const ResourcesPage = lazy(() => import("@/pages/Resources").then((m) => ({ default: m.ResourcesPage })))
const OrganizationsPage = lazy(() => import("@/pages/Organizations").then((m) => ({ default: m.OrganizationsPage })))
const DictsPage = lazy(() => import("@/pages/Dicts").then((m) => ({ default: m.DictsPage })))
const CachePage = lazy(() => import("@/pages/Cache"))

// ===== 平台域（/platform/*）=====
const PlatformDashboardPage = lazy(() => import("@/pages/PlatformDashboard").then((m) => ({ default: m.PlatformDashboardPage })))
const PlatformTenantsPage = lazy(() => import("@/pages/Tenants").then((m) => ({ default: m.TenantsPage })))
const PlatformMenusPage = lazy(() => import("@/pages/Menus").then((m) => ({ default: m.MenusPage })))
const PlatformConfigsPage = lazy(() => import("@/pages/PlatformConfigs").then((m) => ({ default: m.PlatformConfigsPage })))
const PlatformDictsPage = lazy(() => import("@/pages/PlatformDicts").then((m) => ({ default: m.PlatformDictsPage })))

function PageLoader() {
  return (
    <div className="flex h-screen w-full items-center justify-center">
      <div className="h-8 w-8 animate-spin rounded-full border-4 border-muted border-t-primary" />
    </div>
  )
}

/**
 * RequireScope 路由守卫：
 *   - 已登录但 scope 不匹配 → 跳转到该 scope 的默认页（避免 token 跨域串用）
 *   - 未登录 → 跳转到 /login
 */
function RequireScope({ scope, children }: { scope: "tenant" | "platform"; children: React.ReactNode }) {
  const isAuthed = useAuthStore((s) => s.isAuthenticated)
  const currentScope = useAuthStore((s) => s.scope)
  const location = useLocation()
  if (!isAuthed) {
    return <Navigate to={scope === "platform" ? "/platform/login" : "/login"} replace state={{ from: location }} />
  }
  if (currentScope !== scope) {
    return <Navigate to={currentScope === "platform" ? "/platform/dashboard" : "/app/dashboard"} replace />
  }
  return <>{children}</>
}

export function App() {
  return (
    <Suspense fallback={<PageLoader />}>
      <Routes>
        {/* ===== 入口 ===== */}
        <Route path="/" element={<Navigate to="/login" replace />} />

        {/* ===== 登录 ===== */}
        <Route path="/login" element={<TenantLoginPage />} />
        <Route path="/signup" element={<Navigate to="/login" replace />} />
        <Route path="/platform/login" element={<PlatformLoginPage />} />

        {/* ===== 租户域 /app/*（业务） ===== */}
        <Route
          path="/app"
          element={<Navigate to="/app/dashboard" replace />}
        />
        <Route
          path="/app/dashboard"
          element={
            <RequireScope scope="tenant">
              <DashboardPage />
            </RequireScope>
          }
        />
        <Route
          path="/app/users"
          element={
            <RequireScope scope="tenant">
              <UsersPage />
            </RequireScope>
          }
        />
        <Route
          path="/app/roles"
          element={
            <RequireScope scope="tenant">
              <RolesPage />
            </RequireScope>
          }
        />
        <Route
          path="/app/menus"
          element={
            <RequireScope scope="tenant">
              <MenusPage />
            </RequireScope>
          }
        />
        <Route
          path="/app/organizations"
          element={
            <RequireScope scope="tenant">
              <OrganizationsPage />
            </RequireScope>
          }
        />
        <Route
          path="/app/resources"
          element={
            <RequireScope scope="tenant">
              <ResourcesPage />
            </RequireScope>
          }
        />
        <Route
          path="/app/dicts"
          element={
            <RequireScope scope="tenant">
              <DictsPage />
            </RequireScope>
          }
        />
        <Route
          path="/app/configs"
          element={
            <RequireScope scope="tenant">
              <DashboardPage />
            </RequireScope>
          }
        />
        <Route
          path="/app/settings"
          element={
            <RequireScope scope="tenant">
              <SettingsPage />
            </RequireScope>
          }
        />
        <Route
          path="/app/analytics"
          element={
            <RequireScope scope="tenant">
              <AnalyticsPage />
            </RequireScope>
          }
        />
        <Route
          path="/app/projects"
          element={
            <RequireScope scope="tenant">
              <ProjectsPage />
            </RequireScope>
          }
        />
        <Route
          path="/app/team"
          element={
            <RequireScope scope="tenant">
              <TeamPage />
            </RequireScope>
          }
        />
        <Route
          path="/app/frames"
          element={
            <RequireScope scope="tenant">
              <FramesPage />
            </RequireScope>
          }
        />
        <Route
          path="/app/frame-categories"
          element={
            <RequireScope scope="tenant">
              <FrameCategoriesPage />
            </RequireScope>
          }
        />
        <Route
          path="/app/avatars"
          element={
            <RequireScope scope="tenant">
              <AvatarsPage />
            </RequireScope>
          }
        />
        <Route
          path="/app/avatar-categories"
          element={
            <RequireScope scope="tenant">
              <AvatarCategoriesPage />
            </RequireScope>
          }
        />
        <Route
          path="/app/cache"
          element={
            <RequireScope scope="tenant">
              <CachePage />
            </RequireScope>
          }
        />

        {/* ===== 平台域 /platform/*（super_admin） ===== */}
        <Route
          path="/platform"
          element={<Navigate to="/platform/dashboard" replace />}
        />
        <Route
          path="/platform/dashboard"
          element={
            <RequireScope scope="platform">
              <PlatformDashboardPage />
            </RequireScope>
          }
        />
        <Route
          path="/platform/tenants"
          element={
            <RequireScope scope="platform">
              <PlatformTenantsPage />
            </RequireScope>
          }
        />
        <Route
          path="/platform/menus"
          element={
            <RequireScope scope="platform">
              <PlatformMenusPage />
            </RequireScope>
          }
        />
        <Route
          path="/platform/configs"
          element={
            <RequireScope scope="platform">
              <PlatformConfigsPage />
            </RequireScope>
          }
        />
        <Route
          path="/platform/dicts"
          element={
            <RequireScope scope="platform">
              <PlatformDictsPage />
            </RequireScope>
          }
        />

        {/* ===== 兼容期：旧路径 → 新路径 ===== */}
        <Route path="/dashboard" element={<Navigate to="/app/dashboard" replace />} />
        <Route path="/tenants" element={<Navigate to="/platform/tenants" replace />} />
        <Route path="/menus" element={<Navigate to="/app/menus" replace />} />
        <Route path="/dicts" element={<Navigate to="/app/dicts" replace />} />
        <Route path="/configs" element={<Navigate to="/app/configs" replace />} />

        {/* ===== 其他 ===== */}
        <Route path="*" element={<Navigate to="/login" replace />} />
      </Routes>
    </Suspense>
  )
}

export default App