package story_struct

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// =====================================================================================
// 服务实现入口
// 说明: 以下方法均为 *ProjectObj 的方法, 用于满足 interface.go 中定义的两个前端业务接口。
//       与 rootDir/projectID 相关的接口会通过 loadOrInit 在调用前加载工程, 修改后写回磁盘。
//       Simulation/Generation/Optimization 相关状态保存在内存中 (全局 map), 实际项目中可替换为数据库。
// =====================================================================================

// 全局会话与任务存储(进程级, 简单实现, 后续可替换为持久化层)
var (
	simulationStore = make(map[string]*SimulationSession)
	generationStore = make(map[string]*GenerationTask)
	storeMu         sync.RWMutex
)

// helper: 根据 rootDir 加载工程, 失败时返回错误
func loadProject(rootDir string) (*ProjectObj, error) {
	if rootDir == "" {
		return nil, errors.New("工程根目录为空")
	}
	p := &ProjectObj{}
	if err := p.LoadProjectObj(rootDir); err != nil {
		return nil, err
	}
	return p, nil
}

// helper: 构造文件路径(优先使用 ProjectMeta.RootDir)
func resolveRootDir(rootDir, projectID string) string {
	if rootDir != "" {
		return rootDir
	}
	if projectID != "" {
		return projectID
	}
	return ""
}

// =====================================================================================
// FrontendCRUDService 实现
// =====================================================================================

// ListProjects 获取工程列表
func (p *ProjectObj) ListProjects(rootDir string) ([]ProjectMeta, error) {
	if rootDir == "" {
		return nil, errors.New("未输入目录")
	}
	if _, err := os.Stat(rootDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("目录不存在: %s", rootDir)
	}
	list := make([]ProjectMeta, 0)
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		if !strings.HasSuffix(strings.ToLower(info.Name()), ".json") {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		var meta ProjectMeta
		if err := json.Unmarshal(data, &meta); err != nil {
			return nil
		}
		list = append(list, meta)
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("遍历目录失败: %v", err)
	}
	return list, nil
}

// CreateProject 新建工程
func (p *ProjectObj) CreateProject(meta ProjectMeta) (*Project, error) {
	if strings.TrimSpace(meta.Title) == "" {
		return nil, errors.New("未设置标题")
	}
	if len([]rune(meta.Title)) > 200 {
		return nil, errors.New("标题长度不能超过200个字符")
	}
	if meta.Author == "" {
		meta.Author = "未知"
	}
	if meta.CreatedAt.IsZero() {
		meta.CreatedAt = time.Now()
	}
	meta.UpdatedAt = time.Now()
	if meta.Settings == nil {
		meta.Settings = make(map[string]string)
	}
	if meta.EventCount < 0 {
		meta.EventCount = 0
	}
	if meta.EntityCount < 0 {
		meta.EntityCount = 0
	}
	if meta.RootDir == "" {
		meta.RootDir = "./projects"
	}
	if _, err := os.Stat(meta.RootDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("工程根目录不存在: %s", meta.RootDir)
	}

	newP := &ProjectObj{}
	newP.ProjectMeta = meta
	newP.WorldSetting = WorldSetting{
		TimeLabels: make(map[string]string),
		Spine:      []string{},
		WriteRules: []string{},
		WorldRules: []string{},
	}
	newP.Events = make(map[string]Event)
	newP.Entities = make(map[string]Entity)
	newP.Participant = make(map[string]Participant)
	newP.Entity = make(map[string]Entity)

	if err := newP.SaveProjectObj(); err != nil {
		return nil, fmt.Errorf("新建工程失败: %w", err)
	}

	project := &Project{
		Meta:         newP.ProjectMeta,
		WorldSetting: newP.WorldSetting,
		Events:       []Event{},
		Entities:     []Entity{},
	}
	return project, nil
}

// LoadProject 加载完整工程
func (p *ProjectObj) LoadProject(rootDir string) (*Project, error) {
	target, err := loadProject(rootDir)
	if err != nil {
		return nil, err
	}

	events := make([]Event, 0, len(target.Events))
	for _, e := range target.Events {
		events = append(events, e)
	}
	sort.Slice(events, func(i, j int) bool { return events[i].ID < events[j].ID })

	entities := make([]Entity, 0, len(target.Entities))
	for _, e := range target.Entities {
		entities = append(entities, e)
	}
	sort.Slice(entities, func(i, j int) bool { return entities[i].ID < entities[j].ID })

	return &Project{
		Meta:         target.ProjectMeta,
		WorldSetting: target.WorldSetting,
		Events:       events,
		Entities:     entities,
	}, nil
}

// DeleteProject 删除工程
func (p *ProjectObj) DeleteProject(rootDir string, projectID string) error {
	target := resolveRootDir(rootDir, projectID)
	if target == "" {
		return errors.New("未指定工程路径")
	}
	if _, err := os.Stat(target); os.IsNotExist(err) {
		return fmt.Errorf("工程不存在: %s", target)
	}
	if err := os.RemoveAll(target); err != nil {
		return fmt.Errorf("删除工程失败: %w", err)
	}
	logrus.Infof("工程已删除: %s", target)
	return nil
}

