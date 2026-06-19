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

默认连 `http://localhost:8080` 后端（通过 `VITE_API_BASE_URL` 配置）：

```bash
# .env.local
VITE_API_BASE_URL=http://localhost:8087/api/v1
```

> 默认端口是 **8087**（不是 8080），后端 `config/config.yaml` 的 `app.port`。

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
├── api/
│   ├── client.ts        # ApiError + 全部 *Api（userApi / orgApi / dictApi / configApi / flagApi / ...）
│   └── index.ts         # 重导出
├── components/
│   ├── ui/              # shadcn 组件（button / card / dialog / table / select / ...）
│   ├── schema/          # DynamicForm + DynamicTable + showIfEvaluator
│   ├── app-sidebar.tsx  # 侧边栏（按 menuStore 动态生成）
│   ├── page-layout.tsx  # 全局布局（Auth + Sidebar + Header）
│   └── ...
├── locales/{zh-CN,en-US}.ts
├── pages/               # 每个模块一个 .tsx（Users / Roles / Dicts / Configs / Flags / ...）
├── stores/{authStore,menuStore}.ts
├── types/schema.ts      # FormSchema / FormItemSchema / TableSchema / ...
└── App.tsx              # 路由
```

详见 [UI/AGENTS.md](file:///d:/work/xin/XinFramework/UI/AGENTS.md) — 给 AI agent / Codex 看的速查。

## 关键约定

### i18n

- `zh-CN.ts` 是类型源头：`export type LocaleKeys = typeof zhCN`
- 加新 key **先加 zh-CN**（否则 `LocaleKeys` 不包含，TypeScript 报红）
- `en-US.ts` 必须**满足 `LocaleKeys` 类型**，即每个 key 都要有
- 用 `t.<key> || "fallback"` 兜底

### Schema 驱动

- 表单：`FormSchema { items: FormItemSchema[] }`
- 表格：`TableSchema { columns, search?, actions? }`
- 字段类型：`text / number / select / radio / checkbox / switch / date / icon / divider / slot`

### API 客户端

```typescript
import { userApi, configApi, dictApi } from "@/api"

const list = await userApi.list({ page: 1, size: 20 })
const user = await userApi.create({ code: "u001", name: "张三" })

// 配置中心（与后端 /config/* 对应）
const items = await configApi.listItems({ group_code: "site" })

// 字典
const dicts = await dictApi.list({ page: 1, size: 20 })
```

错误抛 `ApiError(status, code, message, data)`。

### Mock 数据（不再静默兜底）

新约定（详见 [UI/AGENTS.md §5.9](file:///d:/work/xin/XinFramework/UI/AGENTS.md#59-mock-兜底约定重要变更)）：

1. `catch` 内必须 `setError(message)`，UI 顶部显示红色错误条
2. mock 仅在 `useMockFallback` 状态为 `true` 时才使用（用户主动勾选）
3. 顶部加"实时数据 / Mock 数据"徽章
4. mock 开关同步到 `localStorage.<key>_use_mock`
5. 错误条带 Retry 按钮

## 新增一个 CRUD 页面

1. 在 `client.ts` 加 `xxxApi = { list, get, create, update, delete }`
2. 在 `zh-CN.ts` 加 `pages.xxx` 块（先加，作为类型源头）
3. 在 `en-US.ts` 同步
4. 在 `App.tsx` 加 `lazy(() => import("@/pages/Xxx"))` + `<Route path="/xxx" element={<XxxPage />} />`
5. 写 `pages/Xxx.tsx`：`fetch + form + table + dialog + error toast + mock toggle`

详见 [UI/AGENTS.md §6](file:///d:/work/xin/XinFramework/UI/AGENTS.md#6-常用配方)。

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

详见 [UI/AGENTS.md §5.1](file:///d:/work/xin/XinFramework/UI/AGENTS.md#51-编码最重要)。

仓库提供 [server/scripts/strip_bom.py](../server/scripts/strip_bom.py) 用于检测 / 剥离 BOM：

```bash
python ../server/scripts/strip_bom.py --check .   # CI gate
python ../server/scripts/strip_bom.py .            # 修复
```

## TypeScript 验证

```bash
.\node_modules\.bin\tsc --noEmit    # 0 错误才能算完成
```
