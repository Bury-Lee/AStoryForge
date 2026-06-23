package core

import (
	"AStoryForge/config"

	"github.com/sashabaranov/go-openai"
)

var Config *config.Config

type AiClientManager struct { //预备字段,可能之后就用这种封装了
	Enable      bool //是否启用,不启用的话当analysis调用MCP的话自动移交人工
	Client      *openai.Client
	History     []openai.ChatCompletionMessage
	ModelName   string
	MaxTokens   int
	Temperature float32
	Tools       []openai.Tool
}

// var AIClient []*openai.Client //也许要再做一层封装
var AIClient []*AiClientManager

var CompilerAI *openai.Client //实际上,最终构建ai建议保持一致,避免文风不一样的问题