// UpdateProjectMeta 更新工程元信息
func (p *ProjectObj) UpdateProjectMeta(rootDir string, projectID string, meta ProjectMeta) (*ProjectMeta, error) {
	target := resolveRootDir(rootDir, projectID)
	if target == "" {
		return nil, errors.New("未指定工程路径")
	}
	loaded, err := loadProject(target)
	if err != nil {
		return nil, err
	}
	// 保留创建时间, 更新更新时间
	if meta.CreatedAt.IsZero() {
		meta.CreatedAt = loaded.ProjectMeta.CreatedAt
	}
	meta.UpdatedAt = time.Now()
	// 保留统计计数, 若未传入
	if meta.EventCount == 0 {
		meta.EventCount = loaded.ProjectMeta.EventCount
	}
	if meta.EntityCount == 0 {
		meta.EntityCount = loaded.ProjectMeta.EntityCount
	}
	if meta.Settings == nil {
		meta.Settings = loaded.ProjectMeta.Settings
	}
	if meta.RootDir == "" {
		meta.RootDir = target
	}
	loaded.ProjectMeta = meta
	if err := loaded.SaveProjectObj(); err != nil {
		return nil, fmt.Errorf("更新工程元信息失败: %w", err)
	}
	return &loaded.ProjectMeta, nil
}

// UpdateProjectSpine 更新叙事链顺序
func (p *ProjectObj) UpdateProjectSpine(rootDir string, projectID string, spine []string) ([]string, error) {
	target := resolveRootDir(rootDir, projectID)
	loaded, err := loadProject(target)
	if err != nil {
		return nil, err
	}
	// 校验 spine 中所有 ID 必须存在于 Events
	for _, id := range spine {
		if _, ok := loaded.Events[id]; !ok {
			return nil, fmt.Errorf("叙事链中存在未知事件 ID: %s", id)
		}
	}
	// 检测环
	visited := make(map[string]bool)
	for _, id := range spine {
		if visited[id] {
			return nil, fmt.Errorf("叙事链存在重复/环,事件 ID: %s", id)
		}
		visited[id] = true
	}
	loaded.WorldSetting.Spine = append([]string{}, spine...)
	loaded.ProjectMeta.UpdatedAt = time.Now()
	if err := loaded.SaveProjectObj(); err != nil {
		return nil, fmt.Errorf("保存叙事链失败: %w", err)
	}
	return loaded.WorldSetting.Spine, nil
}

// SaveProject 保存完整工程
func (p *ProjectObj) SaveProject(rootDir string, project *Project) error {
	if project == nil {
		return errors.New("工程为空")
	}
	target := rootDir
	if target == "" {
		target = project.Meta.RootDir
	}
	if target == "" {
		return errors.New("未指定工程路径")
	}
	loaded, err := loadProject(target)
	if err != nil {
		return err
	}

	loaded.ProjectMeta = project.Meta
	loaded.ProjectMeta.UpdatedAt = time.Now()
	loaded.WorldSetting = project.WorldSetting

	loaded.Events = make(map[string]Event, len(project.Events))
	for _, e := range project.Events {
		if e.ID == "" {
			continue
		}
		loaded.Events[e.ID] = e
	}
	loaded.Entities = make(map[string]Entity, len(project.Entities))
	for _, e := range project.Entities {
		if e.ID == "" {
			continue
		}
		loaded.Entities[e.ID] = e
	}
	loaded.ProjectMeta.EventCount = len(loaded.Events)
	loaded.ProjectMeta.EntityCount = len(loaded.Entities)
	// 同步到冗余字段
	loaded.Entity = make(map[string]Entity, len(loaded.Entities))
	for k, v := range loaded.Entities {
		loaded.Entity[k] = v
	}
	return loaded.SaveProjectObj()
}

// ListEvents 获取工程下全部事件
func (p *ProjectObj) ListEvents(rootDir string, projectID string) ([]Event, error) {
	target := resolveRootDir(rootDir, projectID)
	loaded, err := loadProject(target)
	if err != nil {
		return nil, err
	}
	list := make([]Event, 0, len(loaded.Events))
	for _, e := range loaded.Events {
		list = append(list, e)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })
	return list, nil
}

// GetEvent 获取单个事件详情
func (p *ProjectObj) GetEvent(rootDir string, projectID string, eventID string) (*Event, error) {
	target := resolveRootDir(rootDir, projectID)
	loaded, err := loadProject(target)
	if err != nil {
		return nil, err
	}
	ev, ok := loaded.Events[eventID]
	if !ok {
		return nil, fmt.Errorf("事件不存在: %s", eventID)
	}
	return &ev, nil
}

// ensureID 为空的事件/实体生成 ID
func ensureEventID(ev *Event) {
	if ev.ID == "" {
		ev.ID = fmt.Sprintf("event_%d", time.Now().UnixNano())
	}
	if ev.InEdges == nil {
		ev.InEdges = []EventEdge{}
	}
	if ev.OutEdges == nil {
		ev.OutEdges = []EventEdge{}
	}
	if ev.SubEvents == nil {
		ev.SubEvents = []string{}
	}
	if ev.Participants == nil {
		ev.Participants = []Participant{}
	}
	if ev.SubRules == nil {
		ev.SubRules = []string{}
	}
	if ev.Time == nil {
		ev.Time = map[string]string{}
	}
	if ev.Outcome == nil {
		ev.Outcome = map[string]string{}
	}
}

func ensureEntityID(ent *Entity) {
	if ent.ID == "" {
		ent.ID = fmt.Sprintf("entity_%d", time.Now().UnixNano())
	}
	if ent.Introduction == nil {
		ent.Introduction = []string{}
	}
	if ent.Relationships == nil {
		ent.Relationships = []EntityRelationship{}
	}
	if ent.RuleRefs == nil {
		ent.RuleRefs = []string{}
	}
	if ent.Events == nil {
		ent.Events = []string{}
	}
}

