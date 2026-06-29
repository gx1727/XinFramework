# UI/AGENTS.md

> 给 AI agent 协作者看的 XinFramework 前端设计说明。读这一份，比重新读 24 个 .tsx 快得多。

---

## 1. 技术栈

- **框架**：React 19 + TypeScript 5.9
- **构建**：Vite 7.3 + @vitejs/plugin-react
- **样式**：Tailwind CSS v4（`@tailwindcss/vite` 插件）
- **UI 组件**：shadcn/ui（基于 Radix UI），~25 个组件在 `components/ui/`
- **图标**：lucide-react；另有自研 `IconPicker` 组件
- **路由**：react-router-dom v7（`App.tsx` 集中 lazy + `RequireScope` guard）
- **状态**：zustand v5（`authStore` localStorage 持久化；`menuStore` / `configStore` / `permissionStore` 内存）
- **文案**：仅简体中文；`src/locales/zh-CN.ts`；`t = zhCN` 直接作为静态对象使用
- **HTTP**：原生 `fetch` + `ApiError`（带 JWT 自动 refresh + 指数退避重试）
- **表格**：@tanstack/react-table v8（通用 `data-table.tsx` 封装）
- **图表**：Recharts v3
- **通知**：Sonner v2
- **表单验证**：Zod v4
- **图标选择**：lucide-react 优先

## 2. 目录结构

```
UI/src/
├── api/
│   ├── common.ts          # api() 封装 + ApiError + JWT refresh + 重试
│   ├── index.ts           # 重导出
│   ├── auth.ts            # 认证端点
│   ├── user.ts / role.ts / menu.ts / platformMenu.ts /
│   │   organization.ts / tenant.ts / dict.ts / config.ts /
│   │   resource.ts / frame.ts / frameCategory.ts /
│   │   avatar.ts / avatarCategory.ts / space.ts /
│   │   asset.ts / system.ts
├── components/
│   ├── ui/                # shadcn 组件（~25 个）
│   ├── schema/            # DynamicForm + DynamicTable + showIfEvaluator
│   ├── permission/        # Auth.tsx（按钮级）+ DynamicRouter.tsx（路由级）
│   ├── app-sidebar.tsx    # 侧边栏（按 menuStore 动态生成）
│   ├── data-table.tsx     # tanStack 表格通用封装
│   ├── page-layout.tsx    # 全局布局
│   ├── login-form.tsx     # 共享登录表单
│   ├── identity-picker-dialog.tsx  # 多身份选择对话框
│   ├── tenant-switcher.tsx  # 租户切换
│   └── ...
├── locales/zh-CN.ts       # 简体中文文案（`export const zhCN = { ... }`）
├── pages/                 # 24 个页面文件
├── stores/
│   ├── authStore.ts       # zustand + persist（token / user / scope / identities）
│   ├── menuStore.ts       # 菜单数据（merged platform + tenant）
│   ├── configStore.ts     # 配置中心数据
│   └── permissionStore.ts # 权限数据
├── types/schema.ts        # FormSchema / FormItemSchema / TableSchema / ...
└── App.tsx                # 路由（RequireScope guard）
```

## 3. 关键约定

### 3.1 文案（简体中文）

- 唯一文案源：`src/locales/zh-CN.ts`，导出 `export type LocaleKeys = typeof zhCN`。
- `src/locales/index.ts` 直接把 `zhCN` 重新导出为 `t`：`import { t } from "@/locales"`。
- 用法：`t.pages.users.title`（对象访问，无 hook、无 store）。
- 不再做语言切换、不要 `useTranslation()`；`localeStore.ts` 与 `language-switcher.tsx` 已删除。
- 现有 `t.pages.users?.name || "姓名"` 这种 optional chaining + 兜底写法可以保留；新增文案直接 `t.xxx.yyy` 即可。

### 3.2 Schema 驱动

- 表单：`FormSchema { items: FormItemSchema[] }`；用 `<FormDialog schema={...} initialValues={...} onSubmit={...} />`
- 表格：`TableSchema { columns: ColumnSchema[]; search?: SearchSchema[]; actions?: ActionButton[] }`
- `showIf`：`{ dependsOn, equals, in, ... }` 条件显示
- `options: { label, value }[]`：注意 `value` 是 `string | number`，前端用 `String(value)` 转换

### 3.3 API 客户端模式

