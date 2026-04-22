# 功能验证手册

本文档记录项目所有功能模块及其验证方案，用于快速回归验证。

---

## 1. 配置管理 (pkg/config)

### 功能描述
- 加载 YAML 配置文件
- 支持 `.env` 文件覆盖
- 支持环境变量覆盖（前缀 `XIN_`）
- 优先级：环境变量 > .env > YAML

### 验证步骤

**手动验证：**
```bash
# 1. 验证 YAML 加载
./out/xin run 2>&1 | head -20

# 2. 验证环境变量覆盖
XIN_APP_PORT=9999 XIN_DB_HOST=localhost ./out/xin run 2>&1 | grep -i "starting"
```

**自动化测试用例：**
```go
// Test: 配置加载优先级
func TestConfigLoadPriority(t *testing.T) {
    // 1. 创建测试 YAML
    // 2. 创建 .env 覆盖
    // 3. 设置环境变量覆盖
    // 4. 验证环境变量优先级最高
}
```

### 预期结果
- YAML 配置正确解析
- 环境变量 `XIN_APP_PORT=9999` 覆盖 YAML 中的 `app.port`
- 日志显示实际使用的端口值

---

## 2. 日志系统 (internal/infra/logger)

### 功能描述
- 按天切割日志文件
- 多级别日志支持（DEBUG/INFO/WARN/ERROR）
- 同时输出到 stdout 和文件
- 线程安全

### 验证步骤

**手动验证：**
```bash
# 1. 启动服务，观察日志输出
./out/xin run &

# 2. 触发不同级别日志
curl -s http://localhost:8080/api/v1/health
curl -s -H "X-Tenant-ID: invalid" http://localhost:8080/api/v1/health

# 3. 等待次日，检查日志分割
ls -la out/logs/
cat out/xin.log

# 4. 停止服务
./out/xin stop
```

**自动化测试用例：**
```go
// Test: 日志级别过滤
func TestLogLevelFilter(t *testing.T) {
    logger.Init("./test-logs", "WARN")
    logger.Info("should not appear")
    logger.Warn("should appear")
    // 验证文件只包含 WARN 及以上
}

// Test: 日志按天分割
func TestDailyRotation(t *testing.T) {
    // 模拟跨天，验证新文件创建
}
```

### 预期结果
- 日志同时输出到控制台和 `out/xin.log`
- 午夜后生成新日期日志文件 `YYYY-MM-DD.log`
- 低于设定级别的日志不写入文件

---

## 3. 数据库连接 (internal/infra/db)

### 功能描述
- GORM + PostgreSQL 连接
- 连接池管理
- 租户会话变量隔离（`SET app.tenant_id = ?`）

### 验证步骤

**手动验证：**
```bash
# 1. 配置有效数据库
# 2. 启动服务
./out/xin run &

# 3. 验证数据库连接
# 检查日志中是否有 "database connection established"

# 4. 停止服务
./out/xin stop
```

**自动化测试用例：**
```go
// Test: 数据库连接池
func TestDBPoolSettings(t *testing.T) {
    cfg := &config.DatabaseConfig{
        MaxOpenConns: 25,
        MaxIdleConns: 5,
    }
    db.Init(cfg)
    defer db.Close()

    sqlDB, _ := db.Get().DB()
    assert.Equal(t, 25, sqlDB.MaxOpenConns())
    assert.Equal(t, 5, sqlDB.MaxIdleConns())
}

// Test: 租户会话变量
func TestTenantSessionVariable(t *testing.T) {
    db.Init(cfg)
    defer db.Close()

    db.SetTenantID(123)
    // 验证执行了 SET app.tenant_id = 123

    db.ClearTenantID()
    // 验证执行了 RESET app.tenant_id
}
```

### 预期结果
- 数据库连接成功建立
- 连接池参数生效
- `SetTenantID()` 正确设置 PostgreSQL 会话变量

