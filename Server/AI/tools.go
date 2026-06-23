// Package ai 提供 AI Agent 的核心工具管理能力。
//
// 本文件定义了 Agent 可用的工具（Tool）体系，包括：
//   - ToolDefinition：工具的定义结构（名称、描述、参数模式、执行函数）
//   - agentTools：工具的注册、查询、转换管理器
//   - 角色工厂 NewAgentTools / NewActorTools / NewDirectorTools / NewObserverTools
//     按角色显式注册可用工具, 便于权限控制
//
// 工作流程：
//  1. 定义工具 -> 在 function.go 创建 ToolDefinition 并加入 CRUDToolDefinitions()
//  2. 注册工具 -> 通过角色工厂 NewXxxTools() 显式调用 AddTool 注册
//  3. 获取 OpenAI 格式 -> agentTools.GetTools() -> 供 Function Calling 使用
//  4. 获取 MCP 格式   -> agentTools.RegisterTools() -> 供 MCP 协议调用

package ai

import (
	"context"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sashabaranov/go-openai"
	"github.com/sashabaranov/go-openai/jsonschema"
)

// ToolExecutor 是工具的执行函数签名。
// 它接收 MCP 协议的 CallToolRequest，执行具体逻辑后返回 CallToolResult。
type ToolExecutor func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error)

// ToolDefinition 定义一个完整的 Agent 工具。
// 包含 OpenAI Function Calling 所需的元信息以及实际执行逻辑。
type ToolDefinition struct {
	// Name 工具名称，必须唯一，作为工具的标识符。
	Name string

	// Description 工具描述，OpenAI 模型据此判断何时调用此工具。
	Description string

	// Parameters 工具的 JSON Schema 参数定义，描述工具需要的入参结构。
	Parameters jsonschema.Definition

	// Executor 工具的实际执行函数，收到调用请求后执行具体逻辑。
	Executor ToolExecutor
}

// OpenAITool 将 ToolDefinition 转换为 go-openai 库的 openai.Tool 格式，
// 供 OpenAI Function Calling API 使用。
func (t ToolDefinition) OpenAITool() openai.Tool {
	return openai.Tool{
		Type: openai.ToolTypeFunction,
		Function: &openai.FunctionDefinition{
			Name:        t.Name,
			Description: t.Description,
			Parameters:  t.Parameters,
		},
	}
}

// agentTools 是工具注册管理器，持有所有已注册工具的 map。
// 提供工具注册、查询、格式转换等能力。
type agentTools struct {
	// tools 以工具名为 key 存放所有 ToolDefinition
	tools map[string]ToolDefinition
}

// ensureInit 确保 tools map 已初始化（懒加载模式），
// 避免在未调用 InitAgentTools 时发生 nil map 写入 panic。
func (a *agentTools) ensureInit() {
	if a.tools == nil {
		a.tools = make(map[string]ToolDefinition)
	}
}

// AddTool 注册一个工具到管理器。
// 如果工具名称为空则静默忽略。
// 同名工具会被后注册的覆盖。
func (a *agentTools) AddTool(def ToolDefinition) {
	if def.Name == "" {
		return
	}

	a.ensureInit()
	a.tools[def.Name] = def
}

// GetTools 返回所有已注册工具的 OpenAI Tool 列表，
// 用于提供给 OpenAI API 的 Function Calling 参数。
func (a *agentTools) GetTools() []openai.Tool {
	a.ensureInit()

	tools := make([]openai.Tool, 0, len(a.tools))
	for _, tool := range a.tools {
		tools = append(tools, tool.OpenAITool())
	}

	return tools
}

// RegisterTools 返回所有已注册工具的 Executor 映射（工具名 -> 执行函数），
// 用于 MCP 协议调用时按名称路由到对应的处理函数。
func (a *agentTools) RegisterTools() map[string]func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	a.ensureInit()

	tools := make(map[string]func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error), len(a.tools))
	for k, v := range a.tools {
		tools[k] = v.Executor
	}
	return tools
}

// =====================================================================================
// 角色工厂
// 说明: 每个角色(Agent / 演员 / 导演 / 观察者) 通过显式调用 AddTool 注册各自可用的工具。
//       显式列出而非通配注册, 便于审查与权限控制。
//       新增工具步骤:
//         1. 在 function.go 中实现执行函数, 并加入 CRUDToolDefinitions()
//         2. 在对应角色的工厂中按需添加 mustTool("<name>")
// =====================================================================================

// NewAgentTools 注册全部可用工具(含 TestTool), 作为基础/调试工具集
func NewAgentTools() AITool {
	a := &agentTools{}
	a.tools = make(map[string]ToolDefinition)

	// 调试工具
	a.AddTool(TestToolDefinition())

	// CRUD 工具 - 全部注册
	for _, def := range CRUDToolDefinitions() {
		a.AddTool(def)
	}
	return a
}

