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
	"fmt"
	"log/slog"

	"github.com/sunchao1/hi-im-api/pkg/errno"
	"github.com/sunchao1/hi-im-api/pkg/im/cmd"
	"github.com/sunchao1/hi-im-api/pkg/im/header"
	"github.com/sunchao1/hi-im-usrsvr/internal/group"
	"github.com/sunchao1/hi-im-usrsvr/internal/seqsvr"
	"github.com/sunchao1/hi-im-usrsvr/internal/session"
)

// GroupCreatHandler processes CMD_GROUP_CREAT (beehive UsrSvrGroupCreatHandler).
type GroupCreatHandler struct {
	groups *group.Store
	seq    seqsvr.Client
	sess   *session.Store
	hub    Sender
	log    *slog.Logger
}

// NewGroupCreatHandler creates a GROUP-CREAT handler.
func NewGroupCreatHandler(groups *group.Store, seq seqsvr.Client, sess *session.Store, hub Sender) *GroupCreatHandler {
	return &GroupCreatHandler{
		groups: groups,
		seq:    seq,
		sess:   sess,
		hub:    hub,
		log:    slog.Default(),
	}
}

// Handle implements hubclient.Handler for CMD_GROUP_CREAT.
func (h *GroupCreatHandler) Handle(_ uint32, _ uint32, payload []byte) {
	ctx := context.Background()

	if len(payload) < header.Size {
		h.log.Warn("group-creat: payload too short")
		return
	}
	hdr, err := header.Unmarshal(payload[:header.Size])
	if err != nil {
		h.log.Warn("group-creat: bad header", "err", err)
		return
	}
	if err := hdr.Validate(header.RequireSid); err != nil {
		SendSimpleAck(h.hub, h.log, hdr, cmd.CMD_GROUP_CREAT_ACK, errno.ERR_SVR_HEAD_INVALID, "invalid header")
		return
	}

	req, err := group.ParseCreatRequest(payload[header.Size:])
	if err != nil || req.UID == 0 || req.Name == "" {
		SendSimpleAck(h.hub, h.log, hdr, cmd.CMD_GROUP_CREAT_ACK, errno.ERR_SVR_INVALID_PARAM, "uid/name required")
		return
	}

	attr, err := h.sess.GetSidAttr(ctx, hdr.Sid)
	if err != nil || attr == nil || attr.Uid != req.UID {
		SendSimpleAck(h.hub, h.log, hdr, cmd.CMD_GROUP_CREAT_ACK, errno.ERR_SVR_AUTH_FAIL, "sid not online")
		return
	}

	gid, err := h.seq.AllocGid(ctx)
	if err != nil || gid <= 0 {
		h.log.Warn("group-creat: AllocGid failed", "err", err)
		SendSimpleAck(h.hub, h.log, hdr, cmd.CMD_GROUP_CREAT_ACK, errno.ERR_SYS_REMOTE_SERVICE, "AllocGid failed")
		return
	}

	if err := h.groups.Register(ctx, uint64(gid), req.UID, hdr.Sid, hdr.Nid, req.Name, req.Desc); err != nil {
		h.log.Warn("group-creat: register failed", "err", err)
		SendSimpleAck(h.hub, h.log, hdr, cmd.CMD_GROUP_CREAT_ACK, errno.ERR_SYS_DB, err.Error())
		return
	}

	SendSimpleAck(h.hub, h.log, hdr, cmd.CMD_GROUP_CREAT_ACK, errno.OK, fmt.Sprintf("Ok:%d", gid))
}
