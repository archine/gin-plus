![](https://img.shields.io/badge/version-v3.x-green.svg) &nbsp; ![](https://img.shields.io/badge/version-go1.21-green.svg) &nbsp;  ![](https://img.shields.io/badge/builder-success-green.svg) &nbsp;

> Gin框架的增强版，集成IOC、MVC等概念。只做增强，不做改变，帮您更好的进行Go的Web开发。搭配 [🍳Goland](https://plugins.jetbrains.com/plugin/20652-iocer/versions) 插件可以事半功倍哦！！！😀😀

## 一、前言
详细文档点击前往：[文档](https://eofhs2ef6g.feishu.cn/docx/AXCvdf5jPogZ12xOXHucmgo5nFb)
### 1、安装

* Get
```bash
go get github.com/archine/gin-plus/v3@v3.3.3
```

* Mod
```bash
# go.mod文件加入下面的一条
github.com/archine/gin-plus/v3 v3.3.3

# 命令行在该项目目录下执行
go mod tidy
```
## 二、使用说明

### 1、快速开始

* controller接口
```go
package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/archine/gin-plus/v3/mvc"
	"github.com/archine/gin-plus/v3/resp"
)

type TestController struct {
	// 声明该结构体为控制器
	mvc.Controller
}

// Hello
// @GET(path="/hello") 定义的 get 方法
func (t *TestController) Hello(ctx *gin.Context) {
	resp.Ok(ctx)
}
```

* 启动类
```go
package main

import (
	_ "gin-plus-demo/controller"
	"github.com/archine/gin-plus/v3/application"
)

//go:generate gp-ast
func main() {
	application.Default().Run()
}
```

这时候运行该项目，浏览器访问http://localhost:4006/hello即可

### 2、方法路径前缀
很多时候，我们需要对整个 Controller 里的所有 API 增加访问前缀，这时我们可在 Controller 的结构体注释中通过`@BasePath("/xxx")`来进行声明
```go
package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/archine/gin-plus/v3/mvc"
	"github.com/archine/gin-plus/v3/resp"
)

// TestController 增加固定路径前缀 /test
// @BasePath("/test")
type TestController struct {
	mvc.Controller
}

// Hello
// @GET(path="/hello") 第一个接口
func (t *TestController) Hello(ctx *gin.Context) {
	resp.Json(ctx, "hello world")
}
```
重新启动项目后，浏览器访问http://localhost:4006/test/hello即可

### 3、依赖注入

对结构体中的属性进行依赖注入，下面的例子中，我们为 controller 注入一个 mapper。对 IoC 不熟悉可前往文档查看: 👓[点击前往](http://github.com/archine/ioc)
* mapper
```go
package mapper

import "github.com/archine/ioc"

type TestMapper struct{}

func (t *TestMapper) CreateBean() ioc.Bean {
  return &TestMapper{}
}

// Say 测试依赖注入
func (t *TestMapper) Say() string {
  return "success"
}
```
* controller
```go
package controller

import (
    "gin-plus-demo/mapper"
    "github.com/gin-gonic/gin"
    "github.com/archine/gin-plus/v3/mvc"
    "github.com/archine/gin-plus/v3/resp"
)

type TestController struct {
    mvc.Controller
    TestMapper *mapper.TestMapper
}

// Hello
// @GET(path="/hello") 第一个接口
func (t *TestController) Hello(ctx *gin.Context) {
    // 使用时直接调用即可
    resp.Json(ctx, t.TestMapper.Say())
}
```