// CreateEvent 创建事件节点
func (p *ProjectObj) CreateEvent(rootDir string, projectID string, event Event) (*Event, error) {
	target := resolveRootDir(rootDir, projectID)
	loaded, err := loadProject(target)
	if err != nil {
		return nil, err
	}
	ensureEventID(&event)
	if _, exists := loaded.Events[event.ID]; exists {
		return nil, fmt.Errorf("事件 ID 已存在: %s", event.ID)
	}
	loaded.Events[event.ID] = event
	loaded.ProjectMeta.EventCount = len(loaded.Events)
	loaded.ProjectMeta.UpdatedAt = time.Now()
	if err := loaded.SaveProjectObj(); err != nil {
		return nil, err
	}
	return &event, nil
}

// UpdateEvent 更新事件节点
func (p *ProjectObj) UpdateEvent(rootDir string, projectID string, eventID string, event Event) (*Event, error) {
	target := resolveRootDir(rootDir, projectID)
	loaded, err := loadProject(target)
	if err != nil {
		return nil, err
	}
	if _, ok := loaded.Events[eventID]; !ok {
		return nil, fmt.Errorf("事件不存在: %s", eventID)
	}
	// 保留原始 ID
	event.ID = eventID
	ensureEventID(&event)
	loaded.Events[eventID] = event
	loaded.ProjectMeta.UpdatedAt = time.Now()
	if err := loaded.SaveProjectObj(); err != nil {
		return nil, err
	}
	return &event, nil
}

// DeleteEvent 删除事件节点
func (p *ProjectObj) DeleteEvent(rootDir string, projectID string, eventID string) error {
	target := resolveRootDir(rootDir, projectID)
	loaded, err := loadProject(target)
	if err != nil {
		return err
	}
	if _, ok := loaded.Events[eventID]; !ok {
		return fmt.Errorf("事件不存在: %s", eventID)
	}
	// 清理主链中的引用
	newSpine := make([]string, 0, len(loaded.WorldSetting.Spine))
	for _, id := range loaded.WorldSetting.Spine {
		if id != eventID {
			newSpine = append(newSpine, id)
		}
	}
	loaded.WorldSetting.Spine = newSpine

	// 清理父子关系与因果边
	for id, ev := range loaded.Events {
		if id == eventID {
			continue
		}
		// 子事件
		newSubs := make([]string, 0, len(ev.SubEvents))
		for _, sub := range ev.SubEvents {
			if sub != eventID {
				newSubs = append(newSubs, sub)
			}
		}
		ev.SubEvents = newSubs
		// 入边
		newIn := make([]EventEdge, 0, len(ev.InEdges))
		for _, e := range ev.InEdges {
			if e.Target != eventID {
				newIn = append(newIn, e)
			}
		}
		ev.InEdges = newIn
		// 出边
		newOut := make([]EventEdge, 0, len(ev.OutEdges))
		for _, e := range ev.OutEdges {
			if e.Target != eventID {
				newOut = append(newOut, e)
			}
		}
		ev.OutEdges = newOut
		// 参与者引用
		newParts := make([]Participant, 0, len(ev.Participants))
		for _, part := range ev.Participants {
			newParts = append(newParts, part)
		}
		ev.Participants = newParts
		loaded.Events[id] = ev
	}

	// 清理实体事件索引
	for eid, ent := range loaded.Entities {
		newEvs := make([]string, 0, len(ent.Events))
		for _, ev := range ent.Events {
			if ev != eventID {
				newEvs = append(newEvs, ev)
			}
		}
		ent.Events = newEvs
		loaded.Entities[eid] = ent
	}
	// 同步到冗余字段
	loaded.Entity = make(map[string]Entity, len(loaded.Entities))
	for k, v := range loaded.Entities {
		loaded.Entity[k] = v
	}

	delete(loaded.Events, eventID)
	loaded.ProjectMeta.EventCount = len(loaded.Events)
	loaded.ProjectMeta.UpdatedAt = time.Now()
	return loaded.SaveProjectObj()
}

// SetEventLocked 设置事件锁定状态
func (p *ProjectObj) SetEventLocked(rootDir string, projectID string, eventID string, locked bool) (*Event, error) {
	target := resolveRootDir(rootDir, projectID)
	loaded, err := loadProject(target)
	if err != nil {
		return nil, err
	}
	ev, ok := loaded.Events[eventID]
	if !ok {
		return nil, fmt.Errorf("事件不存在: %s", eventID)
	}
	ev.Locked = locked
	loaded.Events[eventID] = ev
	loaded.ProjectMeta.UpdatedAt = time.Now()
	if err := loaded.SaveProjectObj(); err != nil {
		return nil, err
	}
	return &ev, nil
}

// UpdateEventParticipants 更新事件参与者
func (p *ProjectObj) UpdateEventParticipants(rootDir string, projectID string, eventID string, participants []Participant) (*Event, error) {
	target := resolveRootDir(rootDir, projectID)
	loaded, err := loadProject(target)
	if err != nil {
		return nil, err
	}
	ev, ok := loaded.Events[eventID]
	if !ok {
		return nil, fmt.Errorf("事件不存在: %s", eventID)
	}
	// 校验所有参与者引用的实体存在
	for _, part := range participants {
		if _, ok := loaded.Entities[part.EntityID]; !ok {
			return nil, fmt.Errorf("参与者引用了不存在的实体: %s", part.EntityID)
		}
	}
	ev.Participants = participants
	loaded.Events[eventID] = ev

	// 同步实体的 Events 索引
	for _, part := range participants {
		if ent, ok := loaded.Entities[part.EntityID]; ok {
			found := false
			for _, e := range ent.Events {
				if e == eventID {
					found = true
					break
				}
			}
			if !found {
				ent.Events = append(ent.Events, eventID)
				loaded.Entities[part.EntityID] = ent
			}
		}
	}
	loaded.Entity = make(map[string]Entity, len(loaded.Entities))
	for k, v := range loaded.Entities {
		loaded.Entity[k] = v
	}

	loaded.ProjectMeta.UpdatedAt = time.Now()
	if err := loaded.SaveProjectObj(); err != nil {
		return nil, err
	}
	return &ev, nil
}

