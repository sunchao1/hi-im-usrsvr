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
	"context"
	"errors"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sunchao1/hi-im-api/pkg/rediskey"
)

const roleOwner = 1

var (
	ErrNotFound = errors.New("group not found")
)

// Store manages group metadata in Redis (aligned with beehive lib/chat).
type Store struct {
	rdb *redis.Client
	ttl time.Duration
}

// NewStore creates a group Redis store.
func NewStore(rdb *redis.Client, ttlSec int) *Store {
	return &Store{
		rdb: rdb,
		ttl: time.Duration(ttlSec) * time.Second,
	}
}

func (s *Store) expiryScore() float64 {
	return float64(time.Now().Add(s.ttl).Unix())
}

// Exists reports whether gid is registered in chat:gid:zset.
func (s *Store) Exists(ctx context.Context, gid uint64) (bool, error) {
	score, err := s.rdb.ZScore(ctx, rediskey.GidZSetKey(), strconv.FormatUint(gid, 10)).Result()
	if err == redis.Nil {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return score >= float64(time.Now().Unix()), nil
}

// Register creates a new group and auto-joins the owner online.
func (s *Store) Register(ctx context.Context, gid, ownerUID, sid uint64, nid uint32, name, desc string) error {
	score := s.expiryScore()
	pipe := s.rdb.Pipeline()
	pipe.ZAdd(ctx, rediskey.GidZSetKey(), redis.Z{Score: score, Member: strconv.FormatUint(gid, 10)})
	pipe.HSet(ctx, rediskey.GidRoleTabKey(gid), strconv.FormatUint(ownerUID, 10), roleOwner)
	pipe.HSet(ctx, rediskey.GidInfoTabKey(gid), map[string]interface{}{
		"NAME":  name,
		"DESC":  desc,
		"OWNER": ownerUID,
	})
	_, err := pipe.Exec(ctx)
	if err != nil {
		return err
	}
	return s.JoinOnline(ctx, gid, ownerUID, sid, nid)
}

// JoinOnline adds uid to gid member indexes and records gateway nid for fan-out.
func (s *Store) JoinOnline(ctx context.Context, gid, uid, sid uint64, nid uint32) error {
	ok, err := s.Exists(ctx, gid)
	if err != nil {
		return err
	}
	if !ok {
		return ErrNotFound
	}

	score := s.expiryScore()
	uidKey := rediskey.GidToUidZSetKey(gid)
	nidKey := rediskey.GidToNidZSetKey(gid)

	pipe := s.rdb.Pipeline()
	pipe.ZAdd(ctx, uidKey, redis.Z{Score: score, Member: strconv.FormatUint(uid, 10)})
	pipe.ZAdd(ctx, nidKey, redis.Z{Score: score, Member: strconv.FormatUint(uint64(nid), 10)})
	roleKey := rediskey.GidRoleTabKey(gid)
	pipe.HSetNX(ctx, roleKey, strconv.FormatUint(uid, 10), 3)
	_, err = pipe.Exec(ctx)
	if err != nil {
		return fmt.Errorf("join online: %w", err)
	}
	return nil
}
