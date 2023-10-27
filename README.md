![](https://img.shields.io/badge/version-v3.0.0-green.svg) &nbsp; ![](https://img.shields.io/badge/version-go1.21-green.svg) &nbsp;  ![](https://img.shields.io/badge/builder-success-green.svg) &nbsp;

> 📢📢📢 Gin增强版，集成了IOC、MVC，API定义采用 restful 风格。可帮你快速的进行 web 项目开发，搭配 [🍳Goland](https://plugins.jetbrains.com/plugin/20652-iocer/versions) 插件可以事半功倍哦！！！😀😀

## 一、前言
在线文档点击前往：[文档](https://eofhs2ef6g.feishu.cn/docx/AXCvdf5jPogZ12xOXHucmgo5nFb)
### 1、安装

- Get
```bash
go get github.com/archine/gin-plus/v3@v3.0.0
```

- Mod
```bash
# go.mod文件加入下面的一条
github.com/archine/gin-plus/v3 v3.0.0

# 命令行在该项目目录下执行
go mod tidy
```
## 二、使用说明

### 1、快速开始

- controller接口
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

- 启动类
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

### 5、配置读取

框架默认会读取项目同级目录的 app.yml 文件（可通过 -c 参数指定文件）
* 基础配置
```yaml
log_level: debug # 默认 debug，支持 error、info、trace、warn、panic、fetal、debug
port: 4006 # 默认 4006
max_file_size: 104857600 # 默认 100m，单位字节
```
这些参数框架内部会解析，使用这些参数时，可通过 ``application.Env`` 来获取。

- 自定义配置    

实际开发中，项目配置往往不只是基础配置那些，可能还包括其他配置，这时我们需要在启动时调用 ``ReadConfig()``方法，参数为需要解析到哪个结构体中
```go
package main

import (
  _ "gin-plus-demo/controller"
  "github.com/archine/gin-plus/v3/application"
)

var Conf = &config{}

type config struct {
  // 读取配置文件中的 name 配置，安装了 iocer 插件的话输入 maps 可以快速补全后面的tag
  Name string `mapstructure:"name"`
}

//go:generate gp-ast
func main() {
  application.Default().ReadConfig(Conf).Run()
}
```

### 6、参数校验
对结构体参数进行绑定校验。当我们有多个条件时，我们可以为每个条件单独定义错误信息，格式为条件+Msg，例如：minMsg ，如果未找到，则取 msg，如果也未找到，会使用参数校验默认的 英文信息。项目中通过
``resp.ParamValidation()``调用，💡 如果安装了 IoCer 插件，可输入 **rp** 进行代码快速补全。更多参数校验的关键字， [请参考](https://pkg.go.dev/github.com/go-playground/validator)

```go
package controller

import (
    "github.com/gin-gonic/gin"
    "github.com/archine/gin-plus/v3/mvc"
    "github.com/archine/gin-plus/v3/resp"
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
    if !resp.ParamValidation(ctx, &arg) {
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

**框架使用Demo地址**：[点击前往](https://github.com/archine/gin-plus-demo)