// CreateEventEdge 创建事件因果边
func (p *ProjectObj) CreateEventEdge(rootDir string, projectID string, sourceEventID string, targetEventID string, edgeType EdgeType, mirrorInEdge bool) error {
	target := resolveRootDir(rootDir, projectID)
	loaded, err := loadProject(target)
	if err != nil {
		return err
	}
	src, ok := loaded.Events[sourceEventID]
	if !ok {
		return fmt.Errorf("起点事件不存在: %s", sourceEventID)
	}
	tgt, ok := loaded.Events[targetEventID]
	if !ok {
		return fmt.Errorf("终点事件不存在: %s", targetEventID)
	}
	// 检查重复
	for _, e := range src.OutEdges {
		if e.Target == targetEventID && e.Type == edgeType {
			return fmt.Errorf("边已存在: %s -> %s (%s)", sourceEventID, targetEventID, edgeType)
		}
	}
	src.OutEdges = append(src.OutEdges, EventEdge{Target: targetEventID, Type: edgeType, Discribe: map[string]string{}})
	loaded.Events[sourceEventID] = src

	if mirrorInEdge {
		for _, e := range tgt.InEdges {
			if e.Target == sourceEventID && e.Type == edgeType {
				loaded.Events[targetEventID] = tgt
				loaded.ProjectMeta.UpdatedAt = time.Now()
				return loaded.SaveProjectObj()
			}
		}
		tgt.InEdges = append(tgt.InEdges, EventEdge{Target: sourceEventID, Type: edgeType, Discribe: map[string]string{}})
		loaded.Events[targetEventID] = tgt
	}
	loaded.ProjectMeta.UpdatedAt = time.Now()
	return loaded.SaveProjectObj()
}

// UpdateEventEdge 更新事件因果边
func (p *ProjectObj) UpdateEventEdge(rootDir string, projectID string, sourceEventID string, targetEventID string, newTargetEventID string, edgeType EdgeType, mirrorInEdge bool) error {
	target := resolveRootDir(rootDir, projectID)
	loaded, err := loadProject(target)
	if err != nil {
		return err
	}
	src, ok := loaded.Events[sourceEventID]
	if !ok {
		return fmt.Errorf("起点事件不存在: %s", sourceEventID)
	}
	if _, ok := loaded.Events[newTargetEventID]; !ok {
		return fmt.Errorf("新终点事件不存在: %s", newTargetEventID)
	}
	// 替换 src.OutEdges 中的目标
	found := false
	newOut := make([]EventEdge, 0, len(src.OutEdges))
	for _, e := range src.OutEdges {
		if e.Target == targetEventID && e.Type == edgeType {
			newOut = append(newOut, EventEdge{Target: newTargetEventID, Type: edgeType, Discribe: e.Discribe})
			found = true
		} else {
			newOut = append(newOut, e)
		}
	}
	if !found {
		return fmt.Errorf("未找到原边: %s -> %s (%s)", sourceEventID, targetEventID, edgeType)
	}
	src.OutEdges = newOut
	loaded.Events[sourceEventID] = src

	if mirrorInEdge {
		// 移除旧入边
		if tgt, ok := loaded.Events[targetEventID]; ok {
			newIn := make([]EventEdge, 0, len(tgt.InEdges))
			for _, e := range tgt.InEdges {
				if !(e.Target == sourceEventID && e.Type == edgeType) {
					newIn = append(newIn, e)
				}
			}
			tgt.InEdges = newIn
			loaded.Events[targetEventID] = tgt
		}
		// 添加新入边
		if tgt, ok := loaded.Events[newTargetEventID]; ok {
			tgt.InEdges = append(tgt.InEdges, EventEdge{Target: sourceEventID, Type: edgeType, Discribe: map[string]string{}})
			loaded.Events[newTargetEventID] = tgt
		}
	}
	loaded.ProjectMeta.UpdatedAt = time.Now()
	return loaded.SaveProjectObj()
}

// DeleteEventEdge 删除事件因果边
func (p *ProjectObj) DeleteEventEdge(rootDir string, projectID string, sourceEventID string, targetEventID string, edgeType EdgeType, mirrorInEdge bool) error {
	target := resolveRootDir(rootDir, projectID)
	loaded, err := loadProject(target)
	if err != nil {
		return err
	}
	src, ok := loaded.Events[sourceEventID]
	if !ok {
		return fmt.Errorf("起点事件不存在: %s", sourceEventID)
	}
	newOut := make([]EventEdge, 0, len(src.OutEdges))
	for _, e := range src.OutEdges {
		if !(e.Target == targetEventID && e.Type == edgeType) {
			newOut = append(newOut, e)
		}
	}
	src.OutEdges = newOut
	loaded.Events[sourceEventID] = src

	if mirrorInEdge {
		if tgt, ok := loaded.Events[targetEventID]; ok {
			newIn := make([]EventEdge, 0, len(tgt.InEdges))
			for _, e := range tgt.InEdges {
				if !(e.Target == sourceEventID && e.Type == edgeType) {
					newIn = append(newIn, e)
				}
			}
			tgt.InEdges = newIn
			loaded.Events[targetEventID] = tgt
		}
	}
	loaded.ProjectMeta.UpdatedAt = time.Now()
	return loaded.SaveProjectObj()
}