---

## 4. Redis 缓存 (internal/infra/cache)

### 功能描述
- Redis 客户端连接
- 连接池管理
- 可选优雅降级（`Enabled=false` 时禁用）

### 验证步骤

**手动验证：**
```bash
# 1. 启动 Redis
redis-server &
sleep 1

# 2. 启动服务
./out/xin run &

# 3. 验证 Redis 连接
# 检查日志中是否有 "redis connection established"

# 4. 测试降级
# 设置 redis.enabled: false 重启，验证不报错

# 5. 停止
./out/xin stop
redis-cli shutdown
```

**自动化测试用例：**
```go
// Test: Redis 正常连接
func TestRedisConnection(t *testing.T) {
    cfg := &config.RedisConfig{
        Host:     "localhost",
        Port:     6379,
        Enabled:  true,
        Required: true,
    }
    err := cache.Init(cfg)
    assert.NoError(t, err)
    assert.NotNil(t, cache.Get())

    // 测试 PING
    pong, err := cache.Get().Ping(ctx).Result()
    assert.Equal(t, "PONG", pong)
}

// Test: Redis 禁用降级
func TestRedisDisabled(t *testing.T) {
    cfg := &config.RedisConfig{Enabled: false}
    err := cache.Init(cfg)
    assert.NoError(t, err)
    assert.Nil(t, cache.Get())
}

// Test: Redis 必需模式
func TestRedisRequiredFails(t *testing.T) {
    cfg := &config.RedisConfig{
        Host:     "invalid-host",
        Port:     6379,
        Required: true,
    }
    err := cache.Init(cfg)
    assert.Error(t, err) // 应失败
}
```

### 预期结果
- Redis 连接成功时 `cache.Get()` 返回有效客户端
- `Enabled=false` 时 `cache.Get()` 返回 nil，不影响启动
- `Required=true` 但 Redis 不可用时启动失败

---

## 5. JWT 认证 (pkg/jwt + internal/core/middleware/Auth)

### 功能描述
- JWT Token 生成
- JWT 验证与解析
- 从 Token 提取 user_id、tenant_id、role
- 设置 XinContext.UserID

### 验证步骤

**手动验证：**
```bash
# 1. 生成 Token（需要代码调用或编写测试程序）
# 2. 使用有效 Token 访问受保护路由
curl -H "Authorization: Bearer <token>" http://localhost:8080/api/v1/users

# 3. 使用无效 Token
curl -H "Authorization: Bearer invalid-token" http://localhost:8080/api/v1/users
# 预期返回 401

# 4. 无 Token
curl http://localhost:8080/api/v1/users
# 预期返回 401
```

**自动化测试用例：**
```go
// Test: JWT 生成与验证
func TestJWTGenerateAndValidate(t *testing.T) {
    cfg := &config.JWTConfig{Secret: "test-secret", Expire: 3600}
    token, err := jwt.Generate(cfg, 1, 1, "admin")
    assert.NoError(t, err)
    assert.NotEmpty(t, token)

    // 解析验证
    parsed, err := jwt.Parse(token, cfg.Secret)
    assert.NoError(t, err)
    assert.Equal(t, uint(1), parsed.UserID)
    assert.Equal(t, uint(1), parsed.TenantID)
    assert.Equal(t, "admin", parsed.Role)
}

// Test: JWT 过期验证
func TestJWTExpiration(t *testing.T) {
    cfg := &config.JWTConfig{Secret: "test", Expire: -1} // 已过期
    _, err := jwt.Generate(cfg, 1, 1, "admin")
    assert.Error(t, err)
}

// Test: Auth 中间件
func TestAuthMiddleware(t *testing.T) {
    router := gin.New()
    router.Use(middleware.Auth(cfg))
    router.GET("/test", func(c *gin.Context) {
        ctx := context.New(c)
        assert.Equal(t, uint(1), ctx.UserID)
    })

    // 带有效 Token
    req, _ := http.NewRequest("GET", "/test", nil)
    req.Header.Set("Authorization", "Bearer <valid-token>")
    router.ServeHTTP(w, req)
    assert.Equal(t, 200, w.Code)

    // 无 Token
    req, _ = http.NewRequest("GET", "/test", nil)
    router.ServeHTTP(w, req)
    assert.Equal(t, 401, w.Code)
}
```

