![](https://img.shields.io/badge/version-v4-green.svg) &nbsp; ![](https://img.shields.io/badge/version-go1.23-green.svg) &nbsp;  ![](https://img.shields.io/badge/builder-success-green.svg) &nbsp;

# Gin-Plus

一个基于 Gin 框架的增强库，提供更便捷的 Web 开发体验。详细文档请访问 [Gin-Plus 文档](https://eofhs2ef6g.feishu.cn/docx/UrDxd2p7coj2cWxuWShcYBF3nve)

## 特性

- 🚀 灵活启动  - 支持多种运行模式
- 🧩 依赖注入  - 自动管理对象依赖关系
- 🔧 配置管理  - 自动配置加载和注入
- 📋 MVC模式 - 规范化的路由和控制器定义
- 📝 统一日志  - 内置日志管理系统

## 安装

```bash
go get github.com/archine/gin-plus/v4@v4.1.7
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
