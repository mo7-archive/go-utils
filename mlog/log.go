package mlog

/*

以如下 Api 基于 标准库 slog 封装一个日志库。

myLog := mlog.New(Config{
	Path: "./logs",  // 日志存放路径
	Name: "log",  // 日志文件名前缀
	Level: slog.LevelDebug, // 日志级别，参考 slog 的日志级别
	Stdout: true, // 是否输出到控制台，为 true 则输出到控制台和文件，为 false 则只输出到文件
})

myLog.Info("this is info","user","mo7")
myLog.Warn("this is warn")
myLog.Error("this is error")
myLog.Debug("this is debug")

不同的日志类型会创建不同的文件，name-日志类型-日期.log，如下
log-info-2006-01-02.log
log-warn-2006-01-02.log
log-error-2006-01-02.log
log-debug-2006-01-02.log

日志内容为 json 格式，使用 slog 的 api 格式化
日志文件按天切割，每天一个文件
日志输出为文件的时候 也是采用的 slog 的 api

清理日志文件的Api
myLog.Clear(ClearOpt{
	Clear:   []string{"debug", "warn"}, // 需要清理的日志类型，为空 则清理所有类型
	Before: 7,  // 删除几天之前的日志，默认为 7
})
ClearType 是以文件名为基准的

*/

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
)

// Config: 日志配置
type Config struct {
	Path   string     // 日志存放路径
	Name   string     // 日志文件名前缀
	Level  slog.Level // 日志级别
	Stdout bool       // 是否同时输出到控制台
}

// ClearOpt: 清理选项
type ClearOpt struct {
	Clear  []string // 需要清理的日志类型，为空则清理所有类型
	Before int      // 删除几天之前的日志，默认为 7
}

// Logger: 日志对象
type Logger struct {
	cfg         Config
	mu          sync.Mutex
	files       map[string]*os.File     // 每种类型对应的文件
	loggers     map[string]*slog.Logger // 每种类型的 slog logger
	currentDate string
}

// 默认的日志类型顺序及名称
var logTypes = []string{"debug", "info", "warn", "error"}

// New: 创建 Logger
// 如果 Path/Name 未设置将使用默认值
func New(cfg Config) (l *Logger) {
	if cfg.Path == "" {
		cfg.Path = "./logs"
	}
	if cfg.Name == "" {
		cfg.Name = "log"
	}
	// 注意：slog 的 Level 值中 `LevelInfo == 0`，不能把 0 误当作“未设置”。
	// 不在此处强制覆盖为 Debug，保留调用方传入的值。

	_ = os.MkdirAll(cfg.Path, 0o755)

	l = &Logger{
		cfg:     cfg,
		files:   make(map[string]*os.File),
		loggers: make(map[string]*slog.Logger),
	}
	l.currentDate = time.Now().Format("2006-01-02")

	for _, t := range logTypes {
		f, err := l.openLogFile(t, l.currentDate)
		if err != nil {
			// 如果创建文件失败，退回到 stdout 以免丢日志
			if cfg.Stdout {
				handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.Level})
				l.loggers[t] = slog.New(handler)
			}
			continue
		}
		l.files[t] = f
		var w io.Writer = f
		if cfg.Stdout {
			w = io.MultiWriter(f, os.Stdout)
		}
		handler := slog.NewJSONHandler(w, &slog.HandlerOptions{Level: cfg.Level})
		l.loggers[t] = slog.New(handler)
	}

	return
}

// openLogFile: 根据类型与日期打开对应的日志文件（追加模式）
func (l *Logger) openLogFile(level, date string) (*os.File, error) {
	name := fmt.Sprintf("%s-%s-%s.log", l.cfg.Name, level, date)
	path := filepath.Join(l.cfg.Path, name)
	f, err := os.OpenFile(path, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return nil, err
	}
	return f, nil
}

// ensureDate: 检查日期是否变化，若变化则按天切割并重建 logger
func (l *Logger) ensureDate() {
	nowDate := time.Now().Format("2006-01-02")
	if nowDate == l.currentDate {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	if nowDate == l.currentDate {
		return
	}
	// 关闭旧文件
	for _, f := range l.files {
		if f != nil {
			_ = f.Close()
		}
	}
	l.files = make(map[string]*os.File)
	l.loggers = make(map[string]*slog.Logger)
	l.currentDate = nowDate
	for _, t := range logTypes {
		f, err := l.openLogFile(t, l.currentDate)
		if err != nil {
			if l.cfg.Stdout {
				handler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: l.cfg.Level})
				l.loggers[t] = slog.New(handler)
			}
			continue
		}
		l.files[t] = f
		var w io.Writer = f
		if l.cfg.Stdout {
			w = io.MultiWriter(f, os.Stdout)
		}
		handler := slog.NewJSONHandler(w, &slog.HandlerOptions{Level: l.cfg.Level})
		l.loggers[t] = slog.New(handler)
	}
}

