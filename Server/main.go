package main

import (
	"AStoryForge/core"
	"AStoryForge/flags"
	"AStoryForge/router"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func main() {
	args := flags.ParseFlag()
	core.Config = core.ReadConf(args.Setting)
	core.InitLogrus()
	core.InitAI()
	logrus.Debugf("配置:%+v", core.Config)

	//启动后端路由
	gin.SetMode(core.Config.RunMode)
	engine := gin.Default()

	//路由注册
	router.RegisterRouter(engine)

	//注册路由
	engine.Run(core.Config.WebSocket.Host)
}
