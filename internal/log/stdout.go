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

import "fmt"

type StdoutLogger struct {
	LogLevel LogLevel
}

func (l *StdoutLogger) SetLogLevel(level LogLevel) {
	l.LogLevel = level
}

func (l *StdoutLogger) Trace(s string) {
	if l.LogLevel <= Trace {
		fmt.Println(s)
	}
}

func (l *StdoutLogger) Debug(s string) {
	if l.LogLevel <= Debug {
		fmt.Println(s)
	}
}

func (l *StdoutLogger) Info(s string) {
	if l.LogLevel <= Info {
		fmt.Println(s)
	}
}

func (l *StdoutLogger) Warn(s string) {
	if l.LogLevel <= Warn {
		fmt.Println(s)
	}
}

func (l *StdoutLogger) Error(e error) {
	if l.LogLevel <= Trace {
		fmt.Println(e.Error())
	}
}
