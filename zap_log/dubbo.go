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
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Debug uses fmt.Sprint to construct and log a message.
func (s *ZapLoggerFake) Debug(args ...interface{}) {
	args = append(args, zap.String("traceId", "123456"))
	s.logAndZap(zap.DebugLevel, "", args)
}

// Info uses fmt.Sprint to construct and log a message.
func (s *ZapLoggerFake) Info(args ...interface{}) {
	s.logAndZap(zap.InfoLevel, "", args)
}

// Warn uses fmt.Sprint to construct and log a message.
func (s *ZapLoggerFake) Warn(args ...interface{}) {
	s.logAndZap(zap.WarnLevel, "", args)
}

// Error uses fmt.Sprint to construct and log a message.
func (s *ZapLoggerFake) Error(args ...interface{}) {
	s.logAndZap(zap.ErrorLevel, "", args)
}

// Fatal uses fmt.Sprint to construct and log a message, then calls os.Exit.
func (s *ZapLoggerFake) Fatal(args ...interface{}) {
	s.logAndZap(zap.FatalLevel, "", args)
}

// Debugf uses fmt.Sprintf to log a templated message.
func (s *ZapLoggerFake) Debugf(template string, args ...interface{}) {
	s.log(zap.DebugLevel, template, args, nil)
}

// Infof uses fmt.Sprintf to log a templated message.
func (s *ZapLoggerFake) Infof(template string, args ...interface{}) {
	s.log(zap.InfoLevel, template, args, nil)
}

// Warnf uses fmt.Sprintf to log a templated message.
func (s *ZapLoggerFake) Warnf(template string, args ...interface{}) {
	s.log(zap.WarnLevel, template, args, nil)
}

// Errorf uses fmt.Sprintf to log a templated message.
func (s *ZapLoggerFake) Errorf(template string, args ...interface{}) {
	s.log(zap.ErrorLevel, template, args, nil)
}

// Fatalf uses fmt.Sprintf to log a templated message, then calls os.Exit.
func (s *ZapLoggerFake) Fatalf(template string, args ...interface{}) {
	s.log(zap.FatalLevel, template, args, nil)
}

func (s *ZapLoggerFake) logAndZap(lvl zapcore.Level, template string, fmtArgs []interface{}) {
	s.realLogAndZap(lvl, template, fmtArgs)
}

func (s *ZapLoggerFake) realLogAndZap(lvl zapcore.Level, template string, fmtArgs []interface{}) {

	if lvl < zap.DPanicLevel && !s.logger.Core().Enabled(lvl) {
		return
	}
	args, context := s.sweetenFields(fmtArgs)
	msg := getMessage(template, args)
	if ce := s.logger.Check(lvl, msg); ce != nil {
		ce.Write(context...)
	}
}

func (s *ZapLoggerFake) log(lvl zapcore.Level, template string, fmtArgs []interface{}, context []zap.Field) {
	// If logging at this level is completely disabled, skip the overhead of
	// string formatting.
	if lvl < zap.DPanicLevel && !s.logger.Core().Enabled(lvl) {
		return
	}

	msg := getMessage(template, fmtArgs)
	if ce := s.logger.Check(lvl, msg); ce != nil {
		ce.Write(context...)
	}
}

// getMessage format with Sprint, Sprintf, or neither.
func getMessage(template string, fmtArgs []interface{}) string {
	if len(fmtArgs) == 0 {
		return template
	}

	if template != "" {
		return fmt.Sprintf(template, fmtArgs...)
	}

	if len(fmtArgs) == 1 {
		if str, ok := fmtArgs[0].(string); ok {
			return str
		}
	}
	return fmt.Sprint(fmtArgs...)
}

func (s *ZapLoggerFake) sweetenFields(args []interface{}) ([]interface{}, []zap.Field) {

	if len(args) == 0 {
		return nil, nil
	}

	// Allocate enough space for the worst case; if users pass only structured
	// fields, we shouldn't penalize them with extra allocations.
	fields := make([]zap.Field, 0, len(args))
	var newArgs []interface{}

	for i := 0; i < len(args); i++ {
		// This is a strongly-typed field. Consume it and move on.
		if f, ok := args[i].(zap.Field); ok {
			fields = append(fields, f)
		} else {
			newArgs = append(newArgs, args[i])
		}
	}

	return newArgs, fields
}
