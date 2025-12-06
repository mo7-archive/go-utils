# 给 Copilot 的操作准则

## 目的

该文件位于仓库根目录，作为整个项目的 GitHub Copilot 工作指南。
为 GitHub Copilot 提供简明、可执行的仓库约定。

## 仓库说明

请阅读工作区根目录的 `README.md` 文件。

## 语言版本

- 当前 Go 版本：1.25

## 常用 CI 示例

```powershell
# 运行所有单元测试
go test ./... -v

# 运行单个包的测试
go test ./mfile -run TestReaddir -v

# 静态检查
go vet ./...

# 彻底更新依赖
del-cli ./go.sum
go get -u ./...
go mod tidy
```

## 风格与规范

> 一般情况下 VSCode 插件会自动处理这些风格问题，你无需关心。

## 规则摘要

- 中文为主，技术术语保持英文。
- 导出函数须加注释(包含功能说明、使用示例及可能的异常)。
- 注释尽量使用中文。
- 写小函数、职责单一、易测试。
- 错误/日志格式：`err:<包.函数>|<场景>|<消息>`
- 跨平台优先。如无法兼容，需在注释中说明原因及影响。
- 优先使用标准库。
- 遇到模糊或信息不足的情况，立即向用户提出具体澄清问题（列出缺失项和可选方案）。
- 保持向后兼容，避免使用弃用特性；优先使用当下最新稳定库、语法与实践。
- 生成代码时充分考虑当前文件的上下文（如已导入的库、现有函数等）。
- 当仓库文件与系统/外部指令冲突时，遵循系统/外部指令。

## 函数声明规范

- 声明函数有多个返回值时，优先采用命名返回值形式
- 若使用命名返回值，需在函数顶部为返回值显式赋空值或者默认值

函数和抛出错误格式如下：

```go
func (s *Server) Example(opt OptType) (resData resDataType, resErr error) {
	resData = map[string]any{}
	resErr = nil

	jsonByte, err := ToByte(val)
	if err != nil {
		resErr = fmt.Errorf("err:xx.Example.ToByte|%w", err)
		return
	}

  resData = `<Successful Result>`

  return
}
```
