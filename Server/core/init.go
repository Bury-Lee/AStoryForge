package core

import (
	"AStoryForge/config"
	"fmt"
	"os"

	"go.yaml.in/yaml/v2"
)

const DefultFilePath = "settings.yaml"

func ReadConf(path string) *config.Config {
	if path == "" {
		path = DefultFilePath
	}

	byteData, err := os.ReadFile(path) // Read the config file.
	if err != nil {
		panic(err)
	}
	var c = new(config.Config)
	err = yaml.Unmarshal(byteData, c) // Bind YAML values into the config struct.
	if err != nil {
		panic(fmt.Sprintf("yaml文件格式错误 %s", err))
	}
	return c
}