// AttachSubEvent 建立父子包含关系
func (p *ProjectObj) AttachSubEvent(rootDir string, projectID string, parentEventID string, childEventID string, index int) error {
	target := resolveRootDir(rootDir, projectID)
	loaded, err := loadProject(target)
	if err != nil {
		return err
	}
	parent, ok := loaded.Events[parentEventID]
	if !ok {
		return fmt.Errorf("父事件不存在: %s", parentEventID)
	}
	if _, ok := loaded.Events[childEventID]; !ok {
		return fmt.Errorf("子事件不存在: %s", childEventID)
	}
	// 检查重复
	for _, sub := range parent.SubEvents {
		if sub == childEventID {
			return fmt.Errorf("子事件已挂载: %s", childEventID)
		}
	}
	if index < 0 || index >= len(parent.SubEvents) {
		parent.SubEvents = append(parent.SubEvents, childEventID)
	} else {
		parent.SubEvents = append(parent.SubEvents[:index], append([]string{childEventID}, parent.SubEvents[index:]...)...)
	}
	loaded.Events[parentEventID] = parent
	loaded.ProjectMeta.UpdatedAt = time.Now()
	return loaded.SaveProjectObj()
}

// MoveSubEvent 调整子事件顺序
func (p *ProjectObj) MoveSubEvent(rootDir string, projectID string, parentEventID string, childEventID string, index int) error {
	target := resolveRootDir(rootDir, projectID)
	loaded, err := loadProject(target)
	if err != nil {
		return err
	}
	parent, ok := loaded.Events[parentEventID]
	if !ok {
		return fmt.Errorf("父事件不存在: %s", parentEventID)
	}
	curIdx := -1
	for i, sub := range parent.SubEvents {
		if sub == childEventID {
			curIdx = i
			break
		}
	}
	if curIdx < 0 {
		return fmt.Errorf("子事件未挂载: %s", childEventID)
	}
	parent.SubEvents = append(parent.SubEvents[:curIdx], parent.SubEvents[curIdx+1:]...)
	if index < 0 || index >= len(parent.SubEvents)+1 {
		parent.SubEvents = append(parent.SubEvents, childEventID)
	} else {
		parent.SubEvents = append(parent.SubEvents[:index], append([]string{childEventID}, parent.SubEvents[index:]...)...)
	}
	loaded.Events[parentEventID] = parent
	loaded.ProjectMeta.UpdatedAt = time.Now()
	return loaded.SaveProjectObj()
}

// DetachSubEvent 解除父子包含关系
func (p *ProjectObj) DetachSubEvent(rootDir string, projectID string, parentEventID string, childEventID string) error {
	target := resolveRootDir(rootDir, projectID)
	loaded, err := loadProject(target)
	if err != nil {
		return err
	}
	parent, ok := loaded.Events[parentEventID]
	if !ok {
		return fmt.Errorf("父事件不存在: %s", parentEventID)
	}
	newSubs := make([]string, 0, len(parent.SubEvents))
	for _, sub := range parent.SubEvents {
		if sub != childEventID {
			newSubs = append(newSubs, sub)
		}
	}
	parent.SubEvents = newSubs
	loaded.Events[parentEventID] = parent
	loaded.ProjectMeta.UpdatedAt = time.Now()
	return loaded.SaveProjectObj()
}

// ListEntities 获取工程下全部实体
func (p *ProjectObj) ListEntities(rootDir string, projectID string) ([]Entity, error) {
	target := resolveRootDir(rootDir, projectID)
	loaded, err := loadProject(target)
	if err != nil {
		return nil, err
	}
	list := make([]Entity, 0, len(loaded.Entities))
	for _, e := range loaded.Entities {
		list = append(list, e)
	}
	sort.Slice(list, func(i, j int) bool { return list[i].ID < list[j].ID })
	return list, nil
}

// GetEntity 获取单个实体详情
func (p *ProjectObj) GetEntity(rootDir string, projectID string, entityID string) (*Entity, error) {
	target := resolveRootDir(rootDir, projectID)
	loaded, err := loadProject(target)
	if err != nil {
		return nil, err
	}
	ent, ok := loaded.Entities[entityID]
	if !ok {
		return nil, fmt.Errorf("实体不存在: %s", entityID)
	}
	return &ent, nil
}

// CreateEntity 创建实体节点
func (p *ProjectObj) CreateEntity(rootDir string, projectID string, entity Entity) (*Entity, error) {
	target := resolveRootDir(rootDir, projectID)
	loaded, err := loadProject(target)
	if err != nil {
		return nil, err
	}
	ensureEntityID(&entity)
	if _, exists := loaded.Entities[entity.ID]; exists {
		return nil, fmt.Errorf("实体 ID 已存在: %s", entity.ID)
	}
	loaded.Entities[entity.ID] = entity
	loaded.ProjectMeta.EntityCount = len(loaded.Entities)
	loaded.ProjectMeta.UpdatedAt = time.Now()
	// 同步冗余字段
	loaded.Entity[entity.ID] = entity
	if err := loaded.SaveProjectObj(); err != nil {
		return nil, err
	}
	return &entity, nil
}

// UpdateEntity 更新实体节点
func (p *ProjectObj) UpdateEntity(rootDir string, projectID string, entityID string, entity Entity) (*Entity, error) {
	target := resolveRootDir(rootDir, projectID)
	loaded, err := loadProject(target)
	if err != nil {
		return nil, err
	}
	if _, ok := loaded.Entities[entityID]; !ok {
		return nil, fmt.Errorf("实体不存在: %s", entityID)
	}
	entity.ID = entityID
	ensureEntityID(&entity)
	loaded.Entities[entityID] = entity
	loaded.Entity[entityID] = entity
	loaded.ProjectMeta.UpdatedAt = time.Now()
	if err := loaded.SaveProjectObj(); err != nil {
		return nil, err
	}
	return &entity, nil
}