- `api/common.ts` 中的 `api()` 函数：自动加 `Authorization: Bearer <token>`，自动 401 refresh
- 端点前缀：`${VITE_API_BASE_URL}`（默认 `http://localhost:8087/api/v1`）
- 标准方法：`list(params?) / get(id) / create(data) / update(id, data) / delete(id)`
- 子资源（如 dict items / config items）：`listItems(parentId) / createItem(parentId, data) / updateItem(parentId, id, data) / deleteItem(parentId, id)`
- 返回 `data` 字段；分页用 `PageResponse<T> { list, total, page, size }`
- 错误处理：抛 `ApiError(status, code, message, data)`，前端用 `try { ... } catch (e) { ... }` 兜底
- 后端默认端口 **8087**，通过 `VITE_API_BASE_URL` 配置
- 前端 dev server 端口 **5241**（`vite.config.ts` 配置）

### 3.4 Page 结构模板

```tsx
export function XxxPage() {
  const [list, setList] = useState<Xxx[]>([])
  const [isLoading, setIsLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)
  const [useMockFallback, setUseMockFallback] = useState(false)
  const [isSubmitting, setIsSubmitting] = useState(false)
  const [dialogOpen, setDialogOpen] = useState(false)
  const [dialogMode, setDialogMode] = useState<"add" | "edit">("add")
  const [currentItem, setCurrentItem] = useState<Xxx | null>(null)
  const [deleteOpen, setDeleteOpen] = useState(false)
  const [toDelete, setToDelete] = useState<Xxx | null>(null)

  // 1) fetch + try/catch + setError + 可选 mock fallback（用户主动）
  // 2) form schema (useMemo 依赖 dialogMode)
  // 3) handlers: add / edit / delete confirm / submit / actual delete
  // 4) render: <PageLayout> + Card + Table + FormDialog + 删除确认 Dialog + 顶部 ErrorBar
}
```

### 3.5 左右两栏（树 + 列表）

参考 `Users.tsx`：

- `grid grid-cols-1 lg:grid-cols-[320px_1fr]`
- 左栏：递归 `OrgTreeView` 组件 + 搜索 + 计数徽章
- 右栏：筛选状态条 + 表格
- 选中态：`bg-primary/10 text-primary` + hover

### 3.6 树视图

参考 `Organizations.tsx` 的 `OrgTreeRow`：

- 递归组件 props：`item, level, expandedIds, onToggle, onEdit, onDelete, onAddChild`
- 缩进：`style={{ paddingLeft: `${level * 24}px` }}`
- 展开/收起：`expandedIds.has(item.id)` 或搜索时强制展开
- icon：根用 `NetworkIcon`，子用 `Building2Icon`

### 3.7 字典（dict）专用

- 后端 `dictpkg` 内存缓存，前端一般不直读（除非有 `/api/v1/dicts/:code` 公共端点）
- 维护页 `Dicts.tsx`：左右两栏，左字典列表 + 右字典项管理
- 关键点：
  - 字典编码在编辑时禁用（不可改）
  - 删除字典前会先检查字典项；后端返 409
  - `item_count` 字段在新建/删除字典项时本地同步增减（mock 模式）

### 3.8 表单对话框

- 用 `FormDialog`（来自 `@/components/schema/DynamicForm`）
- 内部用 `formKey` 自增（每次 `open` 触发）来重置表单
- 字段类型：`text / number / select / radio / checkbox / switch / date / datetime / icon / divider / slot`

## 4. 关键文件索引

| 关注点 | 路径 |
|---|---|
| 文案（zh-CN） | `src/locales/zh-CN.ts` |
| 路由 + scope guard | `src/App.tsx` |
| API 封装（base） | `src/api/common.ts` |
| 全部 API 端点 | `src/api/`（18 个文件） |
| 侧边栏（动态） | `src/components/app-sidebar.tsx` |
| 动态表单 | `src/components/schema/DynamicForm.tsx` |
| 动态表格 | `src/components/schema/DynamicTable.tsx` |
| 通用数据表格 | `src/components/data-table.tsx` |
| 全局布局 | `src/components/page-layout.tsx` |
| 按钮级权限 | `src/components/permission/Auth.tsx` |
| 路由级权限 | `src/components/permission/DynamicRouter.tsx` |
| Schema 类型 | `src/types/schema.ts` |
| 用户-组织模板 | `src/pages/Users.tsx` |
| 组织树模板 | `src/pages/Organizations.tsx` |
| 字典维护 | `src/pages/Dicts.tsx` |
| 配置中心模板 | `src/pages/Configs.tsx` |
| 菜单模板 | `src/pages/Menus.tsx` |
| 角色管理模板 | `src/pages/Roles.tsx` |
| 租户管理模板 | `src/pages/Tenants.tsx` |
| Auth store（本地持久化） | `src/stores/authStore.ts` |
| Menu store | `src/stores/menuStore.ts` |
| Config store | `src/stores/configStore.ts` |
| Permission store | `src/stores/permissionStore.ts` |

