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
## 六、贡献指南

欢迎提交 Issue 和 Pull Request 来改进这个项目。

## 许可证

本项目采用 MIT 许可证，详情请查看 [LICENSE](LICENSE) 文件。