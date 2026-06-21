# 快速开始

> 目标：15 分钟内跑通 `xin run`，在浏览器看到 health endpoint 返回 `{"status":"ok"}`。
>
> 文档版本：2026-06（config 重构 + platform_menu/platform_tenant 后）

## 1. 环境要求

| 组件 | 版本 | 用途 |
|---|---|---|
| Go | 1.25+ | 编译 |
| PostgreSQL | 14+ | 主存储，需要 `ltree` 和 `pg_trgm` 扩展 |
| Redis | 7+ | 可选，缓存/会话（`enabled: true` 时启用） |
| Make / Git | 任意 | 构建脚本（`./build.sh`） |

```bash
go version          # go1.25.x 或更新
psql --version      # PostgreSQL 14.x
redis-cli --version # redis 7.x (可选)
```

## 2. 准备 PostgreSQL

### 2.1 安装

```bash
# Ubuntu/Debian
sudo apt install -y postgresql-14 postgresql-contrib
sudo systemctl start postgresql

# macOS
brew install postgresql@14
brew services start postgresql@14

# Windows: 用 EnterpriseDB installer，默认监听 5432
# 或 Docker
docker run --name xin-pg -e POSTGRES_PASSWORD=dev -p 5432:5432 -d postgres:16
```

### 2.2 建库 + 用户

```bash
sudo -u postgres psql
```

```sql
CREATE DATABASE xin;
CREATE USER xin_user WITH PASSWORD 'xin_password';
GRANT ALL PRIVILEGES ON DATABASE xin TO xin_user;
\c xin
GRANT ALL ON SCHEMA public TO xin_user;
```

### 2.3 装扩展

`migrations/framework.sql` 会自动 `CREATE EXTENSION IF NOT EXISTS ltree / pg_trgm`，但需要**超级用户**才能装扩展。如果 `xin_user` 不是 superuser：

```sql
\c xin
CREATE EXTENSION IF NOT EXISTS ltree;
CREATE EXTENSION IF NOT EXISTS pg_trgm;
```

或：

```sql
ALTER USER xin_user WITH SUPERUSER;  -- 仅 dev
```

## 3. 准备 Redis（可选）

```bash
# Ubuntu/Debian
sudo apt install -y redis-server
sudo systemctl start redis-server

# macOS
brew install redis
brew services start redis
```

测试：

```bash
redis-cli ping
# PONG
```

如果**不想用 Redis**，把 `config/config.yaml` 里的 `redis.enabled: false`，框架会：

- 跳过 go-redis 初始化
- session manager 自动切到 DB-backed 实现
- 权限 cache 退化为直接查 DB（慢但可用）

## 4. 克隆 + 编译

```bash
git clone https://github.com/xin/framework.git xin-server
cd xin-server/server

./build.sh                              # Linux/macOS
# 或
.\build.ps1                             # Windows
```

产物：

```
bin/
└── xin                                 # 或 xin.exe
```

## 5. 改配置

默认配置 `config/config.yaml`：

```yaml
app:
  name: xin
  env: dev
  host: 0.0.0.0
  port: 8087

database:
  host: localhost
  port: 5432
  user: xin_user
  password: xin_password
  dbname: xin
  sslmode: disable

redis:
  enabled: true
  required: false
  host: localhost
  port: 6379

jwt:
  secret: your-secret-key    # ⚠️ prod 必改，≥32 字节
  expire: 3600
  refresh_expire: 86400

module: []                   # 空 = alwaysOn + optOut（11 个），optional 不开
```

### 5.1 用环境变量覆盖

任何字段都可以用 `XIN_*` 前缀覆盖（大写 + 下划线），优先级高于 YAML：

```bash
export XIN_APP_PORT=8088
export XIN_DB_HOST=10.0.0.5
export XIN_DB_PASSWORD=secret123
export XIN_JWT_SECRET="$(openssl rand -base64 48)"
```

支持的环境变量清单见 [`framework/pkg/config/config.go`](framework/pkg/config/config.go) `overrideWithEnv()`。

### 5.2 通过 `.env` 文件加载

`config.Load()` 自动尝试读取 `.env`：

```bash
cat > .env <<EOF
XIN_APP_PORT=8088
XIN_DB_PASSWORD=secret123
XIN_JWT_SECRET=...
EOF
```

`.env` 里的值仅当对应环境变量**未设置**时才生效（不覆盖已有的 env）。

## 6. 首次启动

### 6.1 前台跑（看日志）

```bash
go run ./cmd/xin run
# 或编译后跑
./bin/xin run     # Windows: .\bin\xin.exe run
```

预期输出（16 个 module，按配置可能少几个 optional）：

```
2026/06/21 module auth initialized
2026/06/21 module platform_tenant initialized
2026/06/21 module menu initialized
2026/06/21 module organization initialized
2026/06/21 module permission initialized
2026/06/21 module resource initialized
2026/06/21 module role initialized
2026/06/21 module user initialized
2026/06/21 module asset initialized
2026/06/21 module dict initialized
2026/06/21 module system initialized
[GIN-debug] GET /api/v1/health --> ...
2026/06/21 server starting on 0.0.0.0:8087
```

