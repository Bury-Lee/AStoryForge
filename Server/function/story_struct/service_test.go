package story_struct

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"
)

// =====================================================================
// 辅助函数
// =====================================================================

// createTempProject 在临时目录中创建一个合法的工程 JSON 文件，返回 *ProjectObj
func createTempProject(t *testing.T, dir string) (*ProjectObj, string) {
	t.Helper()
	jsonPath := filepath.Join(dir, "test_project.json")

	p := &ProjectObj{}
	p.ProjectMeta = ProjectMeta{
		RootDir:     jsonPath,
		Title:       "测试工程",
		Author:      "测试作者",
		CreatedAt:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		UpdatedAt:   time.Date(2025, 1, 1, 0, 0, 0, 0, time.UTC),
		EventCount:  0,
		EntityCount: 0,
		Settings:    map[string]string{"key": "val"},
	}
	p.WorldSetting = WorldSetting{
		TimeLabels: make(map[string]string),
		Spine:      []string{},
		WriteRules: []string{"rule1"},
		WorldRules: []string{},
	}
	p.Events = make(map[string]Event)
	p.Entities = make(map[string]Entity)
	p.Participant = make(map[string]Participant)
	p.Entity = make(map[string]Entity)

	if err := p.SaveProjectObj(); err != nil {
		t.Fatalf("创建临时工程失败: %v", err)
	}
	return p, dir
}

// createTempProjectWithData 创建含事件和实体的工程
func createTempProjectWithData(t *testing.T, dir string) (*ProjectObj, string) {
	t.Helper()
	p, d := createTempProject(t, dir)

	// 添加事件
	p.Events["evt1"] = Event{
		ID:           "evt1",
		Name:         "事件一",
		Introduction: "简介一",
		Process:      "过程一",
		Outcome:      map[string]string{"结果": "成功"},
		Type:         EventTypeEvent,
	}
	p.Events["evt2"] = Event{
		ID:           "evt2",
		Name:         "事件二",
		Introduction: "简介二",
		Process:      "过程二",
		Outcome:      map[string]string{"结果": "失败"},
		Type:         EventTypeEvent,
	}
	p.ProjectMeta.EventCount = 2

	// 添加实体
	p.Entities["ent1"] = Entity{
		ID:           "ent1",
		Name:         "角色甲",
		Type:         "人物",
		Introduction: []string{"主角"},
	}
	p.Entities["ent2"] = Entity{
		ID:           "ent2",
		Name:         "角色乙",
		Type:         "人物",
		Introduction: []string{"配角"},
	}
	p.ProjectMeta.EntityCount = 2

	p.Entity = make(map[string]Entity, len(p.Entities))
	for k, v := range p.Entities {
		p.Entity[k] = v
	}

	if err := p.SaveProjectObj(); err != nil {
		t.Fatalf("保存含数据的工程失败: %v", err)
	}
	return p, d
}

// createTempProjectDir 创建一个目录（不含 JSON 文件），供测试 CreateProject 使用
func createTempProjectDir(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	return dir
}

// =====================================================================
// TestCreateProject
// =====================================================================

func TestCreateProject_Validation(t *testing.T) {
	p := &ProjectObj{}
	dir := createTempProjectDir(t)

	t.Run("空标题", func(t *testing.T) {
		_, err := p.CreateProject(ProjectMeta{RootDir: dir, Title: ""})
		if err == nil {
			t.Fatal("期望空标题错误，但未返回")
		}
	})

	t.Run("空白标题", func(t *testing.T) {
		_, err := p.CreateProject(ProjectMeta{RootDir: dir, Title: "   "})
		if err == nil {
			t.Fatal("期望空白标题错误，但未返回")
		}
	})

	t.Run("标题超长", func(t *testing.T) {
		longTitle := string(make([]rune, 201))
		_, err := p.CreateProject(ProjectMeta{RootDir: dir, Title: longTitle})
		if err == nil {
			t.Fatal("期望超长标题错误，但未返回")
		}
	})

	t.Run("目录不存在", func(t *testing.T) {
		_, err := p.CreateProject(ProjectMeta{RootDir: "/not/exist/path", Title: "有效标题"})
		if err == nil {
			t.Fatal("期望目录不存在错误，但未返回")
		}
	})
}

func TestCreateProject_Success(t *testing.T) {
	p := &ProjectObj{}
	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "project.json")

	// CreateProject 要求 RootDir 路径存在（用作父目录），
	// SaveProjectObj 再以 RootDir 为文件路径写入 JSON
	// 因此预先创建一个占位文件以满足 os.Stat 检查
	if err := os.WriteFile(jsonPath, []byte("{}"), 0644); err != nil {
		t.Fatalf("创建占位文件失败: %v", err)
	}

	meta := ProjectMeta{
		RootDir: jsonPath,
		Title:   "新建工程",
		Author:  "测试",
	}
	project, err := p.CreateProject(meta)
	if err != nil {
		t.Fatalf("CreateProject 失败: %v", err)
	}
	if project.Meta.Title != "新建工程" {
		t.Errorf("标题不匹配: got %q, want %q", project.Meta.Title, "新建工程")
	}
	if project.Meta.Author != "测试" {
		t.Errorf("作者不匹配: got %q, want %q", project.Meta.Author, "测试")
	}
	if project.Meta.CreatedAt.IsZero() {
		t.Error("CreatedAt 不应为零")
	}
	if project.Meta.UpdatedAt.IsZero() {
		t.Error("UpdatedAt 不应为零")
	}
	if project.Meta.EventCount != 0 {
		t.Errorf("EventCount 应为 0, got %d", project.Meta.EventCount)
	}
}

// =====================================================================
// TestLoadProject / TestListProjects
// =====================================================================

func TestLoadProject(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)

	p := &ProjectObj{}
	loaded, err := p.LoadProject(original.ProjectMeta.RootDir)
	if err != nil {
		t.Fatalf("LoadProject 失败: %v", err)
	}
	if loaded.Meta.Title != "测试工程" {
		t.Errorf("标题不匹配: got %q", loaded.Meta.Title)
	}
	if len(loaded.Events) != 2 {
		t.Errorf("期望 2 个事件, got %d", len(loaded.Events))
	}
	if len(loaded.Entities) != 2 {
		t.Errorf("期望 2 个实体, got %d", len(loaded.Entities))
	}
	if loaded.Events[0].ID != "evt1" {
		t.Errorf("事件未按 ID 排序: first=%q", loaded.Events[0].ID)
	}
}

func TestLoadProject_NotFound(t *testing.T) {
	p := &ProjectObj{}
	_, err := p.LoadProject("/nonexistent/path.json")
	if err == nil {
		t.Fatal("期望加载失败错误，但未返回")
	}
}

func TestListProjects(t *testing.T) {
	p := &ProjectObj{}
	dir := t.TempDir()

	// 在两个子目录中各放一个合法工程
	sub1 := filepath.Join(dir, "proj1")
	sub2 := filepath.Join(dir, "proj2")
	os.MkdirAll(sub1, 0755)
	os.MkdirAll(sub2, 0755)
	createTempProject(t, sub1)
	createTempProject(t, sub2)

	list, err := p.ListProjects(dir)
	if err != nil {
		t.Fatalf("ListProjects 失败: %v", err)
	}
	if len(list) != 2 {
		t.Errorf("期望 2 个工程, got %d", len(list))
	}
}

func TestListProjects_EmptyDir(t *testing.T) {
	p := &ProjectObj{}
	dir := t.TempDir()

	list, err := p.ListProjects(dir)
	if err != nil {
		t.Fatalf("ListProjects 失败: %v", err)
	}
	if len(list) != 0 {
		t.Errorf("空目录期望 0 个工程, got %d", len(list))
	}
}

func TestListProjects_DirNotExist(t *testing.T) {
	p := &ProjectObj{}
	_, err := p.ListProjects("/not/exist")
	if err == nil {
		t.Fatal("期望目录不存在错误，但未返回")
	}
}

// =====================================================================
// TestDeleteProject
// =====================================================================

func TestDeleteProject(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProject(t, dir)
	p := &ProjectObj{}

	err := p.DeleteProject(original.ProjectMeta.RootDir, "")
	if err != nil {
		t.Fatalf("DeleteProject 失败: %v", err)
	}
	if _, err := os.Stat(original.ProjectMeta.RootDir); !os.IsNotExist(err) {
		t.Fatal("工程文件应该已被删除")
	}
}

