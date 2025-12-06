package mjson

import (
	"strings"
	"testing"
)

// go test -v  ./...
// go test -v -run Test_UnmarshalJsonc
func Test_UnmarshalJsonc(t *testing.T) {
	str := `{
		// 这是一个注释
		"key1": "value1", /* 这是另一个注释 */
		"key2": 123, // 数字值\
		"name": "墨七"  
	}`
	var result map[string]any
	err := UnmarshalJsonc([]byte(str), &result)
	if err != nil {
		t.Fatalf("UnmarshalJsonc error: %v", err)
	}
	if result["key1"] != "value1" {
		t.Fatalf("key1 value incorrect: %v", result["key1"])
	}
	if num, ok := result["key2"].(float64); !ok || num != 123 {
		t.Fatalf("key2 value incorrect: %v", result["key2"])
	}
	if result["name"] != "墨七" {
		t.Fatalf("name value incorrect: %v", result["name"])
	}
}

func TestMarshalAndUnmarshal(t *testing.T) {
	obj := map[string]int{
		"a": 1,
		"b": 2,
	}
	b, err := Marshal(obj)
	if err != nil {
		t.Fatalf("Marshal error: %v", err)
	}

	var obj2 map[string]int
	err = Unmarshal(b, &obj2)
	if err != nil {
		t.Fatalf("Unmarshal error: %v", err)
	}

	if obj2["a"] != 1 || obj2["b"] != 2 {
		t.Fatalf("Unmarshal result incorrect: %#v", obj2)
	}
}

// 覆盖 ToByte 的正常与错误路径
func TestToByte_NilAndUnmarshalable(t *testing.T) {
	// nil 应返回错误
	if _, err := ToByte(nil); err == nil {
		t.Fatalf("expected error for nil input")
	}

	// 不可序列化类型（channel）应返回 marshal 错误
	ch := make(chan int)
	if _, err := ToByte(ch); err == nil {
		t.Fatalf("expected marshal error for channel type")
	} else if !strings.Contains(err.Error(), "marshal") && !strings.Contains(err.Error(), "unsupported") {
		t.Fatalf("unexpected error message: %v", err)
	}
}

func TestToStr_And_IndentJson(t *testing.T) {
	// 简单结构体应能被序列化为字符串
	type S struct {
		A int `json:"a"`
		B int `json:"b"`
	}
	s := S{A: 1, B: 2}
	str := ToStr(s)
	if !strings.Contains(str, "\"a\"") || !strings.Contains(str, "1") {
		t.Fatalf("ToStr 返回的 JSON 不包含预期内容: %s", str)
	}

	ind := IndentJson(s)
	if ind == "{}" {
		t.Fatalf("IndentJson 在有效输入上返回了错误默认值: %s", ind)
	}
	if !strings.Contains(ind, "\"a\": 1") || !strings.Contains(ind, "\"b\": 2") {
		t.Fatalf("IndentJson 返回格式不正确: %s", ind)
	}

	// 对不可序列化类型，ToStr/IndentJson 应返回默认 "{}"
	ch := make(chan int)
	if ToStr(ch) != "{}" {
		t.Fatalf("ToStr 对不可序列化类型应返回 '{}'，实际: %s", ToStr(ch))
	}
	if IndentJson(ch) != "{}" {
		t.Fatalf("IndentJson 对不可序列化类型应返回 '{}'，实际: %s", IndentJson(ch))
	}
}

func TestToMap_SuccessAndNil(t *testing.T) {
	// 正常情况：结构体 -> map，数值会被解为 float64
	type T struct {
		X int `json:"x"`
	}
	in := T{X: 5}
	m, err := ToMap(in)
	if err != nil {
		t.Fatalf("ToMap 正常输入返回错误: %v", err)
	}
	v, ok := m["x"]
	if !ok {
		t.Fatalf("ToMap 返回的 map 缺少键 'x'")
	}
	// json.Unmarshal 到 map[string]any 时数字为 float64
	fv, ok := v.(float64)
	if !ok || fv != 5 {
		t.Fatalf("ToMap 返回的值类型和值不正确: %#v", v)
	}

	// nil 输入应返回错误
	if _, err := ToMap(nil); err == nil {
		t.Fatalf("ToMap 对 nil 输入应返回错误")
	}
}

func TestPrintAny_ReturnsIndentString(t *testing.T) {
	type A struct {
		N int `json:"n"`
	}
	a := A{N: 3}
	res := IndentJson(a)
	if !strings.Contains(res, "\"n\": 3") {
		t.Fatalf("PrintAny 返回的字符串不包含预期内容: %s", res)
	}
}
