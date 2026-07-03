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

package imhttp

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/sunchao1/hi-im-api/pkg/errno"
	"github.com/sunchao1/hi-im-usrsvr/internal/seqsvr"
)

// RegisterResponse is the JSON body for GET /im/register.
type RegisterResponse struct {
	UID    uint64 `json:"uid"`
	SID    int64  `json:"sid"`
	Nation string `json:"nation"`
	City   string `json:"city"`
	Town   string `json:"town"`
	Code   int    `json:"code"`
	ErrMsg string `json:"errmsg"`
}

func registerHandler(seq seqsvr.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		uidStr := c.Query("uid")
		nation := c.Query("nation")
		city := c.Query("city")
		town := c.Query("town")

		if uidStr == "" || nation == "" {
			c.JSON(http.StatusBadRequest, RegisterResponse{
				Code:   errno.ERR_SYS_MISS_PARAM,
				ErrMsg: "uid and nation are required",
			})
			return
		}
		uid, err := strconv.ParseUint(uidStr, 10, 64)
		if err != nil || uid == 0 {
			c.JSON(http.StatusBadRequest, RegisterResponse{
				Code:   errno.ERR_SYS_INVALID_PARAM,
				ErrMsg: "invalid uid",
			})
			return
		}

		sid, err := seq.AllocSid(c.Request.Context())
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, RegisterResponse{
				UID:    uid,
				Nation: nation,
				City:   city,
				Town:   town,
				Code:   errno.ERR_SYS_REMOTE_SERVICE,
				ErrMsg: "AllocSid failed",
			})
			return
		}

		c.JSON(http.StatusOK, RegisterResponse{
			UID:    uid,
			SID:    sid,
			Nation: nation,
			City:   city,
			Town:   town,
			Code:   errno.OK,
			ErrMsg: "OK",
		})
	}
}
