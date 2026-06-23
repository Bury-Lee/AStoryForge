package story_struct

//在结构化故事的基础上提供给前端和沙盘推演的接口

// StoryGraph 简短描述前端画布图数据
// 说明: 同时给出叙事链, 节点列表, 边列表
type StoryGraph struct {
	ProjectID string   `json:"project_id"`         // 工程 ID
	Spine     []string `json:"spine"`              // 叙事链
	Events    []Event  `json:"events,omitempty"`   // 节点列表
	Entities  []Entity `json:"entities,omitempty"` // 实体列表
}

// ExtractSourceResult 简短描述文本提取结果,预备字段
// 说明: 前端可直接展示提取出的初稿
type ExtractSourceResult struct {
	Meta         ProjectMeta   `json:"meta"`                   // 提取后的工程元信息
	Events       []Event       `json:"events,omitempty"`       // 提取出的事件
	Entities     []Entity      `json:"entities,omitempty"`     // 提取出的实体
	Edges        []EventEdge   `json:"edges,omitempty"`        // 提取出的事件边
	Participants []Participant `json:"participants,omitempty"` // 提取出的事件参与者
}

// ValidationIssue 简短描述单条校验问题
// 说明: 用于前端高亮问题节点
type ValidationIssue struct {
	Code      string   `json:"code"`                 // 问题代码
	Level     string   `json:"level"`                // 问题级别
	Message   string   `json:"message"`              // 问题说明
	EventIDs  []string `json:"event_ids,omitempty"`  // 关联事件
	EntityIDs []string `json:"entity_ids,omitempty"` // 关联实体
}

// ValidationReport 简短描述工程校验报告
// 说明: 对应编译前的完整检查结果
type ValidationReport struct {
	ProjectID string            `json:"project_id"`       // 工程 ID
	Passed    bool              `json:"passed"`           // 是否通过
	Issues    []ValidationIssue `json:"issues,omitempty"` // 问题列表
}

// OutlinePreview 简短描述大纲预览结果
// 说明: 返回前端可直接显示的 Markdown
type OutlinePreview struct {
	ProjectID string `json:"project_id"` // 工程 ID
	Markdown  string `json:"markdown"`   // 渲染后的 Markdown
}

// CompileResult 简短描述编译结果
// 说明: 同时返回产物路径和校验报告
type CompileResult struct {
	ProjectID  string           `json:"project_id"`         // 工程 ID
	OutputPath string           `json:"output_path"`        // 输出路径
	Markdown   string           `json:"markdown,omitempty"` // 产出的 Markdown
	Report     ValidationReport `json:"report"`             // 校验报告
}

// SimulationStep 简短描述模拟步骤
// 说明: 用于前端回放每一轮推演
type SimulationStep struct {
	ID         string         `json:"id"`                    // 步骤 ID
	Round      int            `json:"round"`                 // 轮次
	ActorID    string         `json:"actor_id,omitempty"`    // 行动实体 ID
	Action     string         `json:"action"`                // 行动描述
	Observer   string         `json:"observer,omitempty"`    // 观察者评语
	Accepted   bool           `json:"accepted"`              // 是否通过
	SceneState map[string]any `json:"scene_state,omitempty"` // 场景状态快照
}

// SimulationSession 简短描述模拟会话
// 说明: 前端可据此展示当前进度和结果
type SimulationSession struct {
	ID          string           `json:"id"`                     // 会话 ID
	ProjectID   string           `json:"project_id"`             // 工程 ID
	Status      SimulationStatus `json:"status"`                 // 会话状态
	Mode        SimulationMode   `json:"mode"`                   // 模拟模式
	Steps       []SimulationStep `json:"steps,omitempty"`        // 推演步骤
	DraftEvents []Event          `json:"draft_events,omitempty"` // 候选事件结果
}

