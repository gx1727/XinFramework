# 部署

> 本文件描述 XinFramework 的构建、运行、守护、监控、CI/CD 流程。

---

## 1. 构建

### 1.1 后端（Go）

#### Linux / Mac

```bash
cd server
./build.sh
# 或手写
go build -o bin/xin ./cmd/xin
go build -o bin/rotate-admin-password ./cmd/rotate-admin-password
```

#### Windows

```powershell
cd server
.\build.ps1
# 或手写
go build -o bin/xin.exe ./cmd/xin
go build -o bin/rotate-admin-password.exe ./cmd/rotate-admin-password
```

构建产物：

| 产物 | 用途 |
|---|---|
| `bin/xin` / `xin.exe` | 主服务（start/stop/run/reload/hot-restart/status） |
| `bin/rotate-admin-password` / `rotate-admin-password.exe` | 独立密码重置工具 |

#### 交叉编译

```bash
# Linux
GOOS=linux GOARCH=amd64 go build -o bin/xin-linux-amd64 ./cmd/xin
GOOS=linux GOARCH=arm64 go build -o bin/xin-linux-arm64 ./cmd/xin

# Windows
GOOS=windows GOARCH=amd64 go build -o bin/xin-windows-amd64.exe ./cmd/xin

# macOS
GOOS=darwin GOARCH=arm64 go build -o bin/xin-darwin-arm64 ./cmd/xin
```

### 1.2 前端（React）

```bash
cd UI
npm install
npm run build
# 产物在 UI/dist/
```

构建配置见 `UI/vite.config.ts`：

- 默认 dev port：5241
- 默认 API 后端：`http://localhost:8087/api/v1`
- 通过 `.env` 覆盖：`VITE_API_BASE_URL` / `VITE_ASSET_BASE_URL`

---

## 2. 配置

### 2.1 配置文件结构

```
server/config/
├── config.yaml        # 主配置（dev 默认值）
├── config.dev.yaml    # dev 覆盖
├── config.prod.yaml   # prod 覆盖（强制 jwt.secret 校验、显式列 module）
└── cms.yaml           # cms 模块私有配置（config.LoadModule("cms", ...)）
```

### 2.2 配置加载顺序

```
1. config.yaml             (基础)
2. config.{app.env}.yaml   (按 env 叠加)
3. .env                    (仅当对应环境变量未设)
4. XIN_* 环境变量          (覆盖)
```

### 2.3 生产环境必填

通过环境变量注入（**不要**写在 `config.prod.yaml`）：

```bash
# 强校验：secret ≥32 字节、非占位符
export XIN_JWT_SECRET=$(openssl rand -base64 48)

# 数据库凭据
export XIN_DATABASE_HOST=...
export XIN_DATABASE_USER=...
export XIN_DATABASE_PASSWORD=...

# Redis
export XIN_REDIS_HOST=...
export XIN_REDIS_PASSWORD=...

# COS（启用时）
export XIN_STORAGE_COS_SECRET_ID=...
export XIN_STORAGE_COS_SECRET_KEY=...

# 微信（启用时）
export XIN_WEIXIN_APPID=...
export XIN_WEIXIN_APPSECRET=...
```

**禁止的占位值**（`config.validateJWTSecret` 在 `app.env=prod` 时检查）：
- `your-secret-key` / `changeme` / `please-change-me` / `secret` / `12345678`
- 空字符串

启动期会 FATAL 退出。生产部署务必通过 CI 注入。

---

## 3. 运行模式

### 3.1 前台（开发用）

```bash
./bin/xin run
# 或
go run ./cmd/xin run
```

### 3.2 守护进程

```bash
# 启动（写 PID 到 ./xin.pid，日志到 ./xin.log）
./bin/xin start

# 查看状态
./bin/xin status
# 输出：
#   PID: 12345
#   最近 5 行日志:
#   ...

# 优雅停止（SIGTERM，30s 超时）
./bin/xin stop

# 重启（stop + start）
./bin/xin restart

# 零停机重载（SIGUSR1）
./bin/xin reload

# 热重启（起新进程 + 停老进程）
./bin/xin hot-restart
```

### 3.3 Windows 服务

Windows 下用 NSSM（Non-Sucking Service Manager）包装：

