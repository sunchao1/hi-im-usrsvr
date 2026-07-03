// Copyright 2026 sunchao1
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package mock

import (
	"context"
	"sync/atomic"
)

// SeqClient is a test double for seqsvr.Client.
type SeqClient struct {
	SidBase int64
	SeqBase int64
	sid     atomic.Int64
	seq     atomic.Int64
}

func (m *SeqClient) AllocSid(_ context.Context) (int64, error) {
	base := m.SidBase
	if base == 0 {
		base = 100000
	}
	return m.sid.Add(1) + base, nil
}

func (m *SeqClient) AllocSeq(_ context.Context, sid int64) (int64, error) {
	base := m.SeqBase
	if base == 0 {
		base = 1
	}
	_ = sid
	return m.seq.Add(1) + base, nil
}

func (m *SeqClient) Close() error { return nil }
