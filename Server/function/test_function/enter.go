package test_function

import "github.com/sirupsen/logrus"

func TestFunction(name string, text string) string {
	// 测试函数
	logrus.Debugf("测试者: %s, 测试参数: %s", name, text)
	return name
}
