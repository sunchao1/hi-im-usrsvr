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

package hub_test

import (
	"context"
	"sync"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	imv1 "github.com/sunchao1/hi-im-api/gen/go/im/v1"
	"github.com/sunchao1/hi-im-api/pkg/im/cmd"
	"github.com/sunchao1/hi-im-api/pkg/im/header"
	"google.golang.org/protobuf/proto"

	"github.com/sunchao1/hi-im-usrsvr/internal/hub"
	"github.com/sunchao1/hi-im-usrsvr/internal/session"
	"github.com/sunchao1/hi-im-usrsvr/test/mock"
)

type stubSender struct {
	sendFn func(cmd, destNid uint32, payload []byte) error
}

func (s *stubSender) AsyncSend(cmd, destNid uint32, payload []byte) error {
	if s.sendFn != nil {
		return s.sendFn(cmd, destNid, payload)
	}
	return nil
}

func TestOnlineHandlerRedisAndAck(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatal(err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	sess := session.NewStore(rdb, 300)
	mockSeq := &mock.SeqClient{SeqBase: 10}

	var ackMu sync.Mutex
	var lastAck *imv1.OnlineAck

	stub := &stubSender{sendFn: func(_ uint32, _ uint32, payload []byte) error {
		if len(payload) < header.Size {
			return nil
		}
		h, _ := header.Unmarshal(payload[:header.Size])
		if h.Cmd != cmd.CMD_ONLINE_ACK {
			return nil
		}
		var ack imv1.OnlineAck
		if err := proto.Unmarshal(payload[header.Size:], &ack); err != nil {
			return err
		}
		ackMu.Lock()
		lastAck = &ack
		ackMu.Unlock()
		return nil
	}}

	onlineH := hub.NewOnlineHandler(sess, mockSeq, stub, true)

	const gatewayCid, sessSid, uid = uint64(555), uint64(100001), uint64(42)
	const gatewayNid uint32 = 20111

	frame, err := hub.PackOnlineFrame(gatewayCid, sessSid, uid, gatewayNid, &imv1.Online{
		Uid: uid,
		Sid: sessSid,
	})
	if err != nil {
		t.Fatal(err)
	}
	onlineH.Handle(cmd.CMD_ONLINE, gatewayNid, frame)

	attr, err := sess.GetSidAttr(context.Background(), sessSid)
	if err != nil {
		t.Fatalf("GetSidAttr: %v", err)
	}
	if attr.Uid != uid || attr.Cid != gatewayCid || attr.Nid != gatewayNid {
		t.Fatalf("unexpected attr: %+v", attr)
	}

	ackMu.Lock()
	ack := lastAck
	ackMu.Unlock()
	if ack == nil {
		t.Fatal("expected ONLINE_ACK")
	}
	if ack.GetSeq() == 0 || ack.GetCode() != 0 {
		t.Fatalf("unexpected ack: %+v", ack)
	}
}
