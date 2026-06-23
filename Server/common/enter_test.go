package common

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"AStoryForge/function/story_struct"
)

// 测试堆栈打印
func TestGetStack(t *testing.T) {
	t.Log("=== 测试堆栈打印 ===")

	// 直接调用 GetStack
	stack := GetStack(1) // skip=1 跳过 GetStack 本身
	t.Logf("堆栈信息:\n%s", string(stack))
}

// 测试多层调用栈
func TestNestedStack(t *testing.T) {
	t.Log("=== 测试多层嵌套调用栈 ===")
	level3()
}

func level3() {
	level2()
}

func level2() {
	level1()
}

func level1() {
	stack := GetStack(1) // skip=1 跳过 level1 本身
	println("===== 完整堆栈 =====")
	println(string(stack))
	println("==================")

}

// 测试 NewStackError 是否打印堆栈
func TestNewStackError(t *testing.T) {
	t.Log("=== 测试 NewStackError 打印堆栈 ===")

	err := Error(10001, "这是一个测试错误")
	t.Logf("错误信息: %v", err)
}

// 测试在真实场景中调用
func TestRealScenario(t *testing.T) {
	t.Log("=== 测试真实场景 ===")

	// 模拟一个函数调用链
	getUser()
}

func getUser() {
	validateUser()
}

func validateUser() {
	checkPermission()
}

func checkPermission() {
	err := Error(10002, "用户权限不足")
	_ = err
}

// 当前的错误信息仅仅只是调用stack函数,似乎有些过于简陋,我认为可以优化 worker 崩溃时的错误信息可读性。当 worker 因 panic 退出时，日志会清晰展示：

// 崩溃发生的具体文件和行号

// 完整的函数调用链

// 触发崩溃的源代码行内容

// 帮助使用者在不依赖调试器的情况下快速定位问题根因。

func TestMakeMermaidFromExample(t *testing.T) {
	wd, err := os.Getwd()
	if err != nil {
		t.Fatalf("获取工作目录失败: %v", err)
	}

	jsonPath := filepath.Join(wd, "..", "function", "story_struct", "example", "星纪元.json")
	mermaidPath := filepath.Join(wd, "..", "function", "story_struct", "example", "星纪元.mermaid")

	data, err := os.ReadFile(jsonPath)
	if err != nil {
		t.Fatalf("读取示例文件失败: %v", err)
	}

	var project story_struct.ProjectObj
	if err := json.Unmarshal(data, &project); err != nil {
		t.Fatalf("反序列化示例文件失败: %v", err)
	}

	mermaid, message := project.MakeMermaid()

	if err := os.WriteFile(mermaidPath, []byte(mermaid), 0644); err != nil {
		t.Fatalf("写入 mermaid 文件失败: %v", err)
	}

	t.Logf("mermaid 已保存到: %s", mermaidPath)
	t.Logf("MakeMermaid 消息: %s", message)
}