按 `Ctrl+C` 优雅退出。

### 6.2 验证

```bash
curl http://localhost:8087/api/v1/health
```

预期响应：

```json
{"code":0,"msg":"ok","data":{"status":"ok"}}
```

## 7. 守护进程模式

```bash
./bin/xin start           # 后台启动，pid 写到 ./xin.pid，日志 → ./xin.log
./bin/xin status          # 看 PID 文件状态
./bin/xin stop            # 发 SIGTERM，等当前请求完成后退出
./bin/xin restart         # stop + start
```

## 8. 跑测试

```bash
go test ./...                       # 全部测试
go test -v ./framework/pkg/...      # 只测 framework 包
go test "-coverprofile=cover.out" ./framework/pkg/middleware/...
go tool cover -func=cover.out       # 看覆盖率
```

当前覆盖情况（2026-06）：

| 包 | 覆盖率 |
|---|---:|
| `framework/pkg/middleware` | ~81% |
| `framework/pkg/plugin` | ~46% |
| `framework/pkg/permission` | ~42% |

详见 [`AGENTS.md`](../AGENTS.md)。

## 9. 常见问题

### 9.1 启动时报 `db pool is not initialized`

排查：

```bash
psql -h localhost -U xin_user -d xin -c 'SELECT 1'
echo $XIN_DB_PASSWORD
./bin/xin run 2>&1 | Select-String -Pattern "db init failed"  # Windows
```

### 9.2 `panic: jwt secret is too short`

```bash
export XIN_JWT_SECRET="$(openssl rand -base64 48)"
```

### 9.3 Redis 拒绝连接但服务不启动

| `redis.required` | 行为 |
|---|---|
| `true` | Redis 不可用 → 启动失败 |
| `false` | Redis 不可用 → log warn，继续启动，session 走 DB |
| `false` 且 `enabled: false` | 完全跳过 Redis 初始化 |

### 9.4 端口被占用

```bash
# Linux/macOS
lsof -i :8087

# Windows
Get-NetTCPConnection -LocalPort 8087
```

### 9.5 迁移没跑

迁移由 `framework/pkg/migrate.Run("migrations")` 触发，按文件名排序：

```bash
ls migrations/
# asset.sql  cms.sql  config.sql  config_alignment.sql
# dict.sql   flag.sql framework.sql  weixin.sql
```

增量迁移用 `<feature>_alignment.sql` 后缀（参考 [database.md §2.1](database.md#21-增量迁移文件命名)）。

### 9.6 JSONB 写入报 `column X is of type jsonb but expression is of type text/bytea`

SQL 加 `::jsonb` cast：

```sql
UPDATE t SET value = $1::jsonb WHERE id = $2
UPDATE t SET value = COALESCE($1::jsonb, value) WHERE id = $2
```

涉及的 JSONB 列（10 列）：

- `db_logs.old_data` / `new_data`
- `config_items.value` / `default_value` / `options` / `validation`
- `dicts.extend` / `dict_items.extend`
- `flag_frames.template_config`
- `tenants.config`

### 9.7 `invalid BOM in the middle of the file (1:4)`

源文件被某个写入脚本重复前置了 UTF-8 BOM：

```bash
python scripts/strip_bom.py --check .     # 检测
python scripts/strip_bom.py .              # 修复
```

### 9.8 gin panic: handlers are already registered for path

public / protected 都挂在 `/api/v1` 同前缀下，路径冲突。**修改方法**：

- public 端用独立前缀（如 `/public/configs`）
- 或者把 public 路由移到 protected 下用 OptionalAuth 替代 Auth

### 9.9 gin panic: ':id' in new path conflicts with existing wildcard ':code'

gin 不允许同 segment 不同 param name。统一用 `:id`（业务层用 `:code` 的放到 query 参数：`?code=`）。

### 9.10 platform module 没生效

- 确认路由在 `adminGroup := protected.Group("/admin", RequirePlatformRole("super_admin"))` 下
- 确认当前账号 `account_roles` 表有 `super_admin` 角色
- super_admin 在 RBAC 中间件层自动 bypass，但 `RequirePlatformRole` 是独立检查

## 10. 下一步

| 你想... | 看 |
|---|---|
| 了解模块怎么注册 | [architecture.md §3](architecture.md#3-模块生命周期init--register--shutdown) |
| 看完整 API 列表 | [api.md](api.md) |
| 添加新的业务模块 | [developing.md](developing.md) |
| 添加新的平台管理模块 | [developing.md §B](developing.md#step-b平台管理模块模板admin-域) |
| 配置 RBAC 权限 | [permissions.md](permissions.md) |
| 部署到生产 | [deployment.md](deployment.md) |