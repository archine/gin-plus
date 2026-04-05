![](https://img.shields.io/badge/version-v4-green.svg) &nbsp; ![](https://img.shields.io/badge/version-go1.23-green.svg) &nbsp;  ![](https://img.shields.io/badge/builder-success-green.svg) &nbsp;

# Gin-Plus

轻量增强 Gin 的工程化框架 — 保持 Gin 兼容性，提供 IoC、MVC、生命周期与模块化能力。目标是“增强，不改动”，帮助团队构建可维护、高可测、可扩展的 Go Web 服务。通过非侵入式扩展，把依赖注入、控制器化与工程化能力加入到 Gin 中，降低样板代码、提升项目规范与可测试性🚀。详细文档请访问 [Gin-Plus 文档](https://eofhs2ef6g.feishu.cn/docx/UrDxd2p7coj2cWxuWShcYBF3nve)

## 特性

- 🚀 与 Gin 无缝兼容：尽量不破坏原有 Gin 行为与调用习惯
- 🧩 IoC / 依赖注入：自动装配服务、控制器与资源，降低手工 wiring 成本
- 🏛 MVC 与控制器风格：约定式路由绑定，控制器方法直观映射 HTTP 接口
- 🔌 中间件体系化：可配置、中继与组合的中间件注册方式
- ♻️ 生命周期与资源管理：支持组件初始化、优雅关闭钩子（DB、缓存、任务等）
- 📦 模块化/插件化：代码按模块组织，方便按需加载与复用
- 🧰 工程化工具集：统一响应格式、全局错误处理、日志适配器、配置注入等
- ⚡ 性能优先：增强体验同时尽量保持 Gin 的高性能和低延迟

## 安装

```bash
go get github.com/archine/gin-plus/v4@latest
```

## 快速开始
### 1、定义Service
```go
package service

import (
	"github.com/archine/gin-plus/v4/component/ioc/bean"
	"github.com/archine/gin-plus/v4/component/ioc"
)

func init() {
	ioc.RegisterBeanDef(&TestService{})
}

type TestService struct {
	bean.Bean
	Age int `value:"${user.age:18}"`
}

func (t *TestService) GetAge() int {
	return t.Age
}
```
### 2、定义Controller

```go
package controller

import (
	"fmt"
	"gin-plus-demo-v4/service"
	"github.com/archine/gin-plus/v4/component/ioc"
	"github.com/archine/gin-plus/v4/resp"
	"github.com/gin-gonic/gin"
)

func init() {
	ioc.RegisterBeanDef(&TestController{})
}

type TestController struct {
	testService *service.TestService `autowire:""`
}

func (t *TestController) SetRoutes(rootGroup *gin.RouterGroup) {
	rootGroup.GET("/test/hello", t.sayHello)
}

func (t *TestController) sayHello(ctx *gin.Context) {
	resp.Json(ctx, t.testService.GetAge())
}
```

### 3、启动应用

```go
package main

import (
	_ "gin-plus-demo-v4/controller"
	ginplus "github.com/archine/gin-plus/v4"
)

func main() {
	ginplus.New().Run(ginplus.ServerMode)
}
```
