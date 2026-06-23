package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"AStoryForge/function/story_struct"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sashabaranov/go-openai/jsonschema"
	"github.com/sirupsen/logrus"
)

// =====================================================================================
// 申请审查与执行工具（观察者/裁判专用）
// 说明: 观察者 Agent 通过 review_proposal 审查演员提交的行为申请,
//       批准后工具内联执行实际修改并将结果写回 proposal.Result。
//       设计理由: 将"审批"和"执行"绑定为一个原子操作, 避免以下竞态:
//       1) 批准后忘了执行 2) 执行时工程已被其他操作改变
//       内部分发器 (executeProposal) 映射 action_type 到 crudService 的对应方法,
//       复用了 story_struct 层已有的写逻辑, 无需重复实现。
// =====================================================================================

// reviewProposalArgs 审查行为申请参数
type reviewProposalArgs struct {
	ProposalID string `json:"proposal_id"`
	Approved   bool   `json:"approved"`
	ReviewNote string `json:"review_note,omitempty"`
}

// ReviewProposalTool 审查行为申请
// 说明: 观察者(裁判)调用此工具审查 pending 状态的申请。
//       approved=true 时自动执行实际修改并写回工程文件。
func ReviewProposalTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logrus.Debugf("调用 review_proposal %+v\n", request)
	var args reviewProposalArgs
	if err := bindArgs(request, &args); err != nil {
		return nil, err
	}
	proposalMu.Lock()
	p, ok := proposalStore[args.ProposalID]
	if !ok {
		proposalMu.Unlock()
		return nil, fmt.Errorf("行为申请不存在: %s", args.ProposalID)
	}
	if p.Status != ProposalStatusPending {
		proposalMu.Unlock()
		return nil, fmt.Errorf("只能审查待审批的申请, 当前状态: %s", p.Status)
	}

	now := time.Now().Format(time.RFC3339)
	p.ReviewNote = args.ReviewNote
	p.ReviewedAt = now

	if args.Approved {
		// 执行实际修改
		resultJSON, errStr := executeProposal(p)
		if errStr != "" {
			p.Status = ProposalStatusRejected
			p.Error = errStr
			p.ReviewedAt = time.Now().Format(time.RFC3339)
			proposalStore[args.ProposalID] = p
			proposalMu.Unlock()
			return mcp.NewToolResultText(formatJSON(p)), nil
		}
		p.Status = ProposalStatusApproved
		p.Result = resultJSON
	} else {
		p.Status = ProposalStatusRejected
	}

	proposalStore[args.ProposalID] = p
	proposalMu.Unlock()
	return mcp.NewToolResultText(formatJSON(p)), nil
}

