package utils

import (
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

func ToolResultToText(result *mcp.CallToolResult) string {
	if result == nil {
		return "" // 无结果时返回空字符串，避免panic
	}

	var parts []string
	// result.Content 是 []interface{} 类型，具体元素可能是 mcp.TextContent 或 *mcp.TextContent
	for _, item := range result.Content {
		switch content := item.(type) {
		case mcp.TextContent:
			parts = append(parts, content.Text)
		case *mcp.TextContent: // 处理指针类型，两种类型都覆盖
			parts = append(parts, content.Text)
		}
		// 如果未来有ImageContent等其他类型，可以在此扩展case，但目前忽略非文本内容
	}
	return strings.Join(parts, "\n")
}

func NewOpenAITool(params jsonschema.Definition, name, Description string) openai.Tool {
	// 使用jsonschema.Definition构建参数结构，这是OpenAI官方库提供的便捷方式

	//参数示例
	// params := jsonschema.Definition{
	// 	Type: jsonschema.Object, // 参数本身是一个JSON对象
	// 	Properties: map[string]jsonschema.Definition{
	// 		"city": {
	// 			Type:        jsonschema.String,  // 属性类型：字符串
	// 			Description: "要查询天气的城市,例如北京或上海", // 给模型看的示例和说明
	// 		},
	// 	},
	// 	Required: []string{"city"}, // 约束模型：调用时必须提供city字段
	// }

	// 返回符合OpenAI规范的工具对象
	return openai.Tool{
		Type: openai.ToolTypeFunction, // 固定值，表示这是一个函数调用工具
		Function: &openai.FunctionDefinition{
			Name:        name,        // 工具名称，模型后续会通过此名称发起调用
			Description: Description, // 模型的决策依据之一，清晰的描述能提高调用准确率
			Parameters:  params,      // 附加的参数JSON Schema
		},
	}
}
