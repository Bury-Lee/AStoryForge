package ai

import (
	"AStoryForge/function/story_struct"
	"context"
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/sirupsen/logrus"
)

//在这里写实际的各种业务逻辑

// TestToolArgs 测试工具参数
type TestToolArgs struct {
	Arg1 string `json:"arg1"`
	Arg2 int    `json:"arg2"`
	Arg3 bool   `json:"arg3"`
}

// 目标:调用这个函数
func TestTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logrus.Debugf("%+v\n", request)

	// 通过 JSON 序列化/反序列化
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
// 工程管理工具参数
// =====================================================================================

type rootDirArgs struct {
	RootDir string `json:"root_dir"`
}

type listProjectsArgs struct {
	RootDir string `json:"root_dir"`
}

type loadProjectArgs struct {
	RootDir string `json:"root_dir"`
}

type createProjectArgs struct {
	Meta story_struct.ProjectMeta `json:"meta"`
}

type deleteProjectArgs struct {
	RootDir   string `json:"root_dir"`
	ProjectID string `json:"project_id"`
}

type updateProjectMetaArgs struct {
	RootDir   string                   `json:"root_dir"`
	ProjectID string                   `json:"project_id"`
	Meta      story_struct.ProjectMeta `json:"meta"`
}

type updateProjectSpineArgs struct {
	RootDir   string   `json:"root_dir"`
	ProjectID string   `json:"project_id"`
	Spine     []string `json:"spine"`
}

type saveProjectArgs struct {
	RootDir string                `json:"root_dir"`
	Project *story_struct.Project `json:"project"`
}

// =====================================================================================
// 事件管理工具参数
// =====================================================================================

type listEventsArgs struct {
	RootDir   string `json:"root_dir"`
	ProjectID string `json:"project_id"`
}

type getEventArgs struct {
	RootDir   string `json:"root_dir"`
	ProjectID string `json:"project_id"`
	EventID   string `json:"event_id"`
}

type createEventArgs struct {
	RootDir   string             `json:"root_dir"`
	ProjectID string             `json:"project_id"`
	Event     story_struct.Event `json:"event"`
}

type updateEventArgs struct {
	RootDir   string             `json:"root_dir"`
	ProjectID string             `json:"project_id"`
	EventID   string             `json:"event_id"`
	Event     story_struct.Event `json:"event"`
}

type deleteEventArgs struct {
	RootDir   string `json:"root_dir"`
	ProjectID string `json:"project_id"`
	EventID   string `json:"event_id"`
}

type setEventLockedArgs struct {
	RootDir   string `json:"root_dir"`
	ProjectID string `json:"project_id"`
	EventID   string `json:"event_id"`
	Locked    bool   `json:"locked"`
}

type updateEventParticipantsArgs struct {
	RootDir      string                     `json:"root_dir"`
	ProjectID    string                     `json:"project_id"`
	EventID      string                     `json:"event_id"`
	Participants []story_struct.Participant `json:"participants"`
}

// =====================================================================================
// 事件边 / 子事件工具参数
// =====================================================================================

type eventEdgeArgs struct {
	RootDir       string                `json:"root_dir"`
	ProjectID     string                `json:"project_id"`
	SourceEventID string                `json:"source_event_id"`
	TargetEventID string                `json:"target_event_id"`
	EdgeType      story_struct.EdgeType `json:"edge_type"`
	MirrorInEdge  bool                  `json:"mirror_in_edge"`
}

type updateEventEdgeArgs struct {
	RootDir          string                `json:"root_dir"`
	ProjectID        string                `json:"project_id"`
	SourceEventID    string                `json:"source_event_id"`
	TargetEventID    string                `json:"target_event_id"`
	NewTargetEventID string                `json:"new_target_event_id"`
	EdgeType         story_struct.EdgeType `json:"edge_type"`
	MirrorInEdge     bool                  `json:"mirror_in_edge"`
}

type subEventArgs struct {
	RootDir       string `json:"root_dir"`
	ProjectID     string `json:"project_id"`
	ParentEventID string `json:"parent_event_id"`
	ChildEventID  string `json:"child_event_id"`
	Index         int    `json:"index"`
}

// =====================================================================================
// 实体管理工具参数
// =====================================================================================

type listEntitiesArgs struct {
	RootDir   string `json:"root_dir"`
	ProjectID string `json:"project_id"`
}

type getEntityArgs struct {
	RootDir   string `json:"root_dir"`
	ProjectID string `json:"project_id"`
	EntityID  string `json:"entity_id"`
}