// executeProposal 执行已批准的申请
// 说明: 根据 action_type 分发到 crudService 的对应写方法。
//       所有写方法复用 story_struct 层已有的加载-修改-保存逻辑。
//       返回值: resultJSON - 执行结果(可为空), errStr - 错误信息(空表示成功)
func executeProposal(p *ActionProposal) (json.RawMessage, string) {
	// 统一解析 payload 中的 rootDir 和 projectID (大部分 CRUD 方法都需要)
	var base struct {
		RootDir   string `json:"root_dir"`
		ProjectID string `json:"project_id"`
	}
	if err := json.Unmarshal(p.Payload, &base); err != nil {
		return nil, fmt.Sprintf("解析 payload 基础字段失败: %v", err)
	}
	rootDir := base.RootDir
	projectID := base.ProjectID

	switch p.ActionType {

	// ========== 工程管理 ==========
	case "create_project":
		var meta struct {
			Meta story_struct.ProjectMeta `json:"meta"`
		}
		if err := json.Unmarshal(p.Payload, &meta); err != nil {
			return nil, fmt.Sprintf("解析 create_project 参数失败: %v", err)
		}
		project, err := crudService.CreateProject(meta.Meta)
		if err != nil {
			return nil, err.Error()
		}
		b, _ := json.Marshal(project)
		return b, ""

	case "update_project_meta":
		var meta struct {
			Meta story_struct.ProjectMeta `json:"meta"`
		}
		if err := json.Unmarshal(p.Payload, &meta); err != nil {
			return nil, fmt.Sprintf("解析 update_project_meta 参数失败: %v", err)
		}
		result, err := crudService.UpdateProjectMeta(rootDir, projectID, meta.Meta)
		if err != nil {
			return nil, err.Error()
		}
		b, _ := json.Marshal(result)
		return b, ""

	case "update_project_spine":
		var spine struct {
			Spine []string `json:"spine"`
		}
		if err := json.Unmarshal(p.Payload, &spine); err != nil {
			return nil, fmt.Sprintf("解析 update_project_spine 参数失败: %v", err)
		}
		result, err := crudService.UpdateProjectSpine(rootDir, projectID, spine.Spine)
		if err != nil {
			return nil, err.Error()
		}
		b, _ := json.Marshal(result)
		return b, ""

	case "save_project":
		var save struct {
			Project *story_struct.Project `json:"project"`
		}
		if err := json.Unmarshal(p.Payload, &save); err != nil {
			return nil, fmt.Sprintf("解析 save_project 参数失败: %v", err)
		}
		if err := crudService.SaveProject(rootDir, save.Project); err != nil {
			return nil, err.Error()
		}
		return json.RawMessage(`"保存成功"`), ""

	// ========== 事件管理 ==========
	case "create_event":
		var ev struct {
			Event story_struct.Event `json:"event"`
		}
		if err := json.Unmarshal(p.Payload, &ev); err != nil {
			return nil, fmt.Sprintf("解析 create_event 参数失败: %v", err)
		}
		result, err := crudService.CreateEvent(rootDir, projectID, ev.Event)
		if err != nil {
			return nil, err.Error()
		}
		b, _ := json.Marshal(result)
		return b, ""

	case "update_event":
		var args struct {
			EventID string            `json:"event_id"`
			Event   story_struct.Event `json:"event"`
		}
		if err := json.Unmarshal(p.Payload, &args); err != nil {
			return nil, fmt.Sprintf("解析 update_event 参数失败: %v", err)
		}
		result, err := crudService.UpdateEvent(rootDir, projectID, args.EventID, args.Event)
		if err != nil {
			return nil, err.Error()
		}
		b, _ := json.Marshal(result)
		return b, ""

	case "delete_event":
		var args struct {
			EventID string `json:"event_id"`
		}
		if err := json.Unmarshal(p.Payload, &args); err != nil {
			return nil, fmt.Sprintf("解析 delete_event 参数失败: %v", err)
		}
		if err := crudService.DeleteEvent(rootDir, projectID, args.EventID); err != nil {
			return nil, err.Error()
		}
		return json.RawMessage(fmt.Sprintf(`"事件已删除: %s"`, args.EventID)), ""

	case "set_event_locked":
		var args struct {
			EventID string `json:"event_id"`
			Locked  bool   `json:"locked"`
		}
		if err := json.Unmarshal(p.Payload, &args); err != nil {
			return nil, fmt.Sprintf("解析 set_event_locked 参数失败: %v", err)
		}
		result, err := crudService.SetEventLocked(rootDir, projectID, args.EventID, args.Locked)
		if err != nil {
			return nil, err.Error()
		}
		b, _ := json.Marshal(result)
		return b, ""

	case "update_event_participants":
		var args struct {
			EventID      string                   `json:"event_id"`
			Participants []story_struct.Participant `json:"participants"`
		}
		if err := json.Unmarshal(p.Payload, &args); err != nil {
			return nil, fmt.Sprintf("解析 update_event_participants 参数失败: %v", err)
		}
		result, err := crudService.UpdateEventParticipants(rootDir, projectID, args.EventID, args.Participants)
		if err != nil {
			return nil, err.Error()
		}
		b, _ := json.Marshal(result)
		return b, ""

	// ========== 事件边管理 ==========
	case "create_event_edge":
		var args struct {
			SourceEventID string              `json:"source_event_id"`
			TargetEventID string              `json:"target_event_id"`
			EdgeType      story_struct.EdgeType `json:"edge_type"`
			MirrorInEdge  bool                `json:"mirror_in_edge"`
		}
		if err := json.Unmarshal(p.Payload, &args); err != nil {
			return nil, fmt.Sprintf("解析 create_event_edge 参数失败: %v", err)
		}
		if err := crudService.CreateEventEdge(rootDir, projectID, args.SourceEventID, args.TargetEventID, args.EdgeType, args.MirrorInEdge); err != nil {
			return nil, err.Error()
		}
		return json.RawMessage(fmt.Sprintf(`"边已创建: %s -> %s"`, args.SourceEventID, args.TargetEventID)), ""

	case "update_event_edge":
		var args struct {
			SourceEventID    string              `json:"source_event_id"`
			TargetEventID    string              `json:"target_event_id"`
			NewTargetEventID string              `json:"new_target_event_id"`
			EdgeType         story_struct.EdgeType `json:"edge_type"`
			MirrorInEdge     bool                `json:"mirror_in_edge"`
		}
		if err := json.Unmarshal(p.Payload, &args); err != nil {
			return nil, fmt.Sprintf("解析 update_event_edge 参数失败: %v", err)
		}
		if err := crudService.UpdateEventEdge(rootDir, projectID, args.SourceEventID, args.TargetEventID, args.NewTargetEventID, args.EdgeType, args.MirrorInEdge); err != nil {
			return nil, err.Error()
		}
		return json.RawMessage(fmt.Sprintf(`"边已更新"`)), ""

	case "delete_event_edge":
		var args struct {
			SourceEventID string              `json:"source_event_id"`
			TargetEventID string              `json:"target_event_id"`
			EdgeType      story_struct.EdgeType `json:"edge_type"`
			MirrorInEdge  bool                `json:"mirror_in_edge"`
		}
		if err := json.Unmarshal(p.Payload, &args); err != nil {
			return nil, fmt.Sprintf("解析 delete_event_edge 参数失败: %v", err)
		}
		if err := crudService.DeleteEventEdge(rootDir, projectID, args.SourceEventID, args.TargetEventID, args.EdgeType, args.MirrorInEdge); err != nil {
			return nil, err.Error()
		}
		return json.RawMessage(fmt.Sprintf(`"边已删除"`)), ""

	// ========== 子事件管理 ==========
	case "attach_sub_event":
		var args struct {
			ParentEventID string `json:"parent_event_id"`
			ChildEventID  string `json:"child_event_id"`
			Index         int    `json:"index"`
		}
		if err := json.Unmarshal(p.Payload, &args); err != nil {
			return nil, fmt.Sprintf("解析 attach_sub_event 参数失败: %v", err)
		}
		if err := crudService.AttachSubEvent(rootDir, projectID, args.ParentEventID, args.ChildEventID, args.Index); err != nil {
			return nil, err.Error()
		}
		return json.RawMessage(fmt.Sprintf(`"子事件已挂载"`)), ""

	case "move_sub_event":
		var args struct {
			ParentEventID string `json:"parent_event_id"`
			ChildEventID  string `json:"child_event_id"`
			Index         int    `json:"index"`
		}
		if err := json.Unmarshal(p.Payload, &args); err != nil {
			return nil, fmt.Sprintf("解析 move_sub_event 参数失败: %v", err)
		}
		if err := crudService.MoveSubEvent(rootDir, projectID, args.ParentEventID, args.ChildEventID, args.Index); err != nil {
			return nil, err.Error()
		}
		return json.RawMessage(fmt.Sprintf(`"子事件已移动"`)), ""

	case "detach_sub_event":
		var args struct {
			ParentEventID string `json:"parent_event_id"`
			ChildEventID  string `json:"child_event_id"`
		}
		if err := json.Unmarshal(p.Payload, &args); err != nil {
			return nil, fmt.Sprintf("解析 detach_sub_event 参数失败: %v", err)
		}
		if err := crudService.DetachSubEvent(rootDir, projectID, args.ParentEventID, args.ChildEventID); err != nil {
			return nil, err.Error()
		}
		return json.RawMessage(fmt.Sprintf(`"子事件已解除"`)), ""

	// ========== 实体管理 ==========
	case "create_entity":
		var ent struct {
			Entity story_struct.Entity `json:"entity"`
		}
		if err := json.Unmarshal(p.Payload, &ent); err != nil {
			return nil, fmt.Sprintf("解析 create_entity 参数失败: %v", err)
		}
		result, err := crudService.CreateEntity(rootDir, projectID, ent.Entity)
		if err != nil {
			return nil, err.Error()
		}
		b, _ := json.Marshal(result)
		return b, ""

	case "update_entity":
		var args struct {
			EntityID string              `json:"entity_id"`
			Entity   story_struct.Entity `json:"entity"`
		}
		if err := json.Unmarshal(p.Payload, &args); err != nil {
			return nil, fmt.Sprintf("解析 update_entity 参数失败: %v", err)
		}
		result, err := crudService.UpdateEntity(rootDir, projectID, args.EntityID, args.Entity)
		if err != nil {
			return nil, err.Error()
		}
		b, _ := json.Marshal(result)
		return b, ""

	case "delete_entity":
		var args struct {
			EntityID string `json:"entity_id"`
		}
		if err := json.Unmarshal(p.Payload, &args); err != nil {
			return nil, fmt.Sprintf("解析 delete_entity 参数失败: %v", err)
		}
		if err := crudService.DeleteEntity(rootDir, projectID, args.EntityID); err != nil {
			return nil, err.Error()
		}
		return json.RawMessage(fmt.Sprintf(`"实体已删除: %s"`, args.EntityID)), ""

	default:
		return nil, fmt.Sprintf("不支持的行为类型: %s", p.ActionType)
	}
}