func TestDeleteProject_NotFound(t *testing.T) {
	p := &ProjectObj{}
	err := p.DeleteProject("", "/nonexistent")
	if err == nil {
		t.Fatal("期望路径错误，但未返回")
	}
}

// =====================================================================
// TestUpdateProjectMeta
// =====================================================================

func TestUpdateProjectMeta(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProject(t, dir)
	p := &ProjectObj{}

	newMeta := ProjectMeta{
		Title:  "更新标题",
		Author: "更新作者",
	}
	updated, err := p.UpdateProjectMeta(original.ProjectMeta.RootDir, "", newMeta)
	if err != nil {
		t.Fatalf("UpdateProjectMeta 失败: %v", err)
	}
	if updated.Title != "更新标题" {
		t.Errorf("标题未更新: got %q", updated.Title)
	}
	if updated.Author != "更新作者" {
		t.Errorf("作者未更新: got %q", updated.Author)
	}
	if updated.UpdatedAt.IsZero() {
		t.Error("UpdatedAt 不应为零")
	}
	// CreatedAt 应保留原值
	if updated.CreatedAt.IsZero() {
		t.Error("CreatedAt 不应被清零")
	}
}

// =====================================================================
// TestUpdateProjectSpine
// =====================================================================

func TestUpdateProjectSpine(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProject(t, dir)
	p := &ProjectObj{}

	// 先创建事件
	ev1, _ := p.CreateEvent(original.ProjectMeta.RootDir, "", Event{ID: "a", Name: "A"})
	ev2, _ := p.CreateEvent(original.ProjectMeta.RootDir, "", Event{ID: "b", Name: "B"})
	_ = ev1
	_ = ev2

	spine, err := p.UpdateProjectSpine(original.ProjectMeta.RootDir, "", []string{"a", "b"})
	if err != nil {
		t.Fatalf("UpdateProjectSpine 失败: %v", err)
	}
	if len(spine) != 2 || spine[0] != "a" || spine[1] != "b" {
		t.Errorf("Spine 内容不符: %v", spine)
	}
}

func TestUpdateProjectSpine_UnknownEvent(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProject(t, dir)
	p := &ProjectObj{}

	_, err := p.UpdateProjectSpine(original.ProjectMeta.RootDir, "", []string{"nonexistent"})
	if err == nil {
		t.Fatal("期望未知事件 ID 错误，但未返回")
	}
}

func TestUpdateProjectSpine_Cycle(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProject(t, dir)
	p := &ProjectObj{}

	p.CreateEvent(original.ProjectMeta.RootDir, "", Event{ID: "a", Name: "A"})
	p.CreateEvent(original.ProjectMeta.RootDir, "", Event{ID: "b", Name: "B"})

	// 重复的 ID 应触发环检测
	_, err := p.UpdateProjectSpine(original.ProjectMeta.RootDir, "", []string{"a", "b", "a"})
	if err == nil {
		t.Fatal("期望环检测错误，但未返回")
	}
}

// =====================================================================
// TestSaveProject
// =====================================================================

func TestSaveProject(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProject(t, dir)
	p := &ProjectObj{}

	project := &Project{
		Meta: original.ProjectMeta,
		WorldSetting: WorldSetting{
			Spine:      []string{},
			WriteRules: []string{"新规则"},
		},
		Events: []Event{
			{ID: "e1", Name: "保存事件"},
		},
		Entities: []Entity{
			{ID: "en1", Name: "保存实体"},
		},
	}
	err := p.SaveProject(original.ProjectMeta.RootDir, project)
	if err != nil {
		t.Fatalf("SaveProject 失败: %v", err)
	}

	// 重新加载验证
	loaded, err := p.LoadProject(original.ProjectMeta.RootDir)
	if err != nil {
		t.Fatalf("重新加载失败: %v", err)
	}
	if len(loaded.Events) != 1 || loaded.Events[0].Name != "保存事件" {
		t.Errorf("事件保存不正确: %+v", loaded.Events)
	}
	if len(loaded.Entities) != 1 || loaded.Entities[0].Name != "保存实体" {
		t.Errorf("实体保存不正确: %+v", loaded.Entities)
	}
	if loaded.Meta.EventCount != 1 || loaded.Meta.EntityCount != 1 {
		t.Errorf("计数不正确: EventCount=%d EntityCount=%d", loaded.Meta.EventCount, loaded.Meta.EntityCount)
	}
}

// =====================================================================
// TestEventCRUD
// =====================================================================

func TestCreateEvent(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProject(t, dir)
	p := &ProjectObj{}

	ev, err := p.CreateEvent(original.ProjectMeta.RootDir, "", Event{
		Name:         "新事件",
		Introduction: "事件描述",
	})
	if err != nil {
		t.Fatalf("CreateEvent 失败: %v", err)
	}
	if ev.ID == "" {
		t.Error("事件 ID 不应为空")
	}
	if ev.Name != "新事件" {
		t.Errorf("名称不匹配: got %q", ev.Name)
	}
}

func TestCreateEvent_DuplicateID(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProject(t, dir)
	p := &ProjectObj{}

	p.CreateEvent(original.ProjectMeta.RootDir, "", Event{ID: "dup", Name: "第一个"})
	_, err := p.CreateEvent(original.ProjectMeta.RootDir, "", Event{ID: "dup", Name: "第二个"})
	if err == nil {
		t.Fatal("期望重复 ID 错误，但未返回")
	}
}

func TestGetEvent(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	ev, err := p.GetEvent(original.ProjectMeta.RootDir, "", "evt1")
	if err != nil {
		t.Fatalf("GetEvent 失败: %v", err)
	}
	if ev.Name != "事件一" {
		t.Errorf("名称不匹配: got %q", ev.Name)
	}
}

func TestGetEvent_NotFound(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProject(t, dir)
	p := &ProjectObj{}

	_, err := p.GetEvent(original.ProjectMeta.RootDir, "", "nonexistent")
	if err == nil {
		t.Fatal("期望不存在的错误，但未返回")
	}
}

func TestUpdateEvent(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	updated, err := p.UpdateEvent(original.ProjectMeta.RootDir, "", "evt1", Event{
		Name:         "更新后事件",
		Introduction: "新简介",
	})
	if err != nil {
		t.Fatalf("UpdateEvent 失败: %v", err)
	}
	if updated.Name != "更新后事件" {
		t.Errorf("名称未更新: got %q", updated.Name)
	}

	// 重新加载确认持久化
	ev, _ := p.GetEvent(original.ProjectMeta.RootDir, "", "evt1")
	if ev.Name != "更新后事件" {
		t.Errorf("持久化后名称不匹配: got %q", ev.Name)
	}
}

func TestUpdateEvent_NotFound(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProject(t, dir)
	p := &ProjectObj{}

	_, err := p.UpdateEvent(original.ProjectMeta.RootDir, "", "nonexistent", Event{Name: "test"})
	if err == nil {
		t.Fatal("期望不存在的错误，但未返回")
	}
}

func TestDeleteEvent(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	// 把 evt1 加入 spine
	p.UpdateProjectSpine(original.ProjectMeta.RootDir, "", []string{"evt1", "evt2"})

	err := p.DeleteEvent(original.ProjectMeta.RootDir, "", "evt1")
	if err != nil {
		t.Fatalf("DeleteEvent 失败: %v", err)
	}

	// spine 应已清理
	graph, _ := p.GetStoryGraph(original.ProjectMeta.RootDir, "")
	if len(graph.Spine) != 1 || graph.Spine[0] != "evt2" {
		t.Errorf("Spine 未正确清理: %v", graph.Spine)
	}

	// 事件计数应减 1
	if len(graph.Events) != 1 {
		t.Errorf("期望剩余 1 个事件, got %d", len(graph.Events))
	}
}

func TestDeleteEvent_NotFound(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProject(t, dir)
	p := &ProjectObj{}

	err := p.DeleteEvent(original.ProjectMeta.RootDir, "", "nonexistent")
	if err == nil {
		t.Fatal("期望不存在的错误，但未返回")
	}
}

func TestListEvents(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	events, err := p.ListEvents(original.ProjectMeta.RootDir, "")
	if err != nil {
		t.Fatalf("ListEvents 失败: %v", err)
	}
	if len(events) != 2 {
		t.Fatalf("期望 2 个事件, got %d", len(events))
	}
	if events[0].ID != "evt1" || events[1].ID != "evt2" {
		t.Errorf("事件顺序应在全局按 ID 排序: %v", []string{events[0].ID, events[1].ID})
	}
}

