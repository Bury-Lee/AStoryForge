package config

type Log struct {
	App      string `yaml:"app"`       // Application name used for log files.
	Dir      string `yaml:"dir"`       // Directory used to store log files.
	LogLevel string `yaml:"log_level"` // Minimum log level to record.
}