// ListPendingProposalsTool 列出所有待审批的行为申请
// 说明: 观察者 Agent 调用此工具获取待办列表, 无需参数。
func ListPendingProposalsTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logrus.Debugf("调用 list_pending_proposals\n")
	proposalMu.RLock()
	defer proposalMu.RUnlock()
	result := make([]*ActionProposal, 0)
	for _, p := range proposalStore {
		if p.Status == ProposalStatusPending {
			result = append(result, p)
		}
	}
	return mcp.NewToolResultText(formatJSON(result)), nil
}

// =====================================================================================
// 行为申请工具定义注册
// 说明: ProposalToolDefinitions 返回所有行为申请相关工具的 ToolDefinition 列表。
//       供角色工厂注册, 与 CRUDToolDefinitions 分开以便按角色粒度控制权限。
//       演员角色仅注册提交/查询类工具, 观察者角色额外注册审查/执行工具。
//       设计理由: 分离定义使得角色工厂的工具注册语义更清晰。
// =====================================================================================

// ProposalToolDefinitions 返回所有行为申请相关工具的 ToolDefinition 列表
func ProposalToolDefinitions() []ToolDefinition {
	return []ToolDefinition{
		// ---------- 演员提交工具 ----------
		{
			Name:        "submit_action_proposal",
			Description: "演员提交行为申请, 表示对工程对象的修改意图。等待观察者审查批准后执行。提案者需提供申请者标识、行为类型(如 create_event/update_entity) 和参数",
			Parameters: objectProp("入参", map[string]jsonschema.Definition{
				"proposer_id": stringProp("申请者标识, 如 Agent ID 或角色名"),
				"action_type": stringProp("行为类型, 支持: create_project, update_project_meta, update_project_spine, save_project, create_event, update_event, delete_event, set_event_locked, update_event_participants, create_event_edge, update_event_edge, delete_event_edge, attach_sub_event, move_sub_event, detach_sub_event, create_entity, update_entity, delete_entity"),
				"target_id":   stringProp("目标 ID(事件/实体 ID), 可选, 部分行为类型需要"),
				"payload":     objectProp("行为参数, JSON 格式, 与对应 CRUD 工具的参数结构一致(包含 root_dir, project_id 等)", nil, nil),
			}, []string{"proposer_id", "action_type", "payload"}),
			Executor: SubmitActionProposalTool,
		},
		{
			Name:        "list_proposals",
			Description: "查询行为申请列表, 可按申请者和状态筛选",
			Parameters: objectProp("入参", map[string]jsonschema.Definition{
				"proposer_id": stringProp("申请者标识, 可选, 留空表示查询所有"),
				"status":      stringProp("申请状态筛选, 可选: pending / approved / rejected / withdrawn"),
			}, nil),
			Executor: ListProposalsTool,
		},
		{
			Name:        "get_proposal",
			Description: "获取单个行为申请的详细信息(含审查结果和执行结果)",
			Parameters: objectProp("入参", map[string]jsonschema.Definition{
				"proposal_id": stringProp("行为申请 ID"),
			}, []string{"proposal_id"}),
			Executor: GetProposalTool,
		},
		{
			Name:        "withdraw_proposal",
			Description: "撤回自己提交的且状态为 pending 的行为申请",
			Parameters: objectProp("入参", map[string]jsonschema.Definition{
				"proposal_id": stringProp("行为申请 ID"),
			}, []string{"proposal_id"}),
			Executor: WithdrawProposalTool,
		},

		// ---------- 观察者审查工具 ----------
		{
			Name:        "review_proposal",
			Description: "审查行为申请: 批准(approved=true)后自动执行修改并写回工程; 拒绝(approved=false)则标注为 rejected",
			Parameters: objectProp("入参", map[string]jsonschema.Definition{
				"proposal_id": stringProp("行为申请 ID"),
				"approved":    boolProp("是否批准: true=批准并执行, false=拒绝"),
				"review_note": stringProp("审查备注, 可选, 描述批准/拒绝的理由"),
			}, []string{"proposal_id", "approved"}),
			Executor: ReviewProposalTool,
		},
		{
			Name:        "list_pending_proposals",
			Description: "列出所有待审批的行为申请(观察者专用)",
			Parameters: objectProp("入参", map[string]jsonschema.Definition{}, nil),
			Executor: ListPendingProposalsTool,
		},
	}
}
