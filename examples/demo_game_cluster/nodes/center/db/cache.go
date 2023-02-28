package db

import (
	"github.com/goburrow/cache"
	"time"
)

var (

	// uid缓存 key:uidKey, value:uid
	uidCache = cache.New(
		cache.WithMaximumSize(65535),
		cache.WithExpireAfterAccess(120*time.Minute),
	)

	// 开发帐号缓存 key:accountName, value:DevAccountTable
	devAccountCache = cache.New(
		cache.WithMaximumSize(65535),
		cache.WithExpireAfterAccess(60*time.Minute),
	)
)

// cache key
const (
	uidKey = "uid.%d.%s" //pid,openId
)