```powershell
# 安装 NSSM 后
nssm install XinFramework "D:\work\xin\XinFramework\server\bin\xin.exe" "run"
nssm set XinFramework AppDirectory "D:\work\xin\XinFramework\server"
nssm set XinFramework AppStdout "D:\work\xin\XinFramework\server\logs\service.log"
nssm set XinFramework AppStderr "D:\work\xin\XinFramework\server\logs\service-error.log"
nssm set XinFramework AppRotateFiles 1
nssm set XinFramework AppRotateBytes 10485760
nssm start XinFramework
```

### 3.4 Linux systemd

仓库提供 `framework/xin-server.service` 模板：

```ini
[Unit]
Description=XinFramework Server
After=network.target postgresql.service

[Service]
Type=simple
User=xin
WorkingDirectory=/opt/xin/XinFramework/server
ExecStart=/opt/xin/XinFramework/server/bin/xin run
Restart=always
RestartSec=5s
LimitNOFILE=65536
Environment="XIN_APP_ENV=prod"
EnvironmentFile=/opt/xin/XinFramework/server/.env

[Install]
WantedBy=multi-user.target
```

启用：

```bash
sudo cp framework/xin-server.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable xin-server
sudo systemctl start xin-server
sudo systemctl status xin-server
```

---

## 4. 环境清单

### 4.1 必备

| 组件 | 版本 | 用途 |
|---|---|---|
| PostgreSQL | 14+ | 主数据库 |
| Go runtime | 1.24+ | 服务端（二进制可独立） |

### 4.2 可选

| 组件 | 用途 | 不可用时 |
|---|---|---|
| Redis 7+ | 缓存 + Session | 降级到 DB session（`auth_sessions` 表） |
| 腾讯云 COS | 对象存储 | 降级到本地文件 `./uploads/` |
| 微信小程序 | 登录 | 模块禁用即可（`module:` 不列 `weixin`） |

### 4.3 前端托管

前端是纯静态文件，可托管在任意 CDN / 反向代理：

- **Vercel / Netlify**：`UI/` 目录直接部署
- **Nginx**：把 `dist/` 挂到 `/var/www/xin`，配置 SPA fallback
- **Caddy**：`file_server` + `try_files`

---

## 5. 反向代理 / Nginx 示例

```nginx
server {
    listen 80;
    server_name xin.example.com;

    # 前端
    location / {
        root /var/www/xin-ui;
        try_files $uri $uri/ /index.html;
    }

    # 后端
    location /api/ {
        proxy_pass http://127.0.0.1:8087;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_set_header X-Request-ID $request_id;
        proxy_read_timeout 60s;
    }

    # 静态文件（uploads / assets）
    location /uploads/ {
        alias /var/xin/uploads/;
        expires 30d;
    }
}
```

HTTPS 用 `certbot --nginx`。

---

## 6. 数据库准备

### 6.1 首次部署

```bash
# 1. 创建用户与数据库
psql -U postgres -c "CREATE USER xin_user WITH PASSWORD '***';"
psql -U postgres -c "CREATE DATABASE xin OWNER xin_user;"

# 2. 启用扩展（迁移文件也会做，但手动更安全）
psql -U xin_user -d xin -c "CREATE EXTENSION IF NOT EXISTS ltree;"
psql -U xin_user -d xin -c "CREATE EXTENSION IF NOT EXISTS pg_trgm;"

# 3. 启动服务（首次自动跑迁移 + bootstrap）
./bin/xin run
```

### 6.2 已运行实例升级

```bash
# 拉新代码
git pull

# 重启服务（启动期自动跑 migrations/ 新文件）
./bin/xin restart
```

迁移是**前向兼容**的——`ALTER TABLE ... ADD COLUMN IF NOT EXISTS` 不会破坏老代码。

### 6.3 备份

```bash
# 全量备份
pg_dump -U xin_user -d xin -Fc -f xin-$(date +%Y%m%d).dump

# 恢复
pg_restore -U xin_user -d xin --clean --if-exists xin-20260624.dump
```

**重要**：定期备份 `db_logs` / `auth_sessions` 等表，按需归档。

---

## 7. 监控

### 7.1 健康检查

```bash
curl http://localhost:8087/api/v1/health
# → {"code":0,"msg":"ok","data":{"status":"ok"}}
```

可挂到 K8s liveness / readiness probe：

