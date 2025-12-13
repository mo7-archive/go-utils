package mtime

// m_time 提供几个常用的时间工具：解析（支持常见字符串与时间戳）、格式化
// 及一些便捷方法（开始/结束时间、加减天数/小时等）。默认尽量基于标准库实现。

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"sync/atomic"
	"time"

	"github.com/araddon/dateparse"
)

// 默认的 token 风格格式
const DefaultToken = "YYYY-MM-DD HH:mm:ss.SSSSSS ±HH:MM"

// MTime 是对 time.Time 的轻量封装，便于链式调用（原名 Time，已重命名以避免与标准库名冲突）。
type MTime struct {
	t time.Time
}

// DefaultLocation 控制数字时间戳解析后使用的时区（默认 UTC）。
// 如果想使用本地时区，调用 SetDefaultLocation(time.Local)。
var defaultLoc atomic.Value

func init() {
	defaultLoc.Store(time.Local)
}

// DefaultLocation 返回当前默认时区（用于数值时间戳解析），默认 UTC。
func DefaultLocation() *time.Location {
	v := defaultLoc.Load()
	loc, ok := v.(*time.Location)
	if ok {
		return loc
	}
	return time.UTC
}

// SetDefaultLocation 设置数字时间戳解析后的默认时区，传入 nil 恢复为 UTC。
func SetDefaultLocation(loc *time.Location) {
	if loc == nil {
		defaultLoc.Store(time.UTC)
		return
	}
	defaultLoc.Store(loc)
}

// Parse 接受多种输入并尝试解析为时间：
// 支持：
// - string（任意可被 dateparse 识别的格式）
// - 数字字符串（表示秒/毫秒/微秒/纳秒 时间戳）
// - 整数类型/无符号整数/浮点数（表示秒或带小数的秒）
// 示例：
//
//	Parse("2020-01-02 15:04:05")
//	Parse(1609459200000)           // 13 位毫秒时间戳
//	Parse("1609459200000")       // 数字字符串（毫秒）
//	Parse(1609459200.123)         // 带小数的秒
//
// ParseString 解析字符串为 Time：
// - 如果字符串为纯数字（可带小数点、可有正负号），优先按时间戳解析（自动识别秒/毫秒/微秒/纳秒/带小数秒）
// - 否则使用第三方解析器尝试解析任意常见时间格式
// 返回解析后的 Time 或解析错误。
func ParseString(s string) (MTime, error) {
	s = strings.TrimSpace(s)
	if s == "" {
		return MTime{}, fmt.Errorf("empty time string")
	}
	if isNumericString(s) {
		if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			return MTime{t: unixFromInt64(i)}, nil
		}
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			return MTime{t: timeFromFloatSeconds(f)}, nil
		}
	}

	loc := DefaultLocation()
	tt, err := dateparse.ParseIn(s, loc)
	if err != nil {
		return MTime{}, err
	}
	return MTime{t: tt}, nil
}

// ParseInt64 按整数时间戳解析并返回 Time。函数会根据数字位数自动判断单位（秒/毫秒/微秒/纳秒）。
// 示例：ParseInt64(1609459200000) // 13 位毫秒
func ParseInt64(n int64) (MTime, error) { return MTime{t: unixFromInt64(n)}, nil }

// ParseFloat64 按带小数的秒数解析为 Time（小数部分表示秒的小数部分）。
// 示例：ParseFloat64(1609459200.123) 表示秒 + 小数秒
func ParseFloat64(f float64) (MTime, error) { return MTime{t: timeFromFloatSeconds(f)}, nil }

