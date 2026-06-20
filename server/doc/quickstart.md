# 快速开始

> 目标：15 分钟内跑通 `xin run`，在浏览器看到 health endpoint 返回 `{"status":"ok"}`。

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

或者给 `xin_user` 装扩展的权限：

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

# 编译（单 module：gx1727.com/xin）
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
  port: 8087                            # 默认端口

database:
  host: localhost
  port: 5432
  user: xin_user
  password: xin_password
  dbname: xin
  sslmode: disable

redis:
  enabled: true
  required: false       # Redis 挂了不阻止启动
  host: localhost
  port: 6379

jwt:
  secret: your-secret-key    # ⚠️ prod 必改， ≥32 字节
  expire: 3600
  refresh_expire: 86400
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

注意：`.env` 里的值仅当对应环境变量**未设置**时才生效（不覆盖已有的 env）。

## 6. 首次启动

### 6.1 前台跑（看日志）

```bash
# 方式 1：直接 go run
go run ./cmd/xin run

# 方式 2：编译后跑
./bin/xin run     # Windows: .\bin\xin.exe run
```

预期输出：

```
2026/06/19 08:33:06 module tenant initialized
2026/06/19 08:33:06 module auth initialized
2026/06/19 08:33:06 module role initialized
... (15 个 module)
[GIN-debug] GET /api/v1/health --> ...
2026/06/19 08:33:06 server starting on 0.0.0.0:8087
```

按 `Ctrl+C` 优雅退出。

### 6.2 验证

另开一个终端：

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
./bin/xin hot-restart     # 不中断服务的热重启（目前等价 restart）
```

## 8. 跑测试

```bash
go test ./...                       # 全部测试
go test -v ./framework/pkg/...      # 只测 framework 包
go test "-coverprofile=cover.out" ./framework/pkg/middleware/...
go tool cover -func=cover.out       # 看覆盖率
```

当前覆盖情况：

| 包 | 覆盖率 |
|---|---:|
| `framework/pkg/middleware` | ~81% |
| `framework/pkg/plugin` | ~46% |
| `framework/pkg/permission` | ~42% |

详见 [`AGENTS.md`](../AGENTS.md)。

## 9. 常见问题

### 9.1 启动时报 `db pool is not initialized`

**原因**：`framework/pkg/db` 的 `Pool` 还是 nil，意味着 `db.Init` 失败了。

排查：

```bash
# 1. 测连通性
psql -h localhost -U xin_user -d xin -c 'SELECT 1'

# 2. 检查 .env / 环境变量是否覆盖了密码
echo $XIN_DB_PASSWORD

# 3. 看日志里的具体错误
./bin/xin run 2>&1 | grep -i "db init failed"
```

### 9.2 `panic: jwt secret is too short`

**原因**：prod 环境强制要求 `jwt.secret` ≥ 32 字节且不是常见占位符。

```bash
export XIN_JWT_SECRET="$(openssl rand -base64 48)"
```

### 9.3 Redis 拒绝连接但服务不启动

检查 `redis.required` 配置：

| 值 | 行为 |
|---|---|
| `true` | Redis 不可用 → 直接退出（启动失败） |
| `false` | Redis 不可用 → log warn，继续启动，session 走 DB |
| `false` 且 `enabled: false` | 完全跳过 Redis 初始化 |

### 9.4 端口被占用

```bash
# Linux/macOS
lsof -i :8087

# Windows
netstat -ano | findstr :8087
```

改 `app.port` 或停掉占用进程。

### 9.5 迁移没跑

迁移由 `framework/pkg/migrate.Run("migrations")` 触发，扫 `./migrations/*.sql` 按文件名顺序执行：

```bash
ls migrations/
# asset.sql  cms.sql  config.sql  dict.sql  flag.sql  framework.sql
```

### 9.6 JSONB 写入报 `column X is of type jsonb but expression is of type text/bytea`

**原因**：pgx 把 Go 的 `string` 当 `text` 发、把 `[]byte` 当 `bytea` 发，但目标列是 `JSONB`。

**修法**：SQL 里显式 `::jsonb` cast：

```sql
-- 错
UPDATE t SET value = $1 WHERE id = $2

-- 对
UPDATE t SET value = $1::jsonb WHERE id = $2
-- 或 COALESCE 场景
UPDATE t SET value = COALESCE($1::jsonb, value) WHERE id = $2
```

涉及的 JSONB 列（已全部加 cast）：

- `db_logs.old_data` / `db_logs.new_data`
- `config_items.value` / `default_value` / `options` / `validation`
- `dicts.extend` / `dict_items.extend`
- `flag_frames.template_config`

### 9.7 `invalid BOM in the middle of the file (1:4)`

**原因**：源文件被某个写入脚本重复前置了 UTF-8 BOM（PowerShell 默认 GBK + 写入带 BOM）。

**修法**：

```bash
python scripts/strip_bom.py --check .     # 检测
python scripts/strip_bom.py .              # 修复（剥所有前导 BOM）
```

详见 [README.md §工具](../README.md)。

## 10. 下一步

| 你想... | 看 |
|---|---|
| 了解模块怎么注册 | [architecture.md](architecture.md#3-模块生命周期init--register--shutdown) |
| 看完整 API 列表 | [api.md](api.md) |
| 添加新的业务模块 | [developing.md](developing.md) |
| 配置 RBAC 权限 | [permissions.md](permissions.md) |
| 部署到生产 | [deployment.md](deployment.md) |