// SimulationApplyResult 简短描述模拟落地结果
// 说明: 返回受影响的事件 ID 列表
type SimulationApplyResult struct {
	SessionID        string   `json:"session_id"`                   // 会话 ID
	AffectedEventIDs []string `json:"affected_event_ids,omitempty"` // 受影响事件
}

// GenerationTask 简短描述正文生成任务
// 说明: 前端可轮询任务状态
type GenerationTask struct {
	ID         string           `json:"id"`                    // 任务 ID
	ProjectID  string           `json:"project_id"`            // 工程 ID
	Status     GenerationStatus `json:"status"`                // 任务状态
	OutputPath string           `json:"output_path,omitempty"` // 输出路径
	Content    string           `json:"content,omitempty"`     // 生成内容
	Error      string           `json:"error,omitempty"`       // 错误信息
}

// OptimizedEventDraft 简短描述单个事件的优化草稿
// 说明: 保留原文和建议稿, 便于前端做对比
type OptimizedEventDraft struct {
	EventID               string `json:"event_id"`                         // 事件 ID
	OriginalIntroduction  string `json:"original_introduction,omitempty"`  // 原简介
	OptimizedIntroduction string `json:"optimized_introduction,omitempty"` // 优化后简介
	OriginalProcess       string `json:"original_process,omitempty"`       // 原过程
	OptimizedProcess      string `json:"optimized_process,omitempty"`      // 优化后过程
	OriginalOutcome       string `json:"original_outcome,omitempty"`       // 原结果
	OptimizedOutcome      string `json:"optimized_outcome,omitempty"`      // 优化后结果
	Comment               string `json:"comment,omitempty"`                // 优化说明
}

// OptimizeNodeExpressionResult 简短描述节点表达优化结果
// 说明: 前端可用于预览批量优化建议
type OptimizeNodeExpressionResult struct {
	ProjectID string                `json:"project_id"`       // 工程 ID
	Drafts    []OptimizedEventDraft `json:"drafts,omitempty"` // 优化草稿列表
}

// ApplyOptimizedNodesResult 简短描述应用优化稿结果
// 说明: 返回实际写回的事件列表
type ApplyOptimizedNodesResult struct {
	ProjectID        string   `json:"project_id"`                   // 工程 ID
	AffectedEventIDs []string `json:"affected_event_ids,omitempty"` // 已更新事件 ID
}

