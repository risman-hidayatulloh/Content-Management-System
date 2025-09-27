package observability

import "go.uber.org/zap"

func NewLogger() *zap.Logger {
	l, _ := zap.NewProduction()
	return l
}
