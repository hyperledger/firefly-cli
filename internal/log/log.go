// Copyright © 2024 Kaleido, Inc.
//
// SPDX-License-Identifier: Apache-2.0
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package log

import (
	"context"
)

type LogLevel int

const (
	Trace LogLevel = iota
	Debug
	Info
	Warn
	Error
)

type Logger interface {
	SetLogLevel(l LogLevel)
	Trace(s string)
	Debug(s string)
	Info(s string)
	Warn(s string)
	Error(e error)
}

type (
	ctxLogKey       struct{}
	ctxVerbosityKey struct{}
)

func WithLogger(ctx context.Context, log Logger) context.Context {
	return context.WithValue(ctx, ctxLogKey{}, log)
}

func LoggerFromContext(ctx context.Context) Logger {
	return ctx.Value(ctxLogKey{}).(Logger)
}

func WithVerbosity(ctx context.Context, verbose bool) context.Context {
	return context.WithValue(ctx, ctxVerbosityKey{}, verbose)
}

func VerbosityFromContext(ctx context.Context) bool {
	return ctx.Value(ctxVerbosityKey{}).(bool)
}