type createEntityArgs struct {
	RootDir   string              `json:"root_dir"`
	ProjectID string              `json:"project_id"`
	Entity    story_struct.Entity `json:"entity"`
}

type updateEntityArgs struct {
	RootDir   string              `json:"root_dir"`
	ProjectID string              `json:"project_id"`
	EntityID  string              `json:"entity_id"`
	Entity    story_struct.Entity `json:"entity"`
}

type deleteEntityArgs struct {
	RootDir   string `json:"root_dir"`
	ProjectID string `json:"project_id"`
	EntityID  string `json:"entity_id"`
}

// =====================================================================================
// 执行 / 编译 / 模拟 / 生成工具参数
// =====================================================================================

type getStoryGraphArgs struct {
	RootDir   string `json:"root_dir"`
	ProjectID string `json:"project_id"`
}

type validateProjectArgs struct {
	RootDir   string `json:"root_dir"`
	ProjectID string `json:"project_id"`
}

type compileArgs struct {
	RootDir      string   `json:"root_dir"`
	ProjectID    string   `json:"project_id"`
	EventIDs     []string `json:"event_ids"`
	OutputPath   string   `json:"output_path"`
	ValidateOnly bool     `json:"validate_only"`
}

type getSimulationArgs struct {
	SessionID string `json:"session_id"`
}

type getGenerationTaskArgs struct {
	TaskID string `json:"task_id"`
}

// =====================================================================================
// 工程管理工具
// =====================================================================================

func ListProjectsTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logrus.Debugf("调用 list_projects %+v\n", request)
	var args listProjectsArgs
	if err := bindArgs(request, &args); err != nil {
		return nil, err
	}
	list, err := crudService.ListProjects(args.RootDir)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(formatJSON(list)), nil
}

func CreateProjectTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logrus.Debugf("调用 create_project %+v\n", request)
	var args createProjectArgs
	if err := bindArgs(request, &args); err != nil {
		return nil, err
	}
	project, err := crudService.CreateProject(args.Meta)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(formatJSON(project)), nil
}

func LoadProjectTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logrus.Debugf("调用 load_project %+v\n", request)
	var args loadProjectArgs
	if err := bindArgs(request, &args); err != nil {
		return nil, err
	}
	project, err := crudService.LoadProject(args.RootDir)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(formatJSON(project)), nil
}

func DeleteProjectTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logrus.Debugf("调用 delete_project %+v\n", request)
	var args deleteProjectArgs
	if err := bindArgs(request, &args); err != nil {
		return nil, err
	}
	if err := crudService.DeleteProject(args.RootDir, args.ProjectID); err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(fmt.Sprintf("工程已删除: %s", resolveTarget(args.RootDir, args.ProjectID))), nil
}

func UpdateProjectMetaTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logrus.Debugf("调用 update_project_meta %+v\n", request)
	var args updateProjectMetaArgs
	if err := bindArgs(request, &args); err != nil {
		return nil, err
	}
	meta, err := crudService.UpdateProjectMeta(args.RootDir, args.ProjectID, args.Meta)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(formatJSON(meta)), nil
}

func UpdateProjectSpineTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logrus.Debugf("调用 update_project_spine %+v\n", request)
	var args updateProjectSpineArgs
	if err := bindArgs(request, &args); err != nil {
		return nil, err
	}
	spine, err := crudService.UpdateProjectSpine(args.RootDir, args.ProjectID, args.Spine)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(formatJSON(spine)), nil
}

func SaveProjectTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logrus.Debugf("调用 save_project %+v\n", request)
	var args saveProjectArgs
	if err := bindArgs(request, &args); err != nil {
		return nil, err
	}
	if err := crudService.SaveProject(args.RootDir, args.Project); err != nil {
		return nil, err
	}
	return mcp.NewToolResultText("工程保存成功"), nil
}

// =====================================================================================
// 事件管理工具
// =====================================================================================

func ListEventsTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logrus.Debugf("调用 list_events %+v\n", request)
	var args listEventsArgs
	if err := bindArgs(request, &args); err != nil {
		return nil, err
	}
	list, err := crudService.ListEvents(args.RootDir, args.ProjectID)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(formatJSON(list)), nil
}

func GetEventTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logrus.Debugf("调用 get_event %+v\n", request)
	var args getEventArgs
	if err := bindArgs(request, &args); err != nil {
		return nil, err
	}
	ev, err := crudService.GetEvent(args.RootDir, args.ProjectID, args.EventID)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(formatJSON(ev)), nil
}

func CreateEventTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logrus.Debugf("调用 create_event %+v\n", request)
	var args createEventArgs
	if err := bindArgs(request, &args); err != nil {
		return nil, err
	}
	ev, err := crudService.CreateEvent(args.RootDir, args.ProjectID, args.Event)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(formatJSON(ev)), nil
}

func UpdateEventTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logrus.Debugf("调用 update_event %+v\n", request)
	var args updateEventArgs
	if err := bindArgs(request, &args); err != nil {
		return nil, err
	}
	ev, err := crudService.UpdateEvent(args.RootDir, args.ProjectID, args.EventID, args.Event)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(formatJSON(ev)), nil
}

func DeleteEventTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logrus.Debugf("调用 delete_event %+v\n", request)
	var args deleteEventArgs
	if err := bindArgs(request, &args); err != nil {
		return nil, err
	}
	if err := crudService.DeleteEvent(args.RootDir, args.ProjectID, args.EventID); err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(fmt.Sprintf("事件已删除: %s", args.EventID)), nil
}

func SetEventLockedTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logrus.Debugf("调用 set_event_locked %+v\n", request)
	var args setEventLockedArgs
	if err := bindArgs(request, &args); err != nil {
		return nil, err
	}
	ev, err := crudService.SetEventLocked(args.RootDir, args.ProjectID, args.EventID, args.Locked)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(formatJSON(ev)), nil
}

func UpdateEventParticipantsTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logrus.Debugf("调用 update_event_participants %+v\n", request)
	var args updateEventParticipantsArgs
	if err := bindArgs(request, &args); err != nil {
		return nil, err
	}
	ev, err := crudService.UpdateEventParticipants(args.RootDir, args.ProjectID, args.EventID, args.Participants)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(formatJSON(ev)), nil
}

// =====================================================================================
// 事件边工具
// =====================================================================================

func CreateEventEdgeTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logrus.Debugf("调用 create_event_edge %+v\n", request)
	var args eventEdgeArgs
	if err := bindArgs(request, &args); err != nil {
		return nil, err
	}
	if err := crudService.CreateEventEdge(args.RootDir, args.ProjectID, args.SourceEventID, args.TargetEventID, args.EdgeType, args.MirrorInEdge); err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(fmt.Sprintf("事件边已创建: %s -%s-> %s", args.SourceEventID, args.EdgeType, args.TargetEventID)), nil
}

func UpdateEventEdgeTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logrus.Debugf("调用 update_event_edge %+v\n", request)
	var args updateEventEdgeArgs
	if err := bindArgs(request, &args); err != nil {
		return nil, err
	}
	if err := crudService.UpdateEventEdge(args.RootDir, args.ProjectID, args.SourceEventID, args.TargetEventID, args.NewTargetEventID, args.EdgeType, args.MirrorInEdge); err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(fmt.Sprintf("事件边已更新: %s -%s-> %s", args.SourceEventID, args.EdgeType, args.NewTargetEventID)), nil
}

func DeleteEventEdgeTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logrus.Debugf("调用 delete_event_edge %+v\n", request)
	var args eventEdgeArgs
	if err := bindArgs(request, &args); err != nil {
		return nil, err
	}
	if err := crudService.DeleteEventEdge(args.RootDir, args.ProjectID, args.SourceEventID, args.TargetEventID, args.EdgeType, args.MirrorInEdge); err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(fmt.Sprintf("事件边已删除: %s -%s-> %s", args.SourceEventID, args.EdgeType, args.TargetEventID)), nil
}

// =====================================================================================
// 子事件工具
// =====================================================================================

func AttachSubEventTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logrus.Debugf("调用 attach_sub_event %+v\n", request)
	var args subEventArgs
	if err := bindArgs(request, &args); err != nil {
		return nil, err
	}
	if err := crudService.AttachSubEvent(args.RootDir, args.ProjectID, args.ParentEventID, args.ChildEventID, args.Index); err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(fmt.Sprintf("子事件已挂载: %s -> %s", args.ChildEventID, args.ParentEventID)), nil
}

func MoveSubEventTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logrus.Debugf("调用 move_sub_event %+v\n", request)
	var args subEventArgs
	if err := bindArgs(request, &args); err != nil {
		return nil, err
	}
	if err := crudService.MoveSubEvent(args.RootDir, args.ProjectID, args.ParentEventID, args.ChildEventID, args.Index); err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(fmt.Sprintf("子事件已移动: %s -> index %d", args.ChildEventID, args.Index)), nil
}

func DetachSubEventTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logrus.Debugf("调用 detach_sub_event %+v\n", request)
	var args subEventArgs
	if err := bindArgs(request, &args); err != nil {
		return nil, err
	}
	if err := crudService.DetachSubEvent(args.RootDir, args.ProjectID, args.ParentEventID, args.ChildEventID); err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(fmt.Sprintf("子事件已解除: %s", args.ChildEventID)), nil
}

// =====================================================================================
// 实体管理工具
// =====================================================================================

func ListEntitiesTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logrus.Debugf("调用 list_entities %+v\n", request)
	var args listEntitiesArgs
	if err := bindArgs(request, &args); err != nil {
		return nil, err
	}
	list, err := crudService.ListEntities(args.RootDir, args.ProjectID)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(formatJSON(list)), nil
}

func GetEntityTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logrus.Debugf("调用 get_entity %+v\n", request)
	var args getEntityArgs
	if err := bindArgs(request, &args); err != nil {
		return nil, err
	}
	ent, err := crudService.GetEntity(args.RootDir, args.ProjectID, args.EntityID)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(formatJSON(ent)), nil
}

func CreateEntityTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logrus.Debugf("调用 create_entity %+v\n", request)
	var args createEntityArgs
	if err := bindArgs(request, &args); err != nil {
		return nil, err
	}
	ent, err := crudService.CreateEntity(args.RootDir, args.ProjectID, args.Entity)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(formatJSON(ent)), nil
}

func UpdateEntityTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logrus.Debugf("调用 update_entity %+v\n", request)
	var args updateEntityArgs
	if err := bindArgs(request, &args); err != nil {
		return nil, err
	}
	ent, err := crudService.UpdateEntity(args.RootDir, args.ProjectID, args.EntityID, args.Entity)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(formatJSON(ent)), nil
}

func DeleteEntityTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logrus.Debugf("调用 delete_entity %+v\n", request)
	var args deleteEntityArgs
	if err := bindArgs(request, &args); err != nil {
		return nil, err
	}
	if err := crudService.DeleteEntity(args.RootDir, args.ProjectID, args.EntityID); err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(fmt.Sprintf("实体已删除: %s", args.EntityID)), nil
}

// =====================================================================================
// 执行 / 编译 / 校验工具
// =====================================================================================

func GetStoryGraphTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logrus.Debugf("调用 get_story_graph %+v\n", request)
	var args getStoryGraphArgs
	if err := bindArgs(request, &args); err != nil {
		return nil, err
	}
	graph, err := crudService.GetStoryGraph(args.RootDir, args.ProjectID)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(formatJSON(graph)), nil
}

func ValidateProjectTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logrus.Debugf("调用 validate_project %+v\n", request)
	var args validateProjectArgs
	if err := bindArgs(request, &args); err != nil {
		return nil, err
	}
	report, err := crudService.ValidateProject(args.RootDir, args.ProjectID)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(formatJSON(report)), nil
}

// MakeMarkdownTool 生成 Markdown 大纲
// 说明: 调用 ProjectObj.MakeMarkdown 内部方法, 加载工程并写入指定输出路径
func MakeMarkdownTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logrus.Debugf("调用 make_markdown %+v\n", request)
	var args compileArgs
	if err := bindArgs(request, &args); err != nil {
		return nil, err
	}

	target := resolveTarget(args.RootDir, args.ProjectID)
	loaded := &story_struct.ProjectObj{}
	if err := loaded.LoadProjectObj(target); err != nil {
		return nil, fmt.Errorf("加载工程失败: %w", err)
	}

	md, err := loaded.MakeMarkdown()
	if err != nil {
		return nil, fmt.Errorf("生成 Markdown 失败: %w", err)
	}

	outputPath := args.OutputPath
	if outputPath == "" {
		outputPath = filepath.Join(filepath.Dir(target), strings.TrimSuffix(filepath.Base(target), filepath.Ext(target))+".md")
	}
	res := &story_struct.CompileResult{
		ProjectID:  args.ProjectID,
		OutputPath: outputPath,
		Markdown:   md,
	}
	return mcp.NewToolResultText(formatJSON(res)), nil
}

