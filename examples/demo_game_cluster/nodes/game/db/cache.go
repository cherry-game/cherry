package db

import (
	"github.com/goburrow/cache"
	"time"
)

var (
	// uid缓存 key:uid, value:playerId
	uidCache = cache.New(
		cache.WithMaximumSize(-1),
		cache.WithExpireAfterAccess(60*time.Minute),
	)

	// 玩家表缓存 key:playerId, value:*PlayerTable
	playerTableCache = cache.New(
		cache.WithMaximumSize(-1),
		cache.WithExpireAfterAccess(60*time.Minute),
	)

	// 玩家昵称缓存 key:playerName, value:playerId
	playerNameCache = cache.New(
		cache.WithMaximumSize(-1),
		cache.WithExpireAfterAccess(60*time.Minute),
	)

	// 英雄表缓存 key:playerId, value:*HeroTable
	// 道具表缓存 key:playerId, value:*ItemTable
)
