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
	"strconv"
	"strings"

	imv1 "github.com/sunchao1/hi-im-api/gen/go/im/v1"
	"github.com/sunchao1/hi-im-api/pkg/errno"
	"github.com/sunchao1/hi-im-api/pkg/im/cmd"
	"github.com/sunchao1/hi-im-api/pkg/im/header"
	"github.com/sunchao1/hi-im-usrsvr/internal/seqsvr"
	"github.com/sunchao1/hi-im-usrsvr/internal/session"
	"google.golang.org/protobuf/proto"
)

// Sender sends IM frames back to gateway via Hub BACKEND.
type Sender interface {
	AsyncSend(cmd, destNid uint32, payload []byte) error
}

// OnlineHandler processes CMD_ONLINE frames from gateway via Hub BACKEND.
type OnlineHandler struct {
	sess        *session.Store
	seq         seqsvr.Client
	hub         Sender
	log         *slog.Logger
	m4SkipToken bool
}

// NewOnlineHandler creates an ONLINE handler.
func NewOnlineHandler(sess *session.Store, seq seqsvr.Client, hub Sender, m4SkipToken bool) *OnlineHandler {
	return &OnlineHandler{
		sess:        sess,
		seq:         seq,
		hub:         hub,
		log:         slog.Default(),
		m4SkipToken: m4SkipToken,
	}
}

// Handle implements hubclient.Handler for CMD_ONLINE.
func (h *OnlineHandler) Handle(_ uint32, _ uint32, payload []byte) {
	ctx := context.Background()

	if len(payload) < header.Size {
		h.log.Warn("online: payload too short", "len", len(payload))
		return
	}
	hdr, err := header.Unmarshal(payload[:header.Size])
	if err != nil {
		h.log.Warn("online: bad header", "err", err)
		return
	}
	if err := hdr.Validate(header.RequireSid); err != nil {
		h.log.Warn("online: invalid header", "err", err)
		return
	}

	var req imv1.Online
	if err := proto.Unmarshal(payload[header.Size:], &req); err != nil {
		h.log.Warn("online: bad body", "err", err)
		h.sendAck(ctx, hdr, &req, 0, errno.ERR_SVR_BODY_INVALID, "invalid body")
		return
	}
	if req.GetUid() == 0 || req.GetSid() == 0 {
		h.sendAck(ctx, hdr, &req, 0, errno.ERR_SVR_MISS_PARAM, "uid/sid required")
		return
	}
	if !h.validateTokenM4(&req) {
		h.sendAck(ctx, hdr, &req, 0, errno.ERR_SVR_AUTH_FAIL, "token invalid")
		return
	}

	seq, err := h.seq.AllocSeq(ctx, int64(req.GetSid()))
	if err != nil {
		h.log.Warn("online: AllocSeq failed", "sid", req.GetSid(), "err", err)
		h.sendAck(ctx, hdr, &req, 0, errno.ERR_SYS_REMOTE_SERVICE, "AllocSeq failed")
		return
	}

	cid := hdr.Cid
	gatewayNid := hdr.Nid

	if old, err := h.sess.GetSidAttr(ctx, req.GetSid()); err == nil && old != nil {
		if old.Nid != gatewayNid || old.Cid != cid {
			_ = h.sess.CleanSession(ctx, req.GetSid(), old.Uid)
		}
	}

	if err := h.sess.SetSidOnline(ctx, req.GetSid(), req.GetUid(), cid, gatewayNid); err != nil {
		h.log.Warn("online: redis failed", "err", err)
		h.sendAck(ctx, hdr, &req, 0, errno.ERR_SYS_DB, "redis failed")
		return
	}

	h.log.Info("online: ok", "sid", req.GetSid(), "uid", req.GetUid(), "cid", cid, "gatewayNid", gatewayNid, "seq", seq)
	h.sendAck(ctx, hdr, &req, uint64(seq), errno.OK, "OK")
}

func (h *OnlineHandler) validateTokenM4(req *imv1.Online) bool {
	if h.m4SkipToken {
		return true
	}
	token := req.GetToken()
	if token == "" {
		return false
	}
	// M4 plaintext token: uid:{uid}:ttl:{ttl}:sid:{sid}:end
	parts := strings.Split(token, ":")
	if len(parts) != 8 || parts[0] != "uid" || parts[2] != "ttl" || parts[4] != "sid" || parts[6] != "end" {
		return false
	}
	uid, err1 := strconv.ParseUint(parts[1], 10, 64)
	sid, err2 := strconv.ParseUint(parts[5], 10, 64)
	if err1 != nil || err2 != nil {
		return false
	}
	return uid == req.GetUid() && sid == req.GetSid()
}

func (h *OnlineHandler) sendAck(_ context.Context, hdr *header.Header, req *imv1.Online, seq uint64, code int, errmsg string) {
	ack := &imv1.OnlineAck{
		Uid:      req.GetUid(),
		Sid:      req.GetSid(),
		Seq:      seq,
		App:      req.GetApp(),
		Version:  req.GetVersion(),
		Terminal: req.GetTerminal(),
		Code:     uint32(code),
		Errmsg:   errmsg,
	}
	body, err := proto.Marshal(ack)
	if err != nil {
		h.log.Warn("online: marshal ack failed", "err", err)
		return
	}

	outHdr := &header.Header{
		Cmd:    cmd.CMD_ONLINE_ACK,
		Length: uint32(len(body)),
		Sid:    req.GetSid(),
		Cid:    hdr.Cid,
		Nid:    hdr.Nid,
		Seq:    seq,
	}
	frame, err := outHdr.Pack()
	if err != nil {
		h.log.Warn("online: pack header failed", "err", err)
		return
	}
	frame = append(frame, body...)

	if err := h.hub.AsyncSend(cmd.CMD_ONLINE_ACK, hdr.Nid, frame); err != nil {
		h.log.Warn("online: AsyncSend failed", "err", err, "gatewayNid", hdr.Nid)
	}
}

// PackOnlineFrame builds an IM ONLINE frame for integration tests.
func PackOnlineFrame(gatewayCid, sessSid, uid uint64, gatewayNid uint32, online *imv1.Online) ([]byte, error) {
	if online == nil {
		online = &imv1.Online{Uid: uid, Sid: sessSid}
	}
	body, err := proto.Marshal(online)
	if err != nil {
		return nil, fmt.Errorf("marshal online: %w", err)
	}
	hdr := &header.Header{
		Cmd:    cmd.CMD_ONLINE,
		Length: uint32(len(body)),
		Sid:    gatewayCid,
		Nid:    gatewayNid,
	}
	buf, err := hdr.Pack()
	if err != nil {
		return nil, err
	}
	return append(buf, body...), nil
}
