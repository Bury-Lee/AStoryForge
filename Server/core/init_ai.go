package core

import (
	"AStoryForge/config"

	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

func InitAIClient(config *config.AI) *openai.Client {
	conf := openai.DefaultConfig(config.ApiKey)
	conf.BaseURL = config.Host
	conf.APIType = openai.APIType(config.APIType)

	client := openai.NewClientWithConfig(conf)
	logrus.Info("模型已加载: " + config.Model)

	if client == nil {
		logrus.Panic("ai连接失败!")
	}
	return client
}

func InitAI() {
	// 初始化结构化ai
	AIClient = make([]*AiClientManager, 0, len(Config.AI))
	for _, config := range Config.AI {
		AIClient = append(AIClient, &AiClientManager{
			Enable:      config.Enable,
			Client:      InitAIClient(&config),
			History:     nil,
			ModelName:   config.Model,
			MaxTokens:   config.MaxTokens,
			Temperature: config.Temperature,
		})
	}
	//初始化最终编译ai
	CompilerAI = InitAIClient(&Config.CompilerAI)
}
