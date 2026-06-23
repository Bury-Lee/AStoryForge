package story_struct

import "time"

//这里是用于结构化故事的核心定义

// ProjectMeta 简短描述 project.json 元信息
// 说明: 只保存工程级配置, 事件和实体独立存储
type WorldSetting struct {
	TimeSystem []string          `json:"time_system,omitempty"` // 时间系统定义,例子: ["year:描述", "month:描述", "day:描述"]这种形式
	TimeLabels map[string]string `json:"time_labels,omitempty"` // 时间标签扩展,例子: map[string]string{"year": "2023", "month": "04-01", "day": "01-01"}
	Spine      []string          `json:"spine,omitempty"`       // 叙事链顺序,表现为事件 ID 列表
	WriteRules []string          `json:"write_rules,omitempty"` // 写作约束
	WorldRules []string          `json:"world_rules,omitempty"` // 世界规则
}

// ProjectSummary 简短描述工程概览,预备字段
// 说明: 用于前端工程列表页
type ProjectMeta struct {
	RootDir     string            `json:"root_dir"`             // 工程所在目录
	Title       string            `json:"title"`                // 工程标题
	Author      string            `json:"author,omitempty"`     // 作者
	Settings    map[string]string `json:"settings,omitempty"`   // 工程设置,表现为键值对
	CreatedAt   time.Time         `json:"created_at,omitempty"` // 创建时间
	UpdatedAt   time.Time         `json:"updated_at,omitempty"` // 更新时间
	EventCount  int               `json:"event_count"`          // 事件总数
	EntityCount int               `json:"entity_count"`         // 实体总数
}

// EventEdge 简短描述事件之间的因果边
// 参数: Target - 指向的目标事件 ID
// 返回:无
// 说明: in_edges 和 out_edges 共用同一结构
type EventEdge struct {
	Target   string            `json:"target"`             // 目标事件 ID
	Type     EdgeType          `json:"type"`               // 边类型
	Discribe map[string]string `json:"discribe,omitempty"` // 边描述,具体的描述其因果逻辑
}

// Participant 简短描述事件参与者
// 说明: state 记录入场快照, impact 记录出场变化
type Participant struct {
	EntityID string            `json:"entity_id"`       // 实体 ID
	State    map[string]string `json:"state,omitempty"` // 描述,包括不限于参与时状态,影响,结果等
}

// Event 简短描述故事事件节点
// 说明: 卷章分层也由 Event 承担, 通过 type 和 sub_events 表达
type Event struct {
	ID           string            `json:"id"`                     // 事件 ID
	Name         string            `json:"name"`                   // 事件名称
	Introduction string            `json:"introduction"`           // 事件概述
	Type         EventType         `json:"type"`                   // 自定义事件类型,可为空
	Time         map[string]string `json:"time,omitempty"`         // 开放时间结构,对应于元标签的时间结构
	SettingRef   string            `json:"setting_ref,omitempty"`  // 所属设定域
	InEdges      []EventEdge       `json:"in_edges,omitempty"`     // 前因边
	OutEdges     []EventEdge       `json:"out_edges,omitempty"`    // 后果边
	SubEvents    []string          `json:"sub_events,omitempty"`   // 子事件 ID
	Participants []Participant     `json:"participants,omitempty"` // 参与者
	SubRules     []string          `json:"sub_rules,omitempty"`    // 子规则,定义AI在写该事件时的约束/注意事项
	Process      string            `json:"process"`                // 事件过程,尽可能详细的事件过程描述
	Outcome      map[string]string `json:"outcome"`                // 事件结果/意义,主要描述其的意义与影响
	Locked       bool              `json:"locked,omitempty"`       // 是否锁定
}

// EntityRelationship 简短描述实体关系
// 说明: 用于表达人物, 组织, 物品之间的多对多关系
type EntityRelationship struct {
	TargetID     string            `json:"target_id"`             // 目标实体 ID
	RelationType string            `json:"relation_type"`         // 关系类型
	Description  map[string]string `json:"description,omitempty"` // 关系描述
}

// Entity 简短描述实体节点
// 说明: introduction 拆成数组, 便于局部更新
type Entity struct {
	ID            string               `json:"id"`                      // 实体 ID
	Name          string               `json:"name"`                    // 实体名称
	Type          string               `json:"type"`                    // 实体类型
	Introduction  []string             `json:"introduction"`            // 实体介绍
	Relationships []EntityRelationship `json:"relationships,omitempty"` // 实体关系
	RuleRefs      []string             `json:"rule_refs,omitempty"`     // 关联规则
	Events        []string             `json:"events,omitempty"`        // 参与事件索引
}

// EventType 简短描述事件节点类型
// 说明: event 为普通事件, container 为纯容器, mixed 为容器加内容
type EventType string

const (
	EventTypeEvent     EventType = "event"
	EventTypeContainer EventType = "container"
	EventTypeMixed     EventType = "mixed"
)

// EdgeType 简短描述事件边类型
// 说明: cause 表示前因, result 表示后果
type EdgeType string

const (
	EdgeTypeCause  EdgeType = "cause"
	EdgeTypeResult EdgeType = "result"
)

// SimulationMode 简短描述模拟模式
// 说明: between 用于填补两事件间空白, inside 用于细化父事件
type SimulationMode string

const (
	SimulationModeBetween SimulationMode = "between"
	SimulationModeInside  SimulationMode = "inside"
)

// SimulationStatus 简短描述模拟会话状态
// 说明: running 进行中, paused 等待人工, completed 已完成, canceled 已取消
type SimulationStatus string

const (
	SimulationStatusRunning   SimulationStatus = "running"
	SimulationStatusPaused    SimulationStatus = "paused"
	SimulationStatusCompleted SimulationStatus = "completed"
	SimulationStatusCanceled  SimulationStatus = "canceled"
)

// SimulationApplyMode 简短描述模拟结果落地方式
// 说明: overwrite 覆盖原事件, append_sub 追加为子事件, append_new 追加为新事件
type SimulationApplyMode string

const (
	SimulationApplyOverwrite SimulationApplyMode = "overwrite"
	SimulationApplySubEvent  SimulationApplyMode = "append_sub"
	SimulationApplyNewEvent  SimulationApplyMode = "append_new"
)

// GenerationStatus 简短描述生成任务状态
// 说明: pending 排队中, running 生成中, completed 已完成, failed 失败
type GenerationStatus string

const (
	GenerationStatusPending   GenerationStatus = "pending"
	GenerationStatusRunning   GenerationStatus = "running"
	GenerationStatusCompleted GenerationStatus = "completed"
	GenerationStatusFailed    GenerationStatus = "failed"
)
