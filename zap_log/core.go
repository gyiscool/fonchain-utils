/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package zap_log

import (
	"github.com/natefinch/lumberjack"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type DuInfo struct {
	Config *Config `yaml:"logger"`
}

type Config struct {
	LumberjackConfig *lumberjack.Logger `yaml:"lumberjack-config"`
	ZapConfig        *zap.Config        `yaml:"zap-config"`
	CallerSkip       int                `yaml:"CallerSkip"`
}

type ZapLoggerFake struct {
	logger *zap.Logger
}

// Debug logs a message at DebugLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Debug(msg string, fields ...zap.Field) {
	zapCoreLogger(GetFakeLogger().GetZapCore(), zap.DebugLevel, msg, fields...)
}

// Info logs a message at InfoLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Info(msg string, fields ...zap.Field) {
	zapCoreLogger(GetFakeLogger().GetZapCore(), zap.InfoLevel, msg, fields...)
}

// Warn logs a message at WarnLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Warn(msg string, fields ...zap.Field) {
	zapCoreLogger(GetFakeLogger().GetZapCore(), zap.WarnLevel, msg, fields...)
}

// Error logs a message at ErrorLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
func Error(msg string, fields ...zap.Field) {
	zapCoreLogger(GetFakeLogger().GetZapCore(), zap.ErrorLevel, msg, fields...)
}

// DPanic logs a message at DPanicLevel. The message includes any fields
// passed at the log site, as well as any fields accumulated on the logger.
//
// If the logger is in development mode, it then panics (DPanic means
// "development panic"). This is useful for catching errors that are
// recoverable, but shouldn't ever happen.
func DPanic(msg string, fields ...zap.Field) {
	zapCoreLogger(GetFakeLogger().GetZapCore(), zap.DPanicLevel, msg, fields...)
}

// Panic logs a message at PanicLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then panics, even if logging at PanicLevel is disabled.
func Panic(msg string, fields ...zap.Field) {
	zapCoreLogger(GetFakeLogger().GetZapCore(), zap.PanicLevel, msg, fields...)
}

// Fatal logs a message at FatalLevel. The message includes any fields passed
// at the log site, as well as any fields accumulated on the logger.
//
// The logger then calls os.Exit(1), even if logging at FatalLevel is
// disabled.
func Fatal(msg string, fields ...zap.Field) {
	zapCoreLogger(GetFakeLogger().GetZapCore(), zap.FatalLevel, msg, fields...)
}

func zapCoreLogger(loggerObj *zap.Logger, level zapcore.Level, msg string, fields ...zap.Field) {
	realZap(loggerObj, level, msg, fields...)
}

func realZap(loggerObj *zap.Logger, level zapcore.Level, msg string, fields ...zap.Field) {
	writeZapCoreLogger(loggerObj, level, msg, fields...)
}

func writeZapCoreLogger(loggerObj *zap.Logger, level zapcore.Level, msg string, fields ...zap.Field) {
	if ce := loggerObj.Check(level, msg); ce != nil {
		ce.Write(fields...)
	}
}

// InitLogger use for init logger by @conf
func initLogger(conf *Config) *ZapLoggerFake {
	var (
		_      *zap.Logger
		config = &Config{}
	)
	if conf == nil || conf.ZapConfig == nil {
		zapLoggerEncoderConfig := zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			EncodeLevel:    zapcore.CapitalColorLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		}
		config.ZapConfig = &zap.Config{
			Level:            zap.NewAtomicLevelAt(zap.InfoLevel),
			Development:      true,
			Encoding:         "console",
			EncoderConfig:    zapLoggerEncoderConfig,
			OutputPaths:      []string{"stderr"},
			ErrorOutputPaths: []string{"stderr"},
		}
	} else {
		config.ZapConfig = conf.ZapConfig
	}

	if conf != nil {
		config.CallerSkip = conf.CallerSkip
	}

	if config.CallerSkip == 0 { //因为包装了两层，所以设置3次
		config.CallerSkip = 2
	}

	var zapLogger *zap.Logger
	if conf == nil || conf.LumberjackConfig == nil {
		zapLogger, _ = config.ZapConfig.Build(zap.AddCaller(), zap.AddCallerSkip(config.CallerSkip))
	} else {
		config.LumberjackConfig = conf.LumberjackConfig
		zapLogger = initZapLoggerWithSyncer(config)
	}

	//zapFakeLog = &ZapLoggerFake{logger: zapLogger}

	return &ZapLoggerFake{logger: zapLogger}
}

func (m *ZapLoggerFake) GetZapCore() *zap.Logger {
	return m.logger
}

func (l *ZapLoggerFake) ZInfo(s string, fields ...zap.Field) {
	l.logger.Info(s, fields...)
}

// initZapLoggerWithSyncer init zap Logger with syncer
func initZapLoggerWithSyncer(conf *Config) *zap.Logger {

	var fields []zapcore.Field

	if len(conf.ZapConfig.InitialFields) > 0 {
		for key, value := range conf.ZapConfig.InitialFields {
			fields = append(fields, zap.Any(key, value))
		}
	}

	core := zapcore.NewCore(
		conf.getEncoder(),
		conf.getLogWriter(),
		zap.NewAtomicLevelAt(conf.ZapConfig.Level.Level()),
	)
	if len(fields) >= 1 {
		core = core.With(fields)
	}

	return zap.New(core, zap.AddCaller(), zap.AddCallerSkip(conf.CallerSkip))
}

// getEncoder get encoder by config, zapcore support json and console encoder
func (c *Config) getEncoder() zapcore.Encoder {
	if c.ZapConfig.Encoding == "json" {
		return zapcore.NewJSONEncoder(c.ZapConfig.EncoderConfig)
	} else if c.ZapConfig.Encoding == "console" {
		return zapcore.NewConsoleEncoder(c.ZapConfig.EncoderConfig)
	}
	return nil
}

// getLogWriter get Lumberjack writer by LumberjackConfig
func (c *Config) getLogWriter() zapcore.WriteSyncer {
	return zapcore.AddSync(c.LumberjackConfig)
}
