# 快速上手

> 本文件描述 XinFramework 的安装、配置、首次启动、常用开发命令。

---

## 1. 前置要求

| 工具 | 版本 | 说明 |
|---|---|---|
| Go | 1.24+（推荐 1.26.2） | 后端编译与运行 |
| Node.js | 20+ | 前端开发与构建 |
| PostgreSQL | 14+ | 主数据库（需要 `ltree` 和 `pg_trgm` 扩展） |
| Redis | 7+（可选） | 缓存 + Session；不可用时降级到 DB session |
| Git | 任意 | 拉取代码 |

**PostgreSQL 扩展**（启动期自动创建）：
- `ltree` — 物化路径（组织树）
- `pg_trgm` — 三元组索引（模糊搜索）

---

## 2. 准备数据库

```bash
# 启动 PostgreSQL（Docker 示例）
docker run --name xin-pg \
  -e POSTGRES_PASSWORD=dev \
  -p 5432:5432 \
  -d postgres:16

# 创建数据库与用户
psql -U postgres -c "CREATE USER xin_user WITH PASSWORD 'xin_password';"
psql -U postgres -c "CREATE DATABASE xin OWNER xin_user;"
psql -U postgres -d xin -c "GRANT ALL PRIVILEGES ON DATABASE xin TO xin_user;"
```

或者直接用 `config/config.yaml` 默认的 `xin_user` / `xin_password` / `xin`。

---

## 3. 后端启动

```bash
# 1. 进入后端目录
cd server

# 2. 准备 .env（首次启动注入 bootstrap admin）
cp .env.example .env

# 3. 启动（前台模式）
go run ./cmd/xin run

# 或构建后跑
go build -o bin/xin ./cmd/xin
./bin/xin run
```

**首次启动会自动**：
- 跑 `migrate.Run` 扫描 `migrations/*.sql`
- 执行 `init_schema.sql` 建表 + RLS policy
- 执行 `init_seed.sql` 插入种子数据
- 执行 `asset.sql` / `cms.sql` / `flag.sql`（如果启用对应模块）

**关键日志**：
```
module auth initialized
module sys_tenant initialized
module sys_user initialized
...
server starting on 0.0.0.0:8087
```

**健康检查**：
```bash
curl http://localhost:8087/api/v1/health
# → {"code":0,"msg":"ok","data":{"status":"ok"}}
```

---

## 4. 前端启动

```bash
# 1. 进入前端目录
cd UI

# 2. 安装依赖
npm install

# 3. 启动开发服务器（:5241）
npm run dev
```

**默认 `.env`**（如需修改）：
```bash
VITE_API_BASE_URL=http://localhost:8087/api/v1
VITE_ASSET_BASE_URL=http://localhost:8087
```

**前端 dev 端口是 5241**（避免 Windows Hyper-V 保留端口范围），后端端口是 8087。

---

## 5. 验证

### 5.1 浏览器访问

打开 `http://localhost:5241`，应该跳转到 `/login`。

### 5.2 默认登录账号

- `/sys/login` → sys 管理后台
- `/login` → 业务后台（需要选择 tenant_id）

### 5.3 API 调试

```bash
# sys 登录（super_admin）
curl -X POST http://localhost:8087/api/v1/auth/sys-login \
  -H "Content-Type: application/json" \
  -d '{"account":"admin","password":"admin123"}'

# 兼容期：旧 /auth/platform-login 仍可用，自动转发到 sys-login

# 业务登录（需要先有 tenant_id）
curl -X POST http://localhost:8087/api/v1/auth/tenant-login \
  -H "Content-Type: application/json" \
  -d '{"account":"admin","password":"admin123","tenant_id":1}'
```

---

## 6. 进程管理命令

`cmd/xin/main.go` 同时是守护命令入口：

```bash
# 启动（写 PID 到 ./xin.pid，日志到 ./xin.log）
./bin/xin start

# 查看状态
./bin/xin status
# → PID: 12345
# → 最近 5 行日志 ...

# 优雅停止（SIGTERM，30s 超时）
./bin/xin stop

# 重启（stop + start）
./bin/xin restart

# 零停机重载（SIGUSR1）
./bin/xin reload

# 热重启（起新进程 + 停老进程）
./bin/xin hot-restart

# 前台运行（开发用）
./bin/xin run
```

Windows 等价物：`build.ps1` 包装 `go build`。

---

## 7. 常用开发命令

```bash
# 类型检查（前端）
cd UI
.\node_modules\.bin\tsc --noEmit

# Lint（前端）
npm run lint

# 格式化（前端）
npm run format

# 构建（前端）
npm run build

# 构建（后端）
cd ../server
go build -o bin/xin ./cmd/xin
go build -o bin/rotate-admin-password ./cmd/rotate-admin-password

# BOM 检测（CI gate）
python scripts/strip_bom.py --check .

# BOM 修复
python scripts/strip_bom.py .
```