// DeleteEntity 删除实体节点
func (p *ProjectObj) DeleteEntity(rootDir string, projectID string, entityID string) error {
	target := resolveRootDir(rootDir, projectID)
	loaded, err := loadProject(target)
	if err != nil {
		return err
	}
	if _, ok := loaded.Entities[entityID]; !ok {
		return fmt.Errorf("实体不存在: %s", entityID)
	}
	// 清理参与者引用
	for id, ev := range loaded.Events {
		newParts := make([]Participant, 0, len(ev.Participants))
		for _, part := range ev.Participants {
			if part.EntityID != entityID {
				newParts = append(newParts, part)
			}
		}
		ev.Participants = newParts
		loaded.Events[id] = ev
	}
	// 清理实体关系引用
	for id, ent := range loaded.Entities {
		if id == entityID {
			continue
		}
		newRels := make([]EntityRelationship, 0, len(ent.Relationships))
		for _, r := range ent.Relationships {
			if r.TargetID != entityID {
				newRels = append(newRels, r)
			}
		}
		ent.Relationships = newRels
		loaded.Entities[id] = ent
	}
	delete(loaded.Entities, entityID)
	delete(loaded.Entity, entityID)
	delete(loaded.Participant, entityID)
	loaded.ProjectMeta.EntityCount = len(loaded.Entities)
	loaded.ProjectMeta.UpdatedAt = time.Now()
	return loaded.SaveProjectObj()
}

// GetSimulation 获取模拟会话详情
func (p *ProjectObj) GetSimulation(sessionID string) (*SimulationSession, error) {
	storeMu.RLock()
	defer storeMu.RUnlock()
	s, ok := simulationStore[sessionID]
	if !ok {
		return nil, fmt.Errorf("模拟会话不存在: %s", sessionID)
	}
	return s, nil
}

// GetGenerationTask 查询正文生成任务
func (p *ProjectObj) GetGenerationTask(taskID string) (*GenerationTask, error) {
	storeMu.RLock()
	defer storeMu.RUnlock()
	t, ok := generationStore[taskID]
	if !ok {
		return nil, fmt.Errorf("生成任务不存在: %s", taskID)
	}
	return t, nil
}

// =====================================================================================
// FrontendExecutionService 实现
// =====================================================================================

// TODO: ExtractFromSource 从原始文本提取结构化工程初稿
// 说明: 此处为占位实现, 实际项目中应调用 AI 模型完成文本到结构化数据的转换。
func (p *ProjectObj) ExtractFromSource(rootDir string, title string, sourceText string, writeRules []string, worldRules []string, autoCreate bool) (*ExtractSourceResult, error) {
	if strings.TrimSpace(title) == "" {
		return nil, errors.New("工程标题为空")
	}
	if strings.TrimSpace(sourceText) == "" {
		return nil, errors.New("原始文本为空")
	}
	meta := ProjectMeta{
		RootDir:   rootDir,
		Title:     title,
		Author:    "AI 提取",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		Settings:  map[string]string{"source": "extract"},
	}
	if autoCreate {
		if _, err := p.CreateProject(meta); err != nil {
			return nil, err
		}
	}
	// 占位返回, 真实场景下应调用 LLM 抽取 events/entities/edges/participants
	return &ExtractSourceResult{
		Meta: meta,
	}, errors.New("ExtractFromSource 需要接入 AI 模型, 当前为占位实现")
}

// GetStoryGraph 获取画布图数据
func (p *ProjectObj) GetStoryGraph(rootDir string, projectID string) (*StoryGraph, error) {
	target := resolveRootDir(rootDir, projectID)
	loaded, err := loadProject(target)
	if err != nil {
		return nil, err
	}
	graph := &StoryGraph{
		ProjectID: projectID,
		Spine:     append([]string{}, loaded.WorldSetting.Spine...),
	}
	for _, e := range loaded.Events {
		graph.Events = append(graph.Events, e)
	}
	sort.Slice(graph.Events, func(i, j int) bool { return graph.Events[i].ID < graph.Events[j].ID })
	for _, e := range loaded.Entities {
		graph.Entities = append(graph.Entities, e)
	}
	sort.Slice(graph.Entities, func(i, j int) bool { return graph.Entities[i].ID < graph.Entities[j].ID })
	return graph, nil
}

