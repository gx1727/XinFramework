# UI/AGENTS.md

> 给"未来的 Codex"看的 XinFramework 前端设计说明。读这一份，比重新读 18 个 .tsx 快得多。

---

## 1. 技术栈

- **构建**：Vite + React 18 + TypeScript
- **样式**：Tailwind CSS + shadcn/ui（`components/ui/`）
- **图标**：lucide-react
- **路由**：react-router-dom v6（`App.tsx` 集中 lazy）
- **状态**：zustand（`stores/authStore`, `stores/menuStore`）
- **文案**：仅简体中文；`UI/src/locales/zh-CN.ts`；`t = zhCN` 直接作为静态对象使用
- **HTTP**：原生 `fetch` + `ApiError`（带 JWT 自动 refresh）
- **图标选择**：lucide-react 优先；shadcn 默认有 30+ 常用图标

## 2. 目录结构

```
UI/src/
├── api/
│   ├── client.ts          # ApiError + 全部 *Api（userApi / orgApi / dictApi / configApi / flagApi / ...）
│   └── index.ts           # 重导出
├── components/
│   ├── ui/                # shadcn 组件（button / card / dialog / table / select / ...）
│   ├── schema/            # DynamicForm + DynamicTable + showIfEvaluator
│   ├── app-sidebar.tsx    # 侧边栏（按 menuStore 动态生成）
│   ├── page-layout.tsx    # 全局布局（Auth + Sidebar + Header）
│   └── ...
├── locales/zh-CN.ts       # 简体中文文案（`t = zhCN`，无 i18n 切换）
├── pages/                 # 每个模块一个 .tsx（Menus / Users / Roles / Dicts / Configs / Flags / ...）
├── stores/{authStore,menuStore}.ts
├── types/schema.ts        # FormSchema / FormItemSchema / TableSchema / ...
└── App.tsx                # 路由
```

## 3. 关键约定

### 3.1 文案（简体中文）

- 唯一文案源：`UI/src/locales/zh-CN.ts`，导出 `export type LocaleKeys = typeof zhCN`。
- `UI/src/locales/index.ts` 直接把 `zhCN` 重新导出为 `t`：`import { t } from "@/locales"`。
- 用法：`t.pages.users.title`（对象访问，无 hook、无 store）。
- 不再做语言切换、不要 `useTranslation()`；`localeStore.ts` 与 `language-switcher.tsx` 已删除。
- 现有 `t.pages.users?.name || "姓名"` 这种 optional chaining + 兜底写法可以保留；新增文案直接 `t.xxx.yyy` 即可。

### 3.2 Schema 驱动

- 表单：`FormSchema { items: FormItemSchema[] }`；用 `<FormDialog schema={...} initialValues={...} onSubmit={...} />`
- 表格：`TableSchema { columns: ColumnSchema[]; search?: SearchSchema[]; actions?: ActionButton[] }`
- `showIf`：`{ dependsOn, equals, in, ... }` 条件显示
- `options: { label, value }[]`：注意 `value` 是 `string | number`，前端用 `String(value)` 转换

### 3.3 API 客户端模式

- 端点前缀：`${VITE_API_BASE_URL}/api/v1/...`
- 标准方法：`list(params?) / get(id) / create(data) / update(id, data) / delete(id)`
- 子资源（如 dict items / config items）：`listItems(parentId) / createItem(parentId, data) / updateItem(parentId, id, data) / deleteItem(parentId, id)`
- 返回 `data` 字段；分页用 `PageResponse<T> { list, total, page, size }`
- 错误处理：抛 `ApiError(status, code, message, data)`，前端用 `try { ... } catch (e) { ... }` 兜底
- 后端默认端口 **8087**（不是 8080），通过 `VITE_API_BASE_URL` 配置

### 3.4 Page 结构模板

