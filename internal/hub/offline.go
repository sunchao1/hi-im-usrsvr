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

	"github.com/sunchao1/hi-im-api/pkg/im/header"
	"github.com/sunchao1/hi-im-usrsvr/internal/session"
)

// OfflineHandler processes CMD_OFFLINE and cleans Redis session state.
type OfflineHandler struct {
	sess *session.Store
	log  *slog.Logger
}

// NewOfflineHandler creates an OFFLINE handler.
func NewOfflineHandler(sess *session.Store) *OfflineHandler {
	return &OfflineHandler{sess: sess, log: slog.Default()}
}

// Handle implements hubclient.Handler for CMD_OFFLINE.
func (h *OfflineHandler) Handle(_ uint32, _ uint32, payload []byte) {
	ctx := context.Background()
	if len(payload) < header.Size {
		h.log.Warn("offline: payload too short")
		return
	}
	hdr, err := header.Unmarshal(payload[:header.Size])
	if err != nil {
		h.log.Warn("offline: bad header", "err", err)
		return
	}

	sid := hdr.Dsid
	if sid == 0 {
		sid = hdr.Sid
	}
	if sid == 0 {
		return
	}

	attr, err := h.sess.GetSidAttr(ctx, sid)
	if err != nil || attr == nil {
		return
	}
	_ = h.sess.CleanSession(ctx, sid, attr.Uid)
}