// MakeMermaidTool 生成 Mermaid 图
// 说明: 调用 ProjectObj.MakeMermaid 内部方法, 加载工程并写入指定输出路径
func MakeMermaidTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logrus.Debugf("调用 make_mermaid %+v\n", request)
	var args compileArgs
	if err := bindArgs(request, &args); err != nil {
		return nil, err
	}

	target := resolveTarget(args.RootDir, args.ProjectID)
	loaded := &story_struct.ProjectObj{}
	if err := loaded.LoadProjectObj(target); err != nil {
		return nil, fmt.Errorf("加载工程失败: %w", err)
	}

	mermaid, err := loaded.MakeMermaid()
	if err != nil {
		return nil, fmt.Errorf("生成 Mermaid 失败: %w", err)
	}

	outputPath := args.OutputPath
	if outputPath == "" {
		outputPath = filepath.Join(filepath.Dir(target), strings.TrimSuffix(filepath.Base(target), filepath.Ext(target))+".mermaid")
	}
	res := &story_struct.CompileResult{
		ProjectID:  args.ProjectID,
		OutputPath: outputPath,
		Markdown:   mermaid,
	}
	return mcp.NewToolResultText(formatJSON(res)), nil
}

// =====================================================================================
// 模拟 / 生成任务查询工具
// =====================================================================================

func GetSimulationTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logrus.Debugf("调用 get_simulation %+v\n", request)
	var args getSimulationArgs
	if err := bindArgs(request, &args); err != nil {
		return nil, err
	}
	session, err := crudService.GetSimulation(args.SessionID)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(formatJSON(session)), nil
}

func GetGenerationTaskTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logrus.Debugf("调用 get_generation_task %+v\n", request)
	var args getGenerationTaskArgs
	if err := bindArgs(request, &args); err != nil {
		return nil, err
	}
	task, err := crudService.GetGenerationTask(args.TaskID)
	if err != nil {
		return nil, err
	}
	return mcp.NewToolResultText(formatJSON(task)), nil
}

// =====================================================================================
// 工具 Schema 与注册信息
// 说明: 此处集中管理所有工具的 JSON Schema 定义, 供 AllAgentTools 注册时使用。
//       复杂类型(事件/实体/工程)仅描述顶层字段, 内部嵌套结构以 story_struct 定义为准。
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

