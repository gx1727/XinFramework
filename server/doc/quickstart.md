# 快速开始

> 5 分钟把 XinFramework 后端跑起来。

## 1. 环境

| 工具 | 版本 | 检查 |
| --- | --- | --- |
| Go | 1.25+ | `go version` |
| Node | 20+（前端） | `node -v` |
| PostgreSQL | 16+ | `psql --version` |
| 操作系统 | Linux / macOS / Windows | —— |

## 2. 启动 PostgreSQL

任选一种：

```bash
# 选项 A：Docker
docker run --name xin-pg \
  -e POSTGRES_USER=xin -e POSTGRES_PASSWORD=dev \
  -e POSTGRES_DB=xin -p 5432:5432 \
  -d postgres:16

# 选项 B：本地服务
# brew install postgresql@16
# pg_ctl -D /usr/local/var/postgresql start
# createuser xin && createdb xin -O xin
# psql -c "ALTER USER xin WITH PASSWORD 'dev';"
```

## 3. 拉代码与编译

```bash
git clone <repo-url> xin
cd xin/server

# 同步 go.work
go work sync

# 编译（首次较慢，会下载依赖）
go build ./...
go vet ./...
```

## 4. 配置

```bash
cp config/config.example.yaml config/config.yaml
# 编辑 config.yaml（下面说明关键字段）
```

`config/config.yaml` 关键字段：

```yaml
app:
  host: 0.0.0.0
  port: 8080
  debug: true

db:
  host: 127.0.0.1
  port: 5432
  user: xin
  password: dev
  database: xin
  max_conns: 20

jwt:
  secret: "change-me-in-production"
  ttl: 3600
  refresh_ttl: 86400

session:
  driver: memory   # memory | redis

storage:
  provider: local  # local | cos
  local_dir: ./uploads
  local_base_url: http://localhost:8080/files

cors:
  allowed_origins:
    - http://localhost:5173
  allowed_methods: [GET, POST, PUT, DELETE, PATCH]
  allowed_headers: [Authorization, Content-Type]

module:
  - auth
  - tenant
  - user
  - role
  - menu
  - resource
  - permission
  - organization
  - dict
  - asset
  - weixin
  - system
  - cms
  - flag
```

`module:` 列表控制运行时启用哪些模块。增删模块：

1. 编辑 `module:` 列表
2. 重启服务

新增模块的代码（app）需要：

1. 在 `apps/<name>/` 建包（详见 [developing.md](file:///d:\work\xin\XinFramework\server\doc\developing.md)）
2. 在 `cmd/xin/main.go` 加 `_ "gx1727.com/xin/apps/<name>"` 行
3. 在 `module:` 列表加名称

## 5. 首次启动

```bash
# 前台运行（开发推荐）
go run ./cmd/xin run

# 输出类似：
# 2026-06-15 10:00:00 [INFO] boot init ok
# 2026-06-15 10:00:00 [INFO] migrations applied (12 files)
# 2026-06-15 10:00:00 [INFO] module auth initialized
# 2026-06-15 10:00:00 [INFO] module tenant initialized
# ...
# 2026-06-15 10:00:00 [INFO] server starting on 0.0.0.0:8080
```

**首次启动**会要求创建默认租户和超级管理员。复制粘贴运行命令里的环境变量：

```bash
# 服务会等待你执行下面的命令：
# XIN_BOOTSTRAP=1 go run ./cmd/xin bootstrap
```

`bootstrap` 命令会要求输入：

| 字段 | 说明 |
| --- | --- |
| `tenant_code` | 默认租户编码（如 `default`） |
| `tenant_name` | 默认租户名称（如 `我的公司`） |
| `admin_account` | 超级管理员账号 |
| `admin_password` | 超级管理员密码（至少 8 位） |

完成后可登录：

```bash
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"tenant_code":"default","account":"admin","password":"xxx"}'
```

返回：

```json
{
  "code": 0,
  "msg": "ok",
  "data": {
    "token": "eyJhbGciOi...",
    "refresh_token": "eyJhbGciOi...",
    "user": { "id": 1, "account": "admin", "name": "管理员" }
  }
}
```

## 6. 启动前端

```bash
cd ../UI
npm install
npm run dev
# Vite 启动在 http://localhost:5173
# 默认走 mock 数据，后端不通也能演示完整 UI
```

## 7. 守护进程（生产）

```bash
# Linux（systemd）
go build -ldflags="-s -w" -o /usr/local/bin/xin ./cmd/xin
cp framework/xin-server.service /etc/systemd/system/
systemctl daemon-reload
systemctl enable --now xin-server
systemctl status xin-server

# 通用（nohup）
go build -ldflags="-s -w" -o ./bin/xin ./cmd/xin
./bin/xin start   # 后台模式（fork + pid）
./bin/xin stop
./bin/xin status
```

详见 [deployment.md](file:///d:\work\xin\XinFramework\server\doc\deployment.md)。

## 8. 常见问题

### Q: 启动报 `password authentication failed for user "xin"`

PostgreSQL 用户密码不对，或用户不存在。检查：

```bash
psql -U postgres -c "\\du xin"
# 如果不存在：
psql -U postgres -c "CREATE USER xin WITH PASSWORD 'dev';"
psql -U postgres -c "CREATE DATABASE xin OWNER xin;"
```

### Q: 报 `relation "users" does not exist`

迁移没跑成功。手动跑：

```bash
psql -U xin -d xin -f migrations/framework.sql
psql -U xin -d xin -f migrations/cms.sql
psql -U xin -d xin -f migrations/flag.sql
```

### Q: 模块启动报 `module X not registered`

`module:` 列表里有名字，但代码里没 import。在 `cmd/xin/main.go` 加 side-effect import：

```go
import _ "gx1727.com/xin/apps/<x>"
```

### Q: 报 `package gx1727.com/xin/framework/internal/...: cannot import internal package`

错把内部模块从 apps 引用了。apps 只能引用 `framework/pkg/...`。详见 [architecture.md §4](file:///d:\work\xin\XinFramework\server\doc\architecture.md#4-跨模块依赖规则)。

### Q: 中文乱码

PowerShell 默认 GBK。详见 [AGENTS.md §10.1](file:///d:\work\xin\XinFramework\server\AGENTS.md#101-编码最重要)。