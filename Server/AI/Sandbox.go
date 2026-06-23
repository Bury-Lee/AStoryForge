package ai

import (
	"AStoryForge/function/story_struct"
	"context"
	"errors"
	"sync"

	"github.com/sirupsen/logrus"
)

// 通过多协程来启动多个智能体来实现并行沙盘。
// 对于实体,提供的是"行为工具"; 对于观察者/导演,提供的是"判断工具"和必要的 CRUD 工具。
var seqlock sync.Mutex // 用于保护顺序执行的互斥锁
var Exit bool = false  // 表示是否需要退出沙盘
var Stop bool = false  // 表示是否需要暂停沙盘

// SandboxAssessment 表示观察者对单步行动的判定结果。
type SandboxAssessment string

const (
	// SandboxAssessmentApproved 表示行动合理,可直接进入下一步状态更新。
	SandboxAssessmentApproved SandboxAssessment = "approved"
	// SandboxAssessmentFlagged 表示行动基本可接受,但需要保留问题提示给用户。
	SandboxAssessmentFlagged SandboxAssessment = "flagged"
	// SandboxAssessmentRejected 表示行动严重不合理,需要触发回退或人工介入。
	SandboxAssessmentRejected SandboxAssessment = "rejected"
)
const (
	演员  string = "演员"
	观察者 string = "观察者"
)

// SandboxRollbackReason 表示本轮需要回退的原因。
type SandboxRollbackReason string

const (
	// SandboxRollbackHardConstraint 表示违反了不可突破的硬约束。
	SandboxRollbackHardConstraint SandboxRollbackReason = "hard_constraint_violation"
	// SandboxRollbackNoProgress 表示连续多轮没有明显推进。
	SandboxRollbackNoProgress SandboxRollbackReason = "no_progress"
	// SandboxRollbackGoalConflict 表示角色行为与自身核心目标冲突。
	SandboxRollbackGoalConflict SandboxRollbackReason = "goal_conflict"
	// SandboxRollbackActionConflict 表示角色之间或角色与场景状态之间存在明显逻辑冲突。
	SandboxRollbackActionConflict SandboxRollbackReason = "action_conflict"
)

const (
	// SandboxDefaultMaxRounds 默认最大推演轮数。
	SandboxDefaultMaxRounds = 50
	// SandboxDefaultMaxRollbackCount 默认最大自动回退次数。
	SandboxDefaultMaxRollbackCount = 3
	// SandboxDefaultMaxNoProgressRounds 默认允许的连续无进展轮数。
	SandboxDefaultMaxNoProgressRounds = 3
)

var ErrSandboxNotImplemented = errors.New("sandbox not implemented")

const (
	// 启动续写后续的agent
	SandboxTypeContinue = "continue"
	// 填充细节的agent
	SandboxTypeFillDetail = "fill_detail"
	// 从原始文件中提取并结构化信息的agent
	SandboxTypeLoadFromRaw = "load_from_raw"
)

