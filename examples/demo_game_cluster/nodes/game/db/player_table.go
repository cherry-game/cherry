package db

import (
	"github.com/cherry-game/cherry/examples/demo_game_cluster/internal/code"
	"github.com/cherry-game/cherry/examples/demo_game_cluster/internal/data"
	"github.com/cherry-game/cherry/examples/demo_game_cluster/internal/guid"
	sessionKey "github.com/cherry-game/cherry/examples/demo_game_cluster/internal/session_key"
	cherryTime "github.com/cherry-game/cherry/extend/time"
	clog "github.com/cherry-game/cherry/logger"
	cproto "github.com/cherry-game/cherry/net/proto"
)

// PlayerTable 角色基础表
type PlayerTable struct {
	PID            int32  `gorm:"column:pid;comment:'平台id'" json:"pid"`
	OpenId         string `gorm:"column:open_id;comment:'平台open_id'" json:"openId"`
	UID            int64  `gorm:"column:uid;comment:'用户id'" json:"uid"`
	ServerId       int32  `gorm:"column:server_id;comment:'创角时的游戏服id'" json:"serverId"`
	MergedServerId int32  `gorm:"column:merged_server_id;comment:'合服后的游戏服id'" json:"mergedServerId"`
	PlayerId       int64  `gorm:"column:player_id;primary_key;comment:'角色id'" json:"playerId"`
	Name           string `gorm:"column:player_name;comment:'角色名称'" json:"name"`
	Gender         int32  `gorm:"column:gender;comment:'角色性别'" json:"gender"`
	Level          int32  `gorm:"column:level;comment:'角色等级'" json:"level"`
	Exp            int64  `gorm:"column:exp;comment:'角色经验'" json:"exp"`
	CreateTime     int64  `gorm:"column:create_time;comment:'创建时间'" json:"createTime"`
}

func (*PlayerTable) TableName() string {
	return "player"
}

// InThisServerId 角色当前正在的游戏服(合服后serverId会变)
func (p *PlayerTable) InThisServerId() int32 {
	if p.MergedServerId > 0 {
		return p.MergedServerId
	}

	return p.ServerId
}

func CreatePlayer(session *cproto.Session, name string, serverId int32, playerInit *data.PlayerInitRow) (*PlayerTable, int32) {
	// 检测是否有重名角色
	if _, found := PlayerNameIsExist(name); found {
		return nil, code.PlayerNameExist
	}

	playerId := guid.Next() // new player id
	pid := session.GetInt32(sessionKey.PID)
	openId := session.GetString(sessionKey.OpenID)

	if session.Uid < 1 || pid < 1 || openId == "" {
		clog.Warnf("create playerTable fail. pid or openId is error. [name = %s, pid = %v, openId = %v]",
			name,
			pid,
			openId,
		)
		return nil, code.PlayerCreateFail
	}

	playerTable := &PlayerTable{
		PID:            pid,
		OpenId:         openId,
		UID:            session.Uid,
		ServerId:       serverId,
		MergedServerId: serverId,
		PlayerId:       playerId,
		Name:           name,
		Gender:         playerInit.Gender,
		Level:          playerInit.Level,
		Exp:            0,
		CreateTime:     cherryTime.Now().ToMillisecond(),
	}

	// 先进缓存
	playerTableCache.Put(playerId, playerTable)
	playerNameCache.Put(name, playerTable.PlayerId) // 缓存角色名
	uidCache.Put(playerTable.UID, playerId)

	// TODO 保存db

	// TODO 初始化角色相关的表
	// 道具表
	// 英雄表

	return playerTable, code.OK
}

// PlayerNameIsExist 玩家角色名全局唯一
func PlayerNameIsExist(playerName string) (int64, bool) {
	val, found := playerNameCache.GetIfPresent(playerName)
	if found {
		playerId := val.(int64)
		return playerId, true
	}

	// TODO 从数据库查，数据存在先保存到 playerNameCache

	return 0, false
}

// GetPlayerIds 批量查询玩家id(过滤不存在的)
func GetPlayerIds(playerIds []int64) []int64 {
	var list []int64

	for _, playerId := range playerIds {
		if _, found := GetPlayerTable(playerId); found {
			list = append(list, playerId)
		}
	}

	return list
}

// GetPlayerName 获取玩家角色名
func GetPlayerName(playerId int64) string {
	playerTable, found := GetPlayerTable(playerId)
	if found == false {
		return ""
	}

	return playerTable.Name
}

func GetPlayerTable(playerId int64) (*PlayerTable, bool) {
	val, found := playerTableCache.GetIfPresent(playerId)
	if found {
		return val.(*PlayerTable), true
	}

	// TODO 从数据库查数据，如果存在则缓存到 playerTableCache
	return nil, false
}

func GetPlayerIdWithUID(uid int64) int64 {
	val, found := uidCache.GetIfPresent(uid)
	if found {
		return val.(int64)
	}

	// TODO 从数据库查数据，如果存在则缓存到 uidCache

	return 0
}