// NewActorTools 注册演员可用工具集
// 说明: 演员负责日常剧情搭建与编辑, 可执行完整的 CRUD 操作(除工程级新建/删除/校验外)
func NewActorTools() AITool {
	a := &agentTools{}
	a.tools = make(map[string]ToolDefinition)

	// === 调试 ===
	a.AddTool(TestToolDefinition())

	// === 工程读 ===
	a.AddTool(mustTool("list_projects"))
	a.AddTool(mustTool("load_project"))

	// === 事件 CRUD ===
	a.AddTool(mustTool("list_events"))
	a.AddTool(mustTool("get_event"))
	a.AddTool(mustTool("create_event"))
	a.AddTool(mustTool("update_event"))
	a.AddTool(mustTool("delete_event"))
	a.AddTool(mustTool("set_event_locked"))
	a.AddTool(mustTool("update_event_participants"))

	// === 事件边 CRUD ===
	a.AddTool(mustTool("create_event_edge"))
	a.AddTool(mustTool("update_event_edge"))
	a.AddTool(mustTool("delete_event_edge"))

	// === 子事件 ===
	a.AddTool(mustTool("attach_sub_event"))
	a.AddTool(mustTool("move_sub_event"))
	a.AddTool(mustTool("detach_sub_event"))

	// === 实体 CRUD ===
	a.AddTool(mustTool("list_entities"))
	a.AddTool(mustTool("get_entity"))
	a.AddTool(mustTool("create_entity"))
	a.AddTool(mustTool("update_entity"))
	a.AddTool(mustTool("delete_entity"))

	// === 画布 / 编译 ===
	a.AddTool(mustTool("get_story_graph"))
	a.AddTool(mustTool("make_markdown"))
	a.AddTool(mustTool("make_mermaid"))

	// === 主链与保存 ===
	a.AddTool(mustTool("update_project_spine"))
	a.AddTool(mustTool("save_project"))

	return a
}

// NewDirectorTools 注册导演可用工具集
// 说明: 导演负责整体编排、校验、推演审查与正文生成管理, 可对工程做高阶管理
func NewDirectorTools() AITool {
	a := &agentTools{}
	a.tools = make(map[string]ToolDefinition)

	// === 调试 ===
	a.AddTool(TestToolDefinition())

	// === 工程读 ===
	a.AddTool(mustTool("list_projects"))
	a.AddTool(mustTool("load_project"))

	// === 工程管理 ===
	a.AddTool(mustTool("create_project"))
	a.AddTool(mustTool("update_project_meta"))
	a.AddTool(mustTool("delete_project"))
	a.AddTool(mustTool("update_project_spine"))

	// === 画布 / 编译 ===
	a.AddTool(mustTool("get_story_graph"))
	a.AddTool(mustTool("make_markdown"))
	a.AddTool(mustTool("make_mermaid"))

	// === 校验 ===
	a.AddTool(mustTool("validate_project"))

	// === 模拟 / 生成任务查询 ===
	a.AddTool(mustTool("get_simulation"))
	a.AddTool(mustTool("get_generation_task"))

	return a
}

// NewObserverTools 注册观察者可用工具集
// 说明: 观察者仅可读取数据与状态, 不允许任何修改/创建/删除操作
func NewObserverTools() AITool {
	a := &agentTools{}
	a.tools = make(map[string]ToolDefinition)

	// === 工程读 ===
	a.AddTool(mustTool("list_projects"))
	a.AddTool(mustTool("load_project"))

	// === 事件读 ===
	a.AddTool(mustTool("list_events"))
	a.AddTool(mustTool("get_event"))

	// === 实体读 ===
	a.AddTool(mustTool("list_entities"))
	a.AddTool(mustTool("get_entity"))

	// === 画布 / 编译 ===
	a.AddTool(mustTool("get_story_graph"))
	a.AddTool(mustTool("make_markdown"))
	a.AddTool(mustTool("make_mermaid"))

	// === 校验 ===
	a.AddTool(mustTool("validate_project"))

	// === 模拟 / 生成任务查询 ===
	a.AddTool(mustTool("get_simulation"))
	a.AddTool(mustTool("get_generation_task"))

	a.AddTool(TestToolDefinition())

	// === 工程读 ===
	a.AddTool(mustTool("list_projects"))
	a.AddTool(mustTool("load_project"))

	// === 工程管理 ===
	a.AddTool(mustTool("create_project"))
	a.AddTool(mustTool("update_project_meta"))
	a.AddTool(mustTool("delete_project"))
	a.AddTool(mustTool("update_project_spine"))

	// === 画布 / 编译 ===
	a.AddTool(mustTool("get_story_graph"))
	a.AddTool(mustTool("make_markdown"))
	a.AddTool(mustTool("make_mermaid"))

	// === 校验 ===
	a.AddTool(mustTool("validate_project"))

	// === 模拟 / 生成任务查询 ===
	a.AddTool(mustTool("get_simulation"))
	a.AddTool(mustTool("get_generation_task"))

	return a
}

// AllAgentTools 保留旧接口签名, 等同 NewAgentTools
func AllAgentTools() *agentTools {
	return NewAgentTools().(*agentTools)
}

// =====================================================================================
// 角色工具集全局实例
// 说明: 通过工厂函数构建, 调用方只能使用 AITool 接口能力, 不再直接访问内部 map
// =====================================================================================

// AgentTools 默认 Agent 工具集, 含全部工具, 供 StartAgent 等通用入口使用
var AgentTools AITool = NewAgentTools()

// 演员Tools 演员角色工具集
var 演员Tools AITool = NewActorTools()

// 导演Tools 导演角色工具集
var 导演Tools AITool = NewDirectorTools()

// 观察者Tools 观察者角色工具集
var 观察者Tools AITool = NewObserverTools()
