package zap_log

import (
	"context"
	"fmt"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm/logger"
)

func NewGormLogger() logger.Interface {
	return &zapLogger{GetFakeLogger().GetZapCore()}
}

type zapLogger struct {
	logger *zap.Logger
}

func (l *zapLogger) LogMode(level logger.LogLevel) logger.Interface {
	newLogger := *l
	return &newLogger
}

func (l *zapLogger) Info(ctx context.Context, s string, args ...interface{}) {
	l.logger.Info(fmt.Sprintf(s, args...))
}

func (l *zapLogger) Warn(ctx context.Context, s string, args ...interface{}) {
	l.logger.Warn(fmt.Sprintf(s, args...))
}

func (l *zapLogger) Error(ctx context.Context, s string, args ...interface{}) {
	l.logger.Error(fmt.Sprintf(s, args...))
}

func (l *zapLogger) Trace(ctx context.Context, begin time.Time, fc func() (string, int64), err error) {
	if err != nil {
		l.Error(ctx, "[%.3fms] [error] %v", time.Since(begin).Seconds()*1000, err)
		return
	}

	sql, rows := fc()
	l.Info(ctx, "[%.3fms] [rows:%v] %s", time.Since(begin).Seconds()*1000, rows, sql)
}