### 预期结果
- 有效 Token 通过验证，受保护路由返回 200
- 无 Token 或无效 Token 返回 401 Unauthorized
- `XinContext.UserID` 正确设置

---

## 6. 租户隔离 (internal/core/middleware/Tenant)

### 功能描述
- 从 `X-Tenant-ID` header 读取租户 ID
- 调用 `db.SetTenantID()` 设置会话变量
- 支持多种 SaaS 模式（shared/schema/database）

### 验证步骤

**手动验证：**
```bash
# 1. 启动服务
./out/xin run &

# 2. 带租户 ID 请求
curl -H "X-Tenant-ID: 123" http://localhost:8080/api/v1/health

# 3. 不带租户 ID 请求
curl http://localhost:8080/api/v1/health

# 4. 带无效租户 ID
curl -H "X-Tenant-ID: invalid" http://localhost:8080/api/v1/health

# 5. 停止
./out/xin stop
```

**自动化测试用例：**
```go
// Test: Tenant 中间件正常提取
func TestTenantMiddleware(t *testing.T) {
    router := gin.New()
    router.Use(middleware.Tenant("shared"))
    router.GET("/test", func(c *gin.Context) {
        ctx := context.New(c)
        assert.Equal(t, uint(123), ctx.TenantID)
    })

    req, _ := http.NewRequest("GET", "/test", nil)
    req.Header.Set("X-Tenant-ID", "123")
    router.ServeHTTP(w, req)
}

// Test: 无租户 ID
func TestTenantMiddlewareEmpty(t *testing.T) {
    router := gin.New()
    router.Use(middleware.Tenant("shared"))
    router.GET("/test", func(c *gin.Context) {
        ctx := context.New(c)
        assert.Equal(t, uint(0), ctx.TenantID)
    })

    req, _ := http.NewRequest("GET", "/test", nil)
    router.ServeHTTP(w, req)
}
```

### 预期结果
- `X-Tenant-ID: 123` 时 `XinContext.TenantID` = 123
- 无 header 时 `TenantID` = 0
- `db.SetTenantID()` 被正确调用

---

## 7. 请求日志 (internal/core/middleware/Logger)

### 功能描述
- 记录请求方法、路径、状态码、延迟
- 使用结构化日志格式

### 验证步骤

**手动验证：**
```bash
# 1. 启动服务
./out/xin run &

# 2. 发起请求
curl http://localhost:8080/api/v1/health

# 3. 检查日志
cat out/xin.log
# 应包含类似: [GIN] 2026/04/20 - GET /api/v1/health 200 ...

# 4. 停止
./out/xin stop
```

**自动化测试用例：**
```go
// Test: 请求日志记录
func TestRequestLogger(t *testing.T) {
    router := gin.New()
    router.Use(middleware.Logger())
    router.GET("/test", func(c *gin.Context) {
        c.Status(200)
    })

    req, _ := http.NewRequest("GET", "/test", nil)
    router.ServeHTTP(w, req)

    // 验证日志输出包含请求信息
}
```

### 预期结果
- 日志包含 `GET /api/v1/health 200` 格式
- 包含响应时间

---

## 8. Panic 恢复 (internal/core/middleware/Recovery)

### 功能描述
- 捕获 panic 防止服务崩溃
- 返回 500 错误
- 记录错误栈

### 验证步骤

