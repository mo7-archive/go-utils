# go-utils

轻量且实用的 Go 通用工具集合，提供文件、JSON、HTTP、日志、时间、路径、字符串、加密、验证、循环/定时等常用功能，便于在日常项目中直接复用。

模块名: `github.com/m-startgo/go-utils`

## 快速安装

```bash
go get -u github.com/m-startgo/go-utils@latest
```

## 安装指定版本

```bash
go get -u github.com/m-startgo/go-utils@v0.4.7
```

## 主要包概览（简略）

- `mfile`：文件读写与追加、目录创建等便捷函数（Write/Append/Read）。
- `mjson`：JSON 编码/解码与美化输出（ToByte/ToStr/IndentJson/ToMap）。
- `mhttp`：简单的 HTTP 请求封装，支持超时、重试、参数/头部、最大响应体大小（Fetch/FetchOptions）。
- `mlog`：简单文件日志库，按日期和级别分文件写入（Info/Warn/Error/Debug）。
- `mmath`：基于 decimal 的数值工具（Sum/Mean 等聚合操作）。
- `mpath`：路径与文件存在性检查等工具函数。
- `mtime`：时间格式化与便捷时间函数（Now/Format 等）。
- `mstr`：字符串模板与拼接工具（TplFormat/Join 等）。
- `murl`：URL 处理辅助函数。
- `mverify`：常用数据校验工具。
- `mcycle`、`mcron`：循环与定时任务相关的辅助实现与测试样例。
- `mencrypt`：加密/解密相关。
- `mudp`：udp 通信和接收封装。

## 示例

### 简单的 HTTP 请求：

```go
import "github.com/m-startgo/go-utils/mhttp"

res, err := NewFetch(FetchOptions{
	URL: "https://uapis.cn/api/v1/answerbook/ask",
	DataMap: map[string]any{
		"question": question,
	},
	Method: "POST",
}).Do()

if err != nil {
	fmt.Println("请求失败:", err)
	return
}

fmt.Println(string(res))

```

### 写日志示例：

```go
import "github.com/m-startgo/go-utils/mlog"

myLog := mlog.New(mlog.Config{
	Path: "./logs",
	Name: "log",
})
myLog.Info("this is info")
myLog.Warn("this is warn")
myLog.Error("this is error")
myLog.Debug("this is debug")

Log.Clear(ClearOpt{
	Type:   []string{"debug", "warn"},
	Before: 7,
})

```

更多细节请查看各子包的源码和测试文件以了解用法与边界行为。
