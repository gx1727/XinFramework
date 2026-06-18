import { lazy, Suspense } from "react"
import { Routes, Route, Navigate } from "react-router-dom"

const LoginPage = lazy(() => import("@/pages/Login").then((m) => ({ default: m.LoginPage })))
const SignupPage = lazy(() => import("@/pages/Signup").then((m) => ({ default: m.SignupPage })))
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
const TenantsPage = lazy(() => import("@/pages/Tenants").then((m) => ({ default: m.TenantsPage })))

function PageLoader() {
  return (
    <div className="flex h-screen w-full items-center justify-center">
      <div className="h-8 w-8 animate-spin rounded-full border-4 border-muted border-t-primary" />
    </div>
  )
}

export function App() {
  return (
    <Suspense fallback={<PageLoader />}>
      <Routes>
        <Route path="/login" element={<LoginPage />} />
        <Route path="/signup" element={<SignupPage />} />
        <Route path="/dashboard" element={<DashboardPage />} />
        <Route path="/users" element={<UsersPage />} />
        <Route path="/roles" element={<RolesPage />} />
        <Route path="/settings" element={<SettingsPage />} />
        <Route path="/analytics" element={<AnalyticsPage />} />
        <Route path="/projects" element={<ProjectsPage />} />
        <Route path="/team" element={<TeamPage />} />
        <Route path="/menus" element={<MenusPage />} />
        <Route path="/resources" element={<ResourcesPage />} />
        <Route path="/organizations" element={<OrganizationsPage />} />
        <Route path="/dicts" element={<DictsPage />} />
        <Route path="/tenants" element={<TenantsPage />} />
        <Route path="/frames" element={<FramesPage />} />
        <Route path="/frame-categories" element={<FrameCategoriesPage />} />
        <Route path="/avatars" element={<AvatarsPage />} />
        <Route path="/avatar-categories" element={<AvatarCategoriesPage />} />
        <Route path="/cache" element={<CachePage />} />
        <Route path="*" element={<Navigate to="/dashboard" replace />} />
      </Routes>
    </Suspense>
  )
}

export default App