// ValidateProject 校验工程结构
func (p *ProjectObj) ValidateProject(rootDir string, projectID string) (*ValidationReport, error) {
	target := resolveRootDir(rootDir, projectID)
	loaded, err := loadProject(target)
	if err != nil {
		return nil, err
	}
	report := &ValidationReport{ProjectID: projectID, Passed: true}

	// 1. 检查叙事链中的事件必须存在
	for _, id := range loaded.WorldSetting.Spine {
		if _, ok := loaded.Events[id]; !ok {
			report.Issues = append(report.Issues, ValidationIssue{
				Code:     "spine_missing_event",
				Level:    "error",
				Message:  fmt.Sprintf("叙事链引用了不存在的事件: %s", id),
				EventIDs: []string{id},
			})
			report.Passed = false
		}
	}

	// 2. 检查事件边与子事件引用
	for id, ev := range loaded.Events {
		for _, e := range ev.InEdges {
			if _, ok := loaded.Events[e.Target]; !ok {
				report.Issues = append(report.Issues, ValidationIssue{
					Code:     "edge_missing_target",
					Level:    "error",
					Message:  fmt.Sprintf("事件 %s 的入边指向不存在的事件: %s", id, e.Target),
					EventIDs: []string{id, e.Target},
				})
				report.Passed = false
			}
		}
		for _, e := range ev.OutEdges {
			if _, ok := loaded.Events[e.Target]; !ok {
				report.Issues = append(report.Issues, ValidationIssue{
					Code:     "edge_missing_target",
					Level:    "error",
					Message:  fmt.Sprintf("事件 %s 的出边指向不存在的事件: %s", id, e.Target),
					EventIDs: []string{id, e.Target},
				})
				report.Passed = false
			}
		}
		for _, sub := range ev.SubEvents {
			if _, ok := loaded.Events[sub]; !ok {
				report.Issues = append(report.Issues, ValidationIssue{
					Code:     "sub_missing_event",
					Level:    "error",
					Message:  fmt.Sprintf("事件 %s 的子事件不存在: %s", id, sub),
					EventIDs: []string{id, sub},
				})
				report.Passed = false
			}
		}
		for _, part := range ev.Participants {
			if _, ok := loaded.Entities[part.EntityID]; !ok {
				report.Issues = append(report.Issues, ValidationIssue{
					Code:      "participant_missing_entity",
					Level:     "error",
					Message:   fmt.Sprintf("事件 %s 的参与者引用了不存在的实体: %s", id, part.EntityID),
					EventIDs:  []string{id},
					EntityIDs: []string{part.EntityID},
				})
				report.Passed = false
			}
		}
	}

	// 3. 因果断裂检测: 所有事件都应至少有 1 条入边或位于 spine 起始位置
	inDegree := make(map[string]int)
	for _, ev := range loaded.Events {
		for _, e := range ev.OutEdges {
			inDegree[e.Target]++
		}
	}
	for id := range loaded.Events {
		if inDegree[id] == 0 {
			isSpineHead := false
			for _, sid := range loaded.WorldSetting.Spine {
				if sid == id {
					isSpineHead = true
					break
				}
			}
			if !isSpineHead {
				report.Issues = append(report.Issues, ValidationIssue{
					Code:     "orphan_event",
					Level:    "warning",
					Message:  fmt.Sprintf("事件 %s 没有入边且不在叙事链头部", id),
					EventIDs: []string{id},
				})
			}
		}
	}

	// 4. 环检测 (在 spine 上)
	spineVisited := make(map[string]bool)
	for _, id := range loaded.WorldSetting.Spine {
		if spineVisited[id] {
			report.Issues = append(report.Issues, ValidationIssue{
				Code:     "spine_cycle",
				Level:    "error",
				Message:  fmt.Sprintf("叙事链存在重复/环: %s", id),
				EventIDs: []string{id},
			})
			report.Passed = false
		}
		spineVisited[id] = true
	}
	return report, nil
}

// writeCompileResult 把编译结果写到磁盘并填充 CompileResult
func (p *ProjectObj) writeCompileResult(target string, markdown string, report *ValidationReport, outputPath string) (*CompileResult, error) {
	res := &CompileResult{
		ProjectID:  target,
		OutputPath: outputPath,
		Markdown:   markdown,
		Report:     *report,
	}
	if outputPath == "" {
		outputPath = filepath.Join(filepath.Dir(target), filepath.Base(target)+".md")
		res.OutputPath = outputPath
	}
	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return nil, fmt.Errorf("创建输出目录失败: %w", err)
	}
	if err := os.WriteFile(outputPath, []byte(markdown), 0644); err != nil {
		return nil, fmt.Errorf("写入输出文件失败: %w", err)
	}
	return res, nil
}

// StartSimulation 开始一次模拟推演
func (p *ProjectObj) StartSimulation(rootDir string, projectID string, mode SimulationMode, startEventID string, endEventID string, parentEventID string, participantIDs []string, userPrompt string, maxRounds int) (*SimulationSession, error) {
	target := resolveRootDir(rootDir, projectID)
	if _, err := loadProject(target); err != nil {
		return nil, err
	}
	if maxRounds <= 0 {
		maxRounds = 5
	}
	session := &SimulationSession{
		ID:          fmt.Sprintf("sim_%d", time.Now().UnixNano()),
		ProjectID:   projectID,
		Status:      SimulationStatusRunning,
		Mode:        mode,
		Steps:       []SimulationStep{},
		DraftEvents: []Event{},
	}
	storeMu.Lock()
	simulationStore[session.ID] = session
	storeMu.Unlock()
	logrus.Infof("启动模拟会话 %s, 模式=%s, 起始=%s, 结束=%s, 父=%s, 参与者=%v, 提示=%q, 最大轮数=%d",
		session.ID, mode, startEventID, endEventID, parentEventID, participantIDs, userPrompt, maxRounds)
	return session, nil
}

// AdvanceSimulation 推进下一轮模拟
func (p *ProjectObj) AdvanceSimulation(sessionID string, userPrompt string, manualActor string) (*SimulationStep, error) {
	storeMu.Lock()
	defer storeMu.Unlock()
	session, ok := simulationStore[sessionID]
	if !ok {
		return nil, fmt.Errorf("模拟会话不存在: %s", sessionID)
	}
	if session.Status != SimulationStatusRunning {
		return nil, fmt.Errorf("模拟会话状态非运行中: %s", session.Status)
	}
	step := &SimulationStep{
		ID:         fmt.Sprintf("step_%d", time.Now().UnixNano()),
		Round:      len(session.Steps) + 1,
		ActorID:    manualActor,
		Action:     userPrompt,
		Observer:   "占位观察者评语, 实际项目中应由 AI 模型生成",
		Accepted:   false,
		SceneState: map[string]any{},
	}
	session.Steps = append(session.Steps, *step)
	return step, nil
}

