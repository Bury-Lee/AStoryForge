package story_struct

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/sirupsen/logrus"
)

// 这里存放工程时的具体实例
type ProjectObj struct {
	ProjectMeta  ProjectMeta            `json:"project_meta"`  // 工程元数据
	WorldSetting WorldSetting           `json:"world_setting"` // 世界设置
	Events       map[string]Event       `json:"events"`        // 事件列表,键为事件ID
	Entities     map[string]Entity      `json:"entities"`      // 实体列表,键为实体ID
	Participant  map[string]Participant `json:"participant"`   // 参与事件索引,键为参与事件ID,值为参与事件
	Entity       map[string]Entity      `json:"entity"`        // 实体列表,键为实体ID
}

var ProjectObject *ProjectObj

func (p *ProjectObj) InitProjectObj(path string) ProjectObj { //初始化函数
	p = &ProjectObj{}
	p.ProjectMeta = ProjectMeta{}
	p.WorldSetting = WorldSetting{}
	p.Events = make(map[string]Event)
	p.Entities = make(map[string]Entity)
	p.Participant = make(map[string]Participant)
	p.Entity = make(map[string]Entity)

	if path == "" {
		if p.ProjectMeta.RootDir == "" {
			logrus.Error("工程文件路径未设置")
			return *p
		}
		path = p.ProjectMeta.RootDir
	}
	// 从本地文件加载
	if err := p.LoadProjectObj(path); err != nil {
		logrus.Errorf("加载工程文件失败: %v", err)
		return *p
	}
	return *p
}

func (p *ProjectObj) DeferProjectMeta() { //程序结束时的工作
	p.SaveProjectObj()
}

func (p *ProjectObj) UpdateProjectObj(ProjectObj ProjectObj) { //TODO:增量更新
	if ProjectObj.ProjectMeta.Title != "" {
		p.ProjectMeta.Title = ProjectObj.ProjectMeta.Title
	}
	//...像这样更新其他字段
}

func (p *ProjectObj) SaveProjectObj() error {
	// 反序列化保存工程对象
	jsonData, err := json.Marshal(p)
	if err != nil {
		logrus.Errorf("序列化工程对象失败: %v", err)
		return err
	}

	// 打印JSON字符串
	logrus.Debugf("保存的工程对象: %s", string(jsonData))

	// 保存到本地文件
	if p.ProjectMeta.RootDir == "" {
		logrus.Error("工程文件路径未设置")
		return fmt.Errorf("工程文件路径未设置")
	}

	// 确保目录存在
	dir := filepath.Dir(p.ProjectMeta.RootDir)
	if err := os.MkdirAll(dir, 0755); err != nil {
		logrus.Errorf("创建目录失败 %s: %v", dir, err)
		return err
	}

	// 写入文件（0644权限：所有者可读写，其他人只读）
	if err := os.WriteFile(p.ProjectMeta.RootDir, jsonData, 0644); err != nil {
		logrus.Errorf("保存工程文件失败 %s: %v", p.ProjectMeta.RootDir, err)
		return fmt.Errorf("保存工程文件失败 %s: %w", p.ProjectMeta.RootDir, err)
	}

	logrus.Infof("工程已保存到: %s", p.ProjectMeta.RootDir)
	return nil
}

func (p *ProjectObj) LoadProjectObj(path string) error {
	// 从本地文件加载
	if path == "" {
		if p.ProjectMeta.RootDir == "" {
			return fmt.Errorf("工程文件路径未设置")
		}
		path = p.ProjectMeta.RootDir
	}
	jsonData, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			// logrus.Warnf("工程文件不存在: %s", path)
			return fmt.Errorf("工程文件不存在: %s", path)
		}
		return fmt.Errorf("读取文件失败: %w", err)
	}

	if err := json.Unmarshal(jsonData, p); err != nil {
		return fmt.Errorf("反序列化失败: %w", err)
	}
	logrus.Debugf("工程 %s 已加载,路径: %s", p.ProjectMeta.Title, path)
	return nil
}

