package flags

import (
	"AStoryForge/core"
	"context"
	"flag"
	"fmt"
	"os"

	_ "embed"

	"github.com/sashabaranov/go-openai"
)

// CLIArg defines the supported command line flags.
type CLIArg struct {
	InitSetting bool   `flag:"initSetting" description:"初始化配置文件"`
	Setting     string `flag:"setting" description:"配置文件路径"`
	ConnectTest bool   `flag:"connectTest" description:"连通性测试"`
	Doc         bool   `flag:"doc" description:"显示帮助文档"`
	Help        bool   `flag:"help" description:"显示帮助文档"`
}

// Embedded sample configuration content.

//go:embed settings.yaml
var defaultSettings string

func ParseFlag() CLIArg {
	var args CLIArg

	// Register command line flags.
	flag.BoolVar(&args.InitSetting, "initSetting", false, "初始化配置文件")
	flag.BoolVar(&args.InitSetting, "i", false, "初始化配置文件（短参数）")

	flag.StringVar(&args.Setting, "setting", "settings.yaml", "配置文件路径")
	flag.StringVar(&args.Setting, "s", "settings.yaml", "配置文件路径（短参数）")

	flag.BoolVar(&args.ConnectTest, "connectTest", false, "连通性测试")
	flag.BoolVar(&args.ConnectTest, "c", false, "连通性测试（短参数）")

	// Customize usage output.
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "用法: %s [选项]\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "选项:\n")
		flag.VisitAll(func(f *flag.Flag) {
			// Only show long flags to avoid duplicate aliases.
			if len(f.Name) > 1 {
				fmt.Fprintf(os.Stderr, "  -%s\n", f.Name)
				fmt.Fprintf(os.Stderr, "    \t%s\n", f.Usage)
			}
		})
		fmt.Fprintf(os.Stderr, "\n示例:\n")
		fmt.Fprintf(os.Stderr, "  %s -initSetting\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -setting=/path/to/config.yaml\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "  %s -connectTest\n", os.Args[0])
	}

	flag.Parse()

	// Execute the selected CLI action.
	switch {
	case args.Help:
		// Show usage information.
		flag.Usage()
		os.Exit(0)
	case args.Doc:
		Doc()
		os.Exit(0)
	case args.InitSetting:
		InitSetting(args.Setting)
	case args.ConnectTest:
		ConnectTest()
	case args.Setting != "settings.yaml":
		// Load the custom config path even when no other action is requested.
		fmt.Printf("配置文件已设置为: %s\n", args.Setting)
		core.Config = core.ReadConf(args.Setting)
	}
	return args
}

// InitSetting writes a sample configuration file into the working directory.
// path: target file path, defaults to settings.yaml.
func InitSetting(path string) {
	if path == "" {
		path = core.DefultFilePath
	}

	// Check whether the target file already exists.
	if _, err := os.Stat(path); err == nil {
		fmt.Printf("配置文件 %s 已存在，是否覆盖？(y/N): ", path)
		var answer string
		fmt.Scanln(&answer)
		if answer != "y" && answer != "Y" {
			fmt.Println("已取消初始化。")
			os.Exit(0)
		}
		fmt.Println("正在覆盖已有配置文件...")
	}

	// Write the embedded sample config.
	err := os.WriteFile(path, []byte(defaultSettings), 0644)
	if err != nil {
		fmt.Fprintf(os.Stderr, "写入配置文件失败: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("示例配置文件已生成: %s\n", path)
	fmt.Println("请根据实际环境修改其中的 host 和 apiKey 等配置项。")
	os.Exit(0)
}

func Doc() {
	// Print a short local help summary.
}

func ConnectTest() {
	fmt.Println("正在测试 AI 模型连通性...")
	core.Config = core.ReadConf("settings.yaml")
	core.InitAI()

	if len(core.AIClient) == 0 {
		fmt.Println("错误: 没有可用的 AI 客户端")
		os.Exit(1)
	}

	fmt.Printf("共加载 %d 个 结构化AI 模型，连通性测试通过\n", len(core.AIClient))
	for i, conf := range core.Config.AI {
		fmt.Printf("  [%d] %s (%s) -> %s\n", i+1, conf.Name, conf.Model, conf.Host)
		message := []openai.ChatCompletionMessage{{
			Role:    openai.ChatMessageRoleUser,
			Content: "当前为连通性测试,请回复\"收到\"",
		},
		}
		resp, err := core.AIClient[i].Client.CreateChatCompletion(
			context.Background(),
			openai.ChatCompletionRequest{
				Model:       conf.Model,
				Messages:    message,
				Temperature: conf.Temperature, // Increase response diversity.
				MaxTokens:   conf.MaxTokens,   // Limit response length.
			},
		)
		if err == nil {
			fmt.Println("ai回复" + resp.Choices[0].Message.Content)
		}
	}

	fmt.Printf("正在测试最终编译AI")
	CplAI := core.Config.CompilerAI
	fmt.Printf("  %s (%s) -> %s\n", CplAI.Name, CplAI.Model, CplAI.Host)
	message := []openai.ChatCompletionMessage{{
		Role:    openai.ChatMessageRoleUser,
		Content: "当前为连通性测试,请回复\"收到\"",
	},
	}
	resp, err := core.CompilerAI.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:       CplAI.Model,
			Messages:    message,
			Temperature: CplAI.Temperature, // Increase response diversity.
			MaxTokens:   CplAI.MaxTokens,   // Limit response length.
		},
	)
	if err == nil {
		fmt.Println("ai回复" + resp.Choices[0].Message.Content)
	}
	os.Exit(0)
}
