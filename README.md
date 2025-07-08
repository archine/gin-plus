![](https://img.shields.io/badge/version-v4.x-green.svg) &nbsp; ![](https://img.shields.io/badge/version-go1.21-green.svg) &nbsp;  ![](https://img.shields.io/badge/builder-success-green.svg) &nbsp;

# Gin-Plus

一个基于 Gin 框架的增强库，提供更便捷的 Web 开发体验。

## 特性

- 🚀 **简化路由配置** - 更直观的路由定义方式
- 🧩 **依赖注入支持** - 内置依赖注入容器，便于解耦与测试
- 📊 **响应封装** - 统一的 API 响应格式

## 一、前言
详细文档点击前往：[文档](https://eofhs2ef6g.feishu.cn/docx/AXCvdf5jPogZ12xOXHucmgo5nFb)

### 1、安装

* Get
```bash
go get github.com/archine/gin-plus/v4@v4.0.0
```

* Mod
```bash
# go.mod文件加入下面的一条
github.com/archine/gin-plus/v4 v4.0.0
```

## 二、快速开始

### 1、基础使用

```go
package main

import (
    "github.com/archine/gin-plus/v4"
    "github.com/gin-gonic/gin"
)

func main() {
    // 创建 gin-plus 应用
    app := ginplus.New()
    
    // 基础路由
    app.GET("/hello", func(c *gin.Context) {
        ginplus.Success(c, "Hello, Gin-Plus!")
    })
    
    // 启动服务
    app.Run(":8080")
}
```

### 2、路由组和中间件

```go
// 创建路由组
api := app.Group("/api/v1")
{
    // 应用中间件
    api.Use(ginplus.Logger(), ginplus.CORS())
    
    // 用户相关路由
    user := api.Group("/user")
    {
        user.POST("/login", loginHandler)
        user.GET("/profile", ginplus.Auth(), profileHandler)
        user.PUT("/profile", ginplus.Auth(), updateProfileHandler)
    }
    
    // 商品相关路由
    product := api.Group("/product")
    {
        product.GET("/list", productListHandler)
        product.POST("/", ginplus.Auth(), createProductHandler)
    }
}
```

### 3、请求参数验证

```go
type LoginRequest struct {
    Username string `json:"username" binding:"required,min=3"`
    Password string `json:"password" binding:"required,min=6"`
}

func loginHandler(c *gin.Context) {
    var req LoginRequest
    
    // 自动绑定和验证参数
    if err := ginplus.BindJSON(c, &req); err != nil {
        ginplus.Error(c, 400, "参数验证失败", err.Error())
        return
    }
    
    // 业务逻辑处理
    token, err := userService.Login(req.Username, req.Password)
    if err != nil {
        ginplus.Error(c, 401, "登录失败", err.Error())
        return
    }
    
    ginplus.Success(c, gin.H{"token": token})
}
```

## 三、高级功能

### 1、统一响应格式

```go
// 成功响应
ginplus.Success(c, data)

// 错误响应
ginplus.Error(c, code, message, details)

// 分页响应
ginplus.Page(c, data, total, page, size)

// 自定义响应
ginplus.Response(c, code, message, data)
```

### 2、配置管理

```go
// 从配置文件加载
config := ginplus.LoadConfig("config.yaml")

// 设置配置
app.SetConfig(config)

// 获取配置值
dbConfig := ginplus.GetConfig("database")
```

### 3、数据库集成

```go
// 初始化数据库
db := ginplus.InitDB(config.Database)

// 自动迁移
ginplus.AutoMigrate(db, &User{}, &Product{})

// 在处理器中使用
func getUserHandler(c *gin.Context) {
    db := ginplus.GetDB(c)
    
    var user User
    if err := db.First(&user, c.Param("id")).Error; err != nil {
        ginplus.Error(c, 404, "用户不存在")
        return
    }
    
    ginplus.Success(c, user)
}
```

## 四、中间件

### 1、内置中间件

```go
// 日志中间件
app.Use(ginplus.Logger())

// CORS 跨域
app.Use(ginplus.CORS())

// 限流中间件
app.Use(ginplus.RateLimit(100, time.Minute))

// JWT 认证
app.Use(ginplus.JWT("your-secret-key"))

// 请求追踪
app.Use(ginplus.RequestID())
```

### 2、自定义中间件

```go
func CustomMiddleware() gin.HandlerFunc {
    return ginplus.Middleware(func(c *gin.Context) {
        // 前置处理
        start := time.Now()
        
        // 执行下一个中间件
        c.Next()
        
        // 后置处理
        latency := time.Since(start)
        ginplus.Logger().Info("Request processed", "latency", latency)
    })
}
```

## 五、部署配置

### 1、生产环境配置

```yaml
# config.yaml
server:
  port: 8080
  mode: release
  
database:
  driver: mysql
  dsn: "user:password@tcp(localhost:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
  
redis:
  addr: "localhost:6379"
  password: ""
  db: 0
  
jwt:
  secret: "your-jwt-secret"
  expire: 7200
```

### 2、Docker 部署

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
COPY --from=builder /app/config.yaml .
CMD ["./main"]
```

## 六、贡献指南

欢迎提交 Issue 和 Pull Request 来改进这个项目。

### 开发环境设置

```bash
# 克隆项目
git clone https://github.com/archine/gin-plus.git

# 安装依赖
go mod download

# 运行测试
go test ./...
```

## 许可证

本项目采用 MIT 许可证，详情请查看 [LICENSE](LICENSE) 文件。