package test_router

import (
	"AStoryForge/common"
	"AStoryForge/function/test_function"

	"github.com/gin-gonic/gin"
)

func TestView(c *gin.Context) {
	common.OkWithMsg("连接成功", c)
}

type FTrequest struct {
	Name string `json:"name"`
	Test string `json:"test"`
}

func FunctionTestView(c *gin.Context) {
	var req FTrequest
	c.ShouldBindJSON(&req)

	name := test_function.TestFunction(req.Name, req.Test)

	common.OkWithMsg(" name: "+name+"测试成功", c)
}