// FrontendCRUDService 简短描述前端增删改查业务接口
// 说明: 覆盖前端从导入, 编辑, 推演, 编译到生成的完整调用面
type FrontendCRUDService interface {
	// ListProjects 获取工程列表
	// 参数: rootDir - 工程根目录
	// 返回: list - 工程元信息列表,error - 查询失败信息
	// 说明: 用于首页或侧边栏展示全部工程及其基础统计信息
	ListProjects(rootDir string) ([]ProjectMeta, error)

	// CreateProject 新建工程
	// 参数: meta - 新建工程元信息
	// 返回: project - 创建后的完整工程,error - 创建失败信息
	// 说明: 前端点击新建工程时调用, 初始化目录, 元信息与默认文件
	CreateProject(meta ProjectMeta) (*Project, error)

	// LoadProject 加载完整工程
	// 参数: rootDir - 工程根目录
	// 返回: project - 完整工程数据,error - 加载失败信息
	// 说明: 用于进入编辑器时一次性加载项目主数据
	LoadProject(rootDir string) (*Project, error)

	// DeleteProject 删除工程
	// 参数: rootDir - 工程根目录, projectID - 工程 ID
	// 返回: error - 删除失败信息
	// 说明: 删除整个工程目录及关联资源, 一般需要前端二次确认,目前是预备功能
	DeleteProject(rootDir string, projectID string) error

	// UpdateProjectMeta 更新工程元信息
	// 参数: rootDir - 工程根目录, projectID - 工程 ID, meta - 更新后的工程元信息
	// 返回: meta - 更新后的工程元信息,error - 更新失败信息
	// 说明: 用于工程设置页统一修改 project.json 中的基础配置
	UpdateProjectMeta(rootDir string, projectID string, meta ProjectMeta) (*ProjectMeta, error)

	// UpdateProjectSpine 更新叙事链顺序
	// 参数: rootDir - 工程根目录, projectID - 工程 ID, spine - 最新的叙事链顺序
	// 返回: spine - 保存后的叙事链,error - 更新失败信息
	// 说明: 对应前端拖拽主链后的整体提交, 叙事顺序与真实时间轴分离,主要给AI的MCP工具使用
	UpdateProjectSpine(rootDir string, projectID string, spine []string) ([]string, error)

	// SaveProject 保存完整工程
	// 参数: rootDir - 工程根目录, project - 待保存工程
	// 返回: error - 保存失败信息
	// 说明: 适合批量修改后的统一提交, 也可作为自动保存入口
	SaveProject(rootDir string, project *Project) error

	// ListEvents 获取工程下全部事件
	// 参数: rootDir - 工程根目录, projectID - 工程 ID
	// 返回: list - 事件列表,error - 查询失败信息
	// 说明: 用于事件列表面板, 搜索面板, 批量选择器等场景
	ListEvents(rootDir string, projectID string) ([]Event, error)

	// GetEvent 获取单个事件详情
	// 参数: rootDir - 工程根目录, projectID - 工程 ID, eventID - 事件 ID
	// 返回: event - 事件详情,error - 查询失败信息
	// 说明: 用于打开节点详情抽屉或事件编辑表单时按需加载
	GetEvent(rootDir string, projectID string, eventID string) (*Event, error)

	// CreateEvent 创建事件节点
	// 参数: rootDir - 工程根目录, projectID - 工程 ID, event - 事件内容
	// 返回: event - 创建后的事件,error - 创建失败信息
	// 说明: 支持普通事件, 容器事件, mixed 事件等多种节点形态
	CreateEvent(rootDir string, projectID string, event Event) (*Event, error)

	// UpdateEvent 更新事件节点
	// 参数: rootDir - 工程根目录, projectID - 工程 ID, eventID - 事件 ID, event - 更新后的事件内容
	// 返回: event - 更新后的事件,error - 更新失败信息
	// 说明: 用于修改事件名称, 简介, 时间, 过程, 结果, 类型等核心字段
	UpdateEvent(rootDir string, projectID string, eventID string, event Event) (*Event, error)

	// DeleteEvent 删除事件节点
	// 参数: rootDir - 工程根目录, projectID - 工程 ID, eventID - 事件 ID
	// 返回: error - 删除失败信息
	// 说明: 删除前需要清理主链, 因果边, 父子关系等关联引用
	DeleteEvent(rootDir string, projectID string, eventID string) error

	// SetEventLocked 设置事件锁定状态
	// 参数: rootDir - 工程根目录, projectID - 工程 ID, eventID - 事件 ID, locked - 锁定标记
	// 返回: event - 更新后的事件,error - 更新失败信息
	// 说明: 锁定后该事件不应被 AI 自动改写, 用于保护用户确认内容
	SetEventLocked(rootDir string, projectID string, eventID string, locked bool) (*Event, error)

	// UpdateEventParticipants 更新事件参与者
	// 参数: rootDir - 工程根目录, projectID - 工程 ID, eventID - 事件 ID, participants - 新的参与者列表
	// 返回: event - 更新后的事件,error - 更新失败信息
	// 说明: 用于维护谁在场, 入场状态, 离场影响等角色弧光数据
	UpdateEventParticipants(rootDir string, projectID string, eventID string, participants []Participant) (*Event, error)

	// CreateEventEdge 创建事件因果边
	// 参数: rootDir - 工程根目录, projectID - 工程 ID, sourceEventID - 起点事件 ID, targetEventID - 终点事件 ID, edgeType - 边类型, mirrorInEdge - 是否回填入边
	// 返回: error - 创建失败信息
	// 说明: 对应画布上拉线建边, 可同步回填目标事件入边
	CreateEventEdge(rootDir string, projectID string, sourceEventID string, targetEventID string, edgeType EdgeType, mirrorInEdge bool) error

	// UpdateEventEdge 更新事件因果边
	// 参数: rootDir - 工程根目录, projectID - 工程 ID, sourceEventID - 原起点事件 ID, targetEventID - 原终点事件 ID, newTargetEventID - 新终点事件 ID, edgeType - 更新后的边类型, mirrorInEdge - 是否同步更新入边
	// 返回: error - 更新失败信息
	// 说明: 用于改线, 切换 cause 或 result 类型, 保持图结构一致
	UpdateEventEdge(rootDir string, projectID string, sourceEventID string, targetEventID string, newTargetEventID string, edgeType EdgeType, mirrorInEdge bool) error

	// DeleteEventEdge 删除事件因果边
	// 参数: rootDir - 工程根目录, projectID - 工程 ID, sourceEventID - 起点事件 ID, targetEventID - 终点事件 ID, edgeType - 边类型, mirrorInEdge - 是否同步删除入边
	// 返回: error - 删除失败信息
	// 说明: 对应画布断线操作, 可同步移除目标节点中的入边引用
	DeleteEventEdge(rootDir string, projectID string, sourceEventID string, targetEventID string, edgeType EdgeType, mirrorInEdge bool) error

	// AttachSubEvent 建立父子包含关系
	// 参数: rootDir - 工程根目录, projectID - 工程 ID, parentEventID - 父事件 ID, childEventID - 子事件 ID, index - 插入位置
	// 返回: error - 建立失败信息
	// 说明: 用于把事件挂入卷, 章, mixed 节点等树形结构中
	AttachSubEvent(rootDir string, projectID string, parentEventID string, childEventID string, index int) error

	// MoveSubEvent 调整子事件顺序
	// 参数: rootDir - 工程根目录, projectID - 工程 ID, parentEventID - 父事件 ID, childEventID - 子事件 ID, index - 新位置
	// 返回: error - 调整失败信息
	// 说明: 对应树结构拖拽排序, 不改变事件内容只调整包含顺序
	MoveSubEvent(rootDir string, projectID string, parentEventID string, childEventID string, index int) error

	// DetachSubEvent 解除父子包含关系
	// 参数: rootDir - 工程根目录, projectID - 工程 ID, parentEventID - 父事件 ID, childEventID - 子事件 ID
	// 返回: error - 解除失败信息
	// 说明: 仅解除挂载关系, 不删除子事件实体本身
	DetachSubEvent(rootDir string, projectID string, parentEventID string, childEventID string) error

	// ListEntities 获取工程下全部实体
	// 参数: rootDir - 工程根目录, projectID - 工程 ID
	// 返回: list - 实体列表,error - 查询失败信息
	// 说明: 用于实体库面板, 参与者选择器, 关系编辑器等场景
	ListEntities(rootDir string, projectID string) ([]Entity, error)

	// GetEntity 获取单个实体详情
	// 参数: rootDir - 工程根目录, projectID - 工程 ID, entityID - 实体 ID
	// 返回: entity - 实体详情,error - 查询失败信息
	// 说明: 用于打开实体详情面板, 查看介绍, 关系, 规则引用等信息
	GetEntity(rootDir string, projectID string, entityID string) (*Entity, error)

	// CreateEntity 创建实体节点
	// 参数: rootDir - 工程根目录, projectID - 工程 ID, entity - 实体内容
	// 返回: entity - 创建后的实体,error - 创建失败信息
	// 说明: 用于新增角色, 组织, 物品, 概念等实体节点
	CreateEntity(rootDir string, projectID string, entity Entity) (*Entity, error)

	// UpdateEntity 更新实体节点
	// 参数: rootDir - 工程根目录, projectID - 工程 ID, entityID - 实体 ID, entity - 更新后的实体内容
	// 返回: entity - 更新后的实体,error - 更新失败信息
	// 说明: 用于修改介绍, 关系, 规则引用和展示信息
	UpdateEntity(rootDir string, projectID string, entityID string, entity Entity) (*Entity, error)

	// DeleteEntity 删除实体节点
	// 参数: rootDir - 工程根目录, projectID - 工程 ID, entityID - 实体 ID
	// 返回: error - 删除失败信息
	// 说明: 删除前应检查参与者引用和实体关系, 避免留下悬空引用
	DeleteEntity(rootDir string, projectID string, entityID string) error

	// GetSimulation 获取模拟会话详情
	// 参数: sessionID - 模拟会话 ID
	// 返回: session - 当前会话与步骤数据,error - 查询失败信息
	// 说明: 用于前端轮询或恢复模拟界面显示
	GetSimulation(sessionID string) (*SimulationSession, error)

	// GetGenerationTask 查询正文生成任务
	// 参数: taskID - 生成任务 ID
	// 返回: task - 当前任务状态与内容,error - 查询失败信息
	// 说明: 用于前端轮询生成进度, 展示结果或错误原因
	GetGenerationTask(taskID string) (*GenerationTask, error)
}

