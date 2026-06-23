package ai

// =====================================================================================
// 设计架构: 行为申请（Action Proposal）系统
// =====================================================================================
//
// 本文件是 AI Agent 工具层的业务逻辑实现, 核心设计遵循 CQS(命令查询分离)模式:
//
// 1. 读操作 vs 写操作分离
//   - 读操作(查询/读取)任何角色可用, 无权限限制。
//   - 写操作(创建/更新/删除)必须经过"行为申请→审查→执行"三段式流程。
//     演员 Agent 提交申请 -> 观察者/裁判 Agent 审查 -> 原子执行。
//
// 2. 三段式设计理由
//   a) 意图与效果解耦: 演员只需表达"想做什么", 无需持有写权限;
//   b) 一致性校验: 观察者在批准前可做冲突检测、规则校验和合理性判断;
//   c) 可追溯性: 所有申请记录(ActionProposal)形成完整决策链,
//      支持沙盘推演回放和复盘;
//   d) 安全边界: 即使演员 Agent 的提示词被注入或行为异常,
//      其破坏范围也被限制在"申请"层面, 观察者可以拦截。
//
// 3. 角色-工具权限矩阵
//    演员 (Actor)      : 读工具 + 行为申请提交工具(submit_action_proposal 等)
//    导演 (Director)   : 所有工具(含直接写操作)
//    观察者 (Observer) : 读工具 + 审查工具(review_proposal, list_pending_proposals)
//
// 4. 执行流程
//    演员提交申请 -> proposalStore 存储 -> 观察者调用 review_proposal ->
//    approved=true -> executeProposal() 内联分发到 crudService 的对应写方法 ->
//    结果写入 proposal.Result -> 完成。
//
//    注意: executeProposal() 绕过 MCP 工具层直接调用 crudService 的方法,
//    因此观察者无需注册具体的写工具(create_event 等), 只需注册 review_proposal。
// =====================================================================================

import (
	"AStoryForge/function/story_struct"
	"context"
	"encoding/json"
	"fmt"

	"sync"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/sirupsen/logrus"
)

// TestToolArgs 测试工具参数
type TestToolArgs struct {
	Arg1 string `json:"arg1"`
	Arg2 int    `json:"arg2"`
	Arg3 bool   `json:"arg3"`
}

// TestTool 测试工具, 用于验证 Agent 工具调用链路是否正常
func TestTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logrus.Debugf("%+v\n", request)

	argsBytes, err := json.Marshal(request.Params.Arguments)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal arguments: %w", err)
	}

	var testReq TestToolArgs
	if err := json.Unmarshal(argsBytes, &testReq); err != nil {
		return nil, fmt.Errorf("invalid arguments: %w", err)
	}

	return mcp.NewToolResultText(fmt.Sprintf("测试成功，结果: %+v", testReq)), nil
}

// =====================================================================================
// 全局 CRUD 服务实例
// 说明: 所有 MCP 工具共享此实例, 内部方法会按 rootDir/projectID 在调用时重新加载工程文件
// =====================================================================================

var crudService = &story_struct.ProjectObj{}

// =====================================================================================
// 行为申请（Action Proposal）系统
// 说明: 演员 Agent 不直接修改工程对象, 而是通过提交"行为申请"来表达意图;
//       观察者/裁判角色的 Agent 负责审查并统一执行批准后的修改。
//       设计理由: 将"修改意图"与"实际写回"解耦——
//       1) 演员无需直接持有写权限, 降低误操作风险;
//       2) 观察者可做一致性校验、冲突检测和裁决, 形成可追溯的决策链;
//       3) 所有申请记录可用于沙盘推演回放和复盘。
// =====================================================================================

// ProposalStatus 行为申请状态
type ProposalStatus string

const (
	ProposalStatusPending   ProposalStatus = "pending"
	ProposalStatusApproved  ProposalStatus = "approved"
	ProposalStatusRejected  ProposalStatus = "rejected"
	ProposalStatusWithdrawn ProposalStatus = "withdrawn"
)