// CRUDToolDefinitions 返回所有 CRUD 工具的 ToolDefinition 列表
// 说明: 供 AllAgentTools 调用, 避免 tools.go 与 function.go 强耦合
func CRUDToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		// ---------- 工程管理 ----------
		{
			Name:        "list_projects",
			Description: "扫描指定目录, 列出所有符合工程元信息格式的 project.json",
			Parameters: objectProp("入参", map[string]jsonschema.Definition{
				"root_dir": stringProp("工程根目录路径"),
			}, []string{"root_dir"}),
			Executor: ListProjectsTool,
		},
		{
			Name:        "create_project",
			Description: "创建一个新工程, 初始化目录与默认文件",
			Parameters: objectProp("入参", map[string]jsonschema.Definition{
				"meta": projectMetaSchema(),
			}, []string{"meta"}),
			Executor: CreateProjectTool,
		},
		{
			Name:        "load_project",
			Description: "加载完整工程数据(元信息、世界设置、事件列表、实体列表)",
			Parameters: objectProp("入参", map[string]jsonschema.Definition{
				"root_dir": stringProp("工程根目录路径"),
			}, []string{"root_dir"}),
			Executor: LoadProjectTool,
		},
		{
			Name:        "delete_project",
			Description: "删除整个工程目录(谨慎操作, 当前为预备功能)",
			Parameters: objectProp("入参", map[string]jsonschema.Definition{
				"root_dir":   stringProp("工程根目录路径"),
				"project_id": stringProp("工程 ID, 与 root_dir 二选一"),
			}, nil),
			Executor: DeleteProjectTool,
		},
		{
			Name:        "update_project_meta",
			Description: "更新工程元信息(标题、作者、设置、统计计数等)",
			Parameters: objectProp("入参", map[string]jsonschema.Definition{
				"root_dir":   stringProp("工程根目录路径"),
				"project_id": stringProp("工程 ID"),
				"meta":       projectMetaSchema(),
			}, []string{"meta"}),
			Executor: UpdateProjectMetaTool,
		},
		{
			Name:        "update_project_spine",
			Description: "更新叙事链顺序, 同步写入工程的世界设置",
			Parameters: objectProp("入参", map[string]jsonschema.Definition{
				"root_dir":   stringProp("工程根目录路径"),
				"project_id": stringProp("工程 ID"),
				"spine":      arrayProp("叙事链上的事件 ID 列表, 顺序即叙事顺序", stringProp("事件 ID")),
			}, []string{"spine"}),
			Executor: UpdateProjectSpineTool,
		},
		{
			Name:        "save_project",
			Description: "保存完整工程, 适合批量修改后的统一提交",
			Parameters: objectProp("入参", map[string]jsonschema.Definition{
				"root_dir": stringProp("工程根目录路径"),
				"project":  objectProp("完整工程数据, 字段见 story_struct.Project", nil, nil),
			}, []string{"project"}),
			Executor: SaveProjectTool,
		},

		// ---------- 事件管理 ----------
		{
			Name:        "list_events",
			Description: "获取工程下全部事件列表, 按 ID 排序",
			Parameters: objectProp("入参", map[string]jsonschema.Definition{
				"root_dir":   stringProp("工程根目录路径"),
				"project_id": stringProp("工程 ID"),
			}, nil),
			Executor: ListEventsTool,
		},
		{
			Name:        "get_event",
			Description: "获取单个事件详情",
			Parameters: objectProp("入参", map[string]jsonschema.Definition{
				"root_dir":   stringProp("工程根目录路径"),
				"project_id": stringProp("工程 ID"),
				"event_id":   stringProp("事件 ID"),
			}, []string{"event_id"}),
			Executor: GetEventTool,
		},
		{
			Name:        "create_event",
			Description: "创建事件节点, 支持普通事件、容器事件、混合事件",
			Parameters: objectProp("入参", map[string]jsonschema.Definition{
				"root_dir":   stringProp("工程根目录路径"),
				"project_id": stringProp("工程 ID"),
				"event":      eventSchema(),
			}, []string{"event"}),
			Executor: CreateEventTool,
		},
		{
			Name:        "update_event",
			Description: "更新事件节点的全部字段(ID 保留)",
			Parameters: objectProp("入参", map[string]jsonschema.Definition{
				"root_dir":   stringProp("工程根目录路径"),
				"project_id": stringProp("工程 ID"),
				"event_id":   stringProp("原事件 ID"),
				"event":      eventSchema(),
			}, []string{"event_id", "event"}),
			Executor: UpdateEventTool,
		},
		{
			Name:        "delete_event",
			Description: "删除事件节点, 自动清理主链、父子关系、因果边与参与者引用",
			Parameters: objectProp("入参", map[string]jsonschema.Definition{
				"root_dir":   stringProp("工程根目录路径"),
				"project_id": stringProp("工程 ID"),
				"event_id":   stringProp("事件 ID"),
			}, []string{"event_id"}),
			Executor: DeleteEventTool,
		},
		{
			Name:        "set_event_locked",
			Description: "设置事件的锁定状态, 锁定后 AI 不应自动改写",
			Parameters: objectProp("入参", map[string]jsonschema.Definition{
				"root_dir":   stringProp("工程根目录路径"),
				"project_id": stringProp("工程 ID"),
				"event_id":   stringProp("事件 ID"),
				"locked":     boolProp("锁定标记"),
			}, []string{"event_id", "locked"}),
			Executor: SetEventLockedTool,
		},
		{
			Name:        "update_event_participants",
			Description: "更新事件的参与者列表, 自动同步实体的 Events 索引",
			Parameters: objectProp("入参", map[string]jsonschema.Definition{
				"root_dir":   stringProp("工程根目录路径"),
				"project_id": stringProp("工程 ID"),
				"event_id":   stringProp("事件 ID"),
				"participants": arrayProp("参与者列表", objectProp("参与者", map[string]jsonschema.Definition{
					"entity_id": stringProp("实体 ID"),
					"state":     objectProp("入场状态快照", map[string]jsonschema.Definition{}, nil),
				}, []string{"entity_id"})),
			}, []string{"event_id", "participants"}),
			Executor: UpdateEventParticipantsTool,
		},

		// ---------- 事件边 ----------
		{
			Name:        "create_event_edge",
			Description: "创建事件因果边(画布拉线建边), 可选同步回填目标入边",
			Parameters: objectProp("入参", map[string]jsonschema.Definition{
				"root_dir":        stringProp("工程根目录路径"),
				"project_id":      stringProp("工程 ID"),
				"source_event_id": stringProp("起点事件 ID"),
				"target_event_id": stringProp("终点事件 ID"),
				"edge_type":       stringProp("边类型: cause | result"),
				"mirror_in_edge":  boolProp("是否回填入边"),
			}, []string{"source_event_id", "target_event_id", "edge_type"}),
			Executor: CreateEventEdgeTool,
		},
		{
			Name:        "update_event_edge",
			Description: "更新事件因果边(改线), 切换 cause/result 或重指向, 保持图结构一致",
			Parameters: objectProp("入参", map[string]jsonschema.Definition{
				"root_dir":            stringProp("工程根目录路径"),
				"project_id":          stringProp("工程 ID"),
				"source_event_id":     stringProp("原起点事件 ID"),
				"target_event_id":     stringProp("原终点事件 ID"),
				"new_target_event_id": stringProp("新终点事件 ID"),
				"edge_type":           stringProp("更新后的边类型: cause | result"),
				"mirror_in_edge":      boolProp("是否同步更新入边"),
			}, []string{"source_event_id", "target_event_id", "new_target_event_id", "edge_type"}),
			Executor: UpdateEventEdgeTool,
		},
		{
			Name:        "delete_event_edge",
			Description: "删除事件因果边(画布断线), 可选同步移除目标入边",
			Parameters: objectProp("入参", map[string]jsonschema.Definition{
				"root_dir":        stringProp("工程根目录路径"),
				"project_id":      stringProp("工程 ID"),
				"source_event_id": stringProp("起点事件 ID"),
				"target_event_id": stringProp("终点事件 ID"),
				"edge_type":       stringProp("边类型: cause | result"),
				"mirror_in_edge":  boolProp("是否同步删除入边"),
			}, []string{"source_event_id", "target_event_id", "edge_type"}),
			Executor: DeleteEventEdgeTool,
		},

		// ---------- 子事件 ----------
		{
			Name:        "attach_sub_event",
			Description: "建立父子包含关系(把子事件挂入卷/章/mixed 节点)",
			Parameters: objectProp("入参", map[string]jsonschema.Definition{
				"root_dir":        stringProp("工程根目录路径"),
				"project_id":      stringProp("工程 ID"),
				"parent_event_id": stringProp("父事件 ID"),
				"child_event_id":  stringProp("子事件 ID"),
				"index":           intProp("插入位置, -1 或越界则追加到末尾"),
			}, []string{"parent_event_id", "child_event_id"}),
			Executor: AttachSubEventTool,
		},
		{
			Name:        "move_sub_event",
			Description: "调整子事件在父节点中的顺序",
			Parameters: objectProp("入参", map[string]jsonschema.Definition{
				"root_dir":        stringProp("工程根目录路径"),
				"project_id":      stringProp("工程 ID"),
				"parent_event_id": stringProp("父事件 ID"),
				"child_event_id":  stringProp("子事件 ID"),
				"index":           intProp("新位置"),
			}, []string{"parent_event_id", "child_event_id", "index"}),
			Executor: MoveSubEventTool,
		},
		{
			Name:        "detach_sub_event",
			Description: "解除父子包含关系, 不删除子事件实体本身",
			Parameters: objectProp("入参", map[string]jsonschema.Definition{
				"root_dir":        stringProp("工程根目录路径"),
				"project_id":      stringProp("工程 ID"),
				"parent_event_id": stringProp("父事件 ID"),
				"child_event_id":  stringProp("子事件 ID"),
			}, []string{"parent_event_id", "child_event_id"}),
			Executor: DetachSubEventTool,
		},

		// ---------- 实体管理 ----------
		{
			Name:        "list_entities",
			Description: "获取工程下全部实体列表, 按 ID 排序",
			Parameters: objectProp("入参", map[string]jsonschema.Definition{
				"root_dir":   stringProp("工程根目录路径"),
				"project_id": stringProp("工程 ID"),
			}, nil),
			Executor: ListEntitiesTool,
		},
		{
			Name:        "get_entity",
			Description: "获取单个实体详情",
			Parameters: objectProp("入参", map[string]jsonschema.Definition{
				"root_dir":   stringProp("工程根目录路径"),
				"project_id": stringProp("工程 ID"),
				"entity_id":  stringProp("实体 ID"),
			}, []string{"entity_id"}),
			Executor: GetEntityTool,
		},
		{
			Name:        "create_entity",
			Description: "创建实体节点(角色、组织、物品、概念等)",
			Parameters: objectProp("入参", map[string]jsonschema.Definition{
				"root_dir":   stringProp("工程根目录路径"),
				"project_id": stringProp("工程 ID"),
				"entity":     entitySchema(),
			}, []string{"entity"}),
			Executor: CreateEntityTool,
		},
		{
			Name:        "update_entity",
			Description: "更新实体节点的全部字段(ID 保留)",
			Parameters: objectProp("入参", map[string]jsonschema.Definition{
				"root_dir":   stringProp("工程根目录路径"),
				"project_id": stringProp("工程 ID"),
				"entity_id":  stringProp("原实体 ID"),
				"entity":     entitySchema(),
			}, []string{"entity_id", "entity"}),
			Executor: UpdateEntityTool,
		},
		{
			Name:        "delete_entity",
			Description: "删除实体节点, 自动清理参与者引用与实体关系",
			Parameters: objectProp("入参", map[string]jsonschema.Definition{
				"root_dir":   stringProp("工程根目录路径"),
				"project_id": stringProp("工程 ID"),
				"entity_id":  stringProp("实体 ID"),
			}, []string{"entity_id"}),
			Executor: DeleteEntityTool,
		},

		// ---------- 执行 / 校验 / 编译 ----------
		{
			Name:        "get_story_graph",
			Description: "获取画布图数据, 同时返回叙事链、事件列表、实体列表",
			Parameters: objectProp("入参", map[string]jsonschema.Definition{
				"root_dir":   stringProp("工程根目录路径"),
				"project_id": stringProp("工程 ID"),
			}, nil),
			Executor: GetStoryGraphTool,
		},
		{
			Name:        "validate_project",
			Description: "校验工程结构(主链缺失、循环引用、悬空边、参与者引用等)",
			Parameters: objectProp("入参", map[string]jsonschema.Definition{
				"root_dir":   stringProp("工程根目录路径"),
				"project_id": stringProp("工程 ID"),
			}, nil),
			Executor: ValidateProjectTool,
		},
		{
			Name:        "make_markdown",
			Description: "将工程编译为结构化 Markdown 大纲, 返回内容与建议输出路径",
			Parameters: objectProp("入参", map[string]jsonschema.Definition{
				"root_dir":      stringProp("工程根目录路径"),
				"project_id":    stringProp("工程 ID"),
				"event_ids":     arrayProp("指定编译的事件 ID 列表, 为空表示全部", stringProp("事件 ID")),
				"output_path":   stringProp("输出文件路径, 为空则自动生成"),
				"validate_only": boolProp("是否仅校验, 当前未使用, 占位"),
			}, nil),
			Executor: MakeMarkdownTool,
		},
		{
			Name:        "make_mermaid",
			Description: "将工程编译为 Mermaid 图, 用于表达事件关系和流程",
			Parameters: objectProp("入参", map[string]jsonschema.Definition{
				"root_dir":      stringProp("工程根目录路径"),
				"project_id":    stringProp("工程 ID"),
				"event_ids":     arrayProp("指定编译的事件 ID 列表, 为空表示全部", stringProp("事件 ID")),
				"output_path":   stringProp("输出文件路径, 为空则自动生成"),
				"validate_only": boolProp("是否仅校验, 当前未使用, 占位"),
			}, nil),
			Executor: MakeMermaidTool,
		},

		// ---------- 模拟 / 生成任务查询 ----------
		{
			Name:        "get_simulation",
			Description: "查询指定模拟会话的状态与步骤数据",
			Parameters: objectProp("入参", map[string]jsonschema.Definition{
				"session_id": stringProp("模拟会话 ID"),
			}, []string{"session_id"}),
			Executor: GetSimulationTool,
		},
		{
			Name:        "get_generation_task",
			Description: "查询指定正文生成任务的状态与内容",
			Parameters: objectProp("入参", map[string]jsonschema.Definition{
				"task_id": stringProp("生成任务 ID"),
			}, []string{"task_id"}),
			Executor: GetGenerationTaskTool,
		},
	}
}

// findTool 按名称从 CRUDToolDefinitions 中查找工具定义
// 说明: 供角色工厂按需注册, 避免重复书写 Schema
func findTool(name string) (ToolDefinition, bool) {
	for _, def := range CRUDToolDefinitions() {
		if def.Name == name {
			return def, true
		}
	}
	return ToolDefinition{}, false
}

// mustTool 按名称获取工具定义, 未找到时 panic(开发期错误, 不应在运行期出现)
func mustTool(name string) ToolDefinition {
	def, ok := findTool(name)
	if !ok {
		panic("AI 工具未注册: " + name)
	}
	return def
}

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