## 5. 踩坑与决策

### 5.1 编码（最重要！）

- **PowerShell 终端默认 GBK**：用 `python -` + here-string 写中文文件会被 mangle 成 `?`。
- 可靠方案：用 `[System.IO.File]::WriteAllText($path, $content, [System.Text.UTF8Encoding]::new($false))` 直接写 UTF-8 无 BOM。
- Vite/esbuild 对 **UTF-8 无 BOM** 期望；**有 BOM** 会出诡异错误。
- 仓库提供 [server/scripts/strip_bom.py](../server/scripts/strip_bom.py)：
  ```bash
  python ../server/scripts/strip_bom.py --check .   # CI gate
  python ../server/scripts/strip_bom.py .            # 修复
  ```

### 5.2 别名导入冲突

- 引入类型时若和本地变量重名，用 `import { type DictItem as Dict } from "@/api"`。
- 否则 `Dict` 既是类型又是本地变量，TypeScript 报 "Duplicate identifier"。

### 5.3 FormDialog 重置

- `FormDialog` 内部用 `formKey` 计数；`open=true` 时自增，触发 `DynamicForm` remount 重置。
- 父组件不要缓存 `initialValues` 到 useMemo 跨对话框生命周期。

### 5.4 子树筛选

- 用 `collectOrgSubtreeIds` 递归收集后代 id（见 `Users.tsx`）。
- 不要在后端用 SQL 子树筛选（除非显式 `?org_subtree=1`），保持接口简洁。

### 5.5 类型扩展

- 增强既有类型（如 `UserItem.org_id: number | null`）时，要同步影响所有 `useState<...[]>` 的初始化。
- 用 `as OrgNode[]` 类型断言很常见（递归树用 `type Tree = X & { children?: Tree[] }`）。

### 5.6 tsc 验证

- 改完跑 `.\node_modules\.bin\tsc --noEmit`（PowerShell 沙箱 `npx` 因签名问题直接调用 tsc 二进制）。
- 0 错误才能算完成。

### 5.7 前端权限

- 路由级：`RequireScope` 组件按 scope（`"tenant"` | `"platform"`）拦截。
- 按钮级：`<Auth action="create">...</Auth>` 包装。
- 资源权限由后端 `middleware.Require` 强制，前端只是隐藏。

### 5.8 多身份登录（Path B）

- 支持两种身份类型：**租户身份**（`scope="tenant"`）和**平台身份**（`scope="platform"`）
- 登录流程：
  1. `POST /auth/login-precheck` → 返回 `{ identities: [{tenant_id, tenant_name, ...}], platform_roles: [...] }`
  2. 用户选择身份 → `POST /auth/select-tenant` 或 `POST /auth/platform-login`
  3. 或 `POST /auth/tenant-login` 直接登录（如果只有一个身份）
- `authStore` 管理 `availableIdentities`、`platformAvailable`、`availablePlatformRoles`、`accountId`
- `switchTenant(tenantId)` 使用 `refresh_token` 无密码切换租户

### 5.9 Mock 兜底约定（重要变更）

- **不再静默 mock 兜底**：之前的模式 `try { api } catch { setX(mockX) }` 已被废弃。
- **新约定**：
  1. catch 内必须 `setError(message)`，UI 顶部显示红色错误条
  2. mock 仅在 `useMockFallback` 状态为 true 时才使用（用户主动勾选）
  3. 顶部加"实时数据 / Mock 数据"徽章明确当前数据源
  4. mock 开关同步到 `localStorage.<key>_use_mock`，跨会话保留
  5. 错误条带 Retry 按钮
- **示例**：见 `Dicts.tsx` 的 `fetchDicts` / `fetchItems` / `useMockFallback` / `error` / `dataSource`。
- 后续页面（Users / Menus / Roles ...）如有 mock fallback 需同步改造。

## 6. 常用配方