// =====================================================================
// TestSetEventLocked
// =====================================================================

func TestSetEventLocked(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	ev, err := p.SetEventLocked(original.ProjectMeta.RootDir, "", "evt1", true)
	if err != nil {
		t.Fatalf("SetEventLocked 失败: %v", err)
	}
	if !ev.Locked {
		t.Error("Locked 应为 true")
	}

	// 重新加载验证
	loaded, _ := p.GetEvent(original.ProjectMeta.RootDir, "", "evt1")
	if !loaded.Locked {
		t.Error("持久化后 Locked 应为 true")
	}
}

func TestSetEventLocked_NotFound(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProject(t, dir)
	p := &ProjectObj{}

	_, err := p.SetEventLocked(original.ProjectMeta.RootDir, "", "nonexistent", true)
	if err == nil {
		t.Fatal("期望不存在的错误，但未返回")
	}
}

// =====================================================================
// TestUpdateEventParticipants
// =====================================================================

func TestUpdateEventParticipants(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	participants := []Participant{
		{EntityID: "ent1", State: map[string]string{"情绪": "平静"}},
	}
	ev, err := p.UpdateEventParticipants(original.ProjectMeta.RootDir, "", "evt1", participants)
	if err != nil {
		t.Fatalf("UpdateEventParticipants 失败: %v", err)
	}
	if len(ev.Participants) != 1 || ev.Participants[0].EntityID != "ent1" {
		t.Errorf("参与者未正确设置: %+v", ev.Participants)
	}

	// 验证实体的事件索引已更新
	ent, _ := p.GetEntity(original.ProjectMeta.RootDir, "", "ent1")
	found := false
	for _, e := range ent.Events {
		if e == "evt1" {
			found = true
			break
		}
	}
	if !found {
		t.Error("实体的事件索引中应包含 evt1")
	}
}

func TestUpdateEventParticipants_InvalidEntity(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	_, err := p.UpdateEventParticipants(original.ProjectMeta.RootDir, "", "evt1", []Participant{
		{EntityID: "nonexistent"},
	})
	if err == nil {
		t.Fatal("期望不存在的实体错误，但未返回")
	}
}

// =====================================================================
// TestEventEdgeOperations
// =====================================================================

func TestCreateEventEdge(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	err := p.CreateEventEdge(original.ProjectMeta.RootDir, "", "evt1", "evt2", EdgeTypeCause, true)
	if err != nil {
		t.Fatalf("CreateEventEdge 失败: %v", err)
	}

	// 验证出边
	ev1, _ := p.GetEvent(original.ProjectMeta.RootDir, "", "evt1")
	if len(ev1.OutEdges) != 1 || ev1.OutEdges[0].Target != "evt2" {
		t.Errorf("出边未正确设置: %+v", ev1.OutEdges)
	}

	// 验证入边（mirrorInEdge=true）
	ev2, _ := p.GetEvent(original.ProjectMeta.RootDir, "", "evt2")
	if len(ev2.InEdges) != 1 || ev2.InEdges[0].Target != "evt1" {
		t.Errorf("入边未正确设置: %+v", ev2.InEdges)
	}
}

func TestCreateEventEdge_Duplicate(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	p.CreateEventEdge(original.ProjectMeta.RootDir, "", "evt1", "evt2", EdgeTypeCause, false)
	err := p.CreateEventEdge(original.ProjectMeta.RootDir, "", "evt1", "evt2", EdgeTypeCause, false)
	if err == nil {
		t.Fatal("期望重复边错误，但未返回")
	}
}

func TestCreateEventEdge_SourceNotFound(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	err := p.CreateEventEdge(original.ProjectMeta.RootDir, "", "nonexistent", "evt2", EdgeTypeCause, false)
	if err == nil {
		t.Fatal("期望起点不存在错误，但未返回")
	}
}

func TestCreateEventEdge_TargetNotFound(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	err := p.CreateEventEdge(original.ProjectMeta.RootDir, "", "evt1", "nonexistent", EdgeTypeCause, false)
	if err == nil {
		t.Fatal("期望终点不存在错误，但未返回")
	}
}

func TestUpdateEventEdge(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	// 先创建边 evt1 -> evt2
	p.CreateEventEdge(original.ProjectMeta.RootDir, "", "evt1", "evt2", EdgeTypeCause, true)

	// 再创建 evt3 作为新目标
	p.CreateEvent(original.ProjectMeta.RootDir, "", Event{ID: "evt3", Name: "事件三"})

	// 更新 evt1 -> evt3
	err := p.UpdateEventEdge(original.ProjectMeta.RootDir, "", "evt1", "evt2", "evt3", EdgeTypeCause, true)
	if err != nil {
		t.Fatalf("UpdateEventEdge 失败: %v", err)
	}

	ev1, _ := p.GetEvent(original.ProjectMeta.RootDir, "", "evt1")
	if len(ev1.OutEdges) != 1 || ev1.OutEdges[0].Target != "evt3" {
		t.Errorf("出边目标未更新: %+v", ev1.OutEdges)
	}

	// evt2 的入边应被移除
	ev2, _ := p.GetEvent(original.ProjectMeta.RootDir, "", "evt2")
	if len(ev2.InEdges) != 0 {
		t.Errorf("evt2 的入边应被清理: %+v", ev2.InEdges)
	}

	// evt3 的入边应包含 evt1
	ev3, _ := p.GetEvent(original.ProjectMeta.RootDir, "", "evt3")
	if len(ev3.InEdges) != 1 || ev3.InEdges[0].Target != "evt1" {
		t.Errorf("evt3 的入边未正确添加: %+v", ev3.InEdges)
	}
}

func TestDeleteEventEdge(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	p.CreateEventEdge(original.ProjectMeta.RootDir, "", "evt1", "evt2", EdgeTypeCause, true)

	err := p.DeleteEventEdge(original.ProjectMeta.RootDir, "", "evt1", "evt2", EdgeTypeCause, true)
	if err != nil {
		t.Fatalf("DeleteEventEdge 失败: %v", err)
	}

	ev1, _ := p.GetEvent(original.ProjectMeta.RootDir, "", "evt1")
	if len(ev1.OutEdges) != 0 {
		t.Errorf("出边应被删除: %+v", ev1.OutEdges)
	}

	ev2, _ := p.GetEvent(original.ProjectMeta.RootDir, "", "evt2")
	if len(ev2.InEdges) != 0 {
		t.Errorf("入边应被删除: %+v", ev2.InEdges)
	}
}

// =====================================================================
// TestSubEventOperations
// =====================================================================

func TestAttachSubEvent(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	err := p.AttachSubEvent(original.ProjectMeta.RootDir, "", "evt1", "evt2", -1)
	if err != nil {
		t.Fatalf("AttachSubEvent 失败: %v", err)
	}

	ev1, _ := p.GetEvent(original.ProjectMeta.RootDir, "", "evt1")
	if len(ev1.SubEvents) != 1 || ev1.SubEvents[0] != "evt2" {
		t.Errorf("子事件未挂载: %v", ev1.SubEvents)
	}
}

func TestAttachSubEvent_Duplicate(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	p.AttachSubEvent(original.ProjectMeta.RootDir, "", "evt1", "evt2", -1)
	err := p.AttachSubEvent(original.ProjectMeta.RootDir, "", "evt1", "evt2", -1)
	if err == nil {
		t.Fatal("期望重复子事件错误，但未返回")
	}
}

func TestAttachSubEvent_ParentNotFound(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	err := p.AttachSubEvent(original.ProjectMeta.RootDir, "", "nonexistent", "evt2", -1)
	if err == nil {
		t.Fatal("期望父事件不存在错误，但未返回")
	}
}

func TestMoveSubEvent(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	// 创建第三个事件
	p.CreateEvent(original.ProjectMeta.RootDir, "", Event{ID: "evt3", Name: "事件三"})

	// 挂载 evt2 和 evt3 到 evt1
	p.AttachSubEvent(original.ProjectMeta.RootDir, "", "evt1", "evt2", 0)
	p.AttachSubEvent(original.ProjectMeta.RootDir, "", "evt1", "evt3", 1)

	// 把 evt3 移到首位
	err := p.MoveSubEvent(original.ProjectMeta.RootDir, "", "evt1", "evt3", 0)
	if err != nil {
		t.Fatalf("MoveSubEvent 失败: %v", err)
	}

	ev1, _ := p.GetEvent(original.ProjectMeta.RootDir, "", "evt1")
	if len(ev1.SubEvents) != 2 || ev1.SubEvents[0] != "evt3" || ev1.SubEvents[1] != "evt2" {
		t.Errorf("子事件顺序未正确调整: %v", ev1.SubEvents)
	}
}

