package mtime

import (
	"strconv"
	"time"

	"github.com/m-startgo/go-utils/mmath"
	"github.com/m-startgo/go-utils/mstr"
)

// Now 返回当前时间的封装
func Now() MTime {
	return MTime{t: time.Now()}
}

// NowUnixMilli 返回当前时间的毫秒时间戳（13 位）
func NowUnixMilli() int64 { return time.Now().UnixNano() / 1e6 }

// NowDefaultString 直接返回当前时间的默认格式字符串 "YYYY-MM-DDTHH:mm:ss"
// 例如: 2020-01-02T15:04:05
func NowDefaultString() string {
	return Now().FormatDefault()
}

// FromStdPtr 将标准库的 *time.Time 转换为本包的 MTime。
// - 如果传入为 nil，返回零值 MTime（即封装的 time.Time 为零值）。
// 示例：
//
//	mt := mtime.FromStdPtr(&time.Now())
//	if mt.IsZero() { ... }
func FromStdPtr(tt *time.Time) MTime {
	if tt == nil {
		return MTime{}
	}
	return MTime{t: *tt}
}

// FormatDefault 返回默认的无参数格式化，格式为 "YYYY-MM-DDTHH:mm:ss"
// 例如: 2020-01-02T15:04:05
func (t MTime) FormatDefault() string {
	return t.Format("YYYY-MM-DDTHH:mm:ss")
}

// UnixMilli 返回以毫秒为单位的时间戳（13 位）
func (t MTime) UnixMilli() int64 {
	return t.t.UnixNano() / 1e6
}

// ParseToTimeWithMillisOffset 解析任意支持的输入为 time.Time，并在结果上加上以毫秒为单位的偏移量。
// 参数:
// - v: 支持 Parse 接受的任意类型（string/数字/浮点等）。
// - offsetMillis: 以毫秒为单位的偏移量，可以为负、正或 0。
// 返回值:
// - 解析并加上偏移后的 time.Time。
// 行为说明:
// - 解析失败时为保持向后兼容，返回 Parse(0) 的结果（即 epoch 对应的 time.Time）。
func ParseToTimeWithMillisOffset(v any, offsetMillis int64) time.Time {
	tm, err := Parse(v)
	if err != nil {
		n, _ := Parse(0)
		return n.ToTime()
	}
	ms := tm.UnixMilli() + offsetMillis
	return time.UnixMilli(ms)
}

// FormatDefaultFrom 将任意支持的输入解析并格式化为默认时间字符串 "YYYY-MM-DDTHH:mm:ss"。
// 解析失败时返回 epoch 的默认格式化字符串以保持向后兼容。
func FormatDefaultFrom(v any) string {
	tm, err := Parse(v)
	if err != nil {
		n, _ := Parse(0)
		return n.FormatDefault()
	}
	return tm.FormatDefault()
}

type GetTimeReturnType struct {
	TimeUnix int64  `bson:"TimeUnix"`
	TimeStr  string `bson:"TimeStr"`
}

func GetTime() (resData GetTimeReturnType) {
	resData.TimeUnix = GetUnixInt64()
	resData.TimeStr = UnixFormat(resData.TimeUnix)
	return
}

func GetUnixInt64() int64 {
	return time.Now().UnixNano() / 1e6
}

// ms=string | int64 毫秒数  return = 2006-01-02T15:04:05
func UnixFormat(ms any) string {
	timeMs := mstr.ToStr(ms)
	if len(timeMs) < 1 {
		timeMs = GetUnix()
	}
	T := MsToTime(timeMs, "0")
	timeStr := T.Format(Lay_ss)
	return timeStr
}

// 获取 13 位毫秒时间戳
func GetUnix() string {
	unix := time.Now().UnixNano() / 1e6
	str := strconv.FormatInt(unix, 10)
	return str
}

func MsToTime(ms any, diff string) time.Time {
	msToStr := mstr.ToStr(ms)

	a, _ := mmath.NewFromString(msToStr)
	b, _ := mmath.NewFromString(diff)
	diffDecimal := a.Add(b)

	msStr := diffDecimal.String()

	msInt, err := strconv.ParseInt(msStr, 10, 64)
	if err != nil {
		return time.Now()
	}
	tm := time.Unix(0, msInt*int64(time.Millisecond))
	return tm
}