// SandboxRequest 定义一次沙盘推演所需的完整参数。
// 这里兼容文档中的两种主流程:
// 1. between: 填补两个事件之间的空白
// 2. inside: 细化某个父事件内部过程
// 3. load_from_raw: 从原始文件中提取并结构化信息
// 4. continue: 继续续写后续的agent
type SandboxRequest struct {
	RootDir   string `json:"root_dir"`   // 工程根目录
	ProjectID string `json:"project_id"` // 工程 ID

	// 推演类型
	SandboxType string `json:"sandbox_type,omitempty"` // 推演类型: continue / fill_detail / load_from_raw

	Description string `json:"description,omitempty"` // 一般为空,如果不为空,则认为是追加了要求或者提示.在类型为原始资料输入时,描述会作为原始资料的组合文本.

	Mode story_struct.SimulationMode `json:"mode"` // 推演模式: between / inside

	StartEventID  string `json:"start_event_id,omitempty"`  // 方式A: 前因事件 ID
	EndEventID    string `json:"end_event_id,omitempty"`    // 方式A: 后果事件 ID
	ParentEventID string `json:"parent_event_id,omitempty"` // 方式B: 被细化的父事件 ID

	ParticipantIDs []string `json:"participant_ids,omitempty"` // 参与实体 ID 列表

	ActorAgentIndexes    []int `json:"actor_agent_indexes,omitempty"`    // 扮演角色的 AI 索引
	ObserverAgentIndexes []int `json:"observer_agent_indexes,omitempty"` // 负责裁决的观察者 AI 索引

	InitialSceneState map[string]any `json:"initial_scene_state,omitempty"` // 初始场景快照
	Goal              string         `json:"goal,omitempty"`                // 本次推演目标
	UserPrompt        string         `json:"user_prompt,omitempty"`         // 用户追加引导语,一般为空

	HardRules  []string `json:"hard_rules,omitempty"`  // 不可违反的规则
	SoftRules  []string `json:"soft_rules,omitempty"`  // 可违反但需标记的规则
	WorldRules []string `json:"world_rules,omitempty"` // 世界观规则

	MaxRounds           int  `json:"max_rounds,omitempty"`             // 最大轮次
	MaxRollbackCount    int  `json:"max_rollback_count,omitempty"`     // 最大自动回退次数
	MaxNoProgressRounds int  `json:"max_no_progress_rounds,omitempty"` // 连续无进展阈值
	AllowUserIntervene  bool `json:"allow_user_intervene,omitempty"`   // 是否允许用户中途介入

	TargetEventID string `json:"target_event_id,omitempty"` // 覆盖/落地目标事件
	InsertIndex   int    `json:"insert_index,omitempty"`    // 作为子事件或新事件插入的位置
}

// Sandbox 开启一次具体的沙盘模拟。
// 当前先完成参数和常量设计,具体执行逻辑后续再补充。
// 参数:上下文,推演类型,请求参数,工程对象
// 返回:模拟会话,错误

func Sandbox(ctx context.Context, req SandboxRequest, obj *story_struct.ProjectObj) (story_struct.SimulationSession, error) {
	//通过Switch语句,根据推演类型,选择不同的流程.

	//组装提示词,分配演员Agent和观察者Agent,为每个参与的实体列表的成员创建一个agent

	//分配n个演员Agent,每个演员Agent负责一个参与的实体列表的成员
	//分配n个观察者Agent,每个观察者Agent负责一个参与的实体列表的成员,负责裁决该成员的行动是否有效,进行1人/多人的投票

	//在操作时记得上锁保证顺序.

	// 每个演员Agent提交了一个工具"甚至可以是跳过"时,本轮回合才算结束.
	// 观察者Agent根据演员Agent的工具结果,进行投票.

	// 以上的上下文信息都会在全局变量里,等待所有人完成一个回合后,再进行下一轮.

	//直到观察者投票表决最终完成,才会结束.
	// 在每个回合开始前检查Stop变量
	for round := 0; round < req.MaxRounds; round++ {
		if Stop {
			logrus.Infof("沙盘已暂停,当前轮次: %d", round)
			break
		}

		// 以上过程会重复进行,直到满足最大轮次或最大回退次数.

		return story_struct.SimulationSession{}, ErrSandboxNotImplemented
	}
	return story_struct.SimulationSession{}, ErrSandboxNotImplemented
}

// sandboxType为推演后续时的流程
func SandboxContinue(ctx context.Context, req SandboxRequest, obj *story_struct.ProjectObj) (story_struct.SimulationSession, error) {
	//组装提示词,分配演员Agent和观察者Agent,为每个参与的实体都创建一次agent任务

	for i, actorID := range req.ParticipantIDs {

		var 演员obj = obj.Entity[actorID]
		var 提示词 = "你是" + 演员obj.Name + "," + "你的特点" + 演员obj.Introduction[1] //......反正把这里组装好
		go StartAgent(提示词, "", i, 演员, SandboxDefaultMaxRounds)
	}

	//每一步都应该要组装SimulationSession
	return story_struct.SimulationSession{}, ErrSandboxNotImplemented
}

//如果可以,为某个角色的每一步行动打分就好了,计算其每一步的平均分,自由选择回退节点,像是git那样,带来更高的容错空间和更灵活的控制能力
//不过目前做不到
