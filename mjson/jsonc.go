package mjson

import (
	"errors"
	"os"
)

/*
解析带注释的 JSON 数据
var result map[string]any
err := UnmarshalJsonc([]byte(str), &result)
*/
func UnmarshalJsonc(data []byte, v any) error {
	data2, err := StripComments(data)
	if err != nil {
		return err
	}
	return Unmarshal(data2, v)
}

/*
读取一个包含注释的 JSON 文件，去除注释后返回标准 JSON 格式的数据。
var result map[string]any
err := ReadyJsonFile("demo.json", &result)
*/
func ReadyJsonFile(filePath string, v any) (resErr error) {
	fileByte, err := os.ReadFile(filePath)
	if err != nil {
		resErr = err
		return
	}
	return UnmarshalJsonc(fileByte, v)
}

func StripComments(data []byte) ([]byte, error) {
	var result []byte
	inString := false       // 是否在双引号字符串中
	inLineComment := false  // 是否在单行注释//中
	inBlockComment := false // 是否在多行注释/* */中
	prevChar := byte(0)     // 前一个字符，用于识别/*和*/

	for _, b := range data {
		// 处理字符串状态（双引号）：转义的"不切换状态
		if b == '"' && !inLineComment && !inBlockComment {
			if prevChar != '\\' {
				inString = !inString
			}
			result = append(result, b)
			prevChar = b
			continue
		}

		// 字符串内的字符直接保留
		if inString {
			result = append(result, b)
			prevChar = b
			continue
		}

		// 处理单行注释：遇到换行符退出
		if inLineComment {
			if b == '\n' {
				inLineComment = false
				result = append(result, b) // 保留换行符
			}
			prevChar = b
			continue
		}

		// 处理多行注释
		if inBlockComment {
			// 检测注释结束符 */
			if prevChar == '*' && b == '/' {
				inBlockComment = false
				prevChar = b
				continue
			}
			// 检测嵌套多行注释（JSONC 不支持）
			if prevChar == '/' && b == '*' {
				return nil, errors.New("nested block comments are not allowed in JSONC")
			}
			prevChar = b
			continue
		}

		// 检测单行注释开始 //
		if prevChar == '/' && b == '/' {
			inLineComment = true
			// 移除已添加的前一个 /
			if len(result) > 0 {
				result = result[:len(result)-1]
			}
			prevChar = b
			continue
		}

		// 检测多行注释开始 /*
		if prevChar == '/' && b == '*' {
			inBlockComment = true
			// 移除已添加的前一个 /
			if len(result) > 0 {
				result = result[:len(result)-1]
			}
			prevChar = b
			continue
		}

		// 非注释字符直接保留
		result = append(result, b)
		prevChar = b
	}

	// 检查未闭合的多行注释
	if inBlockComment {
		return nil, errors.New("unclosed block comment in JSONC")
	}

	return result, nil
}
