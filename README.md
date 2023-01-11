![](https://img.shields.io/badge/version-v2.0.0-green.svg) &nbsp; ![](https://img.shields.io/badge/builder-success-green.svg) &nbsp;

> 📢📢📢 Gin增强版，集成了IOC、MVC设计思想，API定义采用 restful 风格。可帮你快速的进行 web 项目开发，搭配 [🍳Goland](https://plugins.jetbrains.com/plugin/20652-iocer/versions) 插件可以事半功倍哦！！！😀😀

## 一、前言
### 1、安装

- Get
```bash
go get github.com/archine/gin-plus/v2@v2.0.0
```

- Mod
```bash
# go.mod文件加入下面的一条
github.com/archine/gin-plus/v2 v2.0.0

# 命令行在该项目目录下执行
go mod tidy
```

### 2、🌱🌱运行前置条件

- **（1）Goland运行**

运行前，需要做如下配置
![1](https://img-blog.csdnimg.cn/332199e5a62947e8ac80be6d47248e1f.png)
![2](https://img-blog.csdnimg.cn/1b519d9ef06746fd884a71c53db838f6.png)
![3](https://img-blog.csdnimg.cn/c209572927c94cec959c2e0749fb978a.png)


- **（2）其他方式运行**

在执行 `go build` 前，必须先执行 `go generate`

- **（3）项目结构**

项目需要存在 base 目录，同时目录中需要存在 template.go 文件，文件初始化内容为
```go
package base

// 自动生成,请不要编辑

import "github.com/archine/gin-plus/v2/ast"

var Ast = map[string][]*ast.MethodInfo{}
```
## 二、项目使用
本框架声明 API 的方式非常简单，只需在方法的注释中通过如下方式进行声明即可，启动时会自动应用，**需要注意的是，API函数名必须大写**

| 定义方式🍑 | 描述🍎 | 快捷键🍓 |
| --- | --- | --- |
| @GET(path="/hello", globalFunc=true) | Get 请求 | 空白处输入 get |
| @POST(path="/hello", globalFunc=true) | Post 请求 | 空白处输入 post |
| @DELETE(path="/hello", globalFunc=true) | Delete 请求 | 空白处输入 del |
| @PUT(path="/hello", globalFunc=true) | Put 请求 | 空白处输入 put |
| @PATCH(path="/hello", globalFunc=true) | Patch 请求 | 暂无 |
| @HEAD(path="/hello", globalFunc=true) | Head 请求 | 暂无 |
| @OPTIONS(path="/hello", globalFunc=true) | Options 请求 | 暂无 |
| @BasePath("/hello") | 基础路径 | 空白处输入 basep |

其中 `globalFunc`参数为当前 API 是否需要应用全局函数，`true` 表示应用，`false` 表示不应用。
### 1、快速开始

- controller接口
```go
package controller

import (
    "github.com/gin-gonic/gin"
    "github.com/archine/gin-plus/v2/mvc"
    "github.com/archine/gin-plus/v2/resp"
)

type TestController struct {
    // 声明该结构体为控制器
    mvc.Controller
}

func init() {
    // 注册当前控制器到 MVC
    mvc.Register(&TestController{})
}

// Hello
// @GET(path="/hello", globalFunc=true) 定义的 get 方法
func (t *TestController) Hello(ctx *gin.Context) {
    resp.Ok(ctx)
}
```

- 启动类
```go
package main

import (
    _ "gin-plus-demo/controller"
    "github.com/gin-gonic/gin"
    "github.com/archine/gin-plus/v2/ast"
    "github.com/archine/gin-plus/v2/mvc"
    "log"
    "os"
)

//go:generate go run main.go ast
func main() {
    if len(os.Args) > 1 && os.Args[1] == "ast" {
         ast.Parse("controller") // 接口所在的目录
         return
    }
    gin.SetMode(gin.ReleaseMode)
    engine := gin.New()
    // 将 gin 的引擎加入到 mvc 中，第二个参数为依赖注入的开关
    mvc.Apply(engine, true, base.Ast)
    if err := engine.Run(":4006"); err != nil {
        log.Fatalf(err.Error())
    }
}
```

这时候运行该项目，浏览器访问http://localhost:4006/hello即可

![在这里插入图片描述](https://img-blog.csdnimg.cn/27837bfb5714484eac33932392929d7e.png)

### 2、方法路径前缀
很多时候，我们需要对整个 Controller 里的所有 API 增加上固定的前缀，这时我们可在 Controller 的结构体注释中通过`@BasePath("/xxx")`来进行声明
```go
package controller

import (
    "github.com/gin-gonic/gin"
    "github.com/archine/gin-plus/v2/mvc"
    "github.com/archine/gin-plus/v2/resp"
)

// TestController 增加固定路径前缀 /test
// @BasePath("/test")
type TestController struct {
    mvc.Controller
}

func init() {
    mvc.Register(&TestController{})
}

// Hello
// @GET(path="/hello", globalFunc=true) 第一个接口
func (t *TestController) Hello(ctx *gin.Context) {
    resp.Json(ctx, "hello world")
}
```

重新启动项目后，浏览器访问http://localhost:4006/test/hello即可

![在这里插入图片描述](https://img-blog.csdnimg.cn/5d84177e137f4033a7ec517e72579704.png)


### 3、全局函数
全局函数会生效于全部 Controller 中的所有 API，该函数会在调用具体 API 之前触发。这里我们就通过一个日志打印的函数来演示

- controller

这里我们定义两个 API，当我们访问时，都会打印全局函数中的日志，因为都进行了应用，第三个 API 设置为了 false，因此不会应用全局函数
```go
package controller

import (
    "github.com/gin-gonic/gin"
    "github.com/archine/gin-plus/v2/mvc"
    "github.com/archine/gin-plus/v2/resp"
)

type TestController struct {
    mvc.Controller
}

func init() {
    mvc.Register(&TestController{})
}

// Hello
// @GET(path="/hello", globalFunc=true) 第一个接口
func (t *TestController) Hello(ctx *gin.Context) {
    resp.Json(ctx, "hello world")
}

// Hello2
// @GET(path="/hello_2", globalFunc=true) 第二个接口
func (t *TestController) Hello2(ctx *gin.Context) {
    resp.Ok(ctx)
}

// Hello3
// @GET(path="/hello_3", globalFunc=false) 第三个接口，不应用全局函数
func (t *TestController) Hello3(ctx *gin.Context) {
    resp.Ok(ctx)
}
```

- 启动类
```go
package main

import (
    _ "gin-plus-demo/controller"
    "github.com/gin-gonic/gin"
    log "github.com/sirupsen/logrus"
    "github.com/archine/gin-plus/v2/mvc"
    "github.com/archine/gin-plus/v2/plugin"
    "os"
)

//go:generate go run main.go ast
func main() {
    if len(os.Args) > 1 && os.Args[1] == "ast" {
        ast.Parse("controller")
        return
    }
    gin.SetMode(gin.ReleaseMode)
    engine := gin.New()

    // 第四个参数为可变参数，意味着你可以添加多个全局函数
    mvc.Apply(engine, true, base.Ast, func(context *gin.Context) {
        log.Info("我是全局函数")
    })
    if err := engine.Run(":4006"); err != nil {
        log.Fatalf(err.Error())
    }
}
```

这时候，我们通过浏览器访问这三个 API，可以看到只有前两个 API 才会打印全局函数中的日志

![在这里插入图片描述](https://img-blog.csdnimg.cn/7965aa43aa344b1192ce53d5bb38690a.png)


### 4、局部函数
此函数主要应用于某一个具体的 Controller，下面的例子中，定义了两个 API，这里演示只为 Hello函数增加，💡 如果安装了 IoCer 插件，可输入 **callb**进行快速生成
```go
package controller

import (
    "github.com/gin-gonic/gin"
    log "github.com/sirupsen/logrus"
    "github.com/archine/gin-plus/v2/mvc"
    "github.com/archine/gin-plus/v2/resp"
)

type TestController struct {
    mvc.Controller
}

func init() {
    mvc.Register(&TestController{})
}

// CallBefore 前置处理
func (t *TestController) CallBefore(funcName string) []gin.HandlerFunc {
    if funcName == "Hello" { // 这里可通过函数名来控制具体给哪个函数增加局部函数
        return []gin.HandlerFunc{func(context *gin.Context) {
            log.Info("我是局部函数")
        }}
    }
    return nil
}

// Hello
// @GET(path="/hello", globalFunc=true) 第一个接口
func (t *TestController) Hello(ctx *gin.Context) {
    resp.Json(ctx, "hello world")
}

// Hello2
// @GET(path="/hello_2", globalFunc=true) 第二个接口
func (t *TestController) Hello2(ctx *gin.Context) {
    resp.Ok(ctx)
}
```

这时候通过浏览器访问这两个 API ，只有第一个 API 才会打印日志

![在这里插入图片描述](https://img-blog.csdnimg.cn/4043eaa924a041439af1b3e5eaf72802.png)


### 5、依赖注入
对结构体中的属性进行依赖注入，下面的例子中，我们为 controller 注入一个 mapper。对 **IoC** 不熟悉可前往文档查看: [👓点击前往](https://github.com/archine/ioc)

- service
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

- controller
```go
package controller

import (
    "gin-plus-demo/mapper"
    "github.com/gin-gonic/gin"
    "github.com/archine/gin-plus/v2/mvc"
    "github.com/archine/gin-plus/v2/resp"
)

type TestController struct {
    mvc.Controller
    // 注入TestMapper。为了他人直观知道该属性为依赖注入进来的，可在注入的属性右边加入声明（😊建议）
    // 安装了 Iocer 插件的话，可直接在 属性右边 输入 di，可快速生成
    TestMapper *mapper.TestMapper `@autowired:""`
}

func init() {
    mvc.Register(&TestController{})
}

// Hello
// @GET(path="/hello", globalFunc=true) 第一个接口
func (t *TestController) Hello(ctx *gin.Context) {
    // 使用时直接调用即可
    resp.Json(ctx, t.TestMapper.Say())
}
```
### 6、后置处理器
该处理器在 Controller 实例化结束且依赖注入完成后触发，我们可在该函数做其他的一些属性处理，这里例子为 赋值 controller 中的一些私有属性，💡 如果安装了 IoCer 插件，可输入 **pc** 进行快速生成
```go
package controller

import (
    "github.com/gin-gonic/gin"
    "github.com/archine/gin-plus/v2/mvc"
    "github.com/archine/gin-plus/v2/resp"
)

type TestController struct {
    mvc.Controller
    age int
}

func init() {
    mvc.Register(&TestController{})
}

// PostConstruct 初始化私有属性 age 的值
func (t *TestController) PostConstruct() {
    t.age = 100
}
```
### 7、全局异常捕获
在开发中，处理 **error **是个让人头大的问题，很多开发者都是通过一层层的 return，这其实代码很不美观，这里我们提供了全局异常捕获，会对 API 整个调用链进行异常捕获。这时，在碰到 **error **时，可直接采用 panic 的方式，框架中提供了 exception.OrThrow(err)来进行 err 不为 nil 时抛出，💡 如果安装了 IoCer 插件，可输入 **thr** 进行快速生成。下面为应用的例子

```go
package main

import (
    _ "gin-plus-demo/controller"
    "github.com/gin-gonic/gin"
    log "github.com/sirupsen/logrus"
    "github.com/archine/gin-plus/v2/ast"
    "github.com/archine/gin-plus/v2/exception"
    "github.com/archine/gin-plus/v2/mvc"
    "os"
)

//go:generate go run main.go ast
func main() {
    if len(os.Args) > 1 && os.Args[1] == "ast" {
        ast.Parse("controller")
        return
    }
    gin.SetMode(gin.ReleaseMode)
    engine := gin.New()
    // 加入全局异常处理器
    engine.Use(exception.GlobalExceptionInterceptor)
    mvc.Apply(engine, true, base.Ast)
    if err := engine.Run(":4006"); err != nil {
        log.Fatalf(err.Error())
    }
}
```

### 8、日志插件
更改 Gin 中默认得日志插件
```go
package main

import (
	"gin-plus-demo/base"
	_ "gin-plus-demo/controller"
	"github.com/gin-gonic/gin"
	"github.com/archine/gin-plus/v2/ast"
	"github.com/archine/gin-plus/v2/mvc"
	"github.com/archine/gin-plus/v2/plugin"
	"log"
	"os"
)

//go:generate go run main.go ast
func main() {
	if len(os.Args) > 1 && os.Args[1] == "ast" {
		ast.Parse("controller")
		return
	}
	plugin.InitLog("debug") // 先初始化日志级别
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(plugin.LogMiddleware()) // 在运行前加入即可
	mvc.Apply(engine, true, base.Ast)
	if err := engine.Run(":4006"); err != nil {
		log.Fatalf(err.Error())
	}
}
```
## 三、统一返回体
### 1、快速返回
返回 code 和 msg，常用于只告知客户端是否成功，项目中通过`resp.Ok()`调用，💡 如果安装了 IoCer 插件，可输入 **ro** 进行快速生成

```go
package controller

import (
    "github.com/gin-gonic/gin"
    "github.com/archine/gin-plus/v2/mvc"
    "github.com/archine/gin-plus/v2/resp"
)

type TestController struct {
    mvc.Controller
}

func init() {
    mvc.Register(&TestController{})
}

// Hello
// @GET(path="/hello", globalFunc=true) Hello 第一个接口
func (t *TestController) Hello(ctx *gin.Context) {
    // 快速返回
    resp.Ok(ctx)
}
```

- 响应结构
```json
{
  "err_code":0,
  "err_msg":"OK"
}
```
### 2、错误的请求
业务级别异常，返回错误的 code 和 msg，项目中通过`resp.BadRequest()`调用，💡 如果安装了 IoCer 插件，可输入 **rb** 进行快速生成

```go
package controller

import (
    "github.com/gin-gonic/gin"
    "github.com/archine/gin-plus/v2/v2/mvc"
    "github.com/archine/gin-plus/v2/resp"
)

type TestController struct {
    mvc.Controller
}

func init() {
    mvc.Register(&TestController{})
}

// Hello
// @GET(path="/hello", globalFunc=true) 第一个接口
func (t *TestController) Hello(ctx *gin.Context) {
    i := 0
    // 第二个参数为一个 bool 值，满足才会进行错误返回
    if resp.BadRequest(ctx, i == 0,"操作失败") {
        // 💡 满足条件，这里就可以直接 return 了，因为已经响应给客户端
        // 方法即可结束
        return
    }
    resp.Ok(ctx)
}
```

- 响应结构
```json
{
  "err_code":-10400,
  "err_msg":"操作失败"
}
```
### 3、参数校验
对结构体参数进行绑定校验。当我们有多个条件时，我们可以为每个条件单独定义错误信息，格式为条件+Msg，例如：minMsg ，如果未找到，则取 msg，如果也未找到，会使用参数校验默认的 英文信息。项目中通过
`resp.ParamValid()`调用

```go
package controller

import (
    "github.com/gin-gonic/gin"
    "github.com/archine/gin-plus/v2/mvc"
    "github.com/archine/gin-plus/v2/resp"
)

type TestController struct {
    mvc.Controller
}

func init() {
    mvc.Register(&TestController{})
}

type User struct {
    Age  int    `json:"age" binding:"min=10" minMsg:"年龄最小为10"`
    Name string `json:"name" binding:"required" msg:"名字不能为空"`
}

// AddUser
// @POST(path="/add_user", globalFunc=true) 添加用户
func (t *TestController) AddUser(ctx *gin.Context) {
    var arg User
    if resp.ParamValid(ctx, ctx.ShouldBindJSON(&arg), &arg) {
        return
    }
    resp.Ok(ctx)
}
```

- 响应结构
```json
{
    "err_code": -10602,
    "err_msg": "年龄最小为10"
}
```
### 4、携带数据返回
返回 code、msg、data，用于响应数据给客户端。项目中通过`resp.Json()`调用，数据可为任意类型，💡 如果安装了 IoCer 插件，可输入 **rj** 进行快速生成

```go
package controller

import (
    "github.com/gin-gonic/gin"
    "github.com/archine/gin-plus/v2/mvc"
    "github.com/archine/gin-plus/v2/resp"
)

type TestController struct {
    mvc.Controller
}

func init() {
    mvc.Register(&TestController{})
}

// Hello
// @GET(path="/hello", globalFunc=true) 返回数据
func (t *TestController) Hello(ctx *gin.Context) {
    resp.Json(ctx, "数据")
}
```

- 响应结构
```json
{
    "err_code": 0,
    "err_msg": "OK",
    "ret": "数据"
}
```

---

😊💡 其他的返回方法，使用方式类似，这里就不每个介绍了，使用时，可通过查看方法参数的方式来进行使用。 使用中有如何疑问和优化的建议，欢迎联系 😊😊 😊😊

## 拓展

### gin参数校验

看了下网上，这篇文章介绍的比较详细，可以参考: [🔖点击前往](https://blog.csdn.net/IT_DREAM_ER/article/details/106649622)
