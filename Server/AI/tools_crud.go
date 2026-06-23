package ai

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"

	"AStoryForge/function/story_struct"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/sirupsen/logrus"
)

// =====================================================================================
// CRUD 工具参数 / 实现 / 定义
// 说明: 此文件包含所有 CRUD 操作的参数结构体、工具实现函数及工具定义注册。
//       按照工程管理 / 事件管理 / 事件边 / 子事件 / 实体管理 / 执行编译分类。
// =====================================================================================

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
// 工具定义
// =====================================================================================

// CRUDToolDefinitions 返回所有 CRUD 工具的 ToolDefinition 列表
// 说明: 供角色工厂注册, 与 ProposalToolDefinitions 分开以便按角色粒度控制权限。
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

// findTool 按名称从 CRUDToolDefinitions 或 ProposalToolDefinitions 中查找工具定义
// 说明: 供角色工厂按需注册, 避免重复书写 Schema
func findTool(name string) (ToolDefinition, bool) {
	for _, def := range CRUDToolDefinitions() {
		if def.Name == name {
			return def, true
		}
	}
	for _, def := range ProposalToolDefinitions() {
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
