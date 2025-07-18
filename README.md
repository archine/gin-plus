![](https://img.shields.io/badge/version-v4-green.svg) &nbsp; ![](https://img.shields.io/badge/version-go1.23-green.svg) &nbsp;  ![](https://img.shields.io/badge/builder-success-green.svg) &nbsp;

# Gin-Plus

一个基于 Gin 框架的增强库，提供更便捷的 Web 开发体验。

## 特性

- 🚀 路由配置 - 规范路由定义方式
- 🧩 依赖注入 - 简化对象管理
- 🔒 配置注入 - 简化配置读取
- 📊 响应封装 - 规范 API 响应方式

## 一、前言

文档点击前往：[文档](https://eofhs2ef6g.feishu.cn/docx/UrDxd2p7coj2cWxuWShcYBF3nve)

### 1、安装

* Get

```bash
go get github.com/archine/gin-plus/v4@v4.0.6
```

* Mod

```bash
# go.mod文件加入下面的一条
github.com/archine/gin-plus/v4 v4.0.6
```

## 二、快速开始

### 1、基础使用
* API 接口
```go
package controller

import (
	"github.com/archine/gin-plus/v4/component/ioc"
	"github.com/archine/gin-plus/v4/component/mvc"
	"github.com/archine/gin-plus/v4/resp"
	"github.com/gin-gonic/gin"
)

func init() {
	_ = ioc.RegisterBeanDef(&UserController{})
}

type UserController struct {
	mvc.Controller
}

func (u *UserController) SetRoutes(group *gin.RouterGroup) {
	group.GET("/user/list", u.getUserList)
}

// GetUserList retrieves a list of users.
func (u *UserController) getUserList(ctx *gin.Context) {
	resp.Json(ctx, []string{"user1", "user2", "user3"})
}
```
* 启动服务
```go
package main

import (
	_ "gin-plus-demo-v4/controller" // 引入控制器包，确保控制器被注册到 IOC 容器中
	ginplus "github.com/archine/gin-plus/v4"
)

func main() {
	ginplus.Default().Run(ginplus.ServerMode)
}
```

## 贡献指南

欢迎提交 Issue 和 Pull Request 来改进这个项目。

## 许可证

本项目采用 MIT 许可证，详情请查看 [LICENSE](LICENSE) 文件。