**手动验证：**
```bash
# 1. 启动服务
./out/xin run &

# 2. 触发 panic（需要添加测试路由或修改代码）
# 正常情况下无法手动触发

# 3. 检查日志中是否有 panic stack trace
cat out/xin.log

# 4. 停止
./out/xin stop
```

**自动化测试用例：**
```go
// Test: Panic 恢复
func TestRecoveryMiddleware(t *testing.T) {
    router := gin.New()
    router.Use(middleware.Recovery())
    router.GET("/panic", func(c *gin.Context) {
        panic("test panic")
    })

    req, _ := http.NewRequest("GET", "/panic", nil)
    router.ServeHTTP(w, req)

    assert.Equal(t, 500, w.Code)
    assert.Contains(t, w.Body.String(), "Internal Server Error")
}
```

### 预期结果
- Panic 不导致进程退出
- 返回 HTTP 500
- 错误栈写入日志

---

## 9. 请求 ID (internal/core/middleware/RequestID)

### 功能描述
- 生成 UUID 作为请求 ID
- 响应头返回 `X-Request-ID`

### 验证步骤

**手动验证：**
```bash
# 1. 启动服务
./out/xin run &

# 2. 请求，检查响应头
curl -v http://localhost:8080/api/v1/health 2>&1 | grep -i "X-Request-ID"

# 3. 带已有 Request-ID
curl -H "X-Request-ID: my-custom-id" -v http://localhost:8080/api/v1/health 2>&1 | grep -i "X-Request-ID"

# 4. 停止
./out/xin stop
```

**自动化测试用例：```go
// Test: Request ID 生成
func TestRequestIDMiddleware(t *testing.T) {
    router := gin.New()
    router.Use(middleware.RequestID())
    router.GET("/test", func(c *gin.Context) {
        reqID := c.GetHeader("X-Request-ID")
        assert.NotEmpty(t, reqID)
    })

    req, _ := http.NewRequest("GET", "/test", nil)
    router.ServeHTTP(w, req)
    assert.NotEmpty(t, w.Header().Get("X-Request-ID"))
}

// Test: Request ID 透传
func TestRequestIDPassthrough(t *testing.T) {
    router := gin.New()
    router.Use(middleware.RequestID())
    router.GET("/test", func(c *gin.Context) {})

    req, _ := http.NewRequest("GET", "/test", nil)
    req.Header.Set("X-Request-ID", "custom-id")
    router.ServeHTTP(w, req)

    assert.Equal(t, "custom-id", w.Header().Get("X-Request-ID"))
}
```

### 预期结果
- 响应头包含 `X-Request-ID`
- 客户端传入的 ID 被透传

---

## 10. 服务管理命令 (cmd/server)

### 功能描述
| 命令 | 功能 |
|------|------|
| `start` | 守护进程模式启动，写入 PID |
| `stop` | 优雅停止（SIGTERM，30s 超时） |
| `restart` | 重启 |
| `reload` | 热重载（SIGUSR1，Unix） |
| `hot-restart` | 零宕机重启 |
| `status` | 查看状态 |
| `run` | 前台运行 |
| `help` | 帮助 |

### 验证步骤

**手动验证：**
```bash
# 1. 前台运行测试
./out/xin run &
sleep 2
# 观察日志输出

# 2. 停止
./out/xin stop

# 3. daemon 模式启动
./out/xin start
sleep 2
./out/xin status
# 应显示 PID

# 4. 热重载
./out/xin reload

# 5. 重启
./out/xin restart

# 6. 零宕机重启
./out/xin hot-restart

# 7. 停止
./out/xin stop

# 8. 查看帮助
./out/xin help
```

