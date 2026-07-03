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

package session_test

import (
	"context"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/sunchao1/hi-im-api/pkg/rediskey"

	"github.com/sunchao1/hi-im-usrsvr/internal/session"
)

func TestSetSidOnlineAndGetSidAttr(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	store := session.NewStore(rdb, 300)
	ctx := context.Background()

	const sid, uid, cid = uint64(42), uint64(1001), uint64(9001)
	const nid uint32 = 20111

	if err := store.SetSidOnline(ctx, sid, uid, cid, nid); err != nil {
		t.Fatalf("SetSidOnline: %v", err)
	}

	attr, err := store.GetSidAttr(ctx, sid)
	if err != nil {
		t.Fatalf("GetSidAttr: %v", err)
	}
	if attr.Uid != uid || attr.Cid != cid || attr.Nid != nid {
		t.Fatalf("unexpected attr: %+v", attr)
	}

	key := rediskey.SidAttrKey(sid)
	if !mr.Exists(key) {
		t.Fatalf("expected key %s", key)
	}
}

func TestCleanSession(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	store := session.NewStore(rdb, 300)
	ctx := context.Background()

	const sid, uid, cid = uint64(7), uint64(88), uint64(99)
	if err := store.SetSidOnline(ctx, sid, uid, cid, 20111); err != nil {
		t.Fatal(err)
	}
	if err := store.CleanSession(ctx, sid, uid); err != nil {
		t.Fatal(err)
	}
	if _, err := store.GetSidAttr(ctx, sid); err == nil {
		t.Fatal("expected missing attr after clean")
	}
}
