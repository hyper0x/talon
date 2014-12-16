package base

import "hypermind.cn/talon/logging"

// 创建日志记录器。
func NewLogger() logging.Logger {
	return logging.NewSimpleLogger()
}
