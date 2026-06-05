import * as React from "react"
import { Link, useLocation } from "react-router-dom"
import { Button } from "@/components/ui/button"
import {
  SidebarGroup,
  SidebarGroupContent,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
} from "@/components/ui/sidebar"
import { CirclePlusIcon, MailIcon, ChevronDownIcon, ChevronRightIcon } from "lucide-react"

interface NavItem {
  title: string
  url: string
  icon?: React.ReactNode
  children?: NavItem[]
}

interface NavMainProps {
  items: NavItem[]
}

function NavMenuItem({ item, level = 0 }: { item: NavItem; level?: number }) {
  const location = useLocation()
  const [isExpanded, setIsExpanded] = React.useState(false)
  const hasChildren = item.children && item.children.length > 0
  const isActive = location.pathname === item.url
  const isChildActive = item.children?.some((child) => location.pathname.startsWith(child.url)) || false

  const shouldAutoExpand = isChildActive && !isExpanded

  React.useEffect(() => {
    if (shouldAutoExpand) {
      setIsExpanded(true)
    }
  }, [shouldAutoExpand])

  return (
    <>
      <SidebarMenuItem className="relative">
        {hasChildren ? (
          <SidebarMenuButton
            tooltip={item.title}
            isActive={isActive}
            onClick={() => setIsExpanded(!isExpanded)}
            className={`${level > 0 ? "ml-4" : ""}`}
          >
            <div className="flex items-center w-full">
              <div className="flex items-center gap-2 flex-1 min-w-0">
                {item.icon}
                <span className="truncate">{item.title}</span>
              </div>
              <span className="flex-shrink-0">
                {isExpanded ? (
                  <ChevronDownIcon className="h-3 w-3" />
                ) : (
                  <ChevronRightIcon className="h-3 w-3" />
                )}
              </span>
            </div>
          </SidebarMenuButton>
        ) : (
          <SidebarMenuButton
            tooltip={item.title}
            isActive={isActive}
            asChild
            className={`${level > 0 ? "ml-4" : ""}`}
          >
            <Link to={item.url}>
              {item.icon}
              <span className="truncate">{item.title}</span>
            </Link>
          </SidebarMenuButton>
        )}
      </SidebarMenuItem>
      {hasChildren && isExpanded && item.children?.map((child) => (
        <NavMenuItem key={child.url} item={child} level={level + 1} />
      ))}
    </>
  )
}

export function NavMain({
  items,
}: NavMainProps) {
  return (
    <SidebarGroup>
      <SidebarGroupContent className="flex flex-col gap-2">
        <SidebarMenu>
          <SidebarMenuItem className="flex items-center gap-2">
            <SidebarMenuButton
              tooltip="Quick Create"
              className="min-w-8 bg-primary text-primary-foreground duration-200 ease-linear hover:bg-primary/90 hover:text-primary-foreground active:bg-primary/90 active:text-primary-foreground"
            >
              <CirclePlusIcon />
              <span>Quick Create</span>
            </SidebarMenuButton>
            <Button
              size="icon"
              className="size-8 group-data-[collapsible=icon]:opacity-0"
              variant="outline"
            >
              <MailIcon />
              <span className="sr-only">Inbox</span>
            </Button>
          </SidebarMenuItem>
        </SidebarMenu>
        <SidebarMenu>
          {items.map((item) => (
            <NavMenuItem key={item.url} item={item} />
          ))}
        </SidebarMenu>
      </SidebarGroupContent>
    </SidebarGroup>
  )
}