func (p *ProjectObj) Copy() ProjectObj {
	var newProjectObj ProjectObj
	newProjectObj.ProjectMeta = p.ProjectMeta
	newProjectObj.WorldSetting = p.WorldSetting
	newProjectObj.Events = p.Events
	newProjectObj.Entities = p.Entities
	newProjectObj.Participant = p.Participant
	newProjectObj.Entity = p.Entity
	return newProjectObj
}

// MakeMermaid 生成 Mermaid 图
// 将 ProjectObj 中除 ProjectMeta 外的所有信息（WorldSetting、Events、Entities、Entity、Participant）
// 组织为完整的有向图 —— 每个对象的所有字段均写入节点标签，关系通过边表达。
// 前string是Mermaid图字符串,后string是输出给前端的错误信息,如果错误信息为空,则表示没有错误
func (p *ProjectObj) MakeMermaid() (string, error) {
	// 空对象检查
	if p == nil {
		return "", fmt.Errorf("空工程对象")
	}

	// ========== 1. 叙事链环检测 ==========
	Rain := make(map[string]bool)
	for _, e := range p.Events {
		if e.ID == "" {
			continue
		}
		if Rain[e.ID] {
			return "", fmt.Errorf("发现叙事链环,事件ID: %s", e.ID)
		}
		Rain[e.ID] = true
	}

	// ========== 2. 初始化字符串构建器 ==========
	var sb strings.Builder

	// 系统指令
	sb.WriteString("你是一名小说作家,请根据以下大纲完成小说内容:")

	// 写作约束
	if len(p.WorldSetting.WriteRules) > 0 {
		sb.WriteString("- **写作约束**：\n")
		for _, r := range p.WorldSetting.WriteRules {
			sb.WriteString("  - " + r + "\n")
		}
	}

	// 项目标题（markdown 文本，AI 上下文）
	sb.WriteString("# " + p.ProjectMeta.Title + "\n\n")

	// ========== 3. 数据整理：合并实体和事件 ==========
	entityMap := make(map[string]Entity)
	for key, entity := range p.Entities {
		entityID := entity.ID
		if entityID == "" {
			entityID = key
			entity.ID = entityID
		}
		entityMap[entityID] = entity
	}
	for key, entity := range p.Entity {
		entityID := entity.ID
		if entityID == "" {
			entityID = key
			entity.ID = entityID
		}
		if _, ok := entityMap[entityID]; !ok {
			entityMap[entityID] = entity
		}
	}

	eventMap := make(map[string]Event)
	for key, event := range p.Events {
		eventID := event.ID
		if eventID == "" {
			eventID = key
			event.ID = eventID
		}
		eventMap[eventID] = event
	}

	// ========== 4. 提取并排序所有 ID ==========
	eventIDs := make([]string, 0, len(eventMap))
	for eventID := range eventMap {
		eventIDs = append(eventIDs, eventID)
	}
	sort.Strings(eventIDs)

	entityIDs := make([]string, 0, len(entityMap))
	for entityID := range entityMap {
		entityIDs = append(entityIDs, entityID)
	}
	sort.Strings(entityIDs)

	partKeys := make([]string, 0, len(p.Participant))
	for k := range p.Participant {
		partKeys = append(partKeys, k)
	}
	sort.Strings(partKeys)

	// ========== 5. 辅助函数 ==========

	escapeNodeLabel := func(text string) string {
		text = strings.ReplaceAll(text, "\\", "\\\\")
		text = strings.ReplaceAll(text, "\"", "\\\"")
		text = strings.ReplaceAll(text, "\r\n", "<br/>")
		text = strings.ReplaceAll(text, "\n", "<br/>")
		return text
	}

	escapeEdgeLabel := func(text string) string {
		text = strings.ReplaceAll(text, "|", "/")
		text = strings.ReplaceAll(text, "\r\n", " ")
		text = strings.ReplaceAll(text, "\n", " ")
		text = strings.TrimSpace(text)
		return text
	}

	// map[string]string → "k1:v1, k2:v2"
	joinMap := func(m map[string]string) string {
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
			if v := strings.TrimSpace(m[k]); v != "" {
				parts = append(parts, k+":"+v)
			}
		}
		return strings.Join(parts, ", ")
	}

	// []string → "a; b; c"
	joinSlice := func(s []string) string {
		cleaned := make([]string, 0, len(s))
		for _, v := range s {
			if v = strings.TrimSpace(v); v != "" {
				cleaned = append(cleaned, v)
			}
		}
		return strings.Join(cleaned, "; ")
	}

	// map[string]string → "k1:v1 k2:v2"（边标签用）
	joinDetail := func(detail map[string]string) string {
		if len(detail) == 0 {
			return ""
		}
		keys := make([]string, 0, len(detail))
		for key := range detail {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		parts := make([]string, 0, len(keys))
		for _, key := range keys {
			if detail[key] == "" {
				continue
			}
			parts = append(parts, key+":"+detail[key])
		}
		return strings.Join(parts, " ")
	}

	// ========== 6. 节点 & 边管理 ==========

	nodeIDs := make(map[string]string)
	nodeLines := make([]string, 0)
	nodeAdded := make(map[string]bool)
	nodeIndex := 0

	addNode := func(prefix, rawID, label string) {
		if strings.TrimSpace(rawID) == "" {
			return
		}
		key := prefix + ":" + rawID
		if nid, ok := nodeIDs[key]; ok {
			if !nodeAdded[nid] {
				nodeLines = append(nodeLines, fmt.Sprintf("    %s[\"%s\"]", nid, escapeNodeLabel(label)))
				nodeAdded[nid] = true
			}
			return
		}
		nodeIndex++
		nid := fmt.Sprintf("N%d", nodeIndex)
		nodeIDs[key] = nid
		nodeLines = append(nodeLines, fmt.Sprintf("    %s[\"%s\"]", nid, escapeNodeLabel(label)))
		nodeAdded[nid] = true
	}

	addRoundNode := func(prefix, rawID, label string) {
		if strings.TrimSpace(rawID) == "" {
			return
		}
		key := prefix + ":" + rawID
		if nid, ok := nodeIDs[key]; ok {
			if !nodeAdded[nid] {
				nodeLines = append(nodeLines, fmt.Sprintf("    %s([\"%s\"])", nid, escapeNodeLabel(label)))
				nodeAdded[nid] = true
			}
			return
		}
		nodeIndex++
		nid := fmt.Sprintf("N%d", nodeIndex)
		nodeIDs[key] = nid
		nodeLines = append(nodeLines, fmt.Sprintf("    %s([\"%s\"])", nid, escapeNodeLabel(label)))
		nodeAdded[nid] = true
	}

	addStadiumNode := func(prefix, rawID, label string) {
		addRoundNode(prefix, rawID, label)
	}

	ensureNodeExists := func(prefix, rawID string) string {
		key := prefix + ":" + rawID
		if nid, ok := nodeIDs[key]; ok {
			return nid
		}
		nodeIndex++
		nid := fmt.Sprintf("N%d", nodeIndex)
		nodeIDs[key] = nid
		nodeLines = append(nodeLines, fmt.Sprintf("    %s[\"%s\"]", nid, escapeNodeLabel(rawID)))
		nodeAdded[nid] = true
		return nid
	}

	edgeLines := make([]string, 0)
	edgeAdded := make(map[string]bool)

	addEdge := func(srcPrefix, srcID, tgtPrefix, tgtID, label, lineType string) {
		if strings.TrimSpace(srcID) == "" || strings.TrimSpace(tgtID) == "" {
			return
		}
		srcNode := ensureNodeExists(srcPrefix, srcID)
		tgtNode := ensureNodeExists(tgtPrefix, tgtID)
		label = escapeEdgeLabel(label)
		edgeKey := fmt.Sprintf("%s|%s|%s|%s", srcNode, tgtNode, label, lineType)
		if edgeAdded[edgeKey] {
			return
		}
		edgeAdded[edgeKey] = true

		arrow := "-->"
		if lineType == "dotted" {
			arrow = "-.->"
		}
		if label != "" {
			edgeLines = append(edgeLines, fmt.Sprintf("    %s %s|%s| %s", srcNode, arrow, label, tgtNode))
		}
	}

	// ========== 7. 构建节点标签（含全部字段） ==========

	buildWorldSettingLabel := func(ws WorldSetting) string {
		var l strings.Builder
		l.WriteString("WorldSetting")
		if ts := strings.Join(ws.TimeSystem, ", "); ts != "" {
			l.WriteString("<br/>TimeSystem: " + ts)
		}
		if tl := joinMap(ws.TimeLabels); tl != "" {
			l.WriteString("<br/>TimeLabels: " + tl)
		}
		if sp := strings.Join(ws.Spine, ", "); sp != "" {
			l.WriteString("<br/>Spine: " + sp)
		}
		if wr := joinSlice(ws.WriteRules); wr != "" {
			l.WriteString("<br/>WriteRules: " + wr)
		}
		if wr := joinSlice(ws.WorldRules); wr != "" {
			l.WriteString("<br/>WorldRules: " + wr)
		}
		return l.String()
	}

	buildEventLabel := func(ev Event) string {
		var l strings.Builder
		l.WriteString("Event<br/>ID: " + ev.ID)
		if ev.Name != "" {
			l.WriteString("<br/>Name: " + ev.Name)
		}
		if ev.Type != "" {
			l.WriteString("<br/>Type: " + string(ev.Type))
		}
		if ev.Introduction != "" {
			l.WriteString("<br/>Introduction: " + ev.Introduction)
		}
		if ev.Process != "" {
			l.WriteString("<br/>Process: " + ev.Process)
		}
		if tm := joinMap(ev.Time); tm != "" {
			l.WriteString("<br/>Time: " + tm)
		}
		if ev.SettingRef != "" {
			l.WriteString("<br/>SettingRef: " + ev.SettingRef)
		}
		if sr := joinSlice(ev.SubRules); sr != "" {
			l.WriteString("<br/>SubRules: " + sr)
		}
		if oc := joinMap(ev.Outcome); oc != "" {
			l.WriteString("<br/>Outcome: " + oc)
		}
		l.WriteString(fmt.Sprintf("<br/>Locked: %t", ev.Locked))
		return l.String()
	}

	buildEntityLabel := func(ent Entity) string {
		var l strings.Builder
		l.WriteString("Entity<br/>ID: " + ent.ID)
		if ent.Name != "" {
			l.WriteString("<br/>Name: " + ent.Name)
		}
		if ent.Type != "" {
			l.WriteString("<br/>Type: " + ent.Type)
		}
		if len(ent.Introduction) > 0 {
			l.WriteString("<br/>Introduction: " + strings.Join(ent.Introduction, "；"))
		}
		if rr := joinSlice(ent.RuleRefs); rr != "" {
			l.WriteString("<br/>RuleRefs: " + rr)
		}
		return l.String()
	}

	buildParticipantLabel := func(key string, part Participant) string {
		var l strings.Builder
		l.WriteString("Participant<br/>Key: " + key)
		if part.EntityID != "" {
			l.WriteString("<br/>EntityID: " + part.EntityID)
		}
		if st := joinMap(part.State); st != "" {
			l.WriteString("<br/>State: " + st)
		}
		return l.String()
	}

	// ========== Phase 1：创建所有节点（全量标签） ==========

	addStadiumNode("WS", "worldsetting", buildWorldSettingLabel(p.WorldSetting))

	for _, eid := range eventIDs {
		addRoundNode("event", eid, buildEventLabel(eventMap[eid]))
	}

	for _, eid := range entityIDs {
		addNode("entity", eid, buildEntityLabel(entityMap[eid]))
	}

	for _, k := range partKeys {
		addStadiumNode("participant", k, buildParticipantLabel(k, p.Participant[k]))
	}

	// ========== Phase 2：添加所有关系边 ==========

	// 事件关系
	for _, eid := range eventIDs {
		ev := eventMap[eid]

		for _, edge := range ev.InEdges {
			lbl := string(edge.Type)
			if lbl == "" {
				lbl = "前因"
			}
			if d := joinDetail(edge.Discribe); d != "" {
				lbl += " " + d
			}
			addEdge("event", edge.Target, "event", eid, lbl, "solid")
		}

		for _, edge := range ev.OutEdges {
			lbl := string(edge.Type)
			if lbl == "" {
				lbl = "后果"
			}
			if d := joinDetail(edge.Discribe); d != "" {
				lbl += " " + d
			}
			addEdge("event", eid, "event", edge.Target, lbl, "solid")
		}

		for _, subID := range ev.SubEvents {
			addEdge("event", eid, "event", subID, "子事件", "dotted")
		}

		for _, part := range ev.Participants {
			addEdge("entity", part.EntityID, "event", eid, "参与", "dotted")
		}
	}

	// 主线
	for i := 0; i+1 < len(p.WorldSetting.Spine); i++ {
		addEdge("event", p.WorldSetting.Spine[i], "event", p.WorldSetting.Spine[i+1], "主线", "solid")
	}

	// 实体关系
	for _, eid := range entityIDs {
		ent := entityMap[eid]
		for _, rel := range ent.Relationships {
			lbl := rel.RelationType
			if lbl == "" {
				lbl = "关联"
			}
			if d := joinDetail(rel.Description); d != "" {
				lbl += " " + d
			}
			addEdge("entity", eid, "entity", rel.TargetID, lbl, "solid")
		}
		for _, evID := range ent.Events {
			addEdge("entity", eid, "event", evID, "关联事件", "dotted")
		}
	}

	// 顶层 Participant → 实体
	for _, k := range partKeys {
		part := p.Participant[k]
		if part.EntityID != "" {
			addEdge("participant", k, "entity", part.EntityID, "指向", "dotted")
		}
	}

	// ========== 8. 验证 ==========
	if len(edgeLines) == 0 {
		sb.WriteString("```mermaid\ngraph LR\n```\n")
		return sb.String(), fmt.Errorf("无有效边")
	}

	// ========== 9. 输出 Mermaid 图 ==========
	sb.WriteString("```mermaid\n")
	sb.WriteString("graph LR\n")
	sb.WriteString("    classDef event fill:#EAF3FF,stroke:#3B82F6,color:#111827;\n")
	sb.WriteString("    classDef entity fill:#ECFDF5,stroke:#10B981,color:#111827;\n")
	sb.WriteString("    classDef participant fill:#FFF7ED,stroke:#F97316,color:#111827;\n")
	sb.WriteString("    classDef world fill:#F3E8FF,stroke:#A855F7,color:#111827;\n")

	for _, line := range nodeLines {
		sb.WriteString(line + "\n")
	}
	for _, line := range edgeLines {
		sb.WriteString(line + "\n")
	}

	for _, eid := range eventIDs {
		if nid, ok := nodeIDs["event:"+eid]; ok {
			sb.WriteString(fmt.Sprintf("    class %s event;\n", nid))
		}
	}
	for _, eid := range entityIDs {
		if nid, ok := nodeIDs["entity:"+eid]; ok {
			sb.WriteString(fmt.Sprintf("    class %s entity;\n", nid))
		}
	}
	for _, k := range partKeys {
		if nid, ok := nodeIDs["participant:"+k]; ok {
			sb.WriteString(fmt.Sprintf("    class %s participant;\n", nid))
		}
	}
	if nid, ok := nodeIDs["WS:worldsetting"]; ok {
		sb.WriteString(fmt.Sprintf("    class %s world;\n", nid))
	}

	sb.WriteString("```")
	return sb.String(), nil
}

