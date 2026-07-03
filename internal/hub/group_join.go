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
	"errors"
	"fmt"
	"log/slog"

	imv1 "github.com/sunchao1/hi-im-api/gen/go/im/v1"
	"github.com/sunchao1/hi-im-api/pkg/errno"
	"github.com/sunchao1/hi-im-api/pkg/im/cmd"
	"github.com/sunchao1/hi-im-api/pkg/im/header"
	"github.com/sunchao1/hi-im-usrsvr/internal/group"
	"github.com/sunchao1/hi-im-usrsvr/internal/session"
	"google.golang.org/protobuf/proto"
)

// GroupJoinHandler processes CMD_GROUP_JOIN (beehive UsrSvrGroupJoinHandler).
type GroupJoinHandler struct {
	groups *group.Store
	sess   *session.Store
	hub    Sender
	log    *slog.Logger
}

// NewGroupJoinHandler creates a GROUP-JOIN handler.
func NewGroupJoinHandler(groups *group.Store, sess *session.Store, hub Sender) *GroupJoinHandler {
	return &GroupJoinHandler{
		groups: groups,
		sess:   sess,
		hub:    hub,
		log:    slog.Default(),
	}
}

// Handle implements hubclient.Handler for CMD_GROUP_JOIN.
func (h *GroupJoinHandler) Handle(_ uint32, _ uint32, payload []byte) {
	ctx := context.Background()

	if len(payload) < header.Size {
		h.log.Warn("group-join: payload too short")
		return
	}
	hdr, err := header.Unmarshal(payload[:header.Size])
	if err != nil {
		h.log.Warn("group-join: bad header", "err", err)
		return
	}
	if err := hdr.Validate(header.RequireSid); err != nil {
		SendSimpleAck(h.hub, h.log, hdr, cmd.CMD_GROUP_JOIN_ACK, errno.ERR_SVR_HEAD_INVALID, "invalid header")
		return
	}

	req := &imv1.GroupJoin{}
	if err := proto.Unmarshal(payload[header.Size:], req); err != nil {
		SendSimpleAck(h.hub, h.log, hdr, cmd.CMD_GROUP_JOIN_ACK, errno.ERR_SVR_BODY_INVALID, "invalid body")
		return
	}
	if req.GetGid() == 0 || req.GetUid() == 0 {
		SendSimpleAck(h.hub, h.log, hdr, cmd.CMD_GROUP_JOIN_ACK, errno.ERR_SVR_MISS_PARAM, "gid/uid required")
		return
	}

	attr, err := h.sess.GetSidAttr(ctx, hdr.Sid)
	if err != nil || attr == nil || attr.Uid != req.GetUid() {
		SendSimpleAck(h.hub, h.log, hdr, cmd.CMD_GROUP_JOIN_ACK, errno.ERR_SVR_AUTH_FAIL, "sid not online")
		return
	}

	if err := h.groups.JoinOnline(ctx, req.GetGid(), req.GetUid(), hdr.Sid, hdr.Nid); err != nil {
		code := errno.ERR_SVR_CHECK_FAIL
		if errors.Is(err, group.ErrNotFound) {
			code = errno.ERR_SVR_INVALID_PARAM
		}
		SendSimpleAck(h.hub, h.log, hdr, cmd.CMD_GROUP_JOIN_ACK, code, err.Error())
		return
	}

	SendSimpleAck(h.hub, h.log, hdr, cmd.CMD_GROUP_JOIN_ACK, errno.OK, fmt.Sprintf("Ok:%d", req.GetGid()))
	h.log.Info("group-join: ok", "gid", req.GetGid(), "uid", req.GetUid(), "sid", hdr.Sid)
}
