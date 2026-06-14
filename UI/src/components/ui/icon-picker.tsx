import { useState, useMemo, useEffect } from "react"
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog"
import { Button } from "@/components/ui/button"
import { Input } from "@/components/ui/input"
import { SearchIcon, XIcon } from "lucide-react"
import * as LucideIcons from "lucide-react"
import { cn } from "@/lib/utils"

const iconCategories = {
  navigation: [
    "Home", "Menu", "Settings", "User", "Users", "Search", "Bell", "ChevronDown", "ChevronRight",
    "ChevronLeft", "ChevronUp", "ArrowLeft", "ArrowRight", "ArrowUp", "ArrowDown",
    "MoreHorizontal", "MoreVertical", "PanelLeft", "PanelRight", "PanelTop", "PanelBottom",
    "PanelTopClose", "PanelBottomClose", "PanelLeftClose", "PanelRightClose",
    "AppWindow", "Gauge", "Layout", "LayoutDashboard", "LayoutGrid", "LayoutList",
    "LayoutPanelTop", "LayoutPanelLeft", "Sidebar", "SidebarClose", "SidebarOpen",
    "CircleUser", "UserCircle", "LogIn", "LogOut", "Plug", "PlugZap",
  ],
  actions: [
    "Plus", "Minus", "Check", "X", "Edit", "Edit2", "Edit3", "Trash", "Trash2", "Copy",
    "Cut", "Download", "Upload", "Save", "Share", "ExternalLink", "Link", "Lock",
    "Unlock", "Eye", "EyeOff", "RefreshCw", "RotateCcw", "RotateCw",
    "Undo", "Redo", "Clipboard", "ClipboardCheck", "ClipboardList", "ClipboardPaste",
    "FilePlus", "FileMinus", "FileEdit", "FileCheck", "FileX", "FolderPlus",
    "FolderMinus", "FolderEdit", "Archive", "Unarchive", "Move", "MousePointer2",
    "Click", "Hand", "Grab", "Maximize", "Minimize", "ZoomIn", "ZoomOut", "Focus",
    "Square", "Circle", "Triangle", "Hexagon", "Octagon", "SquareStack",
  ],
  content: [
    "File", "FileText", "FilePlus", "FileMinus", "FileEdit", "Folder", "FolderOpen",
    "FolderPlus", "Image", "ImageIcon", "Camera", "Video", "Music", "FileCode",
    "FileArchive", "FileAudio", "FileImage", "FileVideo", "FileUp", "FileDown",
    "FileWarning", "FileX", "Files", "FolderTree", "FolderSymlink", "Newspaper",
    "BookOpen", "Book", "Bookmark", "BookmarkCheck", "BookMarked", "Library",
    "GraduationCap", "School", "Bell", "BellRing", "BellOff", "Megaphone",
    "Speaker", "Volume", "Volume1", "Volume2", "VolumeX", "Music2", "Music3",
    "Headphones", "Mic", "Mic2", "MicOff", "CameraOff", "Video", "VideoOff",
    "Tv", "Tv2", "Monitor", "MonitorPlay", "Projector", "Presentation",
  ],
  communication: [
    "Mail", "Inbox", "Send", "MessageSquare", "MessageCircle", "Phone", "PhoneCall",
    "MapPin", "AtSign", "Share2", "Share", "Forward", "Reply", "ReplyAll",
    "MailCheck", "MailOpen", "MailPlus", "MailX", "MessageCircleDashed",
    "MessageSquareDashed", "MessageSquareReply", "MessagesSquare", "MessagesCircle",
    "PhoneForwarded", "PhoneIncoming", "PhoneMissed", "PhoneOff", "PhoneOutgoing",
    "Contact", "Contact2", "UserCheck", "UserPlus", "UserMinus", "UserX",
    "UserGroup", "UserSearch", "AddressBook", "IdCard",
    "MapPinHouse", "MapPins", "Navigation", "Navigation2", "Globe",
    "Globe2", "Laptop", "Smartphone", "Tablet", "Watch", "Radio",
  ],
  data: [
    "BarChart", "BarChart2", "BarChart3", "BarChartHorizontal", "PieChart", "LineChart",
    "Activity", "TrendingUp", "TrendingDown", "Table", "Database", "Server",
    "Table2", "Columns", "Rows3", "Grid3x3", "Grid2x2", "LayoutGrid",
    "LayoutList", "ChartArea", "ChartBar", "ChartColumn", "ChartLine",
    "ChartPie", "ChartScatter", "GanttChart", "Kanban",
    "TreasureChest", "Box", "Boxes", "Package", "Package2", "PackageCheck",
    "PackageMinus", "PackageOpen", "PackagePlus", "PackageSearch", "PackageX",
    "Container", "Archive", "Bin", "HardDrive", "ExternalDrive", "Save",
    "SaveAll", "Cloud", "CloudDownload", "CloudUpload", "CloudOff", "Sun",
  ],
  items: [
    "Star", "Heart", "Bookmark", "Tag", "Tags", "Package", "Box", "Gift",
    "ShoppingCart", "CreditCard", "DollarSign", "Coins", "Wallet", "Banknote",
    "Receipt", "Calculator", "ShoppingBag", "ShoppingBasket",
    "Trophy", "Medal", "Award", "Crown", "Gem", "Diamond", "Sparkles",
    "StarHalf", "HeartHandshake", "HeartOff", "HeartPulse", "Swords",
    "Shield", "ShieldCheck", "ShieldAlert", "ShieldClose", "ShieldPlus",
    "ShieldMinus", "Umbrella", "Bandage", "Syringe", "Pill",
    "Thermometer", "TestTube", "Microscope", "FlaskConical", "FlaskRound",
    "Atom", "Dna", "Virus", "Flower", "Flower2", "Leaf", "TreeDeciduous",
  ],
  status: [
    "CheckCircle", "CheckCircle2", "XCircle", "AlertCircle", "AlertTriangle",
    "AlertOctagon", "Info", "HelpCircle", "Shield", "ShieldCheck", "ShieldAlert",
    "CheckSquare", "XCircle", "XSquare", "Square", "Circle",
    "CircleDot", "CircleSlash", "CircleCheck", "CircleHelp", "CircleDashed",
    "CircleDotDashed", "TriangleAlert", "TriangleRight",
    "Signal", "SignalHigh", "SignalMedium", "SignalLow", "SignalZero",
    "Battery", "BatteryCharging", "BatteryFull", "BatteryLow", "BatteryMedium",
    "BatteryWarning", "Zap", "ZapOff", "Flashlight", "Sun",
    "Sunset", "Sunrise", "Stars", "Cloud", "CloudRain", "CloudSnow",
  ],
  time: [
    "Clock", "Calendar", "CalendarPlus", "CalendarMinus", "CalendarCheck",
    "Timer", "Watch", "Hourglass", "AlarmClock", "Clock1", "Clock2",
    "Clock3", "Clock4", "Clock5", "Clock6", "Clock7", "Clock8",
    "Clock9", "Clock10", "Clock11", "Clock12", "CalendarClock",
    "CalendarHeart", "CalendarRange", "CalendarSearch", "CalendarX",
    "CalendarDays", "TimerOff",
    "Watches", "Smartwatch",
    "Sunset", "Sunrise", "MoonStar", "Sun", "SunMedium",
    "SunDim", "SunSnow", "CloudSun", "CloudMoon", "CloudLightning",
  ],
  ui: [
    "Grid", "Layout", "LayoutDashboard", "LayoutGrid", "Columns",
    "Rows3", "Sidebar", "SidebarClose", "Maximize", "Minimize", "ZoomIn",
    "ZoomOut", "AlignLeft", "AlignCenter", "AlignRight", "AlignJustify",
    "Bold", "Italic", "Underline", "Strikethrough", "Type", "TextCursor",
    "TextCursorInput", "FontFamily", "FontSize", "Heading", "Heading1",
    "Heading2", "Heading3", "Heading4", "Heading5", "Heading6",
    "List", "ListChecks", "ListEnd", "ListOrdered", "ListPlus", "ListStart",
    "ListTodo", "ListTree", "ListIcon", "MenuSquare",
    "Hamburger", "WrapText", "Write", "PenLine", "Pen",
    "PenTool", "Pencil", "PencilLine", "Highlighter", "Eraser", "Ruler",
    "Sizing", "Scaling", "Move", "MoveHorizontal", "MoveVertical",
  ],
  business: [
    "Briefcase", "Building", "Building2", "Store", "Factory", "Warehouse",
    "Landmark", "Banknote", "Receipt", "Calculator", "Handshake",
    "Coins", "DollarSign", "Euro", "PoundSterling", "Yen", "Currency",
    "CreditCard", "Wallet",
    "Invoice", "FileText", "Percentage", "Percent", "TrendingUp",
    "TrendingDown", "PieChart", "Target", "Crosshair", "Compass", "Map",
    "Ship", "Plane", "PlaneTakeoff", "PlaneLanding", "Train", "TrainFront",
    "Bus", "Car", "Truck", "Van", "Anchor", "Sailboat",
    "Hotel", "Bed", "BedDouble", "BedSingle", "Bath", "Toilet",
  ],
  people: [
    "UserCircle", "UserPlus", "UserMinus", "UserCheck", "UserX",
    "UserGroup", "Contact", "Contact2", "IdCard", "UserCog",
    "UserSearch", "UsersRound", "UserRound", "Group", "UserPen",
    "UserVoice", "UserSpeaker", "ContactRound", "CircleUser",
    "CircleUserRound", "UserBadge",
    "CreditCard", "Person", "PersonStanding",
    "PersonWalking", "PersonRun", "PersonJump", "PersonSit", "Armchair",
    "Baby",
  ],
  misc: [
    "Sparkles", "Zap", "Flashlight", "Key", "KeyRound", "Code", "Code2",
    "Terminal", "Bug", "Target", "Crosshair", "Compass", "Globe", "Map",
    "Navigation", "Binoculars", "Telescope", "Microscope", "TestTube",
    "Atom", "Dna", "Orbit", "Rocket", "Satellite", "Meteor", "Planet",
    "Sun", "Moon", "Star", "Stars", "Cloud", "CloudRain", "CloudSnow",
    "CloudLightning", "CloudSun", "CloudMoon", "Wind", "Snowflake",
    "Umbrella", "Rainbow", "Wave", "Waves", "Mountain", "Glacier",
    "TreePine", "TreeDeciduous", "Tree", "Flower", "Flower2", "Leaf",
    "Pine", "Sprout", "Cherry", "Apple", "Banana", "Carrot", "Coffee",
    "Pizza", "Cake", "Cookie", "CupSoda", "GlassWater", "Beer",
  ],
}

