package story_struct

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// 对于接口的具体实现
func ListProjects(rootDir string) ([]ProjectMeta, error) { //TODO:写单元测试
	// 递归读取目录,扫描是否有json文件,如果有,是否符合工程元信息格式,进行反序列化,添加到列表中
	// 如果符合,则添加到列表中,最后返回列表
	list := make([]ProjectMeta, 0)
	if rootDir == "" {
		return nil, errors.New("未输入目录")
	}

	// 检查目录是否存在
	if _, err := os.Stat(rootDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("目录不存在: %s", rootDir)
	}

	// 遍历目录
	err := filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录，只处理文件
		if info.IsDir() {
			return nil
		}

		// 只处理 .json 文件
		if !strings.HasSuffix(strings.ToLower(info.Name()), ".json") {
			return nil
		}

		// 读取 JSON 文件
		data, err := os.ReadFile(path)
		if err != nil {
			// 记录错误但继续处理其他文件
			fmt.Printf("读取文件失败 %s: %v\n", path, err)
			return nil
		}

		// 尝试反序列化为 ProjectMeta
		var meta ProjectMeta
		err = json.Unmarshal(data, &meta)
		if err != nil {
			// 不是有效的 JSON 或不符合 ProjectMeta 格式，跳过
			return nil
		}

		// 符合条件，添加到列表
		list = append(list, meta)

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("遍历目录失败: %v", err)
	}

	return list, nil
}

// Project 简短描述完整工程聚合
// 说明: Meta 对应 project.json, Events 和 Entities 对应节点目录
type Project struct {
	Meta         ProjectMeta  `json:"meta"`                    // 工程元信息
	WorldSetting WorldSetting `json:"world_setting,omitempty"` // 世界设置
	Events       []Event      `json:"events,omitempty"`        // 事件节点集合
	Entities     []Entity     `json:"entities,omitempty"`      // 实体节点集合
}

func CreateProject(input ProjectMeta) error {
	// 1. 标题完备性检查
	if input.Title == "" {
		return fmt.Errorf("未设置标题")
	}

	// 去除首尾空格后的标题检查
	trimmedTitle := strings.TrimSpace(input.Title)
	if trimmedTitle == "" {
		return fmt.Errorf("标题不能为空白字符")
	}
	input.Title = trimmedTitle

	// 2. 标题长度限制（根据实际需求调整）
	if len([]rune(input.Title)) > 200 {
		return fmt.Errorf("标题长度不能超过200个字符")
	}

	// 3. 作者完备性检查
	if input.Author == "" {
		input.Author = "未知"
	}
	// 作者长度限制
	if len([]rune(input.Author)) > 100 {
		return fmt.Errorf("作者名称长度不能超过100个字符")
	}

	// 5. 时间字段完备性
	var EmptyTime time.Time
	if input.CreatedAt.Equal(EmptyTime) {
		input.CreatedAt = time.Now()
	}
	input.UpdatedAt = time.Now()

	// 6. 设置默认值（可选）
	if input.Settings == nil {
		input.Settings = make(map[string]string) // 避免 nil map
	}

	// 7. 数值字段合法性检查（如果有传入值）
	if input.EventCount < 0 {
		input.EventCount = 0
	}
	if input.EntityCount < 0 {
		input.EntityCount = 0
	}

	// 8. 业务逻辑校验（示例）
	// 检查根目录是否存在（如果需要）
	if input.RootDir != "" {
		if _, err := os.Stat(input.RootDir); os.IsNotExist(err) {
			//TODO:创建目录
			return fmt.Errorf("工程根目录不存在: %s", input.RootDir)
		}
	}
	if input.RootDir == "" { //使用默认目录
		input.RootDir = "./projects"
	}

	var result ProjectObj
	result.ProjectMeta = input

	if err := result.SaveProjectObj(); err != nil {
		return fmt.Errorf("新建工程失败: %w", err)
	}

	return nil
}

// UpdateProjectMeta 更新工程元信息
// 参数: rootDir - 工程根目录, projectID - 工程 ID, meta - 标题, 作者, 时间系统, 规则, 设置等更新内容
// 返回: meta - 更新后的工程元信息,error - 更新失败信息
// 说明: 用于工程设置页统一修改 project.json 中的基础配置
func UpdateProjectMeta(rootDir string, projectID string, meta ProjectMeta) (*ProjectMeta, error) {
	//先读取工程元信息,更新,保存工程元信息
	_ = projectID
	if meta.RootDir == "" {
		meta.RootDir = rootDir
	}
	if meta.RootDir == "" {
		//使用默认目录
		meta.RootDir = "./projects/" + meta.Title
	}
	// 检查根目录是否存在
	if _, err := os.Stat(meta.RootDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("文件不存在: %s", meta.RootDir)
	}

	var result ProjectObj
	result.ProjectMeta = meta
	if err := result.SaveProjectObj(); err != nil {
		return nil, fmt.Errorf("更新工程元信息失败: %w", err)
	}
	result.SaveProjectObj()
	return &result.ProjectMeta, nil
}