### 新增一个 CRUD 页面

1. 在 `api/xxx.ts` 加 `xxxApi = { list, get, create, update, delete }`
2. 在 `App.tsx` 加 `lazy(() => import("@/pages/Xxx"))` + `<Route path="/xxx" element={<XxxPage />} />`
3. 在 `zh-CN.ts` 加 `pages.xxx` 块（无需再同步其他语言）
4. 在 `migrations/init_seed.sql` 加菜单和资源（参考现有 seed 格式）
5. 写 `pages/Xxx.tsx`：
   - fetch + try/catch + setError
   - useMockFallback state（localStorage 持久化）
   - 顶部 ErrorBar + 数据源徽章
   - form schema（useMemo 依赖 dialogMode）
   - table + delete confirm + FormDialog

### 改既有文案

1. 改 `zh-CN.ts`
2. `tsc --noEmit` 验证（确保所有引用点类型仍合法）

### 给后端端点加前端 API

1. `api/xxx.ts` 加新方法（沿用 `api(path, { method, body })` 模板）
2. 类型在 `ApiResponse<T>` / `PageResponse<T>` 上声明返回类型
3. 出错时必须 `setError(...)`，**不要静默 mock fallback**（除非 `useMockFallback=true`）

### 后端模块 ↔ 前端页面映射

| 后端模块 | 前端页面 | 路由 | 前端 API | 后端路径 |
|---|---|---|---|---|
| auth | Login.tsx / TenantLogin.tsx / PlatformLogin.tsx | `/login`, `/platform/login` | `authApi` | `/auth/*` |
| user | Users.tsx | `/app/users` | `userApi` | `/users/*` |
| role | Roles.tsx | `/app/roles` | `roleApi` | `/roles/*` |
| menu（租户域） | Menus.tsx（Tab: 租户菜单） | `/app/menus` | `menuApi` | `/menus/*` |
| menu（平台域） | Menus.tsx（Tab: 平台菜单，仅 super_admin） | `/app/menus` | `platformMenuApi` | `/platform/menus/*` |
| organization | Organizations.tsx | `/app/organizations` | `organizationApi` | `/organizations/*` |
| resource | Resources.tsx | `/app/resources` | `resourceApi` | `/resources/*` |
| asset | Assets.tsx | `/app/asset` | `assetApi` | `/asset/*` |
| dict | Dicts.tsx | `/app/dicts` | `dictApi` | `/dicts/*` |
| config | Configs.tsx / PlatformConfigs.tsx | `/app/configs`, `/platform/configs` | `configApi` | `/configs/*`, `/platform/configs/*`, `/public/configs` |
| flag | Frames.tsx / FrameCategories.tsx / Avatars.tsx / AvatarCategories.tsx | `/app/frames`, etc | `frameApi` 等 | `/flag/*` |
| cms | — | — | — | `/cms/*` |
| **tenants**（仅 super_admin） | Tenants.tsx | `/platform/tenants` | `tenantApi` | `/platform/tenants/*` |
| system | Cache.tsx | `/platform/cache` | `systemApi` | `/platform/system/cache/*` |
| weixin | （无独立页面） | — | — | `/weixin/*` |
| sys_user | （通常通过平台管理页面） | `/platform/users` | — | `/platform/sys-users/*` |
| sys_role | （平台角色） | `/platform/roles` | — | `/platform/sys-roles/*` |
| sys_menu | Menus.tsx（平台 Tab） | `/platform/menus` | `platformMenuApi` | `/platform/menus/*` |

> **关键约定**：
>
> - 前端路由带 scope 前缀：`/app/*`（tenant 域）、`/platform/*`（平台域）
> - 同一前端页面可能调多个后端 API（如 `Menus.tsx` 同时调 `menuApi` 和 `platformMenuApi`）
> - `super_admin` 判断：前端用 `useAuthStore().user?.platform_roles?.includes("super_admin")`；后端用 `RequirePlatformRole("super_admin")` 中间件
>
> **路由约定**（与后端 [server/framework/framework.go](../server/framework/framework.go) 同步）：
>
> | 域 | 前缀 | 说明 |
> |---|---|---|
> | public | `/api/v1/public/*` 或 `/api/v1/<auth>` | 公开读 |
> | tenant（业务） | `/api/v1/*` | 需登录 + tenant_id |
> | platform（super_admin） | `/api/v1/platform/*` | 平台域 CRUD |