// ActionProposal 行为申请结构
// 说明: 演员 Agent 提交的每一个修改意图都记录为此结构,
//
//	观察者审查后更新 Status 和 ReviewNote,
//	审查通过时会执行实际修改并将结果写入 Result 字段。
type ActionProposal struct {
	ID         string          `json:"id"`
	ProposerID string          `json:"proposer_id"`         // 申请者标识(Agent ID / 角色名)
	ActionType string          `json:"action_type"`         // 行为类型, 如 "create_event", "update_entity"
	TargetID   string          `json:"target_id,omitempty"` // 目标 ID(事件/实体 ID), 可选
	Payload    json.RawMessage `json:"payload"`             // 行为参数(与对应工具的参数格式一致)
	Status     ProposalStatus  `json:"status"`
	ReviewNote string          `json:"review_note,omitempty"` // 审查备注
	Result     json.RawMessage `json:"result,omitempty"`      // 执行结果(JSON)
	Error      string          `json:"error,omitempty"`       // 执行错误信息
	CreatedAt  string          `json:"created_at"`            // RFC3339 字符串
	ReviewedAt string          `json:"reviewed_at,omitempty"` // RFC3339 字符串
}

// proposalStore 行为申请存储(进程级, 后续可替换为持久化层)
var (
	proposalStore = make(map[string]*ActionProposal)
	proposalMu    sync.RWMutex
)

// generateProposalID 生成唯一申请 ID
func generateProposalID() string {
	return fmt.Sprintf("prop_%d", time.Now().UnixNano())
}

// =====================================================================================
// 工具通用辅助
// =====================================================================================

// bindArgs 将 MCP 调用参数解析到目标结构体
func bindArgs(request mcp.CallToolRequest, target any) error {
	argsBytes, err := json.Marshal(request.Params.Arguments)
	if err != nil {
		return fmt.Errorf("序列化参数失败: %w", err)
	}
	if err := json.Unmarshal(argsBytes, target); err != nil {
		return fmt.Errorf("参数解析失败: %w", err)
	}
	return nil
}

// formatJSON 把对象格式化为缩进 JSON 字符串
func formatJSON(v any) string {
	b, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return fmt.Sprintf("序列化失败: %v", err)
	}
	return string(b)
}

// resolveTarget 解析工程目标路径,优先取 rootDir, 否则取 projectID
func resolveTarget(rootDir, projectID string) string {
	if rootDir != "" {
		return rootDir
	}
	return projectID
}

// =====================================================================================
// Schema 快捷构造方法
// 说明: 以下辅助函数用于构建 JSON Schema 定义, 供 CRUDToolDefinitions 和
//       ProposalToolDefinitions 使用。
// =====================================================================================

// stringProp 快捷构造字符串参数定义
func stringProp(desc string) jsonschema.Definition {
	return jsonschema.Definition{Type: jsonschema.String, Description: desc}
}

// intProp 快捷构造整数参数定义
func intProp(desc string) jsonschema.Definition {
	return jsonschema.Definition{Type: jsonschema.Integer, Description: desc}
}

// boolProp 快捷构造布尔参数定义
func boolProp(desc string) jsonschema.Definition {
	return jsonschema.Definition{Type: jsonschema.Boolean, Description: desc}
}

// objectProp 快捷构造对象参数定义
func objectProp(desc string, props map[string]jsonschema.Definition, required []string) jsonschema.Definition {
	return jsonschema.Definition{
		Type:        jsonschema.Object,
		Description: desc,
		Properties:  props,
		Required:    required,
	}
}

// arrayProp 快捷构造数组参数定义
func arrayProp(desc string, items jsonschema.Definition) jsonschema.Definition {
	return jsonschema.Definition{
		Type:        jsonschema.Array,
		Description: desc,
		Items:       &items,
	}
}

