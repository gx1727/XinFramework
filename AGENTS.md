# XinFramework — AI Agent 协作总则

> 这是给 AI agent 的**全局规则**。子项目细节请下沉到各自的 `AGENTS.md`:
> - 后端 Go → [`server/AGENTS.md`](server/AGENTS.md)(670 行高密度,**进 server/ 前必读**)
> - 前端 React → [`UI/AGENTS.md`](UI/AGENTS.md)(297 行,**进 UI/ 前必读**)
> - 项目总览 → [`README.md`](README.md)(给**人**看的功能/模块清单)

## 加载规则(就近原则)
- AI 在 `server/**` 工作 → 优先 `server/AGENTS.md`,根 `AGENTS.md` 作为兜底
- AI 在 `UI/**` 工作 → 优先 `UI/AGENTS.md`,根 `AGENTS.md` 作为兜底
- AI 在根目录或跨子项目 → 加载根 `AGENTS.md` + 涉及的子项目 `AGENTS.md`
- 本文件**不重复**子目录的细节;具体技术问题请去对应子目录看

## 项目速览
- 多租户 SaaS 后台框架,monorepo 结构
- 后端 Go 1.24 + Gin + pgx(单一 module `gx1727.com/xin`)
- 前端 React 19 + TypeScript + Vite + shadcn/ui
- 19 个业务模块,统一通过 `framework.Boot()` 注册,按 `cfg.Module` 白名单启停

## 目录边界(铁律)
- **根目录**:仅放配置/文档/`.zed/`/`AGENTS.md`/`README.md`,**不要在这里写 Go/TS 代码**
- **后端代码**:全部在 `server/`,Go module 路径 `gx1727.com/xin`
- **前端代码**:全部在 `UI/`
- 跨子项目改动 → 必须双读 `server/AGENTS.md` + `UI/AGENTS.md`,确保契约一致

## 全局硬约束(任何子项目都生效)

### 1. 文件编码 — UTF-8 无 BOM 且**非 GBK**(最重要)
- PowerShell / 记事本默认 GBK + BOM。某些 **AI agent 文件写入工具**也会无声无息中产生 GBK 字节流(通常无 BOM,strip_bom.py 也查不出来)。
- 任何非合法 UTF-8 字节都会让 Go 报 `illegal UTF-8 encoding`、TS 报 `Invalid byte sequence`。
- **一站式兜底脚本**:`.\fix_and_build.ps1`(仓库根)会自动 扫 GBK→剥 BOM→go build→npm typecheck。任何 batch 改动后可让用户跑一次。
- **细粒度**(只想跑编码扫描):
  - `python server/scripts/fix_encoding.py --check <dir>` —— 检 GBK/GB18030/Big5,带 UTF-8 roundtrip 校验
  - `python server/scripts/strip_bom.py --check <dir>` —— 检 UTF-8 BOM
- 写入文件务必用 UTF-8(无 BOM),见 `UI/AGENTS.md §5.1`
- **案例教训**:2026-06 platform→sys 重构中,sub-agent 编辑 `apps/task/{handler,cron_handler}.go` 后
  遗留 GBK 字节,Go build 跨多个文件连续炸。**根因:agent 写文件工具未强制声明 UTF-8**。
  任何 batch 改动 .go/.ts/.tsx/.json/.yaml/.md 后,**不要等用户报 build 错**,主动执行上面两条扫描。

### 2. 不动未读过的文件
- 打开看清楚再改,别凭文件名猜意图
- 改之前先 `grep` / `read_file` 确认上下文

### 3. 不动敏感文件(除非用户明确要求)
- `.env` / `*.pem` / `*.key` / 任何凭据
- `server/migrations/` 下的 SQL(迁移是不可变的,改它就是改历史)
- 任何 `*_test.go` 中正在跑的断言(除非用户说可以改)

### 4. 不引入新依赖
- Go 改 `go.mod` / 前端改 `package.json` 前必须说一声,给出理由
- 优先复用项目已有依赖

### 5. 保持现有风格
- 看周围 3 个文件模仿命名、缩进、错误处理
- Go:`gofmt` 默认;前端:跟随 `UI/` 的 ESLint/Prettier 配置

### 6. AI agent 写入文件后的强制编码自检
- **写动作可能产生 GBK 字节**(已知污染源:某些 AI 文件工具 / Windows PowerShell 重定向 / 记事本兼容层)。
- 改动 .go/.ts/.tsx/.json/.yaml/.md 后,**agent 应主动建议跑一次**上面 §1 中的两条扫描命令,
  或直接让用户跑 `.\fix_and_build.ps1`。**不要等用户 build 报错再行动**。
- 如果 `fix_encoding.py` 报 `can't decode` 或 `roundtrip-fail`,说明字节流已经损坏到无法安全自动转码,
  此时 **必须 `read_file` + `write_file` 整文件重写**(不能靠 `edit_file` 局部替换 — 局部字节已损坏,grep 也看不到)。
- sub-agent prompt 模板应固定带上:
  "**写入必须 UTF-8 无 BOM,违反 AGENTS.md §1 = bug。batch 末尾必须让用户跑 `python server/scripts/fix_encoding.py --check <改动范围>` 或 `.\fix_and_build.ps1`,发现坏文件立即 fix 或写整文件重写。**"

## 跨子项目改动协议
当一个改动同时影响 server 和 UI(改 API、改字段、改错误码):
1. 读 `server/AGENTS.md §11` 路由清单 + `server/doc/api.md`
2. 读 `UI/AGENTS.md` 了解前端 Schema / 错误条 / i18n 约定
3. **先后端,再前端**;后端改完告知用户"需重启后端"
4. 错误码遵循 `server/AGENTS.md §8` 分段,前端 i18n 同步补文案

## 工作流

| 动作 | 约定 |
|---|---|
| commit message | 中文,`<scope>: <动作> <对象>`,例:`server: 修复 user 分页空指针` |
| 是否 commit | **不主动 commit**,除非用户明确说"提交" |
| 改完自检 | 一键:`.\fix_and_build.ps1`(扫 GBK→剥 BOM→go build→npm typecheck);后端:`go build ./...`;前端:`npm run typecheck`;编码单独检:`python server/scripts/fix_encoding.py --check <dir> && python server/scripts/strip_bom.py --check <dir>` |
| sub-agent prompt 强制项 | 任何改 .go/.ts 任务,prompt 必须含"batch 末尾跑编码扫描"一句(见 §6) |
| 新增模块 | 严格按 `server/AGENTS.md §14` 8 步配方,不要自创结构 |
| 加新 API | 同步更新 `server/doc/api.md` + 前端 `UI/src/api/` |

## 文档同步
- 改了行为/接口 → 同步对应 `doc/*.md`
- 改了 `AGENTS.md` → 提醒用户,规则文件需要人审核
- `AGENTS.md` 本身不写功能描述(那是 README 的事)

## 遇到不确定时
**先问,再动手。** 不要猜测 API 形状、不要补全用户没说过的需求、不要"顺手"重构无关代码。