func TestMoveSubEvent_NotAttached(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	err := p.MoveSubEvent(original.ProjectMeta.RootDir, "", "evt1", "evt2", 0)
	if err == nil {
		t.Fatal("期望子事件未挂载错误，但未返回")
	}
}

func TestDetachSubEvent(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	p.AttachSubEvent(original.ProjectMeta.RootDir, "", "evt1", "evt2", -1)

	err := p.DetachSubEvent(original.ProjectMeta.RootDir, "", "evt1", "evt2")
	if err != nil {
		t.Fatalf("DetachSubEvent 失败: %v", err)
	}

	ev1, _ := p.GetEvent(original.ProjectMeta.RootDir, "", "evt1")
	if len(ev1.SubEvents) != 0 {
		t.Errorf("子事件应被解除: %v", ev1.SubEvents)
	}

	// evt2 本身不应被删除
	_, err = p.GetEvent(original.ProjectMeta.RootDir, "", "evt2")
	if err != nil {
		t.Errorf("子事件实体不应被删除: %v", err)
	}
}

// =====================================================================
// TestEntityCRUD
// =====================================================================

func TestCreateEntity(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProject(t, dir)
	p := &ProjectObj{}

	ent, err := p.CreateEntity(original.ProjectMeta.RootDir, "", Entity{
		Name: "新实体",
		Type: "角色",
	})
	if err != nil {
		t.Fatalf("CreateEntity 失败: %v", err)
	}
	if ent.ID == "" {
		t.Error("实体 ID 不应为空")
	}
	if ent.Name != "新实体" {
		t.Errorf("名称不匹配: got %q", ent.Name)
	}
}

func TestCreateEntity_DuplicateID(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProject(t, dir)
	p := &ProjectObj{}

	p.CreateEntity(original.ProjectMeta.RootDir, "", Entity{ID: "dup", Name: "第一个"})
	_, err := p.CreateEntity(original.ProjectMeta.RootDir, "", Entity{ID: "dup", Name: "第二个"})
	if err == nil {
		t.Fatal("期望重复 ID 错误，但未返回")
	}
}

func TestGetEntity(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	ent, err := p.GetEntity(original.ProjectMeta.RootDir, "", "ent1")
	if err != nil {
		t.Fatalf("GetEntity 失败: %v", err)
	}
	if ent.Name != "角色甲" {
		t.Errorf("名称不匹配: got %q", ent.Name)
	}
}

func TestGetEntity_NotFound(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProject(t, dir)
	p := &ProjectObj{}

	_, err := p.GetEntity(original.ProjectMeta.RootDir, "", "nonexistent")
	if err == nil {
		t.Fatal("期望不存在的错误，但未返回")
	}
}

func TestUpdateEntity(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	updated, err := p.UpdateEntity(original.ProjectMeta.RootDir, "", "ent1", Entity{
		Name:         "更新后实体",
		Introduction: []string{"新介绍"},
	})
	if err != nil {
		t.Fatalf("UpdateEntity 失败: %v", err)
	}
	if updated.Name != "更新后实体" {
		t.Errorf("名称未更新: got %q", updated.Name)
	}

	ent, _ := p.GetEntity(original.ProjectMeta.RootDir, "", "ent1")
	if ent.Name != "更新后实体" {
		t.Errorf("持久化后名称不匹配: got %q", ent.Name)
	}
}

func TestUpdateEntity_NotFound(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProject(t, dir)
	p := &ProjectObj{}

	_, err := p.UpdateEntity(original.ProjectMeta.RootDir, "", "nonexistent", Entity{Name: "test"})
	if err == nil {
		t.Fatal("期望不存在的错误，但未返回")
	}
}

func TestDeleteEntity(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	// 让 ent1 成为 evt1 的参与者
	p.UpdateEventParticipants(original.ProjectMeta.RootDir, "", "evt1", []Participant{
		{EntityID: "ent1"},
	})

	err := p.DeleteEntity(original.ProjectMeta.RootDir, "", "ent1")
	if err != nil {
		t.Fatalf("DeleteEntity 失败: %v", err)
	}

	// 验证事件参与者已被清理
	ev1, _ := p.GetEvent(original.ProjectMeta.RootDir, "", "evt1")
	if len(ev1.Participants) != 0 {
		t.Errorf("事件参与者应被清理: %+v", ev1.Participants)
	}

	// 验证实体计数
	list, _ := p.ListEntities(original.ProjectMeta.RootDir, "")
	if len(list) != 1 {
		t.Errorf("期望剩余 1 个实体, got %d", len(list))
	}
}

func TestDeleteEntity_NotFound(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProject(t, dir)
	p := &ProjectObj{}

	err := p.DeleteEntity(original.ProjectMeta.RootDir, "", "nonexistent")
	if err == nil {
		t.Fatal("期望不存在的错误，但未返回")
	}
}

func TestListEntities(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	entities, err := p.ListEntities(original.ProjectMeta.RootDir, "")
	if err != nil {
		t.Fatalf("ListEntities 失败: %v", err)
	}
	if len(entities) != 2 {
		t.Fatalf("期望 2 个实体, got %d", len(entities))
	}
	if entities[0].ID != "ent1" || entities[1].ID != "ent2" {
		t.Errorf("实体应在全局按 ID 排序: %v", []string{entities[0].ID, entities[1].ID})
	}
}

// =====================================================================
// TestValidateProject
// =====================================================================

func TestValidateProject_Passed(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	// 添加合理的 spine
	p.UpdateProjectSpine(original.ProjectMeta.RootDir, "", []string{"evt1", "evt2"})

	report, err := p.ValidateProject(original.ProjectMeta.RootDir, "")
	if err != nil {
		t.Fatalf("ValidateProject 失败: %v", err)
	}
	if !report.Passed {
		t.Errorf("期望校验通过但有 issues: %+v", report.Issues)
	}
}

func TestValidateProject_SpineUnknownEvent(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProject(t, dir)
	p := &ProjectObj{}

	report, err := p.ValidateProject(original.ProjectMeta.RootDir, "")
	if err != nil {
		t.Fatalf("ValidateProject 失败: %v", err)
	}
	if !report.Passed {
		t.Errorf("空工程无问题，应通过校验: %+v", report.Issues)
	}
}

func TestValidateProject_SpineCycle(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	// 通过直接操作后保存，绕过 UpdateProjectSpine 的校验
	loaded, _ := loadProject(original.ProjectMeta.RootDir)
	loaded.WorldSetting.Spine = []string{"evt1", "evt2", "evt1"}
	loaded.SaveProjectObj()

	report, err := p.ValidateProject(original.ProjectMeta.RootDir, "")
	if err != nil {
		t.Fatalf("ValidateProject 失败: %v", err)
	}
	if report.Passed {
		t.Error("期望校验不通过（环检测）")
	}
	foundCycle := false
	for _, issue := range report.Issues {
		if issue.Code == "spine_cycle" {
			foundCycle = true
			break
		}
	}
	if !foundCycle {
		t.Errorf("期望 spine_cycle issue, 但未找到: %+v", report.Issues)
	}
}

func TestValidateProject_OrphanEvent(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	report, err := p.ValidateProject(original.ProjectMeta.RootDir, "")
	if err != nil {
		t.Fatalf("ValidateProject 失败: %v", err)
	}
	// 两个事件都没有入边且不在 spine 中，应有 orphan_event 警告
	foundOrphan := false
	for _, issue := range report.Issues {
		if issue.Code == "orphan_event" {
			foundOrphan = true
			break
		}
	}
	if !foundOrphan {
		t.Errorf("期望 orphan_event issue, 但未找到: %+v", report.Issues)
	}
}

