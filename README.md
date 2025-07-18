![](https://img.shields.io/badge/version-v4-green.svg) &nbsp; ![](https://img.shields.io/badge/version-go1.23-green.svg) &nbsp;  ![](https://img.shields.io/badge/builder-success-green.svg) &nbsp;

# Gin-Plus

一个基于 Gin 框架的增强库，提供更便捷的 Web 开发体验。支持依赖注入、配置管理、事件驱动等企业级特性。

## 特性

- 🚀 **灵活的应用启动** - 支持多种运行模式
- 🧩 **依赖注入容器** - 自动管理对象依赖关系
- 📋 **MVC控制器** - 规范化的路由和控制器定义
- 🔧 **配置管理** - 自动配置加载和注入
- 📡 **事件驱动** - 应用生命周期事件支持
- 🔒 **中间件支持** - 灵活的中间件注册机制
- 📝 **统一日志** - 内置日志管理系统
- 📨 **标准响应** - 统一的API响应格式

## 安装

```bash
go get github.com/archine/gin-plus/v4@v4.0.7
```

## 快速开始

### 1. 最简单的应用

```go
package main

import ginplus "github.com/archine/gin-plus/v4"

func main() {
    // 创建应用并启动
    ginplus.Default().Run(ginplus.ServerMode)
}
```

## 详细使用指南

### 1. 应用创建与配置

#### 1.1 创建应用的三种方式

```go
package main

import (
    ginplus "github.com/archine/gin-plus/v4"
    "github.com/gin-gonic/gin"
)

func main() {
    // 方式1：使用默认配置
    app1 := ginplus.Default()
    
    // 方式2：创建空应用，手动配置
    app2 := ginplus.New()
    
    // 方式3：创建时配置选项
    app3 := ginplus.New(
        ginplus.WithMiddleware(gin.Logger(), gin.Recovery()),
        ginplus.WithBanner("我的应用 v1.0"),
    )
    
    app1.Run(ginplus.ServerMode)
}
```

#### 1.2 应用运行模式

```go
// InitializeMode: 仅初始化配置和日志
app.Run(ginplus.InitializeMode)

// ContainerMode: 初始化配置、日志和IoC容器，不启动HTTP服务器
app.Run(ginplus.ContainerMode)

// ServerMode: 完整启动模式，启动HTTP服务器
app.Run(ginplus.ServerMode)
```

#### 1.3 链式配置应用

```go
app := ginplus.New().
    With(ginplus.WithMiddleware(gin.Logger())).
    With(ginplus.WithBanner("自定义横幅"))

app.Run(ginplus.ServerMode)
```

### 2. 配置管理

#### 2.1 默认配置提供者

框架默认使用文件配置提供者，自动加载以下配置文件（按优先级排序）：
- `application-{profile}.yml` （如果设置了profile）
- `application.yml`
- `application.yaml`
- `application.json`

配置文件查找路径：
1. 当前工作目录
2. `./config/` 目录
3. `./conf/` 目录

#### 2.2 配置文件示例 (application.yml)

```yaml
server:
  port: 8080
  context-path: /api
  read-timeout: 30s
  write-timeout: 30s
  idle-timeout: 60s

database:
  host: localhost
  port: 3306
  username: root
  password: 123456

app:
  name: gin-plus-demo
  version: 1.0.0
  profile: dev

logging:
  level: info
  format: json
  file: ./logs/app.log
```

#### 2.3 自定义配置提供者

```go
package main

import (
    ginplus "github.com/archine/gin-plus/v4"
    "github.com/archine/gin-plus/v4/component/config"
)

func main() {
    app := ginplus.New(
        // 自定义配置提供者
        ginplus.WithConfigProvider(func() config.Provider {
            return config.NewFileProvider("custom-config.yml") // 指定配置文件
        }),
    )
    
    app.Run(ginplus.ServerMode)
}
```

#### 2.4 在Bean中使用配置

```go
import "github.com/archine/gin-plus/v4/sysconf"

type DatabaseService struct {
    ioc.Bean
    host     string
    port     int
    username string
    password string
}

func (d *DatabaseService) BeanPostConstruct() {
    // 通过全局配置获取配置值
    cfg := sysconf.GlobalProvider
    d.host = cfg.GetString("database.host")
    d.port = cfg.GetInt("database.port")
    d.username = cfg.GetString("database.username")
    d.password = cfg.GetString("database.password")
    
    fmt.Printf("数据库配置: %s:%d\n", d.host, d.port)
}
```

### 3. 日志管理

#### 3.1 默认日志实现

框架内置了日志管理系统，默认使用logrus作为日志实现，支持：
- 多种日志级别（Debug, Info, Warn, Error, Fatal）
- 日志格式配置（text/json）
- 文件输出支持
- 日志轮转

#### 3.2 日志配置

```yaml
logging:
  level: info          # 日志级别：debug, info, warn, error, fatal
  format: json         # 日志格式：text, json
  file: ./logs/app.log # 日志文件路径
  max-size: 100        # 单个日志文件最大大小(MB)
  max-backups: 7       # 保留的旧日志文件数量
  max-age: 30          # 日志文件保留天数
```

#### 3.3 使用日志

```go
import "github.com/archine/gin-plus/v4/syslog"

func (u *UserService) GetUserById(id string) map[string]any {
    syslog.Info("查询用户", "id", id)
    
    user := map[string]any{
        "id":   id,
        "name": "用户" + id,
    }
    
    syslog.Debug("查询结果", "user", user)
    return user
}
```

#### 3.4 自定义日志提供者

```go
import "github.com/archine/gin-plus/v4/component/log"

func main() {
    app := ginplus.New(
        ginplus.WithLogProvider(func() log.Provider {
            // 返回自定义日志实现
            return log.NewLogrusProvider()
        }),
    )
    
    app.Run(ginplus.ServerMode)
}
```

### 4. Server配置

#### 4.1 Server配置项

```yaml
server:
  port: 8080                    # 服务端口，默认8080
  context-path: /api            # 上下文路径，默认为空
  read-timeout: 30s             # 读取超时时间
  write-timeout: 30s            # 写入超时时间
  idle-timeout: 60s             # 空闲超时时间
  max-header-bytes: 1048576     # 最大header字节数
  tls:
    enabled: false              # 是否启用TLS
    cert-file: server.crt       # 证书文件路径
    key-file: server.key        # 私钥文件路径
```

#### 4.2 程序化Server配置

```go
func main() {
    app := ginplus.New(
        ginplus.WithServerConfig(func(cfg *ginplus.ServerConfig) {
            cfg.Port = 9090
            cfg.ContextPath = "/v1"
            cfg.ReadTimeout = time.Second * 60
            cfg.WriteTimeout = time.Second * 60
        }),
    )
    
    app.Run(g
  context-path: /api

database:
  host: localhost
  port: 3306
  username: root
  password: 123456

app:
  name: gin-plus-demo
  version: 1.0.0
```

#### 5.2 自定义配置提供者

```go
package main

import (
    ginplus "github.com/archine/gin-plus/v4"
    "github.com/archine/gin-plus/v4/component/config"
)

func main() {
    app := ginplus.New(
        // 自定义配置提供者
        ginplus.WithConfigProvider(func() config.Provider {
            return config.NewFileProvider() // 或其他自定义实现
        }),
    )
    
    app.Run(ginplus.ServerMode)
}
```

#### 5.3 在Bean中使用配置

```go
type DatabaseService struct {
    ioc.Bean
    host     string
    port     int
    username string
    password string
}

func (d *DatabaseService) BeanPostConstruct() {
    // 通过全局配置获取配置值
    cfg := sysconf.GlobalProvider
    d.host = cfg.GetString("database.host")
    d.port = cfg.GetInt("database.port")
    d.username = cfg.GetString("database.username")
    d.password = cfg.GetString("database.password")
    
    fmt.Printf("数据库配置: %s:%d\n", d.host, d.port)
}
```

### 6. 事件处理

#### 6.1 生命周期事件

```go
package events

import (
    "context"
    "fmt"
    ginplus "github.com/archine/gin-plus/v4"
    "github.com/archine/gin-plus/v4/app"
    "github.com/archine/gin-plus/v4/component/config"
)

type AppLifecycleEvent struct{}

func (e *AppLifecycleEvent) Order() int {
    return 1  // 事件执行优先级，数字越小优先级越高
}

func (e *AppLifecycleEvent) OnStarting() bool {
    fmt.Println("应用启动前的准备工作...")
    // 返回true继续启动，false中止启动
    return true
}

func (e *AppLifecycleEvent) OnStarted() {
    fmt.Println("应用启动完成！")
}

func (e *AppLifecycleEvent) OnStopped(ctx context.Context) {
    fmt.Println("应用停止，执行清理工作...")
}

// 注册事件
func init() {
    app := ginplus.New(
        ginplus.WithEvent(&AppLifecycleEvent{}),
    )
}
```

#### 6.2 配置加载事件

```go
type ConfigEvent struct{}

func (e *ConfigEvent) Order() int {
    return 1
}

func (e *ConfigEvent) OnConfigAfterLoad(cp config.Provider) {
    fmt.Printf("应用名称: %s\n", cp.GetString("app.name"))
    fmt.Printf("应用版本: %s\n", cp.GetString("app.version"))
}
```

#### 6.3 容器刷新事件

```go
type ContainerEvent struct{}

func (e *ContainerEvent) Order() int {
    return 1
}

func (e *ContainerEvent) OnContainerRefreshBefore(ctx app.ApplicationContext) {
    fmt.Println("容器刷新前，注册额外的Bean...")
    
    // 动态注册Bean实例
    customService := &CustomService{}
    ctx.RegisterBean("customService", customService)
}

func (e *ContainerEvent) OnContainerRefreshAfter(ctx app.ApplicationContext) {
    fmt.Println("容器刷新完成！")
    
    // 获取Bean进行验证
    if bean, ok := ctx.GetBean("customService"); ok {
        fmt.Printf("成功获取Bean: %T\n", bean)
    }
}
```

### 7. 完整示例

#### 7.1 项目结构

```
gin-plus-demo/
├── main.go
├── controller/
│   └── user_controller.go
├── service/
│   └── user_service.go
├── events/
│   └── app_events.go
└── application.yml
```

#### 7.2 完整的应用示例

```go
// main.go
package main

import (
    _ "gin-plus-demo/controller"  // 导入控制器包
    _ "gin-plus-demo/events"      // 导入事件包
    ginplus "github.com/archine/gin-plus/v4"
    "github.com/gin-gonic/gin"
)

func main() {
    app := ginplus.New(
        ginplus.WithMiddleware(gin.Logger(), gin.Recovery()),
        ginplus.WithBanner("🚀 Gin-Plus Demo v1.0"),
    )
    
    app.Run(ginplus.ServerMode)
}
```

```go
// service/user_service.go
package service

import (
    "fmt"
    "github.com/archine/gin-plus/v4/component/ioc"
)

func init() {
    _ = ioc.RegisterBeanDef(&UserService{})
}

type UserService struct {
    ioc.Bean
}

func (s *UserService) BeanPostConstruct() {
    fmt.Println("UserService 初始化完成")
}

func (s *UserService) GetUserById(id string) map[string]any {
    return map[string]any{
        "id":   id,
        "name": "用户" + id,
        "age":  25,
    }
}
```

```go
// controller/user_controller.go
package controller

import (
    "gin-plus-demo/service"
    "github.com/archine/gin-plus/v4/component/ioc"
    "github.com/archine/gin-plus/v4/component/mvc"
    "github.com/gin-gonic/gin"
)

func init() {
    _ = ioc.RegisterBeanDef(&UserController{})
}

type UserController struct {
    mvc.Controller
    UserService *service.UserService `autowire:""`
}

func (u *UserController) SetRoutes(group *gin.RouterGroup) {
    group.GET("/user/:id", u.GetUser)
}

func (u *UserController) GetUser(ctx *gin.Context) {
    id := ctx.Param("id")
    user := u.UserService.GetUserById(id)
    ctx.JSON(200, gin.H{
        "code": 200,
        "data": user,
        "message": "success",
    })
}
```

## 常见问题

### Q: 如何确保控制器被正确注册？
A: 在main.go中导入控制器包，确保init函数被执行：
```go
import _ "your-project/controller"
```

### Q: 依赖注入失败怎么办？
A: 检查以下几点：
1. Bean是否正确注册 (`ioc.RegisterBeanDef`)
2. 字段是否添加了 `autowire` 标签
3. 字段类型是否匹配

### Q: 如何自定义配置文件路径？
A: 使用自定义配置提供者：
```go
ginplus.WithConfigProvider(func() config.Provider {
    // 返回自定义配置提供者
    return config.NewFileProvider()
})
```

## 许可证

本项目采用 MIT 许可证，详情请查看 [LICENSE](LICENSE) 文件。