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
	"fmt"

	"github.com/briandowns/spinner"
)

type SpinnerLogger struct {
	Spinner  *spinner.Spinner
	logLevel LogLevel
}

func NewSpinnerLogger(spin *spinner.Spinner) *SpinnerLogger {
	spin.FinalMSG = "done"
	return &SpinnerLogger{
		Spinner: spin,
	}
}

func (l *SpinnerLogger) SetLogLevel(level LogLevel) {
	l.logLevel = level
}

func (l *SpinnerLogger) Trace(s string) {
	if l.logLevel <= Trace && l.Spinner != nil {
		l.Spinner.Suffix = fmt.Sprintf(" %s...", s)
	}
}

func (l *SpinnerLogger) Debug(s string) {
	if l.logLevel <= Debug && l.Spinner != nil {
		l.Spinner.Suffix = fmt.Sprintf(" %s...", s)
	}
}

func (l *SpinnerLogger) Info(s string) {
	if l.logLevel <= Info && l.Spinner != nil {
		l.Spinner.Suffix = fmt.Sprintf(" %s...", s)
	}
}

func (l *SpinnerLogger) Warn(s string) {
	if l.logLevel <= Warn && l.Spinner != nil {
		l.Spinner.Suffix = fmt.Sprintf(" %s...", s)
	}
}

func (l *SpinnerLogger) Error(e error) {
	if l.logLevel <= Error && l.Spinner != nil {
		l.Spinner.Suffix = fmt.Sprintf(" Error: %s...", e.Error())
	}
}