// MakeMarkdown 将工程对象编译为结构化 Markdown 大纲（可直接作为 AI 创作上下文）
// 该方法遍历 ProjectObj 中的所有数据（背景设定、实体、事件链等），
// 生成一个格式化的 Markdown 字符串，供 AI 或人类作者参考以进行小说创作。
func (p *ProjectObj) MakeMarkdown() (string, error) {
	//叙事链环检测
	Rain := make(map[string]bool)
	// 遍历所有事件，检查是否有环
	for _, e := range p.Events {
		if e.ID == "" {
			continue
		}
		if Rain[e.ID] {
			return "", fmt.Errorf("发现叙事链环,事件ID: %s", e.ID)
		}
		Rain[e.ID] = true
	}

	var sb strings.Builder // 使用 strings.Builder 高效拼接字符串

	// 添加一个系统指令，明确告知 AI 的角色和任务
	sb.WriteString("你是一名小说作家,请根据以下大纲完成小说内容:")
	// 输出写作约束（例如“避免现代词汇”、“第三人称”等）
	if len(p.WorldSetting.WriteRules) > 0 {
		sb.WriteString("- **写作约束**：\n")
		for _, r := range p.WorldSetting.WriteRules {
			sb.WriteString("  - " + r + "\n")
		}
	}
	// 输出项目标题，作为 Markdown 的一级标题
	sb.WriteString("# " + p.ProjectMeta.Title + "\n\n")

	// ========== 1. 背景设定 ==========
	sb.WriteString("## 背景设定\n\n")

	// 输出世界运行的核心规则（如魔法体系、物理法则等）
	if len(p.WorldSetting.WorldRules) > 0 {
		sb.WriteString("- **世界规则**：\n")
		for _, r := range p.WorldSetting.WorldRules {
			sb.WriteString("  - " + r + "\n")
		}
	}
	// 输出时间系统定义（如“纪元”、“年份”等维度）
	if len(p.WorldSetting.TimeSystem) > 0 {
		sb.WriteString("- **时间系统**：" + strings.Join(p.WorldSetting.TimeSystem, "、") + "\n")
	}
	// 输出时间标签的具体值（例如“纪元：新生纪元”）
	if len(p.WorldSetting.TimeLabels) > 0 {
		sb.WriteString("- **时间标签**：\n")
		for k, v := range p.WorldSetting.TimeLabels {
			sb.WriteString(fmt.Sprintf("  - %s: %s\n", k, v))
		}
	}

	sb.WriteString("\n")

	// ========== 2. 实体列表 ==========
	sb.WriteString("## 实体列表\n\n")
	if len(p.Entities) > 0 {
		// 生成一个 Markdown 表格，列出所有实体（角色、组织、地点等）
		sb.WriteString("| 实体ID | 名称 | 类型 | 简介 |\n")
		sb.WriteString("|--------|------|------|------|\n")
		for _, e := range p.Entities {
			intro := ""
			if len(e.Introduction) > 0 {
				intro = strings.Join(e.Introduction, "；") // 用分号拼接多条简介
			}
			sb.WriteString(fmt.Sprintf("| %s | %s | %s | %s |\n",
				e.ID, e.Name, e.Type, intro))
		}
		sb.WriteString("\n")
	}

	// 输出实体关系图，便于快速理解角色、组织、物品之间的连接
	relationVisited := make(map[string]bool)
	relationLines := make([]string, 0)
	nodeIDs := make(map[string]string)
	nodeIndex := 0
	getNodeID := func(entityID string) string {
		if nodeID, ok := nodeIDs[entityID]; ok {
			return nodeID
		}
		nodeIndex++
		nodeID := fmt.Sprintf("E%d", nodeIndex)
		nodeIDs[entityID] = nodeID
		return nodeID
	}
	escapeMermaidLabel := func(text string) string {
		return strings.ReplaceAll(text, "\"", "\\\"")
	}
	if len(p.Entities) > 0 {
		for sourceID, entity := range p.Entities {
			sourceName := entity.Name
			if sourceName == "" {
				sourceName = sourceID
			}
			for _, rel := range entity.Relationships {
				if rel.TargetID == "" {
					continue
				}
				target, ok := p.Entities[rel.TargetID]
				if !ok {
					continue
				}
				relationType := rel.RelationType
				if relationType == "" {
					relationType = "关联"
				}
				edgeKey := fmt.Sprintf("%s|%s|%s", sourceID, rel.TargetID, relationType)
				if relationVisited[edgeKey] {
					continue
				}
				relationVisited[edgeKey] = true

				targetName := target.Name
				if targetName == "" {
					targetName = rel.TargetID
				}
				relationLines = append(relationLines, fmt.Sprintf(
					"    %s[\"%s\"] -->|\"%s\"| %s[\"%s\"]",
					getNodeID(sourceID),
					escapeMermaidLabel(sourceName),
					escapeMermaidLabel(relationType),
					getNodeID(rel.TargetID),
					escapeMermaidLabel(targetName),
				))
			}
		}
	}
	if len(relationLines) > 0 {
		sb.WriteString("### 实体关系\n\n")
		sb.WriteString("```mermaid\n")
		sb.WriteString("graph LR\n")
		for _, line := range relationLines {
			sb.WriteString(line + "\n")
		}
		sb.WriteString("```\n\n")
	}

	// ========== 3. 事件链 ==========
	sb.WriteString("## 事件链(请按照如下顺序叙事故事)\n\n")
	visited := make(map[string]bool) // 记录已渲染的事件，防止循环引用和重复渲染

	// 首先按照世界线（Spine）的顺序渲染主事件流
	for _, evID := range p.WorldSetting.Spine {
		if _, ok := p.Events[evID]; !ok {
			continue // 如果事件不存在则跳过
		}
		p.renderEvent(&sb, evID, 0, visited) // 从深度0开始递归渲染
	}

	// 处理可能未被 spine 直接引用但存在于 sub_events 中的独立节点
	// 确保所有事件都被输出（例如未被主流程包含的分支事件）
	//但是一般认为这是用户未用上的,选择不渲染,进行warn警告
	// for evID := range p.Events {
	// 	if !visited[evID] {
	// 		p.renderEvent(&sb, evID, 0, visited)
	// 	}
	// }

	for evID := range p.Events {
		if !visited[evID] {
			sb.WriteString("[warn] 事件未在事件链中:")
			p.renderEvent(&sb, evID, 0, visited)
		}
	}

	// // ========== 4. 写作要求 ==========
	// sb.WriteString("## 写作要求\n\n")
	// if len(p.ProjectMeta.Settings) > 0 {
	// 	sb.WriteString("- **工程设置**：\n")
	// 	for k, v := range p.ProjectMeta.Settings {
	// 		sb.WriteString(fmt.Sprintf("  - %s: %s\n", k, v))
	// 	}
	// }
	// // 再次输出写作约束（与背景设定中的重复，这里强化提示）
	// if len(p.WorldSetting.WriteRules) > 0 {
	// 	sb.WriteString("- **写作要求**：\n")
	// 	for _, r := range p.WorldSetting.WriteRules {
	// 		sb.WriteString("  - " + r + "\n")
	// 	}
	// }
	// // 再次输出世界规则（同样强化提示）
	// if len(p.WorldSetting.WorldRules) > 0 {
	// 	sb.WriteString("- **世界规则**：\n")
	// 	for _, r := range p.WorldSetting.WorldRules {
	// 		sb.WriteString("  - " + r + "\n")
	// 	}
	// }
	// sb.WriteString("\n")

	return sb.String(), nil
}