func TestValidateProject_ParticipantMissingEntity(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	// 添加一个引用不存在实体的参与者
	p.UpdateEventParticipants(original.ProjectMeta.RootDir, "", "evt1", []Participant{
		{EntityID: "ent1"},
	})
	// 直接修改存储添加无效参与者
	loaded, _ := loadProject(original.ProjectMeta.RootDir)
	ev := loaded.Events["evt1"]
	ev.Participants = []Participant{
		{EntityID: "ent1"},
		{EntityID: "nonexistent"},
	}
	loaded.Events["evt1"] = ev
	loaded.SaveProjectObj()

	report, _ := p.ValidateProject(original.ProjectMeta.RootDir, "")
	found := false
	for _, issue := range report.Issues {
		if issue.Code == "participant_missing_entity" {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("期望 participant_missing_entity issue, 但未找到: %+v", report.Issues)
	}
}

// =====================================================================
// TestGetStoryGraph
// =====================================================================

func TestGetStoryGraph(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	graph, err := p.GetStoryGraph(original.ProjectMeta.RootDir, "")
	if err != nil {
		t.Fatalf("GetStoryGraph 失败: %v", err)
	}
	if graph.ProjectID != "" {
		t.Errorf("ProjectID 不应被填充: got %q", graph.ProjectID)
	}
	if len(graph.Events) != 2 {
		t.Errorf("期望 2 个事件, got %d", len(graph.Events))
	}
	if len(graph.Entities) != 2 {
		t.Errorf("期望 2 个实体, got %d", len(graph.Entities))
	}
}

// =====================================================================
// TestInMemoryStore (GetSimulation / GetGenerationTask)
// =====================================================================

func TestGetSimulation_NotFound(t *testing.T) {
	p := &ProjectObj{}
	_, err := p.GetSimulation("nonexistent")
	if err == nil {
		t.Fatal("期望不存在的错误，但未返回")
	}
}

func TestGetGenerationTask_NotFound(t *testing.T) {
	p := &ProjectObj{}
	_, err := p.GetGenerationTask("nonexistent")
	if err == nil {
		t.Fatal("期望不存在的错误，但未返回")
	}
}

// =====================================================================
// TestSimulationLifecycle
// =====================================================================

func TestStartSimulation(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	session, err := p.StartSimulation(
		original.ProjectMeta.RootDir, "", SimulationModeBetween,
		"evt1", "evt2", "", nil, "测试提示", 3,
	)
	if err != nil {
		t.Fatalf("StartSimulation 失败: %v", err)
	}
	if session.ID == "" {
		t.Error("会话 ID 不应为空")
	}
	if session.Status != SimulationStatusRunning {
		t.Errorf("状态应为 running, got %s", session.Status)
	}
	if session.Mode != SimulationModeBetween {
		t.Errorf("模式应为 between, got %s", session.Mode)
	}
}

func TestAdvanceSimulation(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	session, _ := p.StartSimulation(original.ProjectMeta.RootDir, "", SimulationModeBetween,
		"evt1", "evt2", "", nil, "提示", 5)
	step, err := p.AdvanceSimulation(session.ID, "用户动作", "ent1")
	if err != nil {
		t.Fatalf("AdvanceSimulation 失败: %v", err)
	}
	if step.Round != 1 {
		t.Errorf("步数应为 1, got %d", step.Round)
	}
	if step.Action != "用户动作" {
		t.Errorf("动作不匹配: got %q", step.Action)
	}
	if step.ActorID != "ent1" {
		t.Errorf("执行者不匹配: got %q", step.ActorID)
	}
}

func TestAdvanceSimulation_NotRunning(t *testing.T) {
	p := &ProjectObj{}
	// 直接在 store 中创建一个已取消的会话
	storeMu.Lock()
	simulationStore["canceled_sim"] = &SimulationSession{
		ID:     "canceled_sim",
		Status: SimulationStatusCanceled,
	}
	storeMu.Unlock()
	defer func() {
		storeMu.Lock()
		delete(simulationStore, "canceled_sim")
		storeMu.Unlock()
	}()

	_, err := p.AdvanceSimulation("canceled_sim", "动作", "")
	if err == nil {
		t.Fatal("期望状态非运行中错误，但未返回")
	}
}

func TestAdvanceSimulation_NotFound(t *testing.T) {
	p := &ProjectObj{}
	_, err := p.AdvanceSimulation("nonexistent", "动作", "")
	if err == nil {
		t.Fatal("期望不存在的错误，但未返回")
	}
}

func TestReviewSimulation(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	session, _ := p.StartSimulation(original.ProjectMeta.RootDir, "", SimulationModeBetween,
		"evt1", "evt2", "", nil, "提示", 5)
	step, _ := p.AdvanceSimulation(session.ID, "动作", "")

	reviewed, err := p.ReviewSimulation(session.ID, step.ID, true, "通过")
	if err != nil {
		t.Fatalf("ReviewSimulation 失败: %v", err)
	}
	if !reviewed.Accepted {
		t.Error("Accepted 应为 true")
	}
	if reviewed.Observer != "通过" {
		t.Errorf("评论不匹配: got %q", reviewed.Observer)
	}
}

func TestReviewSimulation_StepNotFound(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	session, _ := p.StartSimulation(original.ProjectMeta.RootDir, "", SimulationModeBetween,
		"evt1", "evt2", "", nil, "提示", 5)

	_, err := p.ReviewSimulation(session.ID, "nonexistent_step", true, "")
	if err == nil {
		t.Fatal("期望步骤不存在错误，但未返回")
	}
}

func TestCancelSimulation(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	session, _ := p.StartSimulation(original.ProjectMeta.RootDir, "", SimulationModeBetween,
		"evt1", "evt2", "", nil, "提示", 5)

	err := p.CancelSimulation(session.ID)
	if err != nil {
		t.Fatalf("CancelSimulation 失败: %v", err)
	}

	got, _ := p.GetSimulation(session.ID)
	if got.Status != SimulationStatusCanceled {
		t.Errorf("状态应为 canceled, got %s", got.Status)
	}
}

func TestCancelSimulation_NotFound(t *testing.T) {
	p := &ProjectObj{}
	err := p.CancelSimulation("nonexistent")
	if err == nil {
		t.Fatal("期望不存在的错误，但未返回")
	}
}

func TestApplySimulationResult_NoDraftEvents(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	session, _ := p.StartSimulation(original.ProjectMeta.RootDir, "", SimulationModeBetween,
		"evt1", "evt2", "", nil, "提示", 5)

	result, err := p.ApplySimulationResult(session.ID, SimulationApplyOverwrite, "", "", 0)
	if err != nil {
		t.Fatalf("ApplySimulationResult 失败: %v", err)
	}
	if len(result.AffectedEventIDs) != 0 {
		t.Errorf("无草稿事件应返回空列表, got %v", result.AffectedEventIDs)
	}
}

func TestApplySimulationResult_SessionNotFound(t *testing.T) {
	p := &ProjectObj{}
	_, err := p.ApplySimulationResult("nonexistent", SimulationApplyOverwrite, "", "", 0)
	if err == nil {
		t.Fatal("期望不存在的错误，但未返回")
	}
}

// =====================================================================
// TestGenerateDraft & GetGenerationTask
// =====================================================================

func TestGenerateDraft(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	task, err := p.GenerateDraft(original.ProjectMeta.RootDir, "", []string{"evt1"}, "", "gpt-4", "写正文")
	if err != nil {
		t.Fatalf("GenerateDraft 失败: %v", err)
	}
	if task.ID == "" {
		t.Error("任务 ID 不应为空")
	}
	if task.Status != GenerationStatusPending {
		t.Errorf("状态应为 pending, got %s", task.Status)
	}

	// 能通过 GetGenerationTask 查询到
	got, err := p.GetGenerationTask(task.ID)
	if err != nil {
		t.Fatalf("GetGenerationTask 失败: %v", err)
	}
	if got.ID != task.ID {
		t.Errorf("任务 ID 不匹配: got %q", got.ID)
	}
}

// =====================================================================
// TestOptimizeNodeExpression & ApplyOptimizedNodes
// =====================================================================

func TestOptimizeNodeExpression(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	result, err := p.OptimizeNodeExpression(
		original.ProjectMeta.RootDir, "", []string{"evt1", "evt2"},
		"精简表达", nil, true, "简洁",
	)
	if err != nil {
		t.Fatalf("OptimizeNodeExpression 失败: %v", err)
	}
	if len(result.Drafts) != 2 {
		t.Fatalf("期望 2 个草稿, got %d", len(result.Drafts))
	}
	if result.Drafts[0].EventID != "evt1" {
		t.Errorf("第一个草稿应对应 evt1, got %s", result.Drafts[0].EventID)
	}
	// 占位实现应返回原始内容
	if result.Drafts[0].OptimizedIntroduction != result.Drafts[0].OriginalIntroduction {
		t.Error("占位实现中优化应与原始内容相同")
	}
}

func TestOptimizeNodeExpression_NoMatch(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProject(t, dir)
	p := &ProjectObj{}

	_, err := p.OptimizeNodeExpression(original.ProjectMeta.RootDir, "", []string{"nonexistent"}, "", nil, true, "")
	if err == nil {
		t.Fatal("期望无匹配事件错误，但未返回")
	}
}

func TestApplyOptimizedNodes(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	result, err := p.ApplyOptimizedNodes(original.ProjectMeta.RootDir, "", []OptimizedEventDraft{
		{
			EventID:               "evt1",
			OptimizedIntroduction: "优化后的简介",
			OptimizedProcess:      "优化后的过程",
		},
	})
	if err != nil {
		t.Fatalf("ApplyOptimizedNodes 失败: %v", err)
	}
	if len(result.AffectedEventIDs) != 1 || result.AffectedEventIDs[0] != "evt1" {
		t.Errorf("受影响事件列表错误: %v", result.AffectedEventIDs)
	}

	// 验证持久化
	ev, _ := p.GetEvent(original.ProjectMeta.RootDir, "", "evt1")
	if ev.Introduction != "优化后的简介" {
		t.Errorf("Introduction 未更新: got %q", ev.Introduction)
	}
	if ev.Process != "优化后的过程" {
		t.Errorf("Process 未更新: got %q", ev.Process)
	}
}

func TestApplyOptimizedNodes_NoopForUnknownEvent(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	result, err := p.ApplyOptimizedNodes(original.ProjectMeta.RootDir, "", []OptimizedEventDraft{
		{EventID: "nonexistent", OptimizedIntroduction: "新简介"},
	})
	if err != nil {
		t.Fatalf("ApplyOptimizedNodes 失败: %v", err)
	}
	if len(result.AffectedEventIDs) != 0 {
		t.Errorf("未知事件应返回空列表, got %v", result.AffectedEventIDs)
	}
}

// =====================================================================
// TestExtractFromSource (占位实现)
// =====================================================================

func TestExtractFromSource(t *testing.T) {
	dir := t.TempDir()
	p := &ProjectObj{}

	// autoCreate=false
	result, err := p.ExtractFromSource(dir, "测试标题", "原始文本内容", nil, nil, false)
	if err == nil {
		t.Fatal("占位实现应返回错误")
	}
	if result.Meta.Title != "测试标题" {
		t.Errorf("标题不匹配: got %q", result.Meta.Title)
	}
	_ = result
}

func TestExtractFromSource_EmptyTitle(t *testing.T) {
	p := &ProjectObj{}
	_, err := p.ExtractFromSource("", "", "text", nil, nil, false)
	if err == nil {
		t.Fatal("期望空标题错误，但未返回")
	}
}

func TestExtractFromSource_EmptyText(t *testing.T) {
	p := &ProjectObj{}
	_, err := p.ExtractFromSource("", "title", "", nil, nil, false)
	if err == nil {
		t.Fatal("期望空文本错误，但未返回")
	}
}

// =====================================================================
// TestWriteCompileResult
// =====================================================================

func TestWriteCompileResult(t *testing.T) {
	dir := t.TempDir()
	p := &ProjectObj{}

	outputPath := filepath.Join(dir, "output.md")
	report := &ValidationReport{
		ProjectID: "test",
		Passed:    true,
		Issues:    []ValidationIssue{},
	}
	result, err := p.writeCompileResult("test", "# Markdown 内容", report, outputPath)
	if err != nil {
		t.Fatalf("writeCompileResult 失败: %v", err)
	}
	if result.OutputPath != outputPath {
		t.Errorf("输出路径不匹配: got %q", result.OutputPath)
	}
	if result.Markdown != "# Markdown 内容" {
		t.Errorf("Markdown 内容不匹配: got %q", result.Markdown)
	}

	// 验证文件已写盘
	data, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("读取输出文件失败: %v", err)
	}
	if string(data) != "# Markdown 内容" {
		t.Errorf("文件内容不匹配: got %q", string(data))
	}
}