const popularIcons = [
  "Home", "Menu", "Settings", "User", "Users", "Search", "Plus", "Edit",
  "Trash", "Check", "X", "File", "Folder", "Image", "Mail", "Bell",
  "Dashboard", "Layout", "Grid", "List", "Chart", "Table", "Database",
]

export interface IconOption {
  label: string
  value: string
  category: string
}

const uniqueIcons = new Set<string>()
const allIcons: IconOption[] = []

Object.entries(iconCategories).forEach(([category, icons]) => {
  icons.forEach((iconName) => {
    if (!uniqueIcons.has(iconName)) {
      uniqueIcons.add(iconName)
      allIcons.push({
        label: iconName,
        value: iconName,
        category,
      })
    }
  })
})

interface IconPickerProps {
  value?: string
  onChange: (value: string) => void
  placeholder?: string
  disabled?: boolean
}

export function IconPicker({ value, onChange, placeholder = "选择图标", disabled }: IconPickerProps) {
  const [open, setOpen] = useState(false)
  const [searchTerm, setSearchTerm] = useState("")
  const [activeCategory, setActiveCategory] = useState<string | "all">("all")

  useEffect(() => {
    if (!open) {
      setSearchTerm("")
      setActiveCategory("all")
    }
  }, [open])

  const filteredIcons = useMemo(() => {
    const seenIcons = new Set<string>()
    let icons: IconOption[] = []
    
    if (activeCategory === "popular") {
      allIcons.forEach((icon) => {
        if (popularIcons.includes(icon.value) && !seenIcons.has(icon.value)) {
          seenIcons.add(icon.value)
          icons.push(icon)
        }
      })
    } else if (activeCategory !== "all") {
      allIcons.forEach((icon) => {
        if (icon.category === activeCategory && !seenIcons.has(icon.value)) {
          seenIcons.add(icon.value)
          icons.push(icon)
        }
      })
    } else {
      icons = [...allIcons]
    }
    
    if (searchTerm) {
      const term = searchTerm.toLowerCase()
      icons = icons.filter((icon) =>
        icon.label.toLowerCase().includes(term) ||
        icon.value.toLowerCase().includes(term) ||
        icon.category.toLowerCase().includes(term)
      )
    }
    return icons
  }, [searchTerm, activeCategory])

  const renderIcon = (iconName: string, size: number = 20) => {
    const IconComponent = (LucideIcons as unknown as Record<string, React.ComponentType<{ size?: number; className?: string }>>)[iconName]
    if (IconComponent) {
      return <IconComponent size={size} className="text-foreground" />
    }
    return null
  }

  const handleSelect = (iconValue: string) => {
    onChange(iconValue)
    setOpen(false)
    setSearchTerm("")
  }

  const handleClear = (e: React.MouseEvent) => {
    e.stopPropagation()
    onChange("")
  }

  const categoryLabels: Record<string, string> = {
    navigation: "导航",
    actions: "操作",
    content: "内容",
    communication: "通信",
    data: "数据",
    items: "物品",
    status: "状态",
    time: "时间",
    ui: "界面",
    business: "商业",
    people: "人物",
    misc: "其他",
    popular: "常用",
    all: "全部",
  }

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogTrigger asChild>
        <Button
          variant="outline"
          role="combobox"
          aria-expanded={open}
          className={cn("w-full justify-start font-normal", !value && "text-muted-foreground")}
          disabled={disabled}
        >
          {value ? (
            <>
              <span className="mr-2 flex items-center">
                {renderIcon(value)}
              </span>
              <span className="flex-1 truncate">{value}</span>
              {value && !disabled && (
                <span
                  onClick={handleClear}
                  className="ml-auto p-1 hover:bg-accent rounded"
                >
                  <XIcon className="h-3 w-3" />
                </span>
              )}
            </>
          ) : (
            <>
              <span className="mr-2 opacity-0">
                {renderIcon("Search")}
              </span>
              <span className="flex-1">{placeholder}</span>
            </>
          )}
        </Button>
      </DialogTrigger>
      <DialogContent className="p-6 sm:max-w-[600px]">
        <DialogHeader className="p-0 mb-4">
          <DialogTitle className="text-base">选择图标</DialogTitle>
        </DialogHeader>
        
        <div className="mb-4">
          <div className="relative">
            <SearchIcon className="absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
            <Input
              placeholder="搜索图标名称..."
              value={searchTerm}
              onChange={(e) => setSearchTerm(e.target.value)}
              className="pl-9 h-10"
            />
          </div>
        </div>
        
        <div className="mb-4 overflow-x-auto">
          <div className="flex gap-1.5">
            {["popular", "all", "navigation", "actions", "content", "communication", "data", "items", "status", "time", "ui", "business", "people", "misc"].map((category) => (
              <Button
                key={category}
                variant={activeCategory === category ? "secondary" : "ghost"}
                size="sm"
                onClick={() => setActiveCategory(category)}
                className="h-7 px-2.5 text-xs shrink-0"
              >
                {categoryLabels[category]}
              </Button>
            ))}
          </div>
        </div>
        
        <div className="max-h-[400px] overflow-y-auto">
          <div className="grid grid-cols-8 gap-2">
            {filteredIcons.map((icon) => (
              <button
                key={icon.value}
                onClick={() => handleSelect(icon.value)}
                className={cn(
                  "flex flex-col items-center justify-center p-2 rounded-md hover:bg-accent transition-colors",
                  value === icon.value && "bg-accent ring-2 ring-primary ring-offset-1"
                )}
                title={icon.label}
                type="button"
              >
                <span className="flex items-center justify-center h-6">
                  {renderIcon(icon.value)}
                </span>
                <span className="text-[9px] mt-1 truncate w-full text-center leading-tight">
                  {icon.label}
                </span>
              </button>
            ))}
          </div>
          {filteredIcons.length === 0 && (
            <div className="text-center py-12 text-muted-foreground text-sm">
              <SearchIcon className="mx-auto h-8 w-8 mb-2 opacity-50" />
              <p>未找到匹配的图标</p>
              <p className="text-xs mt-1">尝试其他关键词或分类</p>
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  )
}

export { iconCategories }