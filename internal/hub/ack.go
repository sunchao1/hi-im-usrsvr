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
	"log/slog"

	imv1 "github.com/sunchao1/hi-im-api/gen/go/im/v1"
	"github.com/sunchao1/hi-im-api/pkg/im/header"
	"google.golang.org/protobuf/proto"
)

// SendSimpleAck sends a code/errmsg ACK (GroupJoinAck wire shape) to gateway.
func SendSimpleAck(hub Sender, log *slog.Logger, hdr *header.Header, ackCmd uint32, code int, errmsg string) {
	ack := &imv1.GroupJoinAck{
		Code:   uint32(code),
		Errmsg: errmsg,
	}
	body, err := proto.Marshal(ack)
	if err != nil {
		log.Warn("ack: marshal failed", "err", err)
		return
	}
	outHdr := &header.Header{
		Cmd:    ackCmd,
		Length: uint32(len(body)),
		Sid:    hdr.Sid,
		Cid:    hdr.Cid,
		Nid:    hdr.Nid,
		Seq:    hdr.Seq,
	}
	frame, err := outHdr.Pack()
	if err != nil {
		log.Warn("ack: pack failed", "err", err)
		return
	}
	frame = append(frame, body...)
	if err := hub.AsyncSend(ackCmd, hdr.Nid, frame); err != nil {
		log.Warn("ack: send failed", "err", err, "gatewayNid", hdr.Nid)
	}
}
