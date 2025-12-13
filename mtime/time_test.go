package mtime

import (
	"math"
	"testing"
	"time"
)

// go test -v -run TestLoc
func TestLoc(t *testing.T) {
	// shLoc, err := time.LoadLocation("Asia/Shanghai")
	// if err != nil {
	// 	panic(err)
	// }
	// SetDefaultLocation(shLoc)

	// mytime, _ := ParseInt64(1765643957726)

	// s := mytime.Format(DefaultToken)
}

// TestParseAndFormat 覆盖常见解析与格式化场景。
func TestParseAndFormat(t *testing.T) {
	// ISO-like string
	tt, err := Parse("2020-01-02 15:04:05")
	if err != nil {
		t.Fatalf("Parse error: %v", err)
	}
	got := tt.FormatDefault()
	if got != "2020-01-02T15:04:05" {
		t.Fatalf("unexpected default format: %s", got)
	}

	// Token-based format
	s := tt.Format("YYYY/MM/DD HH:mm:ss")
	if s != "2020/01/02 15:04:05" {
		t.Fatalf("token format mismatch: %s", s)
	}

	// FromTime and ToTime roundtrip
	std := time.Date(2021, 1, 1, 12, 0, 0, 0, time.UTC)
	wrapped := FromTime(std)
	if !wrapped.ToTime().Equal(std) {
		t.Fatalf("FromTime/ToTime roundtrip failed: %v vs %v", wrapped.ToTime(), std)
	}
}

// TestUnixTimestamps 测试不同位数时间戳的解析逻辑（秒/毫秒/微秒/纳秒）
func TestUnixTimestamps(t *testing.T) {
	// seconds
	p1, err := ParseInt64(1609459200) // 2021-01-01T00:00:00Z
	if err != nil {
		t.Fatal(err)
	}
	if p1.FormatDefault()[:4] != "2021" {
		t.Fatalf("expected year 2021, got %s", p1.FormatDefault())
	}

	// milliseconds
	p2, err := ParseInt64(1609459200000)
	if err != nil {
		t.Fatal(err)
	}
	if p2.FormatDefault()[:4] != "2021" {
		t.Fatalf("ms parse failed: %s", p2.FormatDefault())
	}

	// microseconds (16 digits-ish)
	micro := int64(1609459200000000)
	p3, err := ParseInt64(micro)
	if err != nil {
		t.Fatal(err)
	}
	if p3.FormatDefault()[:4] != "2021" {
		t.Fatalf("us parse failed: %s", p3.FormatDefault())
	}

	// nanoseconds (19 digits)
	nano := int64(1609459200000000000)
	p4, err := ParseInt64(nano)
	if err != nil {
		t.Fatal(err)
	}
	if p4.FormatDefault()[:4] != "2021" {
		t.Fatalf("ns parse failed: %s", p4.FormatDefault())
	}

	// negative seconds
	neg, err := ParseInt64(-1)
	if err != nil {
		t.Fatal(err)
	}
	if neg.ToTime().Unix() != -1 {
		t.Fatalf("expected unix -1, got %d", neg.ToTime().Unix())
	}
}

// TestFloatAndOffsets 测试带小数秒与毫秒偏移功能
func TestFloatAndOffsets(t *testing.T) {
	pf, err := ParseFloat64(1609459200.5)
	if err != nil {
		t.Fatal(err)
	}
	// half second -> 500ms
	if pf.ToTime().UnixNano()%1e9 < 499000000 || pf.ToTime().UnixNano()%1e9 > 501000000 {
		t.Fatalf("expected ~0.5s fractional, got %d ns", pf.ToTime().UnixNano()%1e9)
	}

	// ParseToTimeWithMillisOffset: add 1000ms -> +1s
	p, _ := Parse("1609459200000")
	got := ParseToTimeWithMillisOffset(1609459200000, 1000)
	if got.Unix() != p.Unix()+1 {
		t.Fatalf("offset addition failed: %v vs %v", got, p)
	}
}

// TestAddAndDayHelpers 测试加天、加小时、开始/结束日与 IsSameDay
func TestAddAndDayHelpers(t *testing.T) {
	base := FromTime(time.Date(2022, 3, 14, 10, 30, 0, 0, time.UTC))
	if base.AddDays(1).ToTime().Day() != 15 {
		t.Fatalf("AddDays failed")
	}
	if base.AddHours(14).ToTime().Hour() != 0 {
		t.Fatalf("AddHours failed")
	}

	sod := base.StartOfDay()
	if !(sod.ToTime().Hour() == 0 && sod.ToTime().Minute() == 0) {
		t.Fatalf("StartOfDay failed: %v", sod)
	}
	eod := base.EndOfDay()
	if eod.ToTime().Hour() != 23 {
		t.Fatalf("EndOfDay hour not 23: %v", eod)
	}

	// IsSameDay across different times
	a := FromTime(time.Date(2022, 3, 14, 0, 0, 0, 0, time.UTC))
	b := FromTime(time.Date(2022, 3, 14, 23, 59, 59, 0, time.UTC))
	if !a.IsSameDay(b) {
		t.Fatalf("IsSameDay detected difference incorrectly")
	}
}