```tsx
export function XxxPage() {
  const t = useTranslation()
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
  // 2) form schema (useMemo 依赖 dialogMode 和 t)
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
| 文案（zh-CN） | `UI/src/locales/zh-CN.ts` |
| 路由 | `UI/src/App.tsx` |
| API 全部端点 | `UI/src/api/client.ts` |
| 侧边栏（动态） | `UI/src/components/app-sidebar.tsx` |
| 动态表单 | `UI/src/components/schema/DynamicForm.tsx` |
| 动态表格 | `UI/src/components/schema/DynamicTable.tsx` |
| 全局布局 | `UI/src/components/page-layout.tsx` |
| Schema 类型 | `UI/src/types/schema.ts` |
| 用户-组织模板 | `UI/src/pages/Users.tsx` |
| 组织树模板 | `UI/src/pages/Organizations.tsx` |
| 字典维护 | `UI/src/pages/Dicts.tsx` |
| 配置中心模板 | `UI/src/pages/Configs.tsx` |
| 菜单模板 | `UI/src/pages/Menus.tsx` |

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

### 5.5 mock 数据（已废弃静默兜底）

- 见 §5.9 新约定。

### 5.6 类型扩展

- 增强既有类型（如 `UserItem.org_id: number | null`）时，要同步影响所有 `useState<...[]>` 的初始化。
- 用 `as OrgNode[]` 类型断言很常见（递归树用 `type Tree = X & { children?: Tree[] }`）。

### 5.7 tsc 验证

- 改完跑 `.\node_modules\.bin\tsc --noEmit`（PowerShell 沙箱 `npx` 因签名问题直接调用 tsc 二进制）。
- 0 错误才能算完成。

### 5.8 前端权限

- 路由级：`DynamicRouter` 组件按 `useAuthStore.permissions` 拦截。
- 按钮级：`<Auth action="create">...</Auth>` 包装。
- 资源权限由后端 `middleware.Require` 强制，前端只是隐藏。

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

1. 在 `client.ts` 加 `xxxApi = { list, get, create, update, delete }`
2. 在 `App.tsx` 加 `lazy(() => import("@/pages/Xxx"))` + `<Route path="/xxx" element={<XxxPage />} />`
3. 在 `zh-CN.ts` 加 `pages.xxx` 块（无需再同步其他语言）
4. 在 `migrations/framework.sql` 加菜单和资源（参考现有 seed 格式）
5. 写 `pages/Xxx.tsx`：
   - fetch + try/catch + setError
   - useMockFallback state（localStorage 持久化）
   - 顶部 ErrorBar + 数据源徽章
   - form schema（useMemo 依赖 dialogMode，无需把 `t` 放进 deps）
   - table + delete confirm + FormDialog

### 改既有文案

1. 改 `zh-CN.ts`
2. `tsc --noEmit` 验证（确保所有引用点类型仍合法）

### 给后端端点加前端 API

1. `client.ts` 加新方法（沿用 `api(path, { method, body })` 模板）
2. 类型在 `ApiResponse<T>` / `PageResponse<T>` 上声明返回类型
3. 出错时必须 `setError(...)`，**不要静默 mock fallback**（除非 `useMockFallback=true`）

### 后端模块 ↔ 前端页面映射

| 后端模块 | 前端页面 | 路径 |
|---|---|---|
| auth | Login.tsx | `/login` |
| user | Users.tsx | `/users` |
| role | Roles.tsx | `/roles` |
| menu | Menus.tsx | `/menus` |
| organization | Organizations.tsx | `/organizations` |
| resource | Resources.tsx | `/resources` |
| asset | Assets.tsx | `/asset` |
| dict | Dicts.tsx | `/dicts` |
| config | Configs.tsx | `/config` |
| flag | FlagFrames.tsx / FlagSpaces.tsx / ... | `/flag/*` |
| cms | CmsPosts.tsx | `/cms/posts` |
| tenant | Tenants.tsx（仅 super_admin） | `/tenants` |
| system | SystemInfo.tsx | `/system` |
| weixin | （无独立页面） | — |
