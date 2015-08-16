package base

import "hypermind.cn/talon/logging"

// 创建日志记录器。
func NewLogger(logname string) logging.Logger {
	fileLog := logging.NewFileLogger("log", logname, 200, 10)
	fileLog.SetPosition(logging.POSITION_SINGLE)

	fileLog.Initialize()
	return fileLog
}
