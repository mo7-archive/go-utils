package main

import (
	"log/slog"

	"github.com/m-startgo/go-utils/mlog"
)

var myLog *mlog.Logger

func SetLog() {
	// 创建 logger（输出到 ./logs，控制台与文件同时输出，级别为 Debug）
	myLog = mlog.New(mlog.Config{
		Path:  "./logs",
		Name:  "log",
		Level: slog.LevelDebug, // 全部的日志
		// Level: slog.LevelError, // 仅 ERROR
		// Level: slog.LevelInfo, // INFO WARN ERROR
		// Level:  slog.LevelWarn, // WARN ERROR
		Stdout: true,
	})

	// 清理示例：删除 7 天之前的 debug 和 warn 日志
	// fmt.Println("clear error:")

	// 等待短暂时间以保证日志写入（示例用）
	// time.Sleep(200 * time.Millisecond)
	// fmt.Println("demo finished")
}

func main() {
	SetLog()

	type User struct {
		Name   string
		Age    int
		Gender string
	}

	user := User{
		Name:   "mo7",
		Age:    18,
		Gender: "male",
	}

	user2 := User{
		Name:   "张三",
		Age:    23,
		Gender: "sex",
	}

	myLog.Info("this is info")
	myLog.Warn("this is warn")
	myLog.Error("this is error")
	myLog.Debug("this is debug1", user)
	myLog.Debug("this is debug2", user, user2)
	myLog.Debug("this is debug3", user, "user=", user2)
	myLog.Debug("this is debug3", "user=", user, "user2=", user2)

	myLog.Clear(mlog.ClearOpt{
		Clear:  []string{},
		Before: 1,
	})

	myLog.Debug("日志文件已删除")
}