// levelPriority: 给定 slog.Level 返回优先级（数值越大越严重）
func levelPriority(lvl slog.Level) int {
	switch lvl {
	case slog.LevelDebug:
		return 0
	case slog.LevelInfo:
		return 1
	case slog.LevelWarn:
		return 2
	case slog.LevelError:
		return 3
	default:
		return 1
	}
}

// log 内部通用方法
func (l *Logger) log(levelName string, lvl slog.Level, msg string, kv ...any) {
	// 级别过滤
	if levelPriority(lvl) < levelPriority(l.cfg.Level) {
		return
	}
	l.ensureDate()
	l.mu.Lock()
	logger, ok := l.loggers[levelName]
	l.mu.Unlock()
	if !ok || logger == nil {
		// 兜底：输出到 stdout
		if l.cfg.Stdout {
			slog.Default().LogAttrs(context.Background(), lvl, msg, slog.String("level", levelName))
		}
		return
	}
	// 不再按 key/value 配对处理 kv 参数。
	// 将 kv 原样作为单个属性传递给 slog，字段名为 "args"。
	var attrs []slog.Attr
	if len(kv) > 0 {
		if len(kv) == 1 {
			attrs = append(attrs, slog.Any("args", kv[0]))
		} else {
			attrs = append(attrs, slog.Any("args", kv))
		}
	}
	logger.LogAttrs(context.Background(), lvl, msg, attrs...)
}

// Info: 记录 info 日志
func (l *Logger) Info(msg string, kv ...any) {
	l.log("info", slog.LevelInfo, msg, kv...)
}

// Warn: 记录 warn 日志
func (l *Logger) Warn(msg string, kv ...any) {
	l.log("warn", slog.LevelWarn, msg, kv...)
}

// Error: 记录 error 日志
func (l *Logger) Error(msg string, kv ...any) {
	l.log("error", slog.LevelError, msg, kv...)
}

// Debug: 记录 debug 日志
func (l *Logger) Debug(msg string, kv ...any) {
	l.log("debug", slog.LevelDebug, msg, kv...)
}

// Clear: 清理日志文件（按文件名中日期）
func (l *Logger) Clear(opt ClearOpt) (err error) {
	before := 7
	if opt.Before > 0 {
		before = opt.Before
	}
	clearAllTypes := len(opt.Clear) == 0
	typesMap := map[string]struct{}{}
	for _, t := range opt.Clear {
		typesMap[t] = struct{}{}
	}

	files, e := os.ReadDir(l.cfg.Path)
	if e != nil {
		err = fmt.Errorf("err:mlog.Clear|ReadDir|%w", e)
		return
	}
	cutoff := time.Now().AddDate(0, 0, -before)
	// 正则用来匹配文件名中的日期部分 YYYY-MM-DD
	dateRe := regexp.MustCompile(`\d{4}-\d{2}-\d{2}`)
	for _, fi := range files {
		if fi.IsDir() {
			continue
		}
		name := fi.Name()
		// 期望格式: name-type-2006-01-02.log
		if !strings.HasPrefix(name, l.cfg.Name+"-") || !strings.HasSuffix(name, ".log") {
			continue
		}
		// 提取日期（更鲁棒，支持文件名中可能存在多余的短横）
		datePart := dateRe.FindString(name)
		if datePart == "" {
			continue
		}
		d, perr := time.Parse("2006-01-02", datePart)
		if perr != nil {
			continue
		}
		// 提取日志类型：去掉前缀和后缀，再移除日期部分前的短横
		tPart := strings.TrimPrefix(name, l.cfg.Name+"-")
		tPart = strings.TrimSuffix(tPart, ".log")
		t := strings.TrimSuffix(tPart, "-"+datePart)
		if !clearAllTypes {
			if _, ok := typesMap[t]; !ok {
				continue
			}
		}
		if d.Before(cutoff) {
			p := filepath.Join(l.cfg.Path, name)
			if rerr := os.Remove(p); rerr != nil {
				// 非致命，继续删除其他文件
				_ = rerr
			}
		}
	}
	return
}

// Close: 关闭所有打开的日志文件
func (l *Logger) Close() (err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	for _, f := range l.files {
		if f != nil {
			_ = f.Close()
		}
	}
	l.files = nil
	l.loggers = nil
	return
}
