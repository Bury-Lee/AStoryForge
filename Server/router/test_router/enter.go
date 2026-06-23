package test_router

import "github.com/gin-gonic/gin"

func TestRouter(r *gin.RouterGroup) {
	r.GET("/test", TestView)

}