```yaml
livenessProbe:
  httpGet:
    path: /api/v1/health
    port: 8087
  initialDelaySeconds: 10
  periodSeconds: 30

readinessProbe:
  httpGet:
    path: /api/v1/health
    port: 8087
  initialDelaySeconds: 5
  periodSeconds: 10
```

### 7.2 关键指标

通过自定义 admin 端点或日志获取：

| 指标 | 路径 |
|---|---|
| 服务运行时间 | `framework/runtime.go` 暴露 |
| DB 连接池状态 | `pool.Stat()` |
| Redis 连接状态 | `cache.Get() != nil` |
| 慢查询 | 中间件 Logger 输出 |
| 业务错误 | `resp.logResponse` 输出 |

### 7.3 日志

`config/config.yaml`：

```yaml
log:
  dir: logs                  # 日志目录
  level: info                # debug | info | warn | error
```

日志按天切分（`logs/2026-06-24.log`）：

```
logs/
├── 2026-06-17.log
├── 2026-06-18.log
├── audit-2026-06-18.log     # 审计（独立 logger）
├── auth-2026-06-24.log      # 鉴权
├── weixin-2026-06-24.log    # 微信模块
└── ...
```

**日志格式**：

```
[2026-06-24 12:34:56] [INFO] [req-abc123] GET /api/v1/users | 200 | 23.5ms
```

### 7.4 审计

`db_logs` 表是结构化审计；非结构化事件走 logger。

```sql
-- 最近 1 小时的关键操作
SELECT created_at, user_id, action, table_name, record_id
FROM db_logs
WHERE created_at > NOW() - INTERVAL '1 hour'
  AND action IN ('user:delete', 'role:create', 'tenant:purge')
ORDER BY created_at DESC;
```

---

## 8. CI / CD

### 8.1 GitHub Actions 示例

```yaml
name: Build & Test

on: [push, pull_request]

jobs:
  backend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-go@v5
        with:
          go-version: '1.24'

      - name: Build
        working-directory: server
        run: go build -o bin/xin ./cmd/xin

      - name: Test
        working-directory: server
        run: go test ./... || true  # 暂未集成测试

      - name: BOM Check
        run: python server/scripts/strip_bom.py --check .

  frontend:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
      - uses: actions/setup-node@v4
        with:
          node-version: '20'

      - name: Install
        working-directory: UI
        run: npm ci

      - name: Type Check
        working-directory: UI
        run: ./node_modules/.bin/tsc --noEmit

      - name: Build
        working-directory: UI
        run: npm run build
```

### 8.2 发布流程

```bash
# 1. 打 tag
git tag v0.1.0
git push origin v0.1.0

# 2. CI 自动构建多平台二进制
# - xin-linux-amd64
# - xin-linux-arm64
# - xin-darwin-amd64
# - xin-windows-amd64.exe
# - UI dist

# 3. 部署
ssh server "cd /opt/xin && ./xin stop && ./update-bin v0.1.0 && ./xin start"
```

---

## 9. 灰度发布

### 9.1 Blue-Green 部署

```bash
# 1. 起新实例在 8088
XIN_APP_PORT=8088 ./bin/xin start

# 2. 切 nginx upstream
# upstream xin { server 127.0.0.1:8087; server 127.0.0.1:8088; }
# 把 8088 移到 primary，8087 留 1% 流量

# 3. 验证无问题后，停 8087
XIN_APP_PORT=8087 ./bin/xin stop
```

### 9.2 滚动发布（K8s）

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: xin-framework
spec:
  replicas: 3
  strategy:
    type: RollingUpdate
    rollingUpdate:
      maxSurge: 1
      maxUnavailable: 0
  template:
    spec:
      containers:
        - name: xin
          image: registry.example.com/xin-framework:v0.1.0
          ports:
            - containerPort: 8087
          env:
            - name: XIN_APP_ENV
              value: prod
          readinessProbe:
            httpGet:
              path: /api/v1/health
              port: 8087
            initialDelaySeconds: 5
            periodSeconds: 10
          livenessProbe:
            httpGet:
              path: /api/v1/health
              port: 8087
            initialDelaySeconds: 30
            periodSeconds: 30
          resources:
            requests:
              cpu: 100m
              memory: 256Mi
            limits:
              cpu: 1000m
              memory: 1Gi