// TestEdgeCases 测试 MustParse 与 Parse 错误情形，以及 Parse 对不同类型的接受度
func TestEdgeCases(t *testing.T) {
	// MustParse returns zero Time on failure
	m := MustParse("not a time")
	if !m.ToTime().IsZero() {
		t.Fatalf("MustParse expected zero time on failure")
	}

	// Parse with uint64 overflow should error
	_, err := Parse(uint64(math.MaxInt64) + 1)
	if err == nil {
		t.Fatalf("expected error for too large uint64")
	}

	// Parse numeric string
	ps, err := Parse("1609459200000")
	if err != nil {
		t.Fatal(err)
	}
	if ps.Unix() != 1609459200 {
		t.Fatalf("numeric string parse incorrect unix: %d", ps.Unix())
	}

	// FormatDefaultFrom and UnixFormat wrappers
	if FormatDefaultFrom(0) != ParseIntOrPanic(0).FormatDefault() {
		t.Fatalf("FormatDefaultFrom mismatch")
	}

	// tokenToLayout fractional seconds tokens
	tt, _ := Parse("2020-01-02 03:04:05.123456")
	s := tt.Format("YYYY-MM-DD HH:mm:ss.SSSSSS")
	if s[len(s)-6:] != "123456" {
		t.Fatalf("fractional second formatting failed: %s", s)
	}
}

// ParseIntOrPanic is a tiny helper for tests to get Time from integer without repeating error handling.
func ParseIntOrPanic(n int) MTime {
	t, err := Parse(n)
	if err != nil {
		panic(err)
	}
	return t
}

// TestParseToStd 验证 ParseToStd 的行为：
// - 有效输入返回非 nil 且时间值与 Parse 等价
// - 无效输入返回 nil
func TestParseToStd(t *testing.T) {
	// 有效输入
	std := ParseIntOrPanic(1609459200) // 2021-01-01T00:00:00Z
	p := ParseToStd(1609459200)
	if p == nil {
		t.Fatalf("expected non-nil for valid input")
	}
	if !p.Equal(std.ToTime()) {
		t.Fatalf("ParseToStd returned different time: %v vs %v", p, std.ToTime())
	}

	// 无效输入
	invalid := ParseToStd("not a time")
	if invalid != nil {
		t.Fatalf("expected nil for invalid input")
	}
}

// TestFromStdPtr 验证 FromStdPtr 的行为：
// - nil 输入返回零值 MTime
// - 非 nil 返回与原始 time.Time 等价的封装
func TestFromStdPtr(t *testing.T) {
	var nilPtr *time.Time
	mt := FromStdPtr(nilPtr)
	if !mt.ToTime().IsZero() {
		t.Fatalf("FromStdPtr(nil) expected zero time, got %v", mt.ToTime())
	}

	std := time.Date(2021, 1, 2, 3, 4, 5, 0, time.UTC)
	p := &std
	mt2 := FromStdPtr(p)
	if !mt2.ToTime().Equal(std) {
		t.Fatalf("FromStdPtr returned different time: %v vs %v", mt2.ToTime(), std)
	}
}

// TestAge 覆盖 Age 的若干边界场景：无效输入、未来日期、整年与半年内
func TestAge(t *testing.T) {
	// 无效输入应返回 0
	if Age("not-a-date") != 0 {
		t.Fatalf("expected 0 for invalid birthday")
	}

	// 未来日期应返回 0
	future := time.Now().UTC().Add(24 * time.Hour).Format("2006-01-02 15:04:05")
	if Age(future) != 0 {
		t.Fatalf("expected 0 for future birthday, got %d", Age(future))
	}

	// 刚好 1 年前应返回 1
	oneYear := time.Now().UTC().AddDate(-1, 0, 0).Format("2006-01-02 15:04:05")
	if Age(oneYear) != 1 {
		t.Fatalf("expected 1 for one year ago, got %d", Age(oneYear))
	}

	// 半年前应返回 0（小于一年）
	halfYear := time.Now().UTC().AddDate(0, -6, 0).Format("2006-01-02 15:04:05")
	if Age(halfYear) != 0 {
		t.Fatalf("expected 0 for half year ago, got %d", Age(halfYear))
	}
}