// Parse 尝试从任意支持的类型解析为 Time：
// 支持 string、整型、无符号整型、浮点型等。对于不能直接识别的类型，会使用 fmt.Sprintf 作为后备并尝试按数字或字符串解析。
// 返回解析结果或错误。
func Parse(v any) (MTime, error) {
	if v == nil {
		return MTime{}, fmt.Errorf("nil value")
	}
	switch x := v.(type) {
	case string:
		return ParseString(x)
	case int:
		return ParseInt64(int64(x))
	case int8:
		return ParseInt64(int64(x))
	case int16:
		return ParseInt64(int64(x))
	case int32:
		return ParseInt64(int64(x))
	case int64:
		return ParseInt64(x)
	case uint:
		return ParseInt64(int64(x))
	case uint8:
		return ParseInt64(int64(x))
	case uint16:
		return ParseInt64(int64(x))
	case uint32:
		return ParseInt64(int64(x))
	case uint64:
		if x <= math.MaxInt64 {
			return ParseInt64(int64(x))
		}
		return MTime{}, fmt.Errorf("uint64 value too large")
	case float32:
		return ParseFloat64(float64(x))
	case float64:
		return ParseFloat64(x)
	default:
		s := fmt.Sprintf("%v", v)
		if isNumericString(s) {
			if i, err := strconv.ParseInt(s, 10, 64); err == nil {
				return ParseInt64(i)
			}
			if f, err := strconv.ParseFloat(s, 64); err == nil {
				return ParseFloat64(f)
			}
		}
		return MTime{}, fmt.Errorf("unsupported parse type: %T", v)
	}
}

// MustParse 是 Parse 的便捷包装：解析失败时返回零值 Time（不再 panic）。
// 建议在需要明确错误处理的场景使用 Parse，并避免依赖 MustParse 隐式吞错。
func MustParse(v any) MTime {
	t, _ := Parse(v)
	return t
}

// ParseToStd 解析任意支持的输入并直接返回标准库的 *time.Time 指针。
// 该函数是对 Parse 的便捷包装：解析成功时返回底层 time.Time 的指针，解析失败时返回 nil（不返回 error）。
// 注意：返回的 *time.Time 指针指向函数内部包装的 time 值的拷贝，调用方可安全使用。
func ParseToStd(v any) *time.Time {
	mt, err := Parse(v)
	if err != nil {
		return nil
	}
	// 直接返回底层 time.Time 的地址
	return &mt.t
}

// isNumericString 判断字符串是否为“数字样式”的字符串：允许前导 + 或 -，以及单个小数点。
// 返回 true 表示字符串可以被视为数字（整数或小数）。
func isNumericString(s string) bool {
	if s == "" {
		return false
	}
	if s[0] == '+' || s[0] == '-' {
		s = s[1:]
	}
	dot := false
	digits := 0
	for _, c := range s {
		if c == '.' {
			if dot {
				return false
			}
			dot = true
			continue
		}
		if c < '0' || c > '9' {
			return false
		}
		digits++
	}
	return digits > 0
}

// unixFromInt64 根据整数时间戳的位数推断时间单位，并返回对应的 time.Time：
// - >=18 位：视为纳秒
// - >=16 位：视为微秒
// - >=13 位：视为毫秒
// - 否则：视为秒
// 解析后会根据 DefaultLocation 返回 UTC 或指定时区的时间值。
func unixFromInt64(n int64) time.Time {
	// 通过数字位数判断： >=18 纳秒, >=16 微秒, >=13 毫秒, else 秒
	abs := n
	if abs < 0 {
		abs = -abs
	}
	d := len(strconv.FormatInt(abs, 10))
	var tt time.Time
	switch {
	case d >= 18:
		// 纳秒
		tt = time.Unix(0, n)
	case d >= 16:
		// 微秒 -> 转为纳秒
		tt = time.Unix(0, n*1000)
	case d >= 13:
		// 毫秒
		sec := n / 1000
		msec := n % 1000
		tt = time.Unix(sec, msec*1e6)
	default:
		// 秒
		tt = time.Unix(n, 0)
	}
	loc := DefaultLocation()
	if loc == time.UTC {
		return tt.UTC()
	}
	return tt.In(loc)
}

// timeFromFloatSeconds 将带小数的秒数（秒 + 小数）转换为 time.Time，并应用 DefaultLocation。
// 示例：1609459200.123 表示 1609459200 秒 + 0.123 秒。
func timeFromFloatSeconds(f float64) time.Time {
	sec := int64(math.Floor(f))
	frac := f - float64(sec)
	nsec := int64(math.Round(frac * 1e9))
	tt := time.Unix(sec, nsec)
	loc := DefaultLocation()
	if loc == time.UTC {
		return tt.UTC()
	}
	return tt.In(loc)
}

// FromTime 包装一个标准 time.Time 为本包的 MTime 类型，方便链式调用。
func FromTime(tt time.Time) MTime { return MTime{t: tt} }

// ToTime 返回底层的标准 time.Time。
func (t MTime) ToTime() time.Time { return t.t }

