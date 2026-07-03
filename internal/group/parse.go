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

package group

import (
	"fmt"

	"google.golang.org/protobuf/encoding/protowire"
)

// CreatRequest holds decoded GROUP-CREAT body fields.
type CreatRequest struct {
	UID  uint64
	GID  uint64
	Name string
	Desc string
}

// ParseCreatRequest decodes GROUP-CREAT protobuf without generated types.
func ParseCreatRequest(body []byte) (CreatRequest, error) {
	var out CreatRequest
	b := body
	for len(b) > 0 {
		num, typ, n := protowire.ConsumeTag(b)
		if n < 0 {
			return out, fmt.Errorf("bad tag")
		}
		b = b[n:]
		switch num {
		case 1:
			if typ != protowire.VarintType {
				return out, fmt.Errorf("uid type")
			}
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return out, fmt.Errorf("uid value")
			}
			out.UID = v
			b = b[n:]
		case 2:
			if typ != protowire.VarintType {
				return out, fmt.Errorf("gid type")
			}
			v, n := protowire.ConsumeVarint(b)
			if n < 0 {
				return out, fmt.Errorf("gid value")
			}
			out.GID = v
			b = b[n:]
		case 3, 4:
			if typ != protowire.BytesType {
				return out, fmt.Errorf("string type")
			}
			v, n := protowire.ConsumeBytes(b)
			if n < 0 {
				return out, fmt.Errorf("string value")
			}
			if num == 3 {
				out.Name = string(v)
			} else {
				out.Desc = string(v)
			}
			b = b[n:]
		default:
			n = protowire.ConsumeFieldValue(num, typ, b)
			if n < 0 {
				return out, fmt.Errorf("skip field")
			}
			b = b[n:]
		}
	}
	return out, nil
}
