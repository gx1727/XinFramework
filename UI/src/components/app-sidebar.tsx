import * as React from "react"
import { useEffect } from "react"
import { NavDocuments } from "@/components/nav-documents"
import { NavMain } from "@/components/nav-main"
import { NavSecondary } from "@/components/nav-secondary"
import { NavUser } from "@/components/nav-user"
import { useAuthStore } from "@/stores/authStore"
import {
  Sidebar,
  SidebarContent,
  SidebarFooter,
  SidebarHeader,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"
import {
  LayoutDashboardIcon,
  ListIcon,
  ChartBarIcon,
  FolderIcon,
  UsersIcon,
  CameraIcon,
  FileTextIcon,
  Settings2Icon,
  CircleHelpIcon,
  SearchIcon,
  DatabaseIcon,
  FileChartColumnIcon,
  FileIcon,
  CommandIcon,
  ShieldIcon,
  MenuIcon,
  Settings,
  ImageIcon,
  FrameIcon,
  KeyIcon,
} from "lucide-react"
import { t } from "@/locales"
import { useMenuStore, type UnifiedMenuItem } from "@/stores/menuStore"
import { useConfigItem } from "@/stores/configStore"

const iconMap: Record<string, React.ComponentType<{ className?: string }>> = {
  LayoutDashboardIcon,
  ListIcon,
  ChartBarIcon,
  FolderIcon,
  UsersIcon,
  CameraIcon,
  FileTextIcon,
  Settings2Icon,
  CircleHelpIcon,
  SearchIcon,
  DatabaseIcon,
  FileChartColumnIcon,
  FileIcon,
  CommandIcon,
  ShieldIcon,
  MenuIcon,
  SettingsIcon: Settings,
  Settings: Settings,
  ImageIcon,
  FrameIcon,
  KeyIcon,
}

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const { menus, fetchMenus } = useMenuStore()
  const siteName = useConfigItem("site", "site_name") as string | undefined
  const siteLogo = useConfigItem("site", "site_logo") as string | undefined
  const sidebarTitle = siteName || "XinFramework"

  useEffect(() => {
    fetchMenus()
  }, [fetchMenus])

  const buildNavItems = (
    menuList: UnifiedMenuItem[]
  ): {
    title: string
    url: string
    icon?: React.ReactNode
    children?: { title: string; url: string; icon?: React.ReactNode }[]
  }[] => {
    return menuList
      .filter((menu) => menu.parent_id === 0)
      .sort((a, b) => a.sort - b.sort)
      .map((menu) => {
        const IconComp = iconMap[menu.icon || ""] || FileIcon
        const item = {
          title: menu.name,
          url: menu.path || menu.url || "#",
          icon: React.createElement(IconComp),
        }
        if (menu.children && menu.children.length > 0) {
          return {
            ...item,
            children: menu.children
              .sort((a, b) => a.sort - b.sort)
              .map((child) => ({
                title: child.name,
                url: child.path || child.url || "#",
                icon: React.createElement(
                  iconMap[child.icon || ""] || FileIcon
                ),
              })),
          }
        }
        return item
      })
  }

  const navMainItems = buildNavItems(menus)

  const authUser = useAuthStore((s) => s.user)
  // 真实字段映射：real_name 优先 → nickname → code；email 来自 accounts；avatar 来自 users
  const sidebarUser = authUser
    ? {
        name:
          authUser.real_name?.trim() ||
          authUser.nickname?.trim() ||
          authUser.code,
        email: authUser.email?.trim() || authUser.role,
        avatar: authUser.avatar?.trim() || "",
      }
    : { name: "", email: "", avatar: "" }

  const documentsItems = [
    {
      name: t.nav.dataLibrary,
      url: "#",
      icon: <DatabaseIcon />,
    },
    {
      name: t.nav.reports,
      url: "#",
      icon: <FileChartColumnIcon />,
    },
    {
      name: t.nav.wordAssistant,
      url: "#",
      icon: <FileIcon />,
    },
  ]

  const data = {
    user: sidebarUser,
    navSecondary: [
      {
        title: t.nav.settings,
        url: "#",
        icon: <Settings2Icon />,
      },
      {
        title: t.nav.getHelp,
        url: "#",
        icon: <CircleHelpIcon />,
      },
      {
        title: t.nav.search,
        url: "#",
        icon: <SearchIcon />,
      },
    ],
  }

  return (
    <Sidebar collapsible="offcanvas" {...props}>
      <SidebarHeader>
        <SidebarMenu>
          <SidebarMenuItem>
            <SidebarMenuButton
              asChild
              className="data-[slot=sidebar-menu-button]:p-1.5!"
            >
              <a href="#">
                {siteLogo ? (
                  <img
                    src={siteLogo}
                    alt={sidebarTitle}
                    className="size-5! object-contain"
                  />
                ) : (
                  <CommandIcon className="size-5!" />
                )}
                <span className="text-base font-semibold">{sidebarTitle}</span>
              </a>
            </SidebarMenuButton>
          </SidebarMenuItem>
        </SidebarMenu>
      </SidebarHeader>
      <SidebarContent>
        <NavMain items={navMainItems} />
        <NavDocuments items={documentsItems} />
        <NavSecondary items={data.navSecondary} className="mt-auto" />
      </SidebarContent>
      <SidebarFooter>
        <NavUser user={data.user} />
      </SidebarFooter>
    </Sidebar>
  )
}
