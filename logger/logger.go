package logger

import "go.uber.org/zap"

func Init() {
	l, _ := zap.NewDevelopment()
	zap.ReplaceGlobals(l)
}
