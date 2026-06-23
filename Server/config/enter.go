package config

type Config struct {
	RunMode    string    `yaml:"run_mode"`   // Runtime mode: develop or release.
	Log        Log       `yaml:"log"`        // Logging configuration.
	AI         []AI      `yaml:"ai"`         // AI persona configuration list.
	CompilerAI AI        `yaml:"Compilerai"` //用于构建最终小说的AI,推荐使用deepseek,opus等长上下文模型
	WebSocket  WebSocket `yaml:"websocket"`  // WebSocket server configuration.
}
