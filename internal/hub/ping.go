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

package hub

import (
	"context"
	"log/slog"

	"github.com/sunchao1/hi-im-api/pkg/im/cmd"
	"github.com/sunchao1/hi-im-api/pkg/im/header"
	"github.com/sunchao1/hi-im-usrsvr/internal/session"
)

// PingHandler refreshes session TTL and responds with CMD_PONG.
type PingHandler struct {
	sess *session.Store
	hub  Sender
	log  *slog.Logger
}

// NewPingHandler creates a PING handler.
func NewPingHandler(sess *session.Store, hub Sender) *PingHandler {
	return &PingHandler{sess: sess, hub: hub, log: slog.Default()}
}

// Handle implements hubclient.Handler for CMD_PING.
func (h *PingHandler) Handle(_ uint32, _ uint32, payload []byte) {
	ctx := context.Background()
	if len(payload) < header.Size {
		return
	}
	hdr, err := header.Unmarshal(payload[:header.Size])
	if err != nil {
		return
	}

	sid := hdr.Dsid
	if sid == 0 {
		sid = hdr.Sid
	}
	if sid != 0 {
		if attr, err := h.sess.GetSidAttr(ctx, sid); err == nil && attr != nil {
			_ = h.sess.RefreshTTL(ctx, sid, attr.Uid)
		}
	}

	outHdr := &header.Header{
		Cmd:    cmd.CMD_PONG,
		Length: 0,
		Sid:    hdr.Sid,
		Cid:    hdr.Cid,
		Nid:    hdr.Nid,
		Seq:    hdr.Seq,
	}
	frame, err := outHdr.Pack()
	if err != nil {
		return
	}
	_ = h.hub.AsyncSend(cmd.CMD_PONG, hdr.Nid, frame)
}