// renderEvent 递归渲染事件节点（按树深度生成标题）
// 参数:
//   - sb: 字符串构建器，用于累积输出
//   - evID: 当前要渲染的事件ID
//   - depth: 当前事件在事件树中的深度（根深度为0），用于决定Markdown标题级别
//   - visited: 记录已访问过的事件ID，防止无限递归
func (p *ProjectObj) renderEvent(sb *strings.Builder, evID string, depth int, visited map[string]bool) {
	// 如果事件已经渲染过，则直接返回，避免循环引用
	if visited[evID] {
		return
	}
	visited[evID] = true // 标记为已访问

	ev, ok := p.Events[evID]
	if !ok { // 如果事件不存在，安全返回
		return
	}

	// 标题层级：depth=0 → 二级标题(##)，depth=1 → 三级标题(###)，依此类推。
	// 注意：外层 MakeMarkdown 已经使用 "## 事件链"，因此根事件从 ### 开始更合理。
	// 原代码 depth+1 使得 depth=0 输出 #，但那是全局一级标题。为了结构清晰，
	// 这里保持原逻辑，但理解时注意：实际上一级标题已被项目标题占用。
	level := depth + 1
	if level > 6 {
		level = 6 // Markdown 最多支持 6 级标题，超出则固定为6级
	}
	prefix := strings.Repeat("#", level) + " " // 生成如 "### " 的前缀

	// 构造带时间的标题
	timeStr := p.formatTime(ev.Time) // 将时间map转换为可读字符串
	title := ev.Name
	if timeStr != "" {
		title += "（" + timeStr + "）" // 如果存在时间，追加到标题后
	}
	sb.WriteString(prefix + title + "\n\n")

	// 输出事件的概述（简短介绍）
	if ev.Introduction != "" {
		sb.WriteString(ev.Introduction + "\n\n")
	}

	// 输出事件的详细过程（核心剧情）
	if ev.Process != "" {
		sb.WriteString(ev.Process + "\n\n")
	}

	// 输出事件的结果或影响（键值对形式，如“关系改变”、“物品获得”）
	if len(ev.Outcome) > 0 {
		sb.WriteString("**结果 / 影响**：\n")
		for k, v := range ev.Outcome {
			sb.WriteString(fmt.Sprintf("- **%s**：%s\n", k, v))
		}
		sb.WriteString("\n")
	}

	if len(ev.SubRules) > 0 {
		sb.WriteString("**事件写作要求/注意事项**：\n")
		for k, v := range ev.Outcome {
			sb.WriteString(fmt.Sprintf("- **%s**：%s\n", k, v))
		}
		sb.WriteString("\n")
	}

	// 输出参与者列表，包括他们在该事件中的状态（如心情、位置、属性变化）
	if len(ev.Participants) > 0 {
		sb.WriteString("**参与者**：\n")
		for _, part := range ev.Participants {
			// 尝试从实体映射中获取实体名称，如果找不到则回退到实体ID
			entity, ok := p.Entities[part.EntityID]
			entityName := part.EntityID
			if ok {
				entityName = entity.Name
			}
			// 格式化参与者的状态（例如“情绪: 愤怒”，“体力: 50%”）
			stateDesc := ""
			if len(part.State) > 0 {
				parts := make([]string, 0, len(part.State))
				for k, v := range part.State {
					parts = append(parts, k+": "+v)
				}
				stateDesc = "（" + strings.Join(parts, "；") + "）"
			}
			// 如果实体名称与ID不同，则在括号中显示ID以帮助识别
			sb.WriteString(fmt.Sprintf("- **%s**%s%s\n", entityName,
				func() string {
					if entityName != part.EntityID {
						return "（" + part.EntityID + "）"
					}
					return ""
				}(),
				stateDesc))
		}
		sb.WriteString("\n")
	}

	// 如果事件类型是容器(container)或混合(mixed)，则递归渲染其所有子事件
	// 容器类型通常用于组织场景、章节或并行事件组
	if ev.Type == "container" || ev.Type == "mixed" {
		for _, subID := range ev.SubEvents {
			p.renderEvent(sb, subID, depth+1, visited) // 深度+1，子事件标题级别递增
		}
	}
}

// formatTime 根据 TimeSystem 的定义将 Time map 转换为可读字符串
// 例如：TimeSystem = ["纪元", "年份"]，Time = {"纪元": "新生纪元", "年份": "1024"}
// 输出："新生纪元 · 1024"
func (p *ProjectObj) formatTime(timeMap map[string]string) string {
	// 如果没有定义时间系统或者时间map为空，则返回空字符串
	if len(timeMap) == 0 || len(p.WorldSetting.TimeSystem) == 0 {
		return ""
	}
	parts := make([]string, 0, len(p.WorldSetting.TimeSystem))
	// 按照 TimeSystem 中定义的顺序提取对应的值
	for _, key := range p.WorldSetting.TimeSystem {
		if val, ok := timeMap[key]; ok && val != "" {
			parts = append(parts, val)
		}
	}
	// 用 " · " 连接各个时间部分
	return strings.Join(parts, " · ")
}
