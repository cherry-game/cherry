package db

import (
	"github.com/goburrow/cache"
	"time"
)

const (
	maxActorNameNum = 32767
)

var (
	// uid缓存 key:uid, value:actorId
	uidCache = cache.New(
		cache.WithMaximumSize(32767),
		cache.WithExpireAfterAccess(60*time.Minute),
	)

	// 角色表缓存 key:actorId, value:*ActorTable
	actorTableCache = cache.New(
		cache.WithMaximumSize(10000),
		cache.WithExpireAfterAccess(60*time.Minute),
	)

	// 角色名缓存 key:actorName, value:actorId
	actorNameCache = cache.New(
		cache.WithMaximumSize(maxActorNameNum),
		cache.WithExpireAfterAccess(60*time.Minute),
	)

	// 英雄表缓存 key:actorId, value:*HeroTable
	// 道具表缓存 key:actorId, value:*ItemTable
)