// ReviewSimulation 审查某一步模拟结果
func (p *ProjectObj) ReviewSimulation(sessionID string, stepID string, approved bool, comment string) (*SimulationStep, error) {
	storeMu.Lock()
	defer storeMu.Unlock()
	session, ok := simulationStore[sessionID]
	if !ok {
		return nil, fmt.Errorf("模拟会话不存在: %s", sessionID)
	}
	for i := range session.Steps {
		if session.Steps[i].ID == stepID {
			session.Steps[i].Accepted = approved
			if comment != "" {
				session.Steps[i].Observer = comment
			}
			return &session.Steps[i], nil
		}
	}
	return nil, fmt.Errorf("步骤不存在: %s", stepID)
}

// ApplySimulationResult 落地模拟结果
func (p *ProjectObj) ApplySimulationResult(sessionID string, mode SimulationApplyMode, targetEventID string, parentEventID string, insertIndex int) (*SimulationApplyResult, error) {
	storeMu.RLock()
	session, ok := simulationStore[sessionID]
	storeMu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("模拟会话不存在: %s", sessionID)
	}
	if len(session.DraftEvents) == 0 {
		return &SimulationApplyResult{SessionID: sessionID, AffectedEventIDs: []string{}}, nil
	}
	// 占位: 实际应根据 mode 处理 session.DraftEvents
	return &SimulationApplyResult{
		SessionID:        sessionID,
		AffectedEventIDs: []string{},
	}, errors.New("ApplySimulationResult 需要根据 mode 处理草稿事件, 当前为占位实现")
}

// CancelSimulation 取消模拟会话
func (p *ProjectObj) CancelSimulation(sessionID string) error {
	storeMu.Lock()
	defer storeMu.Unlock()
	session, ok := simulationStore[sessionID]
	if !ok {
		return fmt.Errorf("模拟会话不存在: %s", sessionID)
	}
	session.Status = SimulationStatusCanceled
	return nil
}

// GenerateDraft 创建正文生成任务
func (p *ProjectObj) GenerateDraft(rootDir string, projectID string, eventIDs []string, outputPath string, model string, systemPrompt string) (*GenerationTask, error) {
	target := resolveRootDir(rootDir, projectID)
	if _, err := loadProject(target); err != nil {
		return nil, err
	}
	task := &GenerationTask{
		ID:         fmt.Sprintf("gen_%d", time.Now().UnixNano()),
		ProjectID:  projectID,
		Status:     GenerationStatusPending,
		OutputPath: outputPath,
	}
	storeMu.Lock()
	generationStore[task.ID] = task
	storeMu.Unlock()
	logrus.Infof("创建生成任务 %s, model=%s, 系统提示长度=%d, 输出=%s", task.ID, model, len(systemPrompt), outputPath)
	return task, nil
}

// OptimizeNodeExpression 优化一个或多个节点的表达
func (p *ProjectObj) OptimizeNodeExpression(rootDir string, projectID string, eventIDs []string, instruction string, focusFields []string, keepMeaning bool, rewriteStyle string) (*OptimizeNodeExpressionResult, error) {
	target := resolveRootDir(rootDir, projectID)
	loaded, err := loadProject(target)
	if err != nil {
		return nil, err
	}
	result := &OptimizeNodeExpressionResult{ProjectID: projectID, Drafts: []OptimizedEventDraft{}}
	for _, id := range eventIDs {
		ev, ok := loaded.Events[id]
		if !ok {
			continue
		}
		result.Drafts = append(result.Drafts, OptimizedEventDraft{
			EventID:               id,
			OriginalIntroduction:  ev.Introduction,
			OptimizedIntroduction: ev.Introduction,
			OriginalProcess:       ev.Process,
			OptimizedProcess:      ev.Process,
			OriginalOutcome:       formatMap(ev.Outcome),
			OptimizedOutcome:      formatMap(ev.Outcome),
			Comment:               fmt.Sprintf("占位优化结果, instruction=%q, keepMeaning=%v, rewriteStyle=%q", instruction, keepMeaning, rewriteStyle),
		})
	}
	if len(result.Drafts) == 0 {
		return result, errors.New("未匹配到任何有效事件")
	}
	return result, nil
}

func formatMap(m map[string]string) string {
	if len(m) == 0 {
		return ""
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, k := range keys {
		parts = append(parts, fmt.Sprintf("%s:%s", k, m[k]))
	}
	return strings.Join(parts, "; ")
}

// ApplyOptimizedNodes 应用节点优化结果
func (p *ProjectObj) ApplyOptimizedNodes(rootDir string, projectID string, drafts []OptimizedEventDraft) (*ApplyOptimizedNodesResult, error) {
	target := resolveRootDir(rootDir, projectID)
	loaded, err := loadProject(target)
	if err != nil {
		return nil, err
	}
	affected := make([]string, 0, len(drafts))
	for _, d := range drafts {
		ev, ok := loaded.Events[d.EventID]
		if !ok {
			continue
		}
		if d.OptimizedIntroduction != "" {
			ev.Introduction = d.OptimizedIntroduction
		}
		if d.OptimizedProcess != "" {
			ev.Process = d.OptimizedProcess
		}
		loaded.Events[d.EventID] = ev
		affected = append(affected, d.EventID)
	}
	if len(affected) > 0 {
		loaded.ProjectMeta.UpdatedAt = time.Now()
		if err := loaded.SaveProjectObj(); err != nil {
			return nil, err
		}
	}
	return &ApplyOptimizedNodesResult{ProjectID: projectID, AffectedEventIDs: affected}, nil
}

// =====================================================================================
// 接口实现校验常量
// 说明: 编译期检查 *ProjectObj 是否实现了 interface.go 中定义的所有接口。
// =====================================================================================

// var (
// 	_ FrontendCRUDService = (*ProjectObj)(nil)
// )

//TODO:等待验证,写个单元测试