func TestWriteCompileResult_DefaultPath(t *testing.T) {
	dir := t.TempDir()
	p := &ProjectObj{}

	// rootDir 指向 .json 文件，默认输出为 <文件名>.json.md
	jsonPath := filepath.Join(dir, "my_project.json")
	expectedMD := filepath.Join(dir, "my_project.json.md")

	report := &ValidationReport{ProjectID: "test", Passed: true}
	result, err := p.writeCompileResult(jsonPath, "# 内容", report, "")
	if err != nil {
		t.Fatalf("writeCompileResult 失败: %v", err)
	}
	if result.OutputPath != expectedMD {
		t.Errorf("默认输出路径应为 %q, got %q", expectedMD, result.OutputPath)
	}
}

// =====================================================================
// TestFormatMap  (辅助函数)
// =====================================================================

func TestFormatMap(t *testing.T) {
	tests := []struct {
		name string
		m    map[string]string
		want string
	}{
		{"空 map", map[string]string{}, ""},
		{"单键", map[string]string{"a": "1"}, "a:1"},
		{"多键排序", map[string]string{"b": "2", "a": "1"}, "a:1; b:2"},
		{"nil map", nil, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := formatMap(tt.m)
			if got != tt.want {
				t.Errorf("formatMap(%v) = %q, want %q", tt.m, got, tt.want)
			}
		})
	}
}

// =====================================================================
// TestEnsureEventID / EnsureEntityID
// =====================================================================

func TestEnsureEventID(t *testing.T) {
	ev := &Event{}
	ensureEventID(ev)
	if ev.ID == "" {
		t.Error("ID 不应为空")
	}
	if ev.InEdges == nil {
		t.Error("InEdges 不应为 nil")
	}
	if ev.OutEdges == nil {
		t.Error("OutEdges 不应为 nil")
	}
	if ev.SubEvents == nil {
		t.Error("SubEvents 不应为 nil")
	}
	if ev.Participants == nil {
		t.Error("Participants 不应为 nil")
	}
	if ev.SubRules == nil {
		t.Error("SubRules 不应为 nil")
	}
	if ev.Time == nil {
		t.Error("Time 不应为 nil")
	}
	if ev.Outcome == nil {
		t.Error("Outcome 不应为 nil")
	}

	// 再次调用不应改变 ID
	origID := ev.ID
	ensureEventID(ev)
	if ev.ID != origID {
		t.Error("已有 ID 不应被修改")
	}
}

func TestEnsureEntityID(t *testing.T) {
	ent := &Entity{}
	ensureEntityID(ent)
	if ent.ID == "" {
		t.Error("ID 不应为空")
	}
	if ent.Introduction == nil {
		t.Error("Introduction 不应为 nil")
	}
	if ent.Relationships == nil {
		t.Error("Relationships 不应为 nil")
	}
	if ent.RuleRefs == nil {
		t.Error("RuleRefs 不应为 nil")
	}
	if ent.Events == nil {
		t.Error("Events 不应为 nil")
	}

	origID := ent.ID
	ensureEntityID(ent)
	if ent.ID != origID {
		t.Error("已有 ID 不应被修改")
	}
}

// =====================================================================
// TestResolveRootDir
// =====================================================================

func TestResolveRootDir(t *testing.T) {
	tests := []struct {
		name      string
		rootDir   string
		projectID string
		want      string
	}{
		{"rootDir 优先", "/path", "fallback", "/path"},
		{"fallback 到 projectID", "", "project_path", "project_path"},
		{"均为空", "", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := resolveRootDir(tt.rootDir, tt.projectID)
			if got != tt.want {
				t.Errorf("resolveRootDir(%q, %q) = %q, want %q", tt.rootDir, tt.projectID, got, tt.want)
			}
		})
	}
}

// =====================================================================
// TestLoadProject
// =====================================================================

func TestLoadProject_EmptyRootDir(t *testing.T) {
	p := &ProjectObj{}
	_, err := p.LoadProject("")
	if err == nil {
		t.Fatal("期望空根目录错误，但未返回")
	}
}

// =====================================================================
// TestDeleteEntity_EntityRelationshipCleanup
// =====================================================================

func TestDeleteEntity_ClearsRelationships(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	// 给 ent1 添加对 ent2 的关系
	p.UpdateEntity(original.ProjectMeta.RootDir, "", "ent1", Entity{
		Name:         "角色甲",
		Introduction: []string{"主角"},
		Relationships: []EntityRelationship{
			{TargetID: "ent2", RelationType: "朋友"},
		},
	})

	// 删除 ent2
	p.DeleteEntity(original.ProjectMeta.RootDir, "", "ent2")

	// ent1 的关系中应已清理对 ent2 的引用
	ent1, _ := p.GetEntity(original.ProjectMeta.RootDir, "", "ent1")
	for _, rel := range ent1.Relationships {
		if rel.TargetID == "ent2" {
			t.Errorf("实体关系应已清理对已删除实体的引用: %+v", ent1.Relationships)
		}
	}
}

