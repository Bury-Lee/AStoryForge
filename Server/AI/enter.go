package ai

//在这里写关于上下文管理,计划规划等的函数

import (
	"AStoryForge/core"
	"context"
	"fmt"

	"github.com/sashabaranov/go-openai"
	"github.com/sirupsen/logrus"
)

//这里要实现的事情有几样:1,多智能体推演,进行一小剧场,然后可以指定几位"导演",这几位导演进行演员行为的合理性判定,拥有否决,提示等多种权利,
//演员agent注入上下文,并且给予身份,由其同多个其他agent演员交互,进行剧情推演,最终生成一个完整的剧情.

const (
	// 实体
	EntityEntityType = "entity"
	// 旁观者/导演
	ObserverEntityType = "observer"
	// 表示要其他的事情....
)

// input: 输入的用户提示
// prompt: 提示词,用于引导agent行为
// agentIndex: 智能体索引,用于指定要使用的智能体
// 返回值:result: 模型输出的结果进行封装,包含模型输出的内容
// err: 错误信息,用于表示调用失败的原因
// 说明,该函数应该在一个沙盘中调用,类似回合制同步那样,并行的让事件发生,每个事件发生后,需要根据事件的结果,更新上下文,并继续下一次事件发生
func StartAgent(input string, prompt string, Index int, role string, MaxRounds int) (result string, err error) {
	if core.AIClient == nil {
		return "", fmt.Errorf("AIClient未初始化")
	}
	ctx := context.Background()
	tools := AgentTools.GetTools()

	model := core.AIClient[Index]
	// 初始化阶段,填入提示词,工具,skill

	//构造初始的上下文,往后不再主动更改

	for round := 0; round < MaxRounds; round++ {
		//TODO:消息初始化和工具应该放在循环外面
		firstResp, err := model.Client.CreateChatCompletion(ctx, openai.ChatCompletionRequest{
			Model: model.ModelName,
			// Messages: core.History,//TODO:实现上下文管理的结构体
			Tools: tools, // 提供工具能力
			// ToolChoice: "auto" 是默认行为，模型自主决定是否调用工具。
			ToolChoice:  "auto",
			MaxTokens:   model.MaxTokens,
			Temperature: model.Temperature,
		})
		if err != nil {
			logrus.Errorf("Agent %s 调用失败:%v", model.ModelName, err)
			return "", err
		}

		//根据需要启用,是否返回空结果,或无工具调用,则视为结束
		// if len(firstResp.Choices) == 0 {
		// 	logrus.Info("自动恢复模型返回空结果,结束本轮")
		// 	return
		// }

		// if firstResp.Choices[0].Message.Content == "" && len(firstResp.Choices[0].Message.ToolCalls) == 0 {
		// 	logrus.Info("自动恢复模型返回空内容且无工具调用,结束本轮")
		// 	return
		// }

		if len(firstResp.Choices) == 0 {
			continue
		}

		// 3. 可选：记录缓存命中情况（结合你之前的问题）
		if firstResp.Usage.PromptTokensDetails != nil {
			cachedTokens := firstResp.Usage.PromptTokensDetails.CachedTokens
			totalTokens := firstResp.Usage.PromptTokens
			hitRate := float64(cachedTokens) / float64(totalTokens) * 100
			logrus.Debugf("缓存命中率: %.2f%%, 缓存token: %d/%d",
				hitRate, cachedTokens, totalTokens)
		}

		assistantMessage := firstResp.Choices[0].Message
		model.History = append(model.History, assistantMessage)

		// 按需要,可设置为没有工具调用时视为本轮完成
		// if len(assistantMessage.ToolCalls) == 0 {
		// 	logrus.Infof("自动恢复输出:%s", assistantMessage.Content)
		// 	return
		// }
		logrus.Debugf("模型输出:%s", assistantMessage.Content)
		for _, toolCall := range assistantMessage.ToolCalls {
			//如果有工具调用,每一轮都要顺序执行
			seqlock.Lock()
			defer seqlock.Unlock()
			toolResult, toolErr := ExecuteToolCall(ctx, toolCall, AgentTools)
			if toolErr != nil {
				toolResult = fmt.Sprintf("工具执行失败:%v", toolErr)
				logrus.Warnf("工具 %s 执行失败:%v", toolCall.Function.Name, toolErr)
			}
			logrus.Debugf("工具 %s 执行结果:%s", toolCall.Function.Name, toolResult)
			model.History = append(model.History, openai.ChatCompletionMessage{
				Role:       openai.ChatMessageRoleTool,
				ToolCallID: toolCall.ID,
				Content:    toolResult,
				Name:       toolCall.Function.Name,
			})
			//可以主动选择跳过本轮,但是由导演判断是否结束本轮.如下使用工具获取沙盒的包级变量来获取其轮次的结束,判断信息

			//某些特殊的行为类tools调用后直接结束,将结果声明并等待一轮模拟结束
			if toolCall.Function.Name == "？？？" && toolErr == nil { //特殊处理,调用完成函数的话就直接返回结果
				logrus.Debugf("自动恢复结果输出:%s", toolResult)
				//TODO:根据toolResult,判断是否成功,并记录到数据库,以及是否提炼skill
				return toolResult, nil
			}
		}
	}
	logrus.Warnf("Agent %s 达到最大轮次限制(%d),停止继续调用", model.ModelName, MaxRounds)
	return "", fmt.Errorf("Agent %s 达到最大轮次限制(%d),停止继续调用", model.ModelName, MaxRounds)
}

//流程编排:
// 先初始化History,构造一份Mermaid图给ai和对应的角色提示信息
//提供消息追加的方便接口
