package router

import (
	"AStoryForge/core"
	"AStoryForge/middleware"
	"AStoryForge/router/test_router"

	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

func RegisterRouter(r *gin.Engine) *gin.Engine {
	// 注册路由
	if r == nil {
		logrus.Fatalf("程序出错,路由引擎不能为空,请联系开发者")
	}
	r.Static("/web", "static") //静态文件目录,注册个路由到时候写网页?而且还可以用于获取图片//TODO:后期把这个路径移到配置里,不要写死在代码里
	nr := r.Group("/api")      //TODO:测试使用无前缀api,开发完成了要给app组加上/api的前缀,nr := r.Group("/api")这样

	if core.Config.RunMode == "debug" {
		nr.Use(middleware.RequestLogMiddleware)
	}
	test_router.TestRouter(nr)
	return r
}
