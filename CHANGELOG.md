# CHANGELOG

所有重要更改均记录于此。格式：`## vX.Y.Z - YYYY-MM-DD`，下接带 emoji 的变更要点（含 PR/commit 引用）。

---

## v4.3.7 - 2026-04-12

♻️ **重构**

- ♻️ 将 `Event` 接口重命名为 `AppEvent`，并同步更新所有相关用法，精简 `event_manager.go` 代码量（-51 行）([`1e4dd7b`](https://github.com/archine/gin-plus/commit/1e4dd7be6bb6b02dfc543a534919a51bc8d4d024))

---

## v4.3.6 - 2026-04-05

✨ **新特性**

- ✨ 新增可自定义的应用启动 Banner，支持颜色配置，并优化应用初始化逻辑 ([`3dfa343`](https://github.com/archine/gin-plus/commit/3dfa3437e370d37efbabd5aafbdd8e4f7de3184e))

📝 **文档**

- 📝 完善 README，优化项目简介与特性说明 ([`52cd785`](https://github.com/archine/gin-plus/commit/52cd785e404d030ef13f62d3f07aff501de56a2f))

---

## v4.3.5 - 2026-04-04

✨ **新特性**

- ✨ 新增条件注册支持：依赖注入时可按条件（`Condition`）决定是否注册 Bean，实现可选注入 ([`304ee37`](https://github.com/archine/gin-plus/commit/304ee37e55662a9bd2578785403ffb75fa1d210e)) [PR #6](https://github.com/archine/gin-plus/pull/6)
- ✨ 新增 `@Autowired(required=false)` 可选注入与配置值 `required` 校验支持 ([`a361319`](https://github.com/archine/gin-plus/commit/a3613193cc2aac51dd7baf400f5a9c31829f1044)) [PR #6](https://github.com/archine/gin-plus/pull/6)

♻️ **重构**

- ♻️ 移除未使用的 `createPrototypeBean` 函数，精简 Bean 初始化流程 ([`dd6d409`](https://github.com/archine/gin-plus/commit/dd6d409a2c1a8ead5bb67c06f4c357fa9978f211))

---

## v4.3.4 - 2026-03-11

✨ **新特性 / 重构**

- ✨ 重构异常处理体系：新增 `BusinessException`（业务异常）、`StackError`（带堆栈跟踪的错误）、`WrapError`（错误包装），并拆分为独立文件，提升异常可读性与可维护性 ([`b5d572c`](https://github.com/archine/gin-plus/commit/b5d572c67d29881f980a35370814da6d8ca90179))
- ♻️ 删除旧版 `exception/stacktrace/stack.go` 与 `exception/errs.go`，用新结构替代（净增 ~+30 行有效代码）

---

## v4.3.2 - 2026-03-08

⚡️ **性能优化**

- ⚡️ 重命名 `NewStackErr` → `NewStackError`、`WithStack` → `WrapWithStack`，统一命名风格，新增 `Format()` 方法支持 `%+v` 打印堆栈 ([`c0cb0c6`](https://github.com/archine/gin-plus/commit/c0cb0c6bc31f6825534c08d61a749a82db5d8cae))
- ⚡️ 事件管理器（`event_manager`）：缓存类型断言结果、预分配 slice 容量，触发事件实现 **零分配**（0 B/op, 0 allocs/op）
- ✨ `app`：新增初始化状态追踪，防止重复初始化
- ♻️ 新增 `CodedError` 接口与 `BusinessException` 辅助方法，优化 `Wrap()` 函数逻辑

---

> **提示**：如需为新版本生成 Release Notes，只需推送 `v*` tag（例如 `git tag v4.3.8 && git push origin v4.3.8`），GitHub Actions 会自动触发发布流程。
