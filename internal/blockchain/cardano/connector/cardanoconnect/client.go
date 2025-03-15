// Copyright © 2025 IOG Singapore and SundaeSwap, Inc.
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

package cardanoconnect

import (
	"context"
)

type Cardanoconnect struct {
	ctx context.Context
}

func NewCardanoconnect(ctx context.Context) *Cardanoconnect {
	return &Cardanoconnect{
		ctx: ctx,
	}
}

func (c *Cardanoconnect) Name() string {
	return "cardanoconnect"
}

func (c *Cardanoconnect) Port() int {
	return 3000
}
