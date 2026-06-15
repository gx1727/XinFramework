# XinFramework UI

> React 18 + TypeScript + Vite + shadcn/ui。Schema 驱动的 CRUD 框架。

## 快速上手

```bash
# 安装依赖
npm install

# 开发模式（前端 5173）
npm run dev

# 类型检查
npm run typecheck

# 生产构建
npm run build
```

默认连 `http://localhost:8080` 后端（通过 `VITE_API_BASE_URL` 配置）。

```bash
# .env.local
VITE_API_BASE_URL=http://localhost:8080/api/v1
```

## 技术栈

| 关注点 | 选型 |
| --- | --- |
| 构建 | Vite + React 18 + TypeScript |
| 样式 | Tailwind CSS + shadcn/ui（`src/components/ui/`） |
| 图标 | lucide-react |
| 路由 | react-router-dom v6 |
| 状态 | zustand |
| i18n | 自研 |
| HTTP | 原生 `fetch` + `ApiError`（带 JWT 自动 refresh） |

## 目录结构

```
UI/src/
├── api/                # API 客户端
│   ├── client.ts       # ApiError + 所有 *Api
│   └── index.ts        # 重导出
├── components/
│   ├── ui/             # shadcn 组件
│   ├── schema/         # DynamicForm / DynamicTable
│   ├── app-sidebar.tsx # 侧边栏（按 menuStore 动态生成）
│   └── page-layout.tsx # 全局布局
├── locales/
│   ├── zh-CN.ts        # i18n 类型源头
│   └── en-US.ts        # 必须满足 zh-CN 的所有 key
├── pages/              # 每个模块一个 .tsx
├── stores/             # zustand
├── types/schema.ts     # FormSchema / TableSchema
└── App.tsx             # 路由
```

详见 [UI/AGENTS.md](file:///d:\work\xin\XinFramework\UI\AGENTS.md) — 给 AI agent / Codex 看的速查。

## 关键约定

### i18n

- `zh-CN.ts` 是类型源头：`export type LocaleKeys = typeof zhCN`
- 加新 key 先加 `zh-CN.ts`
- 用 `t.<key> || "fallback"` 兜底

### Schema 驱动

- 表单：`FormSchema { items: FormItemSchema[] }`
- 表格：`TableSchema { columns, search?, actions? }`
- 字段类型：`text / number / select / radio / checkbox / switch / date / icon / divider / slot`

### API 客户端

```typescript
import { userApi } from "@/api"

const list = await userApi.list({ page: 1, size: 20 })
const user = await userApi.create({ code: "u001", name: "张三" })
```

错误抛 `ApiError(status, code, message, data)`。

### Mock 数据

每个新页面都加 `mockXxx` 兜底数组。失败时回退，让页面能跑通。详见 [UI/AGENTS.md §5.5](file:///d:\work\xin\XinFramework\UI\AGENTS.md#55-mock-数据)。

## 新增一个 CRUD 页面

1. 在 `client.ts` 加 `xxxApi = { list, get, create, update, delete }`
2. 在 `zh-CN.ts` 加 `pages.xxx` 块（先加，作为类型源头）
3. 在 `en-US.ts` 同步
4. 在 `App.tsx` 加 `lazy(() => import("@/pages/Xxx"))` + `<Route path="/xxx" element={<XxxPage />} />`
5. 写 `pages/Xxx.tsx`：`fetch + form + table + dialog + mock fallback`

详见 [UI/AGENTS.md §6](file:///d:\work\xin\XinFramework\UI\AGENTS.md#6-常用配方)。

## 后端配合

后端在 [server/](../server) 下。详见：

- [server/doc/quickstart.md](../server/doc/quickstart.md) — 启动后端
- [server/doc/api.md](../server/doc/api.md) — API 端点
- [server/doc/developing.md](../server/doc/developing.md) — 新增后端模块