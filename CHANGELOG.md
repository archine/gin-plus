# 📋 变更日志 / CHANGELOG

本文档记录了 **archine/gin-plus** 最近 5 个有重大变动版本的发布说明。

所有版本均遵循 [语义化版本控制](https://semver.org/lang/zh-CN/) 规范。

---

## [v4.3.7] — 2026-04-12

### ♻️ 代码重构

- ♻️ **引入 `Comparable` 事件接口与泛型并发安全 `Registry`**（PR [#7](https://github.com/archine/gin-plus/pull/7)，commit [`518b9ed`](https://github.com/archine/gin-plus/commit/518b9ed5aa4d08f278047d76aeefb4989783fc82)）
  - 以更清晰的排序契约（`Comparable`）替换旧的无类型 `Event` 集合，解决并发竞争问题
  - 使用泛型 `Registry[T Comparable]` 取代单体 `eventManager`，为每种事件类型提供独立注册表
  - 引入懒惰排序（`slices.SortFunc` + `cmp.Compare`），避免在触发时重复排序
  - 保留向后兼容：`type Event = Comparable` 作为废弃别名继续可用
  - 简化触发函数，移除旧的类型缓存与 `ensureSorted` 逻辑

### ✅ 测试

- ✅ 运行 `go test ./...`、`go build ./...` 与 `go vet` 全部通过（PR #7）

---

## [v4.3.6] — 2026-04-05

### ✨ 新特性

- ✨ **可选自动注入（Optional Autowire）**（PR [#6](https://github.com/archine/gin-plus/pull/6)，commit [`74f0ba2`](https://github.com/archine/gin-plus/commit/74f0ba29e0c8281debe1283f013d4ac3176718ab)）
  - `autowire` 标签支持 `,optional` 后缀，Bean 缺失时字段置 `nil` 而非启动崩溃，适用于插件/可选模块场景
  - 按类型注入：`` `autowire:",optional"` ``；按名称注入：`` `autowire:"redisCache,optional"` ``
  - 未知选项（除 `optional` 以外）在启动时立即 panic 并给出清晰提示
- ✨ **必填配置值哨兵（Required Config Sentinel）**（PR #6）
  - 使用 `:?` 作为哨兵默认值，配置键缺失时立即报错，不再静默产生零值
  - 示例：`` `value:"${db.url:?}"` `` → 缺失时启动失败并告知原因

### ♻️ 代码重构

- ♻️ `internal/container/container.go`：为 `AutowireField` 添加 `Optional bool` 字段
- ♻️ `internal/container/process.go`：解析 autowire tag 的 `,optional` 后缀并校验未知选项
- ♻️ `internal/container/injector/config.go`：将 `?` 识别为必填值哨兵，缺失时返回错误

---

## [v4.3.4] — 2026-03-11

### ✨ 新特性

- ✨ **重构异常处理，新增业务异常（`BusinessException`）与堆栈跟踪**（commit [`b5d572c`](https://github.com/archine/gin-plus/commit/b5d572c67d29881f980a35370814da6d8ca90179)）
  - 新增 `CodedError` 接口，统一业务错误码与消息的访问方式
  - 新增 `BusinessException` 辅助方法，简化业务层抛出带错误码的异常
  - 为 `StackError` 添加 `Format()` 方法，支持 `%+v` 格式化输出完整堆栈
  - 优化 `Wrap()` 函数逻辑并完善文档注释

### ♻️ 代码重构

- ♻️ 重命名 `NewStackErr` → `NewStackError` 以提升命名一致性
- ♻️ 重命名 `WithStack` → `WrapWithStack` 保持与惯例一致

---

## [v4.3.2] — 2026-03-07

### ⚡️ 性能优化

- ⚡️ **事件触发零分配**（commit [`c0cb0c6`](https://github.com/archine/gin-plus/commit/c0cb0c6bc31f6825534c08d61a749a82db5d8cae)）
  - 缓存已分类事件，避免重复类型断言；预分配切片容量优化内存使用
  - 基准：**53.64 ns/op，0 B/op，0 allocs/op**
- ⚡️ **应用状态检查近零开销**（commit [`c0cb0c6`](https://github.com/archine/gin-plus/commit/c0cb0c6bc31f6825534c08d61a749a82db5d8cae)）：添加 `atomic.Bool` 实现幂等初始化，基准 **0.22 ns/op，0 allocs/op**
- ⚡️ **对象池操作优化**（commit [`c0cb0c6`](https://github.com/archine/gin-plus/commit/c0cb0c6bc31f6825534c08d61a749a82db5d8cae)）：新增 `GetWithReset()` / `PutWithReset()` 方法，基准 **4.43 ns/op，0 allocs/op**
- ⚡️ **Recovery 中间件优化**（commit [`c0cb0c6`](https://github.com/archine/gin-plus/commit/c0cb0c6bc31f6825534c08d61a749a82db5d8cae)）：正常请求处理基准 **267.7 ns/op**
- ⚡️ **堆栈跟踪内存优化**（commit [`c0cb0c6`](https://github.com/archine/gin-plus/commit/c0cb0c6bc31f6825534c08d61a749a82db5d8cae)）：预分配 512 字节 builder，用 `WriteByte()` 替代单字符 `WriteString()`

### ✨ 新特性

- ✨ 新增 `RecoveryWithMessage` 中间件（commit [`c0cb0c6`](https://github.com/archine/gin-plus/commit/c0cb0c6bc31f6825534c08d61a749a82db5d8cae)），支持自定义错误响应消息
- ✨ 对象池新增 `GetWithReset()` 与 `PutWithReset()`（commit [`c0cb0c6`](https://github.com/archine/gin-plus/commit/c0cb0c6bc31f6825534c08d61a749a82db5d8cae)）

### ♻️ 代码重构

- ♻️ 重命名 `Cors` → `CORS`（commit [`c0cb0c6`](https://github.com/archine/gin-plus/commit/c0cb0c6bc31f6825534c08d61a749a82db5d8cae)），遵循 Go 命名规范
- ♻️ 重命名 `GlobalExceptionInterceptor` → `Recovery`（commit [`c0cb0c6`](https://github.com/archine/gin-plus/commit/c0cb0c6bc31f6825534c08d61a749a82db5d8cae)），语义更清晰
- ♻️ 堆栈跟踪帧数从 8 提升至 16（commit [`c0cb0c6`](https://github.com/archine/gin-plus/commit/c0cb0c6bc31f6825534c08d61a749a82db5d8cae)）

### 📝 文档

- 📝 为中间件、对象池等核心组件添加完整 GoDoc 注释（commit [`c0cb0c6`](https://github.com/archine/gin-plus/commit/c0cb0c6bc31f6825534c08d61a749a82db5d8cae)）

---

## [v4.3.1] — 2026-03-06

### ✨ 新特性

- ✨ **重构日志字段为结构化格式**（commit [`e75f9a3`](https://github.com/archine/gin-plus/commit/e75f9a3d7ea643cfad7a7c033b5b92b1639ac89b)）
  - 日志字段改为结构化键值对，兼容 zap 的 `Field` 类型，便于日志聚合与检索
- ✨ **增强 SSE Writer 初始化**（commit [`e75f9a3`](https://github.com/archine/gin-plus/commit/e75f9a3d7ea643cfad7a7c033b5b92b1639ac89b)）：改进 Server-Sent Events 写入器的初始化流程，提升稳定性与可用性

---

[v4.3.7]: https://github.com/archine/gin-plus/compare/v4.3.6...v4.3.7
[v4.3.6]: https://github.com/archine/gin-plus/compare/v4.3.4...v4.3.6
[v4.3.4]: https://github.com/archine/gin-plus/compare/v4.3.2...v4.3.4
[v4.3.2]: https://github.com/archine/gin-plus/compare/v4.3.1...v4.3.2
[v4.3.1]: https://github.com/archine/gin-plus/compare/v4.3.0...v4.3.1
