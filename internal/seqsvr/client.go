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

package seqsvr

import (
	"context"
	"fmt"
	"time"

	seqv1 "github.com/sunchao1/hi-im-api/gen/go/seq/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Client allocates SID/SEQ via hi-im-seqsvr gRPC.
type Client interface {
	AllocSid(ctx context.Context) (int64, error)
	AllocSeq(ctx context.Context, sid int64) (int64, error)
	Close() error
}

type grpcClient struct {
	conn   *grpc.ClientConn
	client seqv1.SeqServiceClient
}

// Dial connects to seqsvr at addr with a single shared gRPC connection.
func Dial(ctx context.Context, addr string) (Client, error) {
	if addr == "" {
		return nil, fmt.Errorf("empty seqsvr addr")
	}
	dialCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		dialCtx,
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock(),
	)
	if err != nil {
		return nil, fmt.Errorf("dial seqsvr: %w", err)
	}
	return &grpcClient{
		conn:   conn,
		client: seqv1.NewSeqServiceClient(conn),
	}, nil
}

func (c *grpcClient) AllocSid(ctx context.Context) (int64, error) {
	callCtx, cancel := withTimeout(ctx, 3*time.Second)
	defer cancel()
	resp, err := c.client.AllocSid(callCtx, &seqv1.AllocSidRequest{})
	if err != nil {
		return 0, err
	}
	if resp.GetSid() <= 0 {
		return 0, fmt.Errorf("invalid sid %d", resp.GetSid())
	}
	return resp.GetSid(), nil
}

func (c *grpcClient) AllocSeq(ctx context.Context, sid int64) (int64, error) {
	if sid <= 0 {
		return 0, fmt.Errorf("sid must be positive")
	}
	callCtx, cancel := withTimeout(ctx, 3*time.Second)
	defer cancel()
	resp, err := c.client.AllocSeq(callCtx, &seqv1.AllocSeqRequest{Sid: sid})
	if err != nil {
		return 0, err
	}
	if resp.GetSeq() <= 0 {
		return 0, fmt.Errorf("invalid seq %d", resp.GetSeq())
	}
	return resp.GetSeq(), nil
}

func (c *grpcClient) Close() error {
	if c.conn == nil {
		return nil
	}
	return c.conn.Close()
}

func withTimeout(ctx context.Context, d time.Duration) (context.Context, context.CancelFunc) {
	if deadline, ok := ctx.Deadline(); ok && time.Until(deadline) < d {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, d)
}
