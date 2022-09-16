package db

import (
	"github.com/cherry-game/cherry/examples/game_cluster/internal/code"
	"github.com/cherry-game/cherry/examples/game_cluster/internal/data"
	"github.com/cherry-game/cherry/examples/game_cluster/internal/guid"
	sessionKey "github.com/cherry-game/cherry/examples/game_cluster/internal/session_key"
	cherryTime "github.com/cherry-game/cherry/extend/time"
	cs "github.com/cherry-game/cherry/net/session"
)

// ActorTable 角色基础表
type ActorTable struct {
	PID            int32  `gorm:"column:pid;comment:'平台id'" json:"pid"`
	OpenId         string `gorm:"column:open_id;comment:'平台open_id'" json:"openId"`
	UID            int64  `gorm:"column:uid;comment:'用户id'" json:"uid"`
	ServerId       int32  `gorm:"column:server_id;comment:'创角时的游戏服id'" json:"serverId"`
	MergedServerId int32  `gorm:"column:merged_server_id;comment:'合服后的游戏服id'" json:"mergedServerId"`
	ActorId        int64  `gorm:"column:actor_id;primary_key;comment:'角色id'" json:"actorId"`
	Name           string `gorm:"column:actor_name;comment:'角色名称'" json:"name"`
	Gender         int32  `gorm:"column:gender;comment:'角色性别'" json:"gender"`
	Level          int32  `gorm:"column:level;comment:'角色等级'" json:"level"`
	Exp            int64  `gorm:"column:exp;comment:'角色经验'" json:"exp"`
	CreateTime     int64  `gorm:"column:create_time;comment:'创建时间'" json:"createTime"`
}

func (*ActorTable) TableName() string {
	return "actor"
}

// InThisServerId 角色当前正在的游戏服(合服后serverId会变)
func (p *ActorTable) InThisServerId() int32 {
	if p.MergedServerId > 0 {
		return p.MergedServerId
	}

	return p.ServerId
}

func CreateActor(session *cs.Session, name string, serverId int32, actorInit *data.ActorInitRow) (*ActorTable, int32) {
	// 检测是否有重名角色
	if _, found := ActorNameIsExist(name); found {
		return nil, code.ActorNameExist
	}

	actorId := guid.Next() // new actor id
	pid := sessionKey.GetPID(session)
	openId := sessionKey.GetOpenId(session)
	uid := session.UID()

	if uid < 1 || pid < 1 || openId == "" {
		session.Warnf("create actor fail. pid or openId is error. [name = %s, sessionData = %v]", name, session.Data())
		return nil, code.ActorCreateFail
	}

	actor := &ActorTable{
		PID:            pid,
		OpenId:         openId,
		UID:            uid,
		ServerId:       serverId,
		MergedServerId: serverId,
		ActorId:        actorId,
		Name:           name,
		Gender:         actorInit.Gender,
		Level:          actorInit.Level,
		Exp:            0,
		CreateTime:     cherryTime.Now().ToMillisecond(),
	}

	// 先进缓存
	actorTableCache.Put(actorId, actor)
	actorNameCache.Put(name, actor.ActorId) // 缓存角色名
	uidCache.Put(uid, actorId)

	// TODO 保存db

	// TODO 初始化角色相关的表
	// 道具表
	// 英雄表

	return actor, code.OK
}

// ActorNameIsExist 角色名全局唯一
func ActorNameIsExist(actorName string) (int64, bool) {
	val, found := actorNameCache.GetIfPresent(actorName)
	if found {
		actorId := val.(int64)
		return actorId, true
	}

	// TODO 从数据库查，数据存在先保存到 actorNameCache

	return 0, false
}

// GetActorIds 批量查询角色id(过滤不存在的)
func GetActorIds(actorIds []int64) []int64 {
	var list []int64

	for _, actorId := range actorIds {
		if _, found := GetActorTable(actorId); found {
			list = append(list, actorId)
		}
	}

	return list
}

// GetActorName 获取角色名
func GetActorName(actorId int64) string {
	actorTable, found := GetActorTable(actorId)
	if found == false {
		return ""
	}

	return actorTable.Name
}

func GetActorTable(actorId int64) (*ActorTable, bool) {
	val, found := actorTableCache.GetIfPresent(actorId)
	if found {
		return val.(*ActorTable), true
	}

	// TODO 从数据库查数据，如果存在则缓存到 actorTableCache
	return nil, false
}

func GetActorIdWithUID(serverId int32, uid int64) int64 {
	val, found := uidCache.GetIfPresent(uid)
	if found {
		return val.(int64)
	}

	// TODO 从数据库查数据，如果存在则缓存到 uidCache

	return 0
}
