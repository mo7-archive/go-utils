package main

import (
	"fmt"

	"github.com/m-startgo/go-utils/mjson"
)

func main() {
	// 读取文件
	var data map[string]any
	err := mjson.ReadyJsonFile("demo.json", &data)
	if err != nil {
		fmt.Println("读取文件出错：", err)
		return
	}

	fmt.Println("解析结果", data["test_metadata"].(map[string]any)["name"])
}
