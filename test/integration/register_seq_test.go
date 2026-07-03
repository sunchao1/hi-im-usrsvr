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

//go:build integration

package integration_test

import (
	"context"
	"testing"

	"github.com/sunchao1/hi-im-usrsvr/test/mock"
)

func TestRegisterSeqMonotonic(t *testing.T) {
	seqClient := &mock.SeqClient{SidBase: 200000}
	s1, err := seqClient.AllocSid(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	s2, err := seqClient.AllocSid(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if s2 <= s1 {
		t.Fatalf("expected monotonic sid: %d then %d", s1, s2)
	}
}