// eventSchema 事件节点 JSON Schema 描述
func eventSchema() jsonschema.Definition {
	return objectProp("事件节点结构, 字段含义见 story_struct.Event",
		map[string]jsonschema.Definition{
			"id":           stringProp("事件 ID, 为空时自动生成"),
			"name":         stringProp("事件名称"),
			"introduction": stringProp("事件概述"),
			"type":         stringProp("事件类型: event | container | mixed"),
			"time":         objectProp("开放时间结构, key 为时间标签", map[string]jsonschema.Definition{}, nil),
			"setting_ref":  stringProp("所属设定域"),
			"in_edges":     arrayProp("前因边列表", objectProp("边", map[string]jsonschema.Definition{"target": stringProp("目标事件 ID"), "type": stringProp("cause | result")}, nil)),
			"out_edges":    arrayProp("后果边列表", objectProp("边", map[string]jsonschema.Definition{"target": stringProp("目标事件 ID"), "type": stringProp("cause | result")}, nil)),
			"sub_events":   arrayProp("子事件 ID 列表", stringProp("子事件 ID")),
			"participants": arrayProp("参与者列表", objectProp("参与者", map[string]jsonschema.Definition{"entity_id": stringProp("实体 ID"), "state": objectProp("入场状态快照", map[string]jsonschema.Definition{}, nil)}, nil)),
			"sub_rules":    arrayProp("子规则, AI 写作时的约束/注意事项", stringProp("规则条目")),
			"process":      stringProp("事件过程, 详细描述"),
			"outcome":      objectProp("事件结果/意义", map[string]jsonschema.Definition{}, nil),
			"locked":       boolProp("是否锁定"),
		}, nil)
}

// entitySchema 实体节点 JSON Schema 描述
func entitySchema() jsonschema.Definition {
	return objectProp("实体节点结构, 字段含义见 story_struct.Entity",
		map[string]jsonschema.Definition{
			"id":            stringProp("实体 ID, 为空时自动生成"),
			"name":          stringProp("实体名称"),
			"type":          stringProp("实体类型, 如 character / org / item"),
			"introduction":  arrayProp("实体介绍, 多段文本", stringProp("介绍段落")),
			"relationships": arrayProp("实体关系列表", objectProp("关系", map[string]jsonschema.Definition{"target_id": stringProp("目标实体 ID"), "relation_type": stringProp("关系类型"), "description": objectProp("关系描述", map[string]jsonschema.Definition{}, nil)}, nil)),
			"rule_refs":     arrayProp("关联规则 ID", stringProp("规则 ID")),
			"events":        arrayProp("参与事件索引", stringProp("事件 ID")),
		}, nil)
}

// projectMetaSchema 工程元信息 Schema
func projectMetaSchema() jsonschema.Definition {
	return objectProp("工程元信息, 见 story_struct.ProjectMeta",
		map[string]jsonschema.Definition{
			"root_dir":     stringProp("工程所在目录"),
			"title":        stringProp("工程标题"),
			"author":       stringProp("作者"),
			"settings":     objectProp("工程设置, 表现为键值对", map[string]jsonschema.Definition{}, nil),
			"created_at":   stringProp("创建时间, RFC3339 字符串"),
			"updated_at":   stringProp("更新时间, RFC3339 字符串"),
			"event_count":  intProp("事件总数"),
			"entity_count": intProp("实体总数"),
		}, []string{"title"})
}

// =====================================================================================
// TestTool 工具定义
// 说明: 提供给角色工厂注册的调试工具定义
// =====================================================================================

// TestToolDefinition 返回内置的 TestTool 工具定义, 供角色工厂复用
func TestToolDefinition() ToolDefinition {
	return ToolDefinition{
		Name:        "TestTool",
		Description: "测试工具，用于验证 Agent 工具调用链路是否正常",
		Parameters: jsonschema.Definition{
			Type: jsonschema.Object,
			Properties: map[string]jsonschema.Definition{
				"arg1": {Type: jsonschema.String, Description: "字符串参数，接受任意文本输入"},
				"arg2": {Type: jsonschema.Integer, Description: "整数参数"},
				"arg3": {Type: jsonschema.Boolean, Description: "布尔参数"},
			},
		},
		Executor: TestTool,
	}
}
