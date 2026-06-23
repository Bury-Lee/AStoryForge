package utils

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// FormatJSON 将未格式化的JSON字符串美化为格式化输出
// 输入: rawJSON - 未格式化的JSON字符串
// 输出: 格式化后的JSON字符串和可能的错误
func FormatJSON(rawJSON string) (string, error) {
	var prettyJSON bytes.Buffer

	// 将字符串解析为interface{}，以便重新编码
	var data interface{}
	err := json.Unmarshal([]byte(rawJSON), &data)
	if err != nil {
		return "", fmt.Errorf("解析JSON失败: %v", err)
	}

	// 编码为格式化JSON，缩进2个空格
	encoder := json.NewEncoder(&prettyJSON)
	encoder.SetIndent("", "  ")
	encoder.SetEscapeHTML(false) // 保持原始字符，不转义HTML
	err = encoder.Encode(data)
	if err != nil {
		return "", fmt.Errorf("格式化JSON失败: %v", err)
	}

	return prettyJSON.String(), nil
}

// FormatJSONWithIndent 自定义缩进的格式化函数
// indent: 缩进字符串，如 "  " 或 "\t"
func FormatJSONWithIndent(rawJSON string, indent string) (string, error) {
	var prettyJSON bytes.Buffer

	var data interface{}
	err := json.Unmarshal([]byte(rawJSON), &data)
	if err != nil {
		return "", fmt.Errorf("解析JSON失败: %v", err)
	}

	encoder := json.NewEncoder(&prettyJSON)
	encoder.SetIndent("", indent)
	encoder.SetEscapeHTML(false)
	err = encoder.Encode(data)
	if err != nil {
		return "", fmt.Errorf("格式化JSON失败: %v", err)
	}

	return prettyJSON.String(), nil
}

// FormatJSONSimple 使用json.Indent的简单版本
func FormatJSONSimple(rawJSON string) (string, error) {
	var prettyJSON bytes.Buffer
	err := json.Indent(&prettyJSON, []byte(rawJSON), "", "  ")
	if err != nil {
		return "", fmt.Errorf("格式化JSON失败: %v", err)
	}
	return prettyJSON.String(), nil
}
