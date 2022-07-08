package log

import "go.uber.org/zap"

var logger *zap.Logger
var sugar *zap.SugaredLogger

func init() {
	logger, _ = zap.NewProduction()
	defer logger.Sync()
	sugar = logger.Sugar()
}

func Fatalf(format string, v ...any) {
	sugar.Fatalf(format, v)
}

func Printf(format string, v ...any) {
	sugar.Infof(format, v)
}

func Errorf(format string, v ...any) {
	sugar.Errorf(format, v)
}

func Logger() *zap.Logger {
	return logger
}
