package ai

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/sirupsen/logrus"
)

// =====================================================================================
// 行为申请工具（演员专用）
// 说明: 演员 Agent 通过以下工具提交修改意图, 不直接操作工程对象。
//       submit_action_proposal 是唯一入口, 入参中的 payload 保持与对应 CRUD 工具一致的
//       JSON 格式, 以便观察者直接复用解析逻辑。
//       设计理由: 一个通用提交入口替代多个细粒度写工具, 降低演员 Agent 的学习成本;
//       审查者只需在一个地方做路由分发, 不必为每种操作单独实现审查接口。
// =====================================================================================

// submitProposalArgs 提交行为申请参数
type submitProposalArgs struct {
	ProposerID string          `json:"proposer_id"`
	ActionType string          `json:"action_type"`
	TargetID   string          `json:"target_id,omitempty"`
	Payload    json.RawMessage `json:"payload"`
}

// SubmitActionProposalTool 提交行为申请
// 说明: 演员 Agent 调用此工具表达对工程对象的修改意图。
//       提交后 status 为 pending, 等待观察者审查。
func SubmitActionProposalTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logrus.Debugf("调用 submit_action_proposal %+v\n", request)
	var args submitProposalArgs
	if err := bindArgs(request, &args); err != nil {
		return nil, err
	}
	if args.ProposerID == "" {
		return nil, fmt.Errorf("申请者标识不能为空")
	}
	if args.ActionType == "" {
		return nil, fmt.Errorf("行为类型不能为空")
	}
	proposal := &ActionProposal{
		ID:         generateProposalID(),
		ProposerID: args.ProposerID,
		ActionType: args.ActionType,
		TargetID:   args.TargetID,
		Payload:    args.Payload,
		Status:     ProposalStatusPending,
		CreatedAt:  time.Now().Format(time.RFC3339),
	}
	proposalMu.Lock()
	proposalStore[proposal.ID] = proposal
	proposalMu.Unlock()
	return mcp.NewToolResultText(formatJSON(proposal)), nil
}

// listProposalsArgs 查询行为申请列表参数
type listProposalsArgs struct {
	ProposerID string         `json:"proposer_id"`
	Status     ProposalStatus `json:"status,omitempty"`
}

// ListProposalsTool 查询行为申请列表
// 说明: 按申请者和(可选)状态筛选, 返回匹配的申请列表。
func ListProposalsTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logrus.Debugf("调用 list_proposals %+v\n", request)
	var args listProposalsArgs
	if err := bindArgs(request, &args); err != nil {
		return nil, err
	}
	proposalMu.RLock()
	defer proposalMu.RUnlock()
	result := make([]*ActionProposal, 0)
	for _, p := range proposalStore {
		if args.ProposerID != "" && p.ProposerID != args.ProposerID {
			continue
		}
		if args.Status != "" && p.Status != args.Status {
			continue
		}
		result = append(result, p)
	}
	return mcp.NewToolResultText(formatJSON(result)), nil
}

// getProposalArgs 查询单个行为申请参数
type getProposalArgs struct {
	ProposalID string `json:"proposal_id"`
}

// GetProposalTool 获取单个行为申请详情
func GetProposalTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logrus.Debugf("调用 get_proposal %+v\n", request)
	var args getProposalArgs
	if err := bindArgs(request, &args); err != nil {
		return nil, err
	}
	proposalMu.RLock()
	p, ok := proposalStore[args.ProposalID]
	proposalMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("行为申请不存在: %s", args.ProposalID)
	}
	return mcp.NewToolResultText(formatJSON(p)), nil
}

// withdrawProposalArgs 撤回行为申请参数
type withdrawProposalArgs struct {
	ProposalID string `json:"proposal_id"`
}

// WithdrawProposalTool 撤回行为申请
// 说明: 仅允许撤回自己提交的且状态为 pending 的申请。
func WithdrawProposalTool(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	logrus.Debugf("调用 withdraw_proposal %+v\n", request)
	var args withdrawProposalArgs
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
		return nil, fmt.Errorf("只能撤回待审批的申请, 当前状态: %s", p.Status)
	}
	p.Status = ProposalStatusWithdrawn
	proposalStore[args.ProposalID] = p
	proposalMu.Unlock()
	return mcp.NewToolResultText(formatJSON(p)), nil
}