---

## 8. 配置文件

`server/config/config.yaml` 主配置（dev 默认值），`config.dev.yaml` / `config.prod.yaml` 按 `app.env` 叠加。

加载顺序：主文件 → `config.{env}.yaml` → `.env` → 环境变量 `XIN_*`（覆盖）。

### 8.1 关键配置项

```yaml
app:
  name: xin
  env: dev                    # dev | prod（prod 强校验 JWT secret）
  host: 0.0.0.0
  port: 8087

database:
  host: localhost
  port: 5432
  user: xin_user
  password: xin_password
  dbname: xin
  sslmode: disable
  max_open_conns: 100
  max_idle_conns: 20
  conn_max_lifetime_sec: 300
  conn_max_idle_time_sec: 60

redis:
  host: localhost
  port: 6379
  enabled: true              # false → 关闭；true+required=false → 降级
  required: false            # true → 不可用时启动失败

jwt:
  secret: your-secret-key    # prod 强制 ≥32 字节非占位
  expire: 3600               # 1h
  refresh_expire: 86400      # 24h

storage:
  provider: local            # local | cos
  local_dir: ./uploads
  local_base_url: /uploads
  cos_url: ""                # cos：https://<bucket>.cos.<region>.myqcloud.com
  cos_secret_id: ""          # 建议用 XIN_STORAGE_COS_SECRET_ID 注入
  cos_secret_key: ""
  cos_base_url: ""

log:
  dir: logs
  level: info                # debug | info | warn | error

module:
  - weixin                   # 默认启用：alwaysOn + optOut
  - cms                      # 可选：删除即不启用
  - flag
```

### 8.2 生产环境必填

通过环境变量注入（**不要**写在 `config.prod.yaml`）：

```bash
# 强校验：secret ≥32 字节、非占位符
export XIN_JWT_SECRET=$(openssl rand -base64 48)

# 数据库凭据
export XIN_DATABASE_HOST=...
export XIN_DATABASE_USER=...
export XIN_DATABASE_PASSWORD=...

# 存储凭据（如启用 COS）
export XIN_STORAGE_COS_SECRET_ID=...
export XIN_STORAGE_COS_SECRET_KEY=...

# 微信（启用 weixin 模块时）
export XIN_WEIXIN_APPID=...
export XIN_WEIXIN_APPSECRET=...
```

**禁止的占位值**（`config.validateJWTSecret` 在 `app.env=prod` 时检查）：
- `your-secret-key` / `changeme` / `please-change-me` / `secret` / `12345678`
- 空字符串

---

## 9. 常见问题

### 9.1 启动失败：JWT secret 校验失败

```
[FATAL] jwt secret invalid: must not be placeholder
```

→ 生产环境必须设置非占位、≥32 字节的 `XIN_JWT_SECRET`。

### 9.2 启动失败：extension not found

```
ERROR: extension "ltree" is not available
```

→ 需要 PostgreSQL 安装时带 `ltree` 扩展（`postgres:16` 默认包含）。

### 9.3 Redis 不可用

```
[INFO] redis ping failed, falling back to DB session
```

→ 正常：`redis.required=false` 时自动降级。改为 `required=true` 才会启动失败。

### 9.4 启动报 RLS policy 错误

```
ERROR: policy "tenant_isolation_policy" already exists
```

→ 迁移文件已用 `DROP POLICY IF EXISTS` 幂等保护。检查是否手动改过 schema。

### 9.5 前端端口冲突

Vite 默认 5173，但本项目用 **5241**（避免 Windows Hyper-V 保留端口）。如需修改：编辑 `UI/vite.config.ts` 的 `server.port`。

### 9.6 前端 API 报 CORS 错误

检查 `config/config.yaml` 的 `cors.allow_origins` 是否包含前端源：

```yaml
cors:
  enabled: true
  allow_origins:
    - "http://localhost:5241"
    - "http://127.0.0.1:5241"
  allow_credentials: true     # 注意：true 时 allow_origins 不能为 "*"
```

### 9.7 启动期卡住

检查 `migrations/` 目录是否可读；首次启动会按文件名字母序跑所有 SQL，时间略长属正常。

---

## 10. 下一步

- [architecture.md](./architecture.md) — 理解代码层架构
- [modules.md](./modules.md) — 查看 19 个模块清单
- [database.md](./database.md) — 理解表结构与 RLS
- [developing.md](./developing.md) — 新增模块的 8 步流程
