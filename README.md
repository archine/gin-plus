![](https://img.shields.io/badge/version-v2.1.1-green.svg) &nbsp; ![](https://img.shields.io/badge/builder-success-green.svg) &nbsp;

> 📢📢📢 Gin增强版，集成了IOC、MVC，API定义采用 restful 风格。可帮你快速的进行 web 项目开发，搭配 [🍳Goland](https://plugins.jetbrains.com/plugin/20652-iocer/versions) 插件可以事半功倍哦！！！😀😀

## 一、前言
如果访问 github 网络不是很好，建议前往在线文档：[在线文档](https://eofhs2ef6g.feishu.cn/docx/AXCvdf5jPogZ12xOXHucmgo5nFb)
### 1、安装

- Get
```bash
go get github.com/archine/gin-plus/v2@v2.1.1
```

- Mod
```bash
# go.mod文件加入下面的一条
github.com/archine/gin-plus/v2 v2.1.1

# 命令行在该项目目录下执行
go mod tidy
```
- 安装 ast 解析工具
```shell
# 可将 latest 指定为具体版本
go install github.com/archine/gin-plus/v2/ast/mvc@latest
```
>  ❗ v2.1.0 版本开始需要安装此工具，确保 gopath 的 bin 目录有加入到系统环境变量中     

使用时可以直接在命令行执行
```
# 参数非必填，默认解析当前命令执行所在目录中的 controller 目录下的所有 go 文件
mvc <scan dir>
```
也可通过在启动类上加上注释，这时候就可以通过 go generate来执行
```
//go:generate mvc <scan dir>
func main() {
    application.Default().Run()
}
```    

执行结束后，会在对应的扫描目录生成 controller_init.go 文件，请勿编辑 ❌，如果目录下的 API 定义发生了更改，如更换了 请求路径，请求方式等，一定要重新执行哦

### 2、🎁小技巧

使用 Goland 进行开发时，可以按照下方的教程配置一下，就不需要每次修改了 API，都手动执行 ``go generate ``     

![generate](https://user-images.githubusercontent.com/35919643/221461839-eea974bd-72f1-474c-b72a-3dccd55b797b.gif)
      
      
## 二、项目使用
本框架声明 API 的方式非常简单，只需在方法的注释中通过如下方式进行声明即可，启动时会自动应用，**需要注意的是，API函数名必须大写**

| 定义方式🍑 | 描述🍎 | 快捷键🍓 |
| --- | --- | --- |
| @GET(path="/hello") | Get 请求 | 空白处输入 get |
| @POST(path="/hello") | Post 请求 | 空白处输入 post |
| @DELETE(path="/hello") | Delete 请求 | 空白处输入 del |
| @PUT(path="/hello") | Put 请求 | 空白处输入 put |
| @PATCH(path="/hello") | Patch 请求 | 暂无 |
| @HEAD(path="/hello") | Head 请求 | 暂无 |
| @OPTIONS(path="/hello") | Options 请求 | 暂无 |
| @BasePath("/hello") | 基础路径 | 空白处输入 basep |

> ❗ v2.1.0 版本开始 API 定义移除了 GlobalFunc 参数
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

// Hello
// @GET(path="/hello") 定义的 get 方法
func (t *TestController) Hello(ctx *gin.Context) {
	resp.Ok(ctx)
}
```
> ❗ v2.1.0 版本开始不需要手动在每个 Controller 中通过 init() 方法注册

- 启动类
```go
package main

import (
	_ "gin-plus-demo/controller"
	"github.com/archine/gin-plus/v2/application"
)

//go:generate mvc
func main() {
	application.Default().Run()
}
```

这时候运行该项目，浏览器访问http://localhost:4006/hello即可

![在这里插入图片描述](https://img-blog.csdnimg.cn/27837bfb5714484eac33932392929d7e.png)

### 2、方法路径前缀
很多时候，我们需要对整个 Controller 里的所有 API 增加访问前缀，这时我们可在 Controller 的结构体注释中通过`@BasePath("/xxx")`来进行声明
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

// Hello
// @GET(path="/hello") 第一个接口
func (t *TestController) Hello(ctx *gin.Context) {
	resp.Json(ctx, "hello world")
}
```

重新启动项目后，浏览器访问http://localhost:4006/test/hello即可

![在这里插入图片描述](https://img-blog.csdnimg.cn/5d84177e137f4033a7ec517e72579704.png)


### 3、API 接口拦截器
对项目 API 方法进行拦截，通过拦截器可以对访问进行逻辑化处理。如：登录校验、日志打印等等。。。。

- controller
```go
package controller

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/archine/gin-plus/v2/mvc"
)

// UserController
// @BasePath("/user")
type UserController struct {
	mvc.Controller
}

// UserList
// @GET(path="/list") API描述
func (u *UserController) UserList(ctx *gin.Context) {
	fmt.Println("正在执行API方法")
}
```
- 定义拦截器
  需要实现 MethodInterceptor 接口
```go
package intercptor

type TestInterceptor struct {}

// Predicate 过滤条件，true 表示拦截
func (t *TestInterceptor) Predicate(request *http.Request) bool {
    return true
}

// PreHandle 方法调用前
func (t *TestInterceptor) PreHandle(ctx *gin.Context) {
    // 方法中通过调用 ctx.Abort() 可中断当前客户端请求
    // 😊 中断时记得响应给客户端哦
    fmt.Println("前置处理器")
}

// PostHandle 访问调用后
func (t *TestInterceptor) PostHandle(ctx *gin.Context) {
    // 方法中通过调用 ctx.Abort() 可中断当前客户端请求
    // 😊 中断时记得响应给客户端哦
    fmt.Println("后置处理器")
}
```
- 应用拦截器
  只需要在启动类中添加进去即可，拦截器为可变参数，因此可以添加多个
```go
package main

import (
   _ "gin-plus-demo/controller"
   "github.com/archine/gin-plus/v2/application"
)

//go:generate mvc
func main() {
   application.Default().Run(&TestInterceptor{})
}
```
这时候，我们通过浏览器访问这三个 API，可以看到只有前两个 API 才会打印全局函数中的日志

![image](https://user-images.githubusercontent.com/35919643/221462946-92f04e47-c800-48dc-ac50-e0e261204320.png)

### 4、依赖注入前事件
在执行依赖注入前触发，此时项目运行环境中无任何 bean，意味着你不能在此步骤中处理任何要获取 bean 的逻辑。该事件为同步，因此 阻塞性事件需要通过新的 协程处理，否则会影响整个流程
```go
package main

import (
  _ "gin-plus-demo/controller"
  "github.com/archine/gin-plus/v2/application"
)

//go:generate mvc
func main() {
  application.Default().ApplyBefore(func() {
    fmt.Println("注入前逻辑")
  }).Run()
}
```

### 5、启动前事件
项目运行最后一个事件， 依赖注入已执行完毕，即将启动，意味着你可以在这里执行任意逻辑。该事件为同步，因此 阻塞性事件需要通过新的 协程处理，否则会影响整个流程。 在启动类进行加入
```go
package main

import (
   _ "gin-plus-demo/controller"
   "github.com/archine/gin-plus/v2/application"
)

//go:generate mvc
func main() {
    application.Default().StartBefore(func() {
       fmt.Println("启动前逻辑")
    }).Run()
}
```
### 6、依赖注入

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
    "github.com/archine/gin-plus/v2/mvc"
    "github.com/archine/gin-plus/v2/resp"
)

type TestController struct {
    mvc.Controller
    // 注入TestMapper。为了他人直观知道该属性为依赖注入进来的，可在注入的属性右边加入声明（😊建议）
    // 安装了 Iocer 插件的话，可直接在 属性右边 输入 di，可快速生成
    TestMapper *mapper.TestMapper `@autowired:""`
}

// Hello
// @GET(path="/hello") 第一个接口
func (t *TestController) Hello(ctx *gin.Context) {
    // 使用时直接调用即可
    resp.Json(ctx, t.TestMapper.Say())
}
```
### 7、Controller构造后置处理

该处理器在 Controller 实例化结束且依赖注入完成后触发，我们可在该函数做其他的一些属性处理，这里例子为
赋值 controller 中的一些私有属性，💡 如果安装了 IoCer 插件，可输入 pc 进行快速生成
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

// PostConstruct 初始化私有属性 age 的值
func (t *TestController) PostConstruct() {
  t.age = 100
}
```

### 8、配置读取

框架默认会读取项目同级目录的 app.yml 文件（可通过 -c 参数指定文件）
* 基础配置
```yaml
log_level: debug # 默认 debug，支持 error、info、trace、warn、panic、fetal、debug
port: 4006 # 默认 4006
max_file_size: 104857600 # 默认 100m，单位字节
```
这些参数框架内部会解析，使用这些参数时，可通过 ``application.Env`` 来获取  

- 自定义配置
  实际开发中，项目配置往往不只是基础配置那些，可能还包括其他配置，这时我们需要在启动时调用 ``ReadConfig()``方法，参数为需要解析到哪个结构体中
```go
package main

import (
  _ "gin-plus-demo/controller"
  "github.com/archine/gin-plus/v2/application"
)

var Conf = &config{}

type config struct {
  // 读取配置文件中的 name 配置
  Name string `mapstructure:"name"`
}

//go:generate mvc
func main() {
  application.Default().ReadConfig(Conf).Run()
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

// Hello
// @GET(path="/hello") Hello 第一个接口
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
    "github.com/archine/gin-plus/v2/mvc"
    "github.com/archine/gin-plus/v2/resp"
)

type TestController struct {
    mvc.Controller
}

// Hello
// @GET(path="/hello") 第一个接口
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
`resp.ParamValid()`调用。更多参数校验的关键字， [请参考](https://pkg.go.dev/github.com/go-playground/validator)

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

type User struct {
    Age  int    `json:"age" binding:"min=10" minMsg:"年龄最小为10"`
    Name string `json:"name" binding:"required" msg:"名字不能为空"`
}

// AddUser
// @POST(path="/user") 添加用户
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

// Hello
// @GET(path="/hello") 返回数据
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


**框架使用Demo地址**：[点击前往](https://github.com/archine/gin-plus-demo)