// FrontendExecutionService 简短描述前端检查/保存/执行业务接口
// 说明: 覆盖前端从导入, 编辑, 推演, 编译到生成的完整调用面
type FrontendExecutionService interface {
	// ExtractFromSource 从原始文本提取结构化工程初稿
	// 参数: rootDir - 工程根目录, title - 工程标题, sourceText - 原始文本, writeRules - 附加写作约束, worldRules - 附加世界规则, autoCreate - 是否自动建工程
	// 返回: result - 提取出的元信息, 事件, 实体初稿,error - 提取失败信息
	// 说明: 交由 AI 模型处理, 可串联多次 MCP 或多阶段提取流程完成初始结构化构建
	ExtractFromSource(rootDir string, title string, sourceText string, writeRules []string, worldRules []string, autoCreate bool) (*ExtractSourceResult, error)

	// GetStoryGraph 获取画布图数据
	// 参数: rootDir - 工程根目录, projectID - 工程 ID
	// 返回: graph - 画布所需图数据,error - 获取失败信息
	// 说明: 用于节点编辑器初始化, 一次性返回主链, 事件, 实体等信息
	GetStoryGraph(rootDir string, projectID string) (*StoryGraph, error)

	// ValidateProject 校验工程结构
	// 参数: rootDir - 工程根目录, projectID - 工程 ID
	// 返回: report - 校验报告,error - 校验失败信息
	// 说明: 检查循环引用, 缺失引用, 因果断裂, 时间矛盾等结构问题
	ValidateProject(rootDir string, projectID string) (*ValidationReport, error)

	// CompileProject 编译工程为 Markdown 或中间产物
	// 参数: rootDir - 工程根目录, projectID - 工程 ID, eventIDs - 编译范围, outputPath - 输出路径, validateOnly - 是否仅校验
	// 返回: result - 编译结果与校验报告,error - 编译失败信息
	// 说明: 对应正式编译流程, 可用于导出结构化大纲或生成后续步骤输入
	MakeMarkdown(rootDir string, projectID string, eventIDs []string, outputPath string, validateOnly bool) (*CompileResult, error)

	// MakeMermaid 生成 Mermaid 图,用于输入给ai模型,表达事件关系和流程,逻辑等信息
	// 参数: rootDir - 工程根目录, projectID - 工程 ID, eventIDs - 编译范围, outputPath - 输出路径, validateOnly - 是否仅校验
	// 返回: result - 编译结果与校验报告,error - 编译失败信息
	// 说明: 对应正式编译流程, 可用于导出结构化大纲或生成后续步骤输入
	MakeMermaid(rootDir string, projectID string, eventIDs []string, outputPath string, validateOnly bool) (*CompileResult, error)

	// StartSimulation 开始一次模拟推演
	// 参数: rootDir - 工程根目录, projectID - 工程 ID, mode - 模拟模式, startEventID - 起始事件 ID, endEventID - 结束事件 ID, parentEventID - 父事件 ID, participantIDs - 参与实体 ID, userPrompt - 用户提示, maxRounds - 最大轮数
	// 返回: session - 新建的模拟会话,error - 启动失败信息
	// 说明: 支持填补两事件间空白或细化父事件内部过程两类主要场景
	StartSimulation(rootDir string, projectID string, mode SimulationMode, startEventID string, endEventID string, parentEventID string, participantIDs []string, userPrompt string, maxRounds int) (*SimulationSession, error)

	// AdvanceSimulation 推进下一轮模拟
	// 参数: sessionID - 会话 ID, userPrompt - 用户引导语, manualActor - 手动指定行动者
	// 返回: step - 新产生的模拟步骤,error - 推进失败信息
	// 说明: 支持 AI 自动推进, 也支持用户中途介入修正推演方向
	AdvanceSimulation(sessionID string, userPrompt string, manualActor string) (*SimulationStep, error)

	// ReviewSimulation 审查某一步模拟结果
	// 参数: sessionID - 会话 ID, stepID - 步骤 ID, approved - 审核结论, comment - 审查备注
	// 返回: step - 审查后的步骤,error - 审查失败信息
	// 说明: 用于人工确认, 驳回, 标记异常, 与观察者机制配合使用
	ReviewSimulation(sessionID string, stepID string, approved bool, comment string) (*SimulationStep, error)

	// ApplySimulationResult 落地模拟结果
	// 参数: sessionID - 会话 ID, mode - 落地模式, targetEventID - 目标事件 ID, parentEventID - 父事件 ID, insertIndex - 插入位置
	// 返回: result - 受影响事件列表,error - 落地失败信息
	// 说明: 支持覆盖原事件, 挂为子事件, 新增独立事件三种导出方式
	ApplySimulationResult(sessionID string, mode SimulationApplyMode, targetEventID string, parentEventID string, insertIndex int) (*SimulationApplyResult, error)

	// CancelSimulation 取消模拟会话
	// 参数: sessionID - 模拟会话 ID
	// 返回: error - 取消失败信息
	// 说明: 用于用户主动终止当前推演, 释放后续资源占用
	CancelSimulation(sessionID string) error

	// GenerateDraft 创建正文生成任务
	// 参数: rootDir - 工程根目录, projectID - 工程 ID, eventIDs - 指定生成范围, outputPath - 输出路径, model - 使用模型, systemPrompt - 附加提示词
	// 返回: task - 新建的生成任务,error - 创建失败信息
	// 说明: 基于编译后的大纲调用大模型生成正文, 适合异步执行
	GenerateDraft(rootDir string, projectID string, eventIDs []string, outputPath string, model string, systemPrompt string) (*GenerationTask, error)

	// OptimizeNodeExpression 优化一个或多个节点的表达
	// 参数: rootDir - 工程根目录, projectID - 工程 ID, eventIDs - 待优化节点, instruction - 风格要求, focusFields - 优化重点, keepMeaning - 是否保持原意, rewriteStyle - 改写风格
	// 返回: result - 各节点的优化草稿,error - 优化失败信息
	// 说明: 适合对 introduction, process, outcome 做润色增强, 默认不直接覆盖原数据
	OptimizeNodeExpression(rootDir string, projectID string, eventIDs []string, instruction string, focusFields []string, keepMeaning bool, rewriteStyle string) (*OptimizeNodeExpressionResult, error)

	// ApplyOptimizedNodes 应用节点优化结果
	// 参数: rootDir - 工程根目录, projectID - 工程 ID, drafts - 需要写回的优化草稿列表
	// 返回: result - 实际受影响的事件列表,error - 应用失败信息
	// 说明: 前端确认后再调用, 支持只应用批量建议中的部分节点
	ApplyOptimizedNodes(rootDir string, projectID string, drafts []OptimizedEventDraft) (*ApplyOptimizedNodesResult, error)
}
