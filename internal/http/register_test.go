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

package imhttp_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/sunchao1/hi-im-api/pkg/errno"

	imhttp "github.com/sunchao1/hi-im-usrsvr/internal/http"
	"github.com/sunchao1/hi-im-usrsvr/test/mock"
)

func TestRegisterHTTP(t *testing.T) {
	seqClient := &mock.SeqClient{SidBase: 100000}
	r := imhttp.NewRouter(imhttp.RegisterDeps{Seq: seqClient})

	req := httptest.NewRequest(http.MethodGet, "/im/register?uid=1&nation=86&city=beijing", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}

	var resp imhttp.RegisterResponse
	if err := json.NewDecoder(rec.Body).Decode(&resp); err != nil {
		t.Fatal(err)
	}
	if resp.Code != errno.OK {
		t.Fatalf("code=%d errmsg=%s", resp.Code, resp.ErrMsg)
	}
	if resp.SID <= 0 {
		t.Fatalf("expected sid>0, got %d", resp.SID)
	}
	if resp.UID != 1 || resp.Nation != "86" || resp.City != "beijing" {
		t.Fatalf("unexpected resp: %+v", resp)
	}
}

func TestRegisterMissingParams(t *testing.T) {
	seqClient := &mock.SeqClient{}
	r := imhttp.NewRouter(imhttp.RegisterDeps{Seq: seqClient})

	req := httptest.NewRequest(http.MethodGet, "/im/register?uid=1", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", rec.Code)
	}
}