**自动化测试用例：**
```go
// Test: PID 文件管理
func TestPidFile(t *testing.T) {
    // 启动服务
    exec.Command("./out/xin", "start").Run()
    time.Sleep(time.Second)

    // 验证 PID 文件存在
    data, err := os.ReadFile("./xin.pid")
    assert.NoError(t, err)

    pid, err := strconv.Atoi(string(data))
    assert.NoError(t, err)
    assert.True(t, processExists(pid))

    // 停止服务
    exec.Command("./out/xin", "stop").Run()
    time.Sleep(time.Second)

    // 验证 PID 文件删除
    _, err = os.ReadFile("./xin.pid")
    assert.True(t, os.IsNotExist(err))
}

// Test: 状态命令
func TestStatus(t *testing.T) {
    // 启动服务
    exec.Command("./out/xin", "start").Run()
    defer exec.Command("./out/xin", "stop").Run()

    output, _ := exec.Command("./out/xin", "status").Output()
    assert.Contains(t, string(output), "PID:")
    assert.Contains(t, string(output), "running")
}
```

### 预期结果
- `start` 创建 PID 文件，服务后台运行
- `stop` 优雅终止服务
- `status` 显示运行状态和 PID
- PID 文件在服务停止后删除

---

## 11. 健康检查 (api/v1)

### 功能描述
- `GET /api/v1/health` 返回 `{"status": "ok"}`

### 验证步骤

**手动验证：**
```bash
# 1. 启动服务
./out/xin run &

# 2. 请求健康检查
curl http://localhost:8080/api/v1/health
# 预期: {"status":"ok"}

# 3. 停止
./out/xin stop
```

**自动化测试用例：```go
// Test: 健康检查端点
func TestHealthEndpoint(t *testing.T) {
    router := gin.New()
    api_v1.RegisterRoutes(router)

    req, _ := http.NewRequest("GET", "/api/v1/health", nil)
    w := httptest.NewRecorder()
    router.ServeHTTP(w, req)

    assert.Equal(t, 200, w.Code)
    assert.Contains(t, w.Body.String(), "ok")
}
```

### 预期结果
- HTTP 200
- JSON 响应 `{"status":"ok"}`

---

## 12. 统一响应格式 (pkg/resp)

### 功能描述
| 函数 | HTTP 状态码 | 业务码 |
|------|-------------|--------|
| `Success` | 200 | 0 |
| `Error` | 200 | 自定义 |
| `Unauthorized` | 401 | 401 |
| `Forbidden` | 403 | 403 |
| `BadRequest` | 400 | 400 |
| `NotFound` | 404 | 404 |
| `ServerError` | 500 | 500 |
| `Paginate` | 200 | 0 |

### 验证步骤

**手动验证：**
```bash
# 需要路由支持，以下为结构验证
curl -v http://localhost:8080/api/v1/health 2>&1
# 响应格式: {"code":0,"msg":"ok","data":{"status":"ok"}}
```

**自动化测试用例：```go
// Test: Success 响应
func TestSuccessResponse(t *testing.T) {
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)

    resp.Success(c, map[string]string{"key": "value"})

    assert.Equal(t, 200, w.Code)
    assert.Contains(t, w.Body.String(), `"code":0`)
    assert.Contains(t, w.Body.String(), `"msg":"ok"`)
}

// Test: Error 响应
func TestErrorResponse(t *testing.T) {
    w := httptest.NewRecorder()
    c, _ := gin.CreateTestContext(w)

    resp.Error(c, 1001, "custom error")

    assert.Equal(t, 200, w.Code) // 业务错误仍是 200
    assert.Contains(t, w.Body.String(), `"code":1001`)
    assert.Contains(t, w.Body.String(), `"msg":"custom error"`)
}

// Test: HTTP 错误响应
func TestHTTPErrorResponses(t *testing.T) {
    testCases := []struct {
        fn         func(*gin.Context, string)
        expCode    int
        expMsgCode int
    }{
        {resp.Unauthorized, 401, 401},
        {resp.Forbidden, 403, 403},
        {resp.BadRequest, 400, 400},
        {resp.NotFound, 404, 404},
        {resp.ServerError, 500, 500},
    }

    for _, tc := range testCases {
        w := httptest.NewRecorder()
        c, _ := gin.CreateTestContext(w)
        tc.fn(c, "test")

        assert.Equal(t, tc.expCode, w.Code)
        assert.Contains(t, w.Body.String(), fmt.Sprintf(`"code":%d`, tc.expMsgCode))
    }
}
```