```

---

## 10. 运维命令

### 10.1 重置 admin 密码

```bash
# 用独立工具
./bin/rotate-admin-password -tenant 1 -account admin -new "new-password"
```

或通过 SQL（需要 argon2id 哈希）：

```sql
-- 1. 用工具生成哈希
go run ./cmd/rotate-admin-password -h
# Usage: ... -hash-only -password "new-password"

-- 2. UPDATE
UPDATE accounts SET password = '$argon2id$v=19$m=...,t=...,p=4$...' 
WHERE phone = '13800138000';
```

### 10.2 清缓存

```bash
curl -X POST http://localhost:8087/api/v1/system/clear-cache \
  -H "Authorization: Bearer <super_admin_token>"
```

### 10.3 查 server info

```bash
curl http://localhost:8087/api/v1/system/server-info \
  -H "Authorization: Bearer <super_admin_token>"
```

### 10.4 远程信号

```bash
# SIGUSR1 - 零停机重载（重新加载配置）
kill -USR1 $(cat server/xin.pid)

# SIGTERM - 优雅停止
kill -TERM $(cat server/xin.pid)
```

---

## 11. 安全清单

部署前确认：

- [ ] `app.env=prod`（强制 JWT secret 校验）
- [ ] `XIN_JWT_SECRET` ≥32 字节、非占位符
- [ ] 数据库密码强（≥16 字符）
- [ ] Redis 启用 `requirepass`（生产环境）
- [ ] COS / 微信凭据用 env 注入，不在 yaml
- [ ] `cors.allow_origins` 列具体源（不是 `*`）
- [ ] HTTPS 终止在反代（nginx / caddy / cloud LB）
- [ ] DB / Redis 不对外暴露（仅本机或内网）
- [ ] 防火墙：仅 80/443 对外
- [ ] 日志轮转（logrotate）
- [ ] 监控告警：CPU / 内存 / 磁盘 / DB 连接数 / 慢查询
- [ ] 备份策略：DB 全量每日，增量每 6h

---

## 12. 故障排查

### 12.1 启动失败

| 现象 | 排查 |
|---|---|
| `FATAL jwt secret invalid` | 检查 `XIN_JWT_SECRET` |
| `connection refused` (PG) | DB 启动？端口对？防火墙？ |
| `extension "ltree" not found` | PG 安装包未带 contrib（`postgresql-contrib`） |
| `migrations failed` | 看具体 SQL；常见：JSONB cast 缺 / 表已存在 |
| `module X not enabled` | `cfg.module` 缺该模块；optOut 也要显式列 |

### 12.2 运行期

| 现象 | 排查 |
|---|---|
| 业务 5xx | `db_logs` 查具体错；`logResponse` 输出 |
| 慢查询 | `EXPLAIN ANALYZE`；检查索引；考虑加缓存 |
| 内存增长 | goroutine 泄漏（`pprof`）；连接未关闭 |
| 502/504 | 反代超时；后端 OOM；DB 慢 |
| 403 持续 | 角色权限码；session 失效（重新登录） |
| 401 持续 | JWT secret 改了（老 token 全失效，预期） |

### 12.3 数据问题

| 现象 | 排查 |
|---|---|
| 看不到数据 | RLS 拒？tenant_id 错？DataScope 过滤？ |
| 数据重复 | 唯一约束是否被绕过（直接 INSERT 缺 `ON CONFLICT`） |
| 软删但还在列表 | `is_deleted = FALSE` 谓词；Repository 漏写 |

---

## 13. 升级到新版本

```bash
# 1. 备份
pg_dump -U xin_user -d xin -Fc -f backup-$(date +%Y%m%d).dump

# 2. 拉新代码
git pull origin main

# 3. 编译
go build -o bin/xin.new ./cmd/xin
go build -o bin/rotate-admin-password.new ./cmd/rotate-admin-password

# 4. 原子替换
mv bin/xin bin/xin.bak
mv bin/xin.new bin/xin
mv bin/rotate-admin-password bin/rotate-admin-password.bak
mv bin/rotate-admin-password.new bin/rotate-admin-password

# 5. 重启（启动期自动跑 migrations/ 新文件）
./bin/xin stop
./bin/xin start

# 6. 验证
curl http://localhost:8087/api/v1/health

# 7. 保留旧版本 1 周
# 确认无问题后删除 xin.bak
```