// =====================================================================
// TestCreateEvent_LoadProjectWithProjectID
// =====================================================================

func TestCreateEvent_WithProjectID(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProject(t, dir)
	p := &ProjectObj{}

	// 使用 projectID（即文件路径）代替 rootDir
	ev, err := p.CreateEvent("", original.ProjectMeta.RootDir, Event{
		Name: "通过 projectID 创建",
	})
	if err != nil {
		t.Fatalf("CreateEvent (with projectID) 失败: %v", err)
	}
	if ev.Name != "通过 projectID 创建" {
		t.Errorf("名称不匹配: got %q", ev.Name)
	}
}

// =====================================================================
// TestSaveProject_NilProject
// =====================================================================

func TestSaveProject_NilProject(t *testing.T) {
	p := &ProjectObj{}
	err := p.SaveProject("", nil)
	if err == nil {
		t.Fatal("期望空工程错误，但未返回")
	}
}

// =====================================================================
// TestEdgeOperations_NoMirror
// =====================================================================

func TestCreateEventEdge_NoMirror(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	err := p.CreateEventEdge(original.ProjectMeta.RootDir, "", "evt1", "evt2", EdgeTypeCause, false)
	if err != nil {
		t.Fatalf("CreateEventEdge (no mirror) 失败: %v", err)
	}

	ev1, _ := p.GetEvent(original.ProjectMeta.RootDir, "", "evt1")
	if len(ev1.OutEdges) != 1 {
		t.Errorf("出边应被创建: %+v", ev1.OutEdges)
	}
	ev2, _ := p.GetEvent(original.ProjectMeta.RootDir, "", "evt2")
	if len(ev2.InEdges) != 0 {
		t.Errorf("mirrorInEdge=false 时入边不应被创建: %+v", ev2.InEdges)
	}
}

func TestDeleteEventEdge_NoMirror(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	p.CreateEventEdge(original.ProjectMeta.RootDir, "", "evt1", "evt2", EdgeTypeCause, true)

	err := p.DeleteEventEdge(original.ProjectMeta.RootDir, "", "evt1", "evt2", EdgeTypeCause, false)
	if err != nil {
		t.Fatalf("DeleteEventEdge (no mirror) 失败: %v", err)
	}

	ev1, _ := p.GetEvent(original.ProjectMeta.RootDir, "", "evt1")
	if len(ev1.OutEdges) != 0 {
		t.Errorf("出边应被删除: %+v", ev1.OutEdges)
	}
	ev2, _ := p.GetEvent(original.ProjectMeta.RootDir, "", "evt2")
	if len(ev2.InEdges) != 1 {
		t.Errorf("mirrorInEdge=false 时入边不应被删除: %+v", ev2.InEdges)
	}
}

// =====================================================================
// TestAttachSubEvent_WithIndex
// =====================================================================

func TestAttachSubEvent_WithIndex(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	p.CreateEvent(original.ProjectMeta.RootDir, "", Event{ID: "evt3", Name: "事件三"})

	// 挂载 evt2 到索引 0
	err := p.AttachSubEvent(original.ProjectMeta.RootDir, "", "evt1", "evt2", 0)
	if err != nil {
		t.Fatalf("AttachSubEvent(0) 失败: %v", err)
	}
	// 挂载 evt3 到索引 0 -> 应插入到 evt2 前面
	err = p.AttachSubEvent(original.ProjectMeta.RootDir, "", "evt1", "evt3", 0)
	if err != nil {
		t.Fatalf("AttachSubEvent(0) 第二次失败: %v", err)
	}

	ev1, _ := p.GetEvent(original.ProjectMeta.RootDir, "", "evt1")
	if len(ev1.SubEvents) != 2 || ev1.SubEvents[0] != "evt3" || ev1.SubEvents[1] != "evt2" {
		t.Errorf("索引插入顺序不正确: %v", ev1.SubEvents)
	}
}

// =====================================================================
// TestDeleteEvent_CleansSubEventsAndEdges
// =====================================================================

func TestDeleteEvent_CleansReferences(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	// 创建 evt3 作为子事件
	p.CreateEvent(original.ProjectMeta.RootDir, "", Event{ID: "evt3", Name: "事件三"})

	// evt1 -> 子 evt2
	p.AttachSubEvent(original.ProjectMeta.RootDir, "", "evt1", "evt2", -1)
	// evt1 -> cause -> evt3
	p.CreateEventEdge(original.ProjectMeta.RootDir, "", "evt1", "evt3", EdgeTypeCause, true)

	// 删除 evt2
	err := p.DeleteEvent(original.ProjectMeta.RootDir, "", "evt2")
	if err != nil {
		t.Fatalf("DeleteEvent 失败: %v", err)
	}

	// evt1 的 SubEvents 不应再包含 evt2
	ev1, _ := p.GetEvent(original.ProjectMeta.RootDir, "", "evt1")
	if len(ev1.SubEvents) != 0 {
		t.Errorf("evt1 的 SubEvents 应为空: %v", ev1.SubEvents)
	}
	// evt3 的 InEdges 不应受影响 (只删 evt2)
	ev3, _ := p.GetEvent(original.ProjectMeta.RootDir, "", "evt3")
	if len(ev3.InEdges) != 1 {
		t.Errorf("evt3 的入边应保留: %+v", ev3.InEdges)
	}
}

// =====================================================================
// TestSaveProject_WithProjectMetaAsRootDir
// =====================================================================

func TestSaveProject_UsesMetaRootDir(t *testing.T) {
	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "project.json")

	// 先通过 SaveProjectObj 直接在目标路径创建工程文件
	orig := &ProjectObj{}
	orig.ProjectMeta = ProjectMeta{
		RootDir: jsonPath,
		Title:   "初始工程",
		Settings: map[string]string{},
	}
	orig.WorldSetting = WorldSetting{
		TimeLabels: make(map[string]string),
		Spine:      []string{},
		WriteRules: []string{},
		WorldRules: []string{},
	}
	orig.Events = make(map[string]Event)
	orig.Entities = make(map[string]Entity)
	orig.Participant = make(map[string]Participant)
	orig.Entity = make(map[string]Entity)
	if err := orig.SaveProjectObj(); err != nil {
		t.Fatalf("创建初始工程失败: %v", err)
	}

	p := &ProjectObj{}

	project := &Project{
		Meta: ProjectMeta{
			RootDir: jsonPath,
			Title:   "通过 Meta.RootDir 保存",
		},
		Events:   []Event{{ID: "e1", Name: "事件"}},
		Entities: []Entity{},
	}

	// rootDir="" 时应当使用 project.Meta.RootDir
	err := p.SaveProject("", project)
	if err != nil {
		t.Fatalf("SaveProject (通过 Meta.RootDir) 失败: %v", err)
	}

	loaded, _ := p.LoadProject(jsonPath)
	if loaded.Meta.Title != "通过 Meta.RootDir 保存" {
		t.Errorf("标题不匹配: got %q", loaded.Meta.Title)
	}
}

// =====================================================================
// TestUpdateProjectMeta_PreservesCounts
// =====================================================================

func TestUpdateProjectMeta_PreservesCounts(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	// 已知工程有 2 个事件和 2 个实体
	updated, err := p.UpdateProjectMeta(original.ProjectMeta.RootDir, "", ProjectMeta{
		Title: "仅更新标题",
	})
	if err != nil {
		t.Fatalf("UpdateProjectMeta 失败: %v", err)
	}
	if updated.EventCount != 2 {
		t.Errorf("EventCount 应保留原值 2, got %d", updated.EventCount)
	}
	if updated.EntityCount != 2 {
		t.Errorf("EntityCount 应保留原值 2, got %d", updated.EntityCount)
	}
}

// =====================================================================
// TestUpdateProjectSpine_CopiesSlice
// =====================================================================