### 预期结果
- 所有响应包含 `code`, `msg`, `data` 字段
- HTTP 状态码与响应函数对应

---

## 13. 启动初始化 (internal/core/boot)

### 功能描述
顺序初始化：Logger → DB → Cache → Server

### 验证步骤

**手动验证：**
```bash
# 1. 正常启动
./out/xin run &
sleep 3
# 检查日志顺序: logger init → db init → cache init → server start

# 2. 数据库配置错误启动失败
# 修改 config.yaml 为无效数据库
./out/xin run
# 预期: "boot init failed: ..."

# 3. Redis 配置错误（Required=true）启动失败
# 修改 redis.required: true 但 Redis 不可用
./out/xin run
# 预期: "boot init failed: ..."
```

**自动化测试用例：```go
// Test: 完整初始化流程
func TestBootInitSequence(t *testing.T) {
    // 使用内存数据库和 mock Redis
    cfg := &config.Config{
        App:     config.AppConfig{Port: 18080},
        Log:     config.LogConfig{Dir: "./test-logs", Level: "debug"},
        // ... 其他配置
    }

    srv, err := boot.Init(cfg)
    assert.NoError(t, err)
    assert.NotNil(t, srv)

    // 验证各组件已初始化
    assert.NotNil(t, db.Get())
    assert.NotNil(t, cache.Get())

    boot.Shutdown()
}

// Test: 初始化失败处理
func TestBootInitFailure(t *testing.T) {
    cfg := &config.Config{
        Database: config.DatabaseConfig{Host: "invalid"},
    }

    _, err := boot.Init(cfg)
    assert.Error(t, err)
    assert.Contains(t, err.Error(), "boot init failed")
}
```

### 预期结果
- 启动日志按顺序显示各组件初始化
- 失败时返回明确错误信息

---

## 14. 用户模型 (internal/module/user)

### 功能描述
定义 User, Role, Permission, Tenant 模型

### 验证步骤

**自动化测试用例：```go
// Test: User 模型字段
func TestUserModel(t *testing.T) {
    u := user.User{
        ID:       1,
        TenantID: 10,
        Username: "testuser",
        Email:    "test@example.com",
        Status:   1,
    }

    assert.Equal(t, uint(1), u.ID)
    assert.Equal(t, uint(10), u.TenantID)
    assert.Equal(t, "testuser", u.Username)
}

// Test: Tenant 模型
func TestTenantModel(t *testing.T) {
    t := user.Tenant{
        ID:    1,
        Name:  "Test Corp",
        Code:  "testcorp",
        Plan:  "enterprise",
        Status: 1,
    }

    assert.Equal(t, "enterprise", t.Plan)
}

// Test: GORM Tags
func TestModelTags(t *testing.T) {
    // 验证 table name
    assert.Equal(t, "users", (&user.User{}).TableName())
    assert.Equal(t, "tenants", (&user.Tenant{}).TableName())
}
```

### 预期结果
- TableName 返回正确的表名
- 字段标签正确

---

## 运行所有验证

```bash
# 1. 构建
./build.sh

# 2. 准备环境
# - PostgreSQL 运行在 localhost:5432
# - Redis 运行在 localhost:6379
# - 创建测试数据库

# 3. 启动服务（后台）
./out/xin start

# 4. 运行测试
go test -v ./...

# 5. 手动端到端测试
curl http://localhost:8080/api/v1/health
./out/xin status
./out/xin stop
```

---

## 更新日志

| 日期 | 更新内容 |
|------|----------|
| 2026-04-20 | 初始版本，记录 14 个功能模块及验证方案 |