// Format 使用自定义 token 或默认 token 将 Time 格式化为字符串：
// - 传入空字符串或 DefaultToken 使用默认格式
// - 传入类似 "YYYY-MM-DD HH:mm:ss" 的 token，将被映射为 Go 的 layout 进行格式化
func (t MTime) Format(token string) string {
	if token == "" {
		token = DefaultToken
	}
	layout := tokenToLayout(token)
	return t.t.Format(layout)
}

// String 实现 fmt.Stringer，等价于使用默认 token 的 Format。
func (t MTime) String() string { return t.Format(DefaultToken) }

// Add 返回在当前 Time 上加上指定的 duration 后的新 Time（不修改原值）。
func (t MTime) Add(d time.Duration) MTime { return MTime{t: t.t.Add(d)} }

// AddDays 在当前 Time 上增加指定天数（整天）。
func (t MTime) AddDays(days int) MTime {
	return MTime{t: t.t.Add(time.Duration(days) * 24 * time.Hour)}
}

// AddHours 在当前 Time 上增加指定小时数（整小时）。
func (t MTime) AddHours(h int) MTime { return MTime{t: t.t.Add(time.Duration(h) * time.Hour)} }

// StartOfDay 返回当前时间对应日期的 00:00:00（同一时区）。
func (t MTime) StartOfDay() MTime {
	y, m, d := t.t.Date()
	loc := t.t.Location()
	return MTime{t: time.Date(y, m, d, 0, 0, 0, 0, loc)}
}

// EndOfDay 返回当前时间对应日期的末时刻（23:59:59.999999），精度到微秒，保留原时区。
func (t MTime) EndOfDay() MTime {
	y, m, d := t.t.Date()
	loc := t.t.Location()
	const endOfDayNsec = 999999 * int(time.Microsecond)
	return MTime{t: time.Date(y, m, d, 23, 59, 59, endOfDayNsec, loc)}
}

// IsSameDay 判断两个 Time 是否处于同一日期（按各自时区的年月日判断）。
func (t MTime) IsSameDay(o MTime) bool {
	y1, m1, d1 := t.t.Date()
	y2, m2, d2 := o.t.Date()
	return y1 == y2 && m1 == m2 && d1 == d2
}

// Diff 返回 t - o 的 duration（time.Duration）。
func (t MTime) Diff(o MTime) time.Duration { return t.t.Sub(o.t) }

// Unix 返回秒级时间戳（等价于 time.Time.Unix）。
func (t MTime) Unix() int64 { return t.t.Unix() }

// UnixNano 返回纳秒级时间戳（等价于 time.Time.UnixNano）。
func (t MTime) UnixNano() int64 { return t.t.UnixNano() }

// UTC 返回转换为 UTC 时区的新 Time。
func (t MTime) UTC() MTime { return MTime{t: t.t.UTC()} }

// Local 返回转换为本地时区的新 Time。
func (t MTime) Local() MTime { return MTime{t: t.t.Local()} }

// tokenToLayout 将常见 token 映射为 Go 的时间 layout，支持部分 dayjs 风格 token：
// 支持 token 示例：YYYY, MM, DD, HH, mm, ss, SSS, SSSSSS, ±HH:MM
func tokenToLayout(token string) string {
	// 优先处理 SSSSSS
	r := token
	// 保证 ±HH:MM 映射为 -07:00（Go layout 的时区偏移形式）
	r = strings.ReplaceAll(r, "±HH:MM", "-07:00")
	// 有顺序地替换其他 token
	replacements := []struct{ old, new string }{
		{"YYYY", "2006"},
		{"MM", "01"},
		{"DD", "02"},
		{"HH", "15"},
		{"hh", "15"},
		{"mm", "04"},
		{"ss", "05"},
		{"SSSSSS", "000000"},
		{"SSS", "000"},
	}
	for _, rp := range replacements {
		r = strings.ReplaceAll(r, rp.old, rp.new)
	}
	return r
}

// MillisPer... 常量表示毫秒级别的时间长度，命名更语义化且便于直接使用。
const (
	MillisPerSecond int64 = 1000
	MillisPerMinute int64 = 60 * MillisPerSecond
	MillisPerHour   int64 = 60 * MillisPerMinute
	MillisPerDay    int64 = 24 * MillisPerHour
)
