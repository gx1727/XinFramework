# XinFramework UI

> React 19 + TypeScript + Vite + shadcn/ui。Schema 驱动的 CRUD 框架。

## 快速上手

```bash
# 安装依赖
npm install

# 开发模式（前端 5241）
npm run dev

# 类型检查
npm run typecheck

# 生产构建
npm run build
```

默认连 `http://localhost:8087/api/v1` 后端（通过 `VITE_API_BASE_URL` 配置）：

```bash
# .env
VITE_API_BASE_URL=http://localhost:8087/api/v1
VITE_ASSET_BASE_URL=http://localhost:8087
```

> 前端 dev server 端口是 **5241**（避免 Windows Hyper-V 保留端口范围），后端端口 **8087**。

## 技术栈

| 关注点 | 选型 |
| --- | --- |
| 构建 | Vite 7 + React 19 + TypeScript 5.9 |
| 样式 | Tailwind CSS v4 + shadcn/ui（`src/components/ui/`） |
| 图标 | lucide-react |
| 路由 | react-router-dom v7（scope-based guard） |
| 状态 | zustand v5（authStore 持久化到 localStorage） |
| i18n | 自研（仅 zh-CN，类型安全） |
| HTTP | 原生 `fetch` + `ApiError`（带 JWT 自动 refresh + 重试） |
| 表格 | @tanstack/react-table v8 |
| 图表 | Recharts v3 |
| 表单验证 | Zod v4 |
| 通知 | Sonner |

## 目录结构

```
UI/src/
├── api/
│   ├── common.ts        # fetch 封装 + ApiError + JWT refresh + 重试
│   ├── index.ts         # 重导出
│   ├── auth.ts          # 认证 API
│   ├── user.ts / role.ts / menu.ts / ...
│   └── ...              # 19 个 API 模块
├── components/
│   ├── ui/              # shadcn 组件（button / card / dialog / table / select / ...）
│   ├── schema/          # DynamicForm + DynamicTable + showIfEvaluator
│   ├── app-sidebar.tsx  # 侧边栏（按 menuStore 动态生成）
│   ├── page-layout.tsx  # 全局布局
│   ├── permission/      # Auth.tsx（按钮级权限）+ DynamicRouter.tsx（路由级）
│   └── ...
├── locales/{zh-CN}.ts   # 简体中文文案源
├── pages/               # 24 个页面（Users / Roles / Dicts / Configs / Flags / ...）
├── stores/{authStore,menuStore,configStore,permissionStore}.ts
├── types/schema.ts      # FormSchema / FormItemSchema / TableSchema / ...
└── App.tsx              # 路由（lazy load + RequireScope guard）
```

详见 [UI/AGENTS.md](AGENTS.md) — 给 AI agent / Codex 看的速查。

## 关键约定

### i18n

- `zh-CN.ts` 是类型源头：`export type LocaleKeys = typeof zhCN`
- 加新 key **先加 zh-CN**（否则 `LocaleKeys` 不包含，TypeScript 报红）
- 用 `t.<key>` 对象访问，无 hook、无 store

### Schema 驱动

- 表单：`FormSchema { items: FormItemSchema[] }`
- 表格：`TableSchema { columns, search?, actions? }`
- 字段类型：`text / number / select / radio / checkbox / switch / date / icon / divider / slot`

### API 客户端

```typescript
import { userApi, configApi, dictApi } from "@/api"

const list = await userApi.list({ page: 1, size: 20 })
const user = await userApi.create({ code: "u001", name: "张三" })

// 配置中心（与后端三域路由对应）
const groups = await configApi.listGroups()                    // GET /api/v1/configs
const items  = await configApi.listItemsByGroup(groupId)        // GET /api/v1/configs/:id/items
const pub    = await configApi.getPublic("site")                // GET /api/v1/public/configs
// 平台域（super_admin）
await configApi.createPlatformGroup({ code: "site", name: "站点" }) // POST /api/v1/platform/configs

// 字典
const dicts = await dictApi.list({ page: 1, size: 20 })
```

错误抛 `ApiError(status, code, message, data)`。

### Mock 数据（不再静默兜底）

新约定（详见 [UI/AGENTS.md §5.9](AGENTS.md#59-mock-兜底约定重要变更)）：

1. `catch` 内必须 `setError(message)`，UI 顶部显示红色错误条
2. mock 仅在 `useMockFallback` 状态为 `true` 时才使用（用户主动勾选）
3. 顶部加"实时数据 / Mock 数据"徽章
4. mock 开关同步到 `localStorage.<key>_use_mock`
5. 错误条带 Retry 按钮

## 新增一个 CRUD 页面

1. 在 `api/xxx.ts` 加 `xxxApi = { list, get, create, update, delete }`
2. 在 `zh-CN.ts` 加 `pages.xxx` 块（先加，作为类型源头）
3. 在 `App.tsx` 加 `lazy(() => import("@/pages/Xxx"))` + `<Route path="/xxx" element={<XxxPage />} />`
4. 写 `pages/Xxx.tsx`：`fetch + form + table + dialog + error toast + mock toggle`

详见 [UI/AGENTS.md §6](AGENTS.md#6-常用配方)。

## 后端配合

后端在 [server/](../server) 下。详见：

- [server/doc/quickstart.md](../server/doc/quickstart.md) — 启动后端
- [server/doc/api.md](../server/doc/api.md) — API 端点
- [server/doc/developing.md](../server/doc/developing.md) — 新增后端模块
- [server/doc/database.md](../server/doc/database.md) — 数据库表结构

## 编码

**所有源文件 UTF-8 无 BOM**。PowerShell 默认 GBK，写入用：

```powershell
[System.IO.File]::WriteAllText($path, $content, [System.Text.UTF8Encoding]::new($false))
```

详见 [UI/AGENTS.md §5.1](AGENTS.md#51-编码最重要)。

仓库提供 [server/scripts/strip_bom.py](../server/scripts/strip_bom.py) 用于检测 / 剥离 BOM：

```bash
python ../server/scripts/strip_bom.py --check .   # CI gate
python ../server/scripts/strip_bom.py .            # 修复
```

## TypeScript 验证

```bash
.\node_modules\.bin\tsc --noEmit    # 0 错误才能算完成
```
