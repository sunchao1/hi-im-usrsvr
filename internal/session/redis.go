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

package session

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sunchao1/hi-im-api/pkg/rediskey"
)

const (
	fieldCID = "cid"
	fieldUID = "uid"
	fieldNID = "nid"

	// chatSidZSetKey matches beehive chat:sid:zset (see hi-im-api doc §8).
	chatSidZSetKey = "chat:sid:zset"
	// uidToSidSetKey matches beehive im:uid:{uid}:to:sid:set.
	uidToSidSetKey = "im:uid:%d:to:sid:set"
)

// SidAttr holds decoded session attributes from Redis.
type SidAttr struct {
	Cid uint64
	Uid uint64
	Nid uint32
}

// Store manages IM session state in Redis.
type Store struct {
	rdb *redis.Client
	ttl time.Duration
}

// NewStore creates a session store backed by Redis.
func NewStore(rdb *redis.Client, ttlSec int) *Store {
	return &Store{
		rdb: rdb,
		ttl: time.Duration(ttlSec) * time.Second,
	}
}

// GetSidAttr reads im:sid:{sid}:attr hash fields.
func (s *Store) GetSidAttr(ctx context.Context, sid uint64) (*SidAttr, error) {
	key := rediskey.SidAttrKey(sid)
	vals, err := s.rdb.HGetAll(ctx, key).Result()
	if err != nil {
		return nil, err
	}
	if len(vals) == 0 {
		return nil, redis.Nil
	}
	attr, err := parseSidAttr(vals)
	if err != nil {
		return nil, err
	}
	return attr, nil
}

// SetSidOnline writes session online state via Redis pipeline.
func (s *Store) SetSidOnline(ctx context.Context, sid, uid, cid uint64, nid uint32) error {
	attrKey := rediskey.SidAttrKey(sid)
	uidSetKey := fmt.Sprintf(uidToSidSetKey, uid)
	score := float64(time.Now().Unix())

	pipe := s.rdb.Pipeline()
	pipe.HSet(ctx, attrKey, map[string]interface{}{
		fieldCID: cid,
		fieldUID: uid,
		fieldNID: nid,
	})
	pipe.Expire(ctx, attrKey, s.ttl)
	pipe.ZAdd(ctx, chatSidZSetKey, redis.Z{Score: score, Member: sid})
	pipe.Expire(ctx, chatSidZSetKey, s.ttl)
	pipe.SAdd(ctx, uidSetKey, sid)
	pipe.Expire(ctx, uidSetKey, s.ttl)
	_, err := pipe.Exec(ctx)
	return err
}

// CleanSession removes session keys for the given sid and uid.
func (s *Store) CleanSession(ctx context.Context, sid, uid uint64) error {
	attrKey := rediskey.SidAttrKey(sid)
	uidSetKey := fmt.Sprintf(uidToSidSetKey, uid)

	pipe := s.rdb.Pipeline()
	pipe.Del(ctx, attrKey)
	pipe.ZRem(ctx, chatSidZSetKey, sid)
	pipe.SRem(ctx, uidSetKey, sid)
	_, err := pipe.Exec(ctx)
	return err
}

// RefreshTTL extends session key TTL on heartbeat.
func (s *Store) RefreshTTL(ctx context.Context, sid, uid uint64) error {
	attrKey := rediskey.SidAttrKey(sid)
	uidSetKey := fmt.Sprintf(uidToSidSetKey, uid)

	pipe := s.rdb.Pipeline()
	pipe.Expire(ctx, attrKey, s.ttl)
	pipe.Expire(ctx, chatSidZSetKey, s.ttl)
	pipe.Expire(ctx, uidSetKey, s.ttl)
	_, err := pipe.Exec(ctx)
	return err
}

func parseSidAttr(vals map[string]string) (*SidAttr, error) {
	cid, err := strconv.ParseUint(vals[fieldCID], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse cid: %w", err)
	}
	uid, err := strconv.ParseUint(vals[fieldUID], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("parse uid: %w", err)
	}
	nid64, err := strconv.ParseUint(vals[fieldNID], 10, 32)
	if err != nil {
		return nil, fmt.Errorf("parse nid: %w", err)
	}
	return &SidAttr{Cid: cid, Uid: uid, Nid: uint32(nid64)}, nil
}
