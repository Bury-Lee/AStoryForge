package ai

import (
	"AStoryForge/core"
	"sync"
)

//沙盒的上下文管理

// 这里提供上下文管理,流程控制的功能
func ClearHistory(agent core.AiClientManager) {

}

// 表示一次演员Agent的行为申请.
type Action struct {
	ActorIndex int               `json:"actor_index"` // debug参数,演员Agent的索引,从0开始计数
	Actor      string            `json:"actor"`       // debug参数,演员Agent的名称
	Action     string            `json:"action"`      // 具体的行为行动
	Params     map[string]string `json:"params"`      // 行动的详细标注
	// 例如: 为某个角色添加一个事件,则需要指定事件的类型,事件的内容,事件的时间等.
	Thinking string `json:"think"` // 行动时的想法,思考过程等.当action和params为空时,则认为是跳过.当action和params不为空时,则认为是执行该行动.当action为空时,则认为是角色仅仅在进行思考和心理活动.
}
type ActionList struct {
	Action []Action   //列表
	mu     sync.Mutex //并发保护
}

var actionList []Action

// 在原本的上下文基础上添加新的消息
func AppendHistory(agent core.AiClientManager, message string) {

}
