![](https://img.shields.io/badge/version-v1.0.3-green.svg) &nbsp; ![](https://img.shields.io/badge/builder-success-green.svg) &nbsp;

> 📢📢📢 Gin框架基础包，集成了IOC 和 MVC，默认提供了一些插件，统一返回结构。IoC的具体使用请参考文档: [点击跳转](https://github.com/archine/ioc)

## 一、前言

### 1、🚀🚀安装

- Get

```shell
go get github.com/archine/gin-plus@v1.0.3
```

- Mod

```shell
# go.mod文件加入下面的一条
github.com/archine/gin-plus v1.0.3
# 命令行在该项目目录下执行
go mod tidy
```

## 二、项目使用

### 1、快速开始

* controller接口

```go
package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/archine/gin-plus/mvc"
	"github.com/archine/gin-plus/resp"
)

type TestController struct {
	mvc.Controller // 需要组合 mvc 中的controller
}

func init() {
	t := &TestController{}

	t.Get("/hello", t.hello, false) // get 方法

	mvc.Register(t) // 注册当前controller到mvc容器中
}

func (t *TestController) hello(ctx *gin.Context) {
	resp.Json(ctx, "Hello world") // resp 是一个快速返回的工具，后面第三章会说明
}
```

* 启动类

```go
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/archine/gin-plus/mvc"
	_ "hj-common-test/gintest/controller" // controller包路径一定要引用，否则api不会缓存
	"log"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	// 将 gin 的引擎加入到 mvc 中，true表示开启依赖注入，后面会讲到
	mvc.Apply(engine, true)

	if err := engine.Run(":8080"); err != nil {
		log.Fatalf(err.Error())
	}
}

```

启动完成后通过访问 ``localhost:8080/hello`` 即可

### 2、方法组

在写controller方法时，如果每次单独调用 Get、Post 等方法，方法多的话，看着比较乱，这时候可以运用方法组，下面以 GetGroup 为例

```go
package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/archine/gin-plus/mvc"
	"github.com/archine/gin-plus/resp"
)

type TestController struct {
	mvc.Controller
}

func init() {
	t := &TestController{}

	t.GetGroup([]*mvc.ApiInfo{
		{"/hello", t.hello, false},
		{"/say", t.say, false},
	})

	mvc.Register(t)
}

func (t *TestController) hello(ctx *gin.Context) {
	resp.Json(ctx, "Hello world")
}

func (t *TestController) say(ctx *gin.Context) {
	resp.Json(ctx, "hi")
}
```

### 3、方法访问路径前缀

```go
package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/archine/gin-plus/mvc"
	"github.com/archine/gin-plus/resp"
)

type TestController struct {
	mvc.Controller // 需要组合 mvc 中的controller
}

func init() {
	t := &TestController{}

	t.Prefix("/test").
		Get("/hello", t.hello, false) // 最终的访问路径为: /test/hello

	mvc.Register(t)
}

func (t *TestController) hello(ctx *gin.Context) {
	resp.Json(ctx, "Hello world")
}
```

### 4、api全局函数

只对当前controller的所有API有效

```go
package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/archine/gin-plus/mvc"
	"github.com/archine/gin-plus/resp"
)

type TestController struct {
	mvc.Controller
}

func init() {
	t := &TestController{}

	t.GlobalFunc(global).
		Get("/hello", t.hello, true) // true表示当前 Api 要使用全局函数

	mvc.Register(t)
}

func (t *TestController) hello(ctx *gin.Context) {
	resp.Json(ctx, "Hello world")
}

// 全局函数
func global(ctx *gin.Context) {
	fmt.Println("啦啦啦")
}
```

### 5、项目全局函数

会对项目中的所有 controller 的API 生效，需要在 Apply 时 设置，优先级最高。如果某个 controller 中也配置了全局函数，那么会追加在项目全局函数后面

```go
package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/archine/gin-plus/mvc"
	"github.com/archine/gin-plus/resp"
)

type TestController struct {
	mvc.Controller
}

func init() {
	t := &TestController{}

	t.GlobalFunc(controllerGlobal).
		Get("/hello", t.hello, true)

	mvc.Register(t)
}

func (t *TestController) hello(ctx *gin.Context) {
	resp.Json(ctx, "Hello world")
}

func controllerGlobal(ctx *gin.Context) {
	fmt.Println("controller函数")
}
```

```go
package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/archine/gin-plus/mvc"
	_ "hj-common-test/gintest/controller"
	"log"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	mvc.Apply(engine, false, global) // Apply 时，传入全局函数

	if err := engine.Run(":8080"); err != nil {
		log.Fatalf(err.Error())
	}
}

func global(ctx *gin.Context) {
	fmt.Println("我是全局函数")
}
```

这时启动运行，访问控制台会打印两条日志

```shell
我是全局函数
controller函数
```
### 6、后置处理器
该处理器会在当前 controller 的所有 API 应用到 gin 之前触发（可以理解为当你调用当前 controller 的构造器后触发），如果搭配了依赖注入，那么会在当前 controller 完成依赖注入后触发。
通过该处理器可以进一步修饰你的 controller 属性
```go
package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/archine/gin-plus/mvc"
	"github.com/archine/gin-plus/resp"
)

type TestController struct {
	mvc.Controller
	age int // 定义了 controller 的全局属性，通过后置处理器进行赋值
}

func (t *TestController) PostConstruct() {
	fmt.Println("后置处理器, 初始化controller的属性 age")
	t.age = 10
}

func init() {
	t := &TestController{}
	t.Get("/hello", t.hello, false)
	mvc.Register(t)
}

func (t *TestController) hello(ctx *gin.Context) {
	resp.Json(ctx, t.age)
}
```
### 7、搭配依赖注入

这里为了方便，依赖的属性直接和 controller 定义在一个文件里，依赖注入不熟悉用法的话前往: [IOC文档](http://gitlab.avatarworks.com/servers/component/hj-ioc)

```go
package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/archine/gin-plus/mvc"
	"github.com/archine/gin-plus/resp"
	ioc "gitlab.avatarworks.com/servers/component/hj-ioc"
)

// 模拟的 Service
type TestService struct {
}

func (t *TestService) CreateBean() ioc.Bean {
	return &TestService{}
}

type TestController struct {
	mvc.Controller
	TestService *TestService // 注入需要的Service
}

func init() {
	t := &TestController{}
	t.Get("/hello", t.hello, false)
	mvc.Register(t)
}

func (t *TestController) hello(ctx *gin.Context) {
	resp.Json(ctx, "Hello world")
}
```

* 启动类

```go
package main

import (
	"github.com/gin-gonic/gin"
	"github.com/archine/gin-plus/mvc"
	_ "hj-common-test/gintest/controller"
	"log"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()

	mvc.Apply(engine, true) // autowired 参数要设置为 true，否则不会自动注入

	if err := engine.Run(":8080"); err != nil {
		log.Fatalf(err.Error())
	}
}
```

## 三、API返回工具

参数中的 ctx 为 gin 的 context

### 1、直接返回

不包含任何数据，只返回 code 和 message

```go
package main

type TestController struct {
	mvc.Controller
}

func init() {
	t := &TestController{}
	t.Get("/hello", t.hello, false)
	mvc.Register(t)
}

func (t *TestController) hello(ctx *gin.Context) {
	resp.Ok(ctx)
}
```

### 2、携带数据

返回数据，包含 code、message、ret

```go
package main

import "github.com/archine/gin-plus/resp"

type TestController struct {
	mvc.Controller
}

func init() {
	t := &TestController{}
	t.Get("/hello", t.hello, false)
	mvc.Register(t)
}

func (t *TestController) hello(ctx *gin.Context) {
	resp.Json(ctx, "这里可以任何数据")
}
```

### 3、参数校验

参数通过 ``binding`` 标签，有哪些标签可以查看第四章的说明。

```go
package controller

import (
	"github.com/gin-gonic/gin"
	"github.com/archine/gin-plus/mvc"
	"github.com/archine/gin-plus/resp"
)

type Arg struct {
	// 定义了参数的两个限制，一个必填，一个最大长度,同时定义了两个msg，如果没有对应 标签名+Msg 的话，默认会取 msg，
	name string `json:"name" binding:"required,max=50" msg:"名称必填" maxMsg:"长度最大为50"`
}

type TestController struct {
	mvc.Controller
}

func init() {
	t := &TestController{}
	t.Get("/hello", t.hello, false)
	mvc.Register(t)
}

func (t *TestController) hello(ctx *gin.Context) {
	var arg Arg
	if resp.ParamValid(ctx, ctx.ShouldBindJSON(&arg), arg) {
		// 如果为true，表示校验出现错误，直接return结束该 API 即可，前端会收到响应
		return
	}
	resp.Ok(ctx)
}
```
```json
{
  code: -10602,
  message: "名称必填"
}
```
### 4、错误的请求
```go
func (t *TestController) hello(ctx *gin.Context) {
	// 第二个参数是个 bool 值，满足条件会返回给前端错误信息
	if resp.BadRequest(ctx,true, "错误") {
		// 满足条件就直接 return 结束该 API
		return
	}
}
```

> 其他的错误请求，都和 3、4 案例一样，满足条件会返回给你 true，直接return 即可 结束该API，前端会收到响应，这里就不一一举例了
## 四、插件
### 1、全局异常捕获
对一次完整的 gin 链路处理过程中的全部 painc 进行捕获，用法参考

```go
package main

import (
	"github.com/archine/gin-plus/plugin"
	"github.com/gin-gonic/gin"
	_ "hj-common-test/gintest/controller"
	"log"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	engine.Use(plugin.GlobalExceptionInterceptor) // 加入插件

	if err := engine.Run(":8080"); err != nil {
		log.Fatalf(err.Error())
	}
}

```
### 2、日志
美化控制台打印日志
```go
package main

import (
	"github.com/archine/gin-plus/plugin"
	"github.com/gin-gonic/gin"
	"github.com/archine/gin-plus/mvc"
	_ "hj-common-test/gintest/controller"
	"log"
)

func main() {
	gin.SetMode(gin.ReleaseMode)
	engine := gin.New()
	plugin.InitLog("debug") // 配置日志级别
	engine.Use(plugin.LogMiddleware()) // 加入插件

	if err := engine.Run(":8080"); err != nil {
		log.Fatalf(err.Error())
	}
}
```
## 五、拓展

### gin参数校验

看了下网上，这篇文章介绍的比较详细，可以参考: [🔖点击前往](https://blog.csdn.net/IT_DREAM_ER/article/details/106649622)
