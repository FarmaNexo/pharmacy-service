// pkg/logger/logger.go
package logger

import (
	"go.uber.org/zap"
)

func NewLogger(environment string) (*zap.Logger, error) {
	switch environment {
	case "production", "uat":
		return zap.NewProduction()
	default:
		return zap.NewDevelopment()
	}
}