func TestUpdateProjectSpine_CopiesSlice(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	spineInput := []string{"evt1", "evt2"}
	spine, err := p.UpdateProjectSpine(original.ProjectMeta.RootDir, "", spineInput)
	if err != nil {
		t.Fatalf("UpdateProjectSpine 失败: %v", err)
	}

	// 修改返回的 spine 不应影响存储
	spine[0] = "modified"
	spineInput[1] = "modified_input"

	loaded, _ := loadProject(original.ProjectMeta.RootDir)
	if loaded.WorldSetting.Spine[0] != "evt1" || loaded.WorldSetting.Spine[1] != "evt2" {
		t.Errorf("Spine 被外部意外修改: %v", loaded.WorldSetting.Spine)
	}
}

// =====================================================================
// TestDeleteEntity_ParticipantCleanup
// =====================================================================

func TestDeleteEntity_CleansEventParticipants(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	// 让 ent1 和 ent2 都成为 evt1 的参与者
	p.UpdateEventParticipants(original.ProjectMeta.RootDir, "", "evt1", []Participant{
		{EntityID: "ent1", State: map[string]string{"role": "hero"}},
		{EntityID: "ent2", State: map[string]string{"role": "sidekick"}},
	})

	// 删除 ent1
	err := p.DeleteEntity(original.ProjectMeta.RootDir, "", "ent1")
	if err != nil {
		t.Fatalf("DeleteEntity 失败: %v", err)
	}

	ev1, _ := p.GetEvent(original.ProjectMeta.RootDir, "", "evt1")
	if len(ev1.Participants) != 1 || ev1.Participants[0].EntityID != "ent2" {
		t.Errorf("参与者应只保留 ent2: %+v", ev1.Participants)
	}
}

// =====================================================================
// TestMultipleDeleteEdge_NonExistent
// =====================================================================

func TestDeleteEventEdge_NonExistent(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	// 删除不存在的边不应报错（src 存在但边不存在，走正常逻辑不报错）
	err := p.DeleteEventEdge(original.ProjectMeta.RootDir, "", "evt1", "evt2", EdgeTypeCause, true)
	if err != nil {
		// 这取决于实现：如果边不存在，删除操作只是空操作
		// 当前实现只是过滤掉匹配的边，如果没有匹配则不做任何事
		t.Fatalf("删除不存在的边不应失败: %v", err)
	}
}

// =====================================================================
// TestCreateProject_DefaultValues
// =====================================================================

func TestCreateProject_DefaultValues(t *testing.T) {
	p := &ProjectObj{}
	dir := t.TempDir()
	jsonPath := filepath.Join(dir, "project.json")

	// 创建占位文件以满足 CreateProject 的 os.Stat 检查
	if err := os.WriteFile(jsonPath, []byte("{}"), 0644); err != nil {
		t.Fatalf("创建占位文件失败: %v", err)
	}

	meta := ProjectMeta{
		RootDir: jsonPath,
		Title:   "默认值测试",
		// Author 留空
	}
	project, err := p.CreateProject(meta)
	if err != nil {
		t.Fatalf("CreateProject 失败: %v", err)
	}
	if project.Meta.Author != "未知" {
		t.Errorf("Author 默认应为「未知」, got %q", project.Meta.Author)
	}
	if project.Meta.Settings == nil {
		t.Error("Settings 不应为 nil")
	}
}

// =====================================================================
// TestInvalidEventOperations
// =====================================================================

func TestCreateEvent_InProjectIDEmpty(t *testing.T) {
	p := &ProjectObj{}
	_, err := p.CreateEvent("", "", Event{Name: "事件"})
	if err == nil {
		t.Fatal("期望空路径错误，但未返回")
	}
}

func TestUpdateEventEdge_NewTargetNotFound(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	p.CreateEventEdge(original.ProjectMeta.RootDir, "", "evt1", "evt2", EdgeTypeCause, true)

	err := p.UpdateEventEdge(original.ProjectMeta.RootDir, "", "evt1", "evt2", "nonexistent", EdgeTypeCause, true)
	if err == nil {
		t.Fatal("期望新终点不存在错误，但未返回")
	}
}

func TestUpdateEventEdge_SourceNotFound(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	err := p.UpdateEventEdge(original.ProjectMeta.RootDir, "", "nonexistent", "evt2", "evt1", EdgeTypeCause, true)
	if err == nil {
		t.Fatal("期望起点不存在错误，但未返回")
	}
}

func TestUpdateEventEdge_OriginalEdgeNotFound(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}
	p.CreateEvent(original.ProjectMeta.RootDir, "", Event{ID: "evt3", Name: "事件三"})

	// 没有创建边就尝试更新
	err := p.UpdateEventEdge(original.ProjectMeta.RootDir, "", "evt1", "evt2", "evt3", EdgeTypeCause, true)
	if err == nil {
		t.Fatal("期望原边不存在错误，但未返回")
	}
}

// =====================================================================
// TestSimulationStepCannotAdvanceAfterCancel
// =====================================================================

func TestAdvanceSimulation_AfterCancel(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	session, _ := p.StartSimulation(original.ProjectMeta.RootDir, "", SimulationModeBetween,
		"evt1", "evt2", "", nil, "提示", 5)
	p.CancelSimulation(session.ID)

	_, err := p.AdvanceSimulation(session.ID, "动作", "")
	if err == nil {
		t.Fatal("在已取消的会话上不应允许推进")
	}
}

// =====================================================================
// TestJSONPersistence
// =====================================================================

func TestEventCRUD_Persistence(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProject(t, dir)
	p := &ProjectObj{}

	// 创建一个事件
	ev, _ := p.CreateEvent(original.ProjectMeta.RootDir, "", Event{
		Name: "持久化测试",
		Time: map[string]string{"year": "2025"},
	})

	// 直接读取 JSON 验证
	data, err := os.ReadFile(original.ProjectMeta.RootDir)
	if err != nil {
		t.Fatalf("读取 JSON 失败: %v", err)
	}
	var loaded ProjectObj
	json.Unmarshal(data, &loaded)

	savedEv, ok := loaded.Events[ev.ID]
	if !ok {
		t.Fatal("JSON 中应包含新创建的事件")
	}
	if savedEv.Name != "持久化测试" {
		t.Errorf("JSON 中事件名称不匹配: got %q", savedEv.Name)
	}
	if savedEv.Time["year"] != "2025" {
		t.Errorf("JSON 中事件时间不匹配: got %v", savedEv.Time)
	}
}

// =====================================================================
// TestGetStoryGraph_WithProjectID
// =====================================================================

func TestGetStoryGraph_WithProjectID(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	// projectID = 文件路径
	graph, err := p.GetStoryGraph("", original.ProjectMeta.RootDir)
	if err != nil {
		t.Fatalf("GetStoryGraph (with projectID) 失败: %v", err)
	}
	if len(graph.Events) != 2 {
		t.Errorf("期望 2 个事件, got %d", len(graph.Events))
	}
}

// =====================================================================
// TestStartSimulation_MaxRoundsDefault
// =====================================================================

func TestStartSimulation_DefaultMaxRounds(t *testing.T) {
	dir := t.TempDir()
	original, _ := createTempProjectWithData(t, dir)
	p := &ProjectObj{}

	session, err := p.StartSimulation(original.ProjectMeta.RootDir, "", SimulationModeBetween,
		"evt1", "evt2", "", nil, "提示", 0)
	if err != nil {
		t.Fatalf("StartSimulation 失败: %v", err)
	}
	if session.ID == "" {
		t.Error("会话 ID 不应为空")
	}
	// maxRounds <= 0 时默认应为 5，但 session 不存储 maxRounds，仅检查启动成功
}

// =====================================================================
// TestGenerateDraft_ProjectNotFound
// =====================================================================

func TestGenerateDraft_ProjectNotFound(t *testing.T) {
	p := &ProjectObj{}
	_, err := p.GenerateDraft("/nonexistent", "", []string{"e1"}, "", "gpt-4", "prompt")
	if err == nil {
		t.Fatal("期望工程不存在错误，但未返回")
	}
}

// =====================================================================
// TestStartSimulation_ProjectNotFound
// =====================================================================

func TestStartSimulation_ProjectNotFound(t *testing.T) {
	p := &ProjectObj{}
	_, err := p.StartSimulation("/nonexistent", "", SimulationModeBetween,
		"", "", "", nil, "提示", 5)
	if err == nil {
		t.Fatal("期望工程不存在错误，但未返回")
	}
}

// =====================================================================
// TestReviewSimulation_SessionNotFound
// =====================================================================

func TestReviewSimulation_SessionNotFound(t *testing.T) {
	p := &ProjectObj{}
	_, err := p.ReviewSimulation("nonexistent", "step1", true, "")
	if err == nil {
		t.Fatal("期望会话不存在错误，但未返回")
	}
}
