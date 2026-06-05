import * as React from "react"
import { useEffect } from "react"
import { NavDocuments } from "@/components/nav-documents"
import { NavMain } from "@/components/nav-main"
import { NavSecondary } from "@/components/nav-secondary"
import { NavUser } from "@/components/nav-user"
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
} from "lucide-react"
import { useTranslation } from "@/locales"
import { useMenuStore } from "@/stores/menuStore"
import type { MenuItem } from "@/stores/menuStore"

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
  ShieldIcon,
  MenuIcon,
  SettingsIcon: Settings,
  Settings: Settings,
  ImageIcon,
  FrameIcon,
}

export function AppSidebar({ ...props }: React.ComponentProps<typeof Sidebar>) {
  const t = useTranslation()
  const { menus, fetchMenus } = useMenuStore()

  useEffect(() => {
    fetchMenus()
  }, [fetchMenus])

  const buildNavItems = (menuList: MenuItem[]): { title: string; url: string; icon?: React.ReactNode; children?: { title: string; url: string; icon?: React.ReactNode }[] }[] => {
    return menuList
      .filter((menu) => menu.parent_id === 0)
      .sort((a, b) => a.sort - b.sort)
      .map((menu) => {
        const item = {
          title: menu.name,
          url: menu.path || menu.url || "#",
          icon: iconMap[menu.icon || ""] ? React.createElement(iconMap[menu.icon || ""]) : <FileIcon />,
        }
        if (menu.children && menu.children.length > 0) {
          return {
            ...item,
            children: menu.children
              .sort((a, b) => a.sort - b.sort)
              .map((child) => ({
                title: child.name,
                url: child.path || child.url || "#",
                icon: iconMap[child.icon || ""] ? React.createElement(iconMap[child.icon || ""]) : <FileIcon />,
              })),
          }
        }
        return item
      })
  }

  const navMainItems = buildNavItems(menus)

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
    user: {
      name: "shadcn",
      email: "m@example.com",
      avatar: "/avatars/shadcn.jpg",
    },
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
                <CommandIcon className="size-5!" />
                <span className="text-base font-semibold">Acme Inc.</span>
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