package main

import (
	"fmt"
	cherryError "github.com/cherry-game/cherry/error"
	"github.com/cherry-game/cherry/examples/demo_game_cluster/internal/code"
	"github.com/cherry-game/cherry/examples/demo_game_cluster/internal/pb"
	cherryHttp "github.com/cherry-game/cherry/extend/http"
	cherryTime "github.com/cherry-game/cherry/extend/time"
	cherryLogger "github.com/cherry-game/cherry/logger"
	cherryClient "github.com/cherry-game/cherry/net/parser/pomelo/client"
	jsoniter "github.com/json-iterator/go"
	"math/rand"
	"time"
)

type (
	// Robot client robot
	Robot struct {
		*cherryClient.Client
		PrintLog   bool
		Token      string
		ServerId   int32
		PID        int32
		UID        int64
		OpenId     string
		PlayerId   int64
		PlayerName string
		StartTime  cherryTime.CherryTime
	}
)

func New(client *cherryClient.Client) *Robot {
	return &Robot{
		Client: client,
	}
}

// GetToken  http登录获取token对象
// http://172.16.124.137/login?pid=2126003&account=test1&password=test1
func (p *Robot) GetToken(url string, pid, userName, password string) error {
	// http登陆获取token json对象
	requestURL := fmt.Sprintf("%s/login", url)
	jsonBytes, _, err := cherryHttp.GET(requestURL, map[string]string{
		"pid":      pid,      //sdk包id
		"account":  userName, //帐号名
		"password": password, //密码
	})

	if err != nil {
		return err
	}

	// 转换json对象
	rsp := code.Result{}
	if err = jsoniter.Unmarshal(jsonBytes, &rsp); err != nil {
		return err
	}

	if code.IsFail(rsp.Code) {
		return cherryError.Errorf("get Token fail. [message = %s]", rsp.Message)
	}

	// 获取token值
	p.Token = rsp.Data.(string)
	p.TagName = fmt.Sprintf("%s_%s", pid, userName)
	p.StartTime = cherryTime.Now()

	return nil
}

// UserLogin 用户登录对某游戏服
func (p *Robot) UserLogin(serverId int32) error {
	route := "gate.user.login"

	p.Debugf("[%s] [UserLogin] request ServerID = %d", p.TagName, serverId)

	msg, err := p.Request(route, &pb.LoginRequest{
		ServerId: serverId,
		Token:    p.Token,
		Params:   nil,
	})

	if err != nil {
		return err
	}

	p.ServerId = serverId

	rsp := &pb.LoginResponse{}
	err = p.Serializer().Unmarshal(msg.Data, rsp)
	if err != nil {
		return err
	}

	p.UID = rsp.Uid
	p.PID = rsp.Pid
	p.OpenId = rsp.OpenId

	p.Debugf("[%s] [UserLogin] response = %+v", p.TagName, rsp)
	return nil
}

// PlayerSelect 查看玩家列表
func (p *Robot) PlayerSelect() error {
	route := "game.player.select"

	msg, err := p.Request(route, &pb.None{})
	if err != nil {
		return err
	}

	rsp := &pb.PlayerSelectResponse{}
	err = p.Serializer().Unmarshal(msg.Data, rsp)
	if err != nil {
		return err
	}

	if len(rsp.List) < 1 {
		p.Debugf("[%s] not found player list.", p.TagName)
		return nil
	}

	p.PlayerId = rsp.List[0].PlayerId
	p.PlayerName = rsp.List[0].PlayerName

	p.Debugf("[%s] [PlayerSelect] response PlayerId = %d,PlayerName = %s", p.TagName, p.PlayerId, p.PlayerName)

	return nil
}

// ActorCreate 创建角色
func (p *Robot) ActorCreate() error {
	if p.PlayerId > 0 {
		p.Debugf("[%s] deny create actor", p.TagName)
		return nil
	}

	route := "game.player.create"
	gender := rand.Int31n(1)

	req := &pb.PlayerCreateRequest{
		PlayerName: "p" + p.OpenId,
		Gender:     gender,
	}

	msg, err := p.Request(route, req)
	if err != nil {
		return err
	}

	rsp := &pb.PlayerCreateResponse{}
	err = p.Serializer().Unmarshal(msg.Data, rsp)
	if err != nil {
		return err
	}

	p.PlayerId = rsp.Player.PlayerId
	p.PlayerName = rsp.Player.PlayerName

	p.Debugf("[%s] [ActorCreate] PlayerId = %d,ActorName = %s", p.TagName, p.PlayerId, p.PlayerName)

	return nil
}

// ActorEnter 角色进入游戏
func (p *Robot) ActorEnter() error {
	route := "game.player.enter"
	req := &pb.Int64{
		Value: p.PlayerId,
	}

	msg, err := p.Request(route, req)
	if err != nil {
		return err
	}

	rsp := &pb.PlayerEnterResponse{}
	err = p.Serializer().Unmarshal(msg.Data, rsp)
	if err != nil {
		return err
	}

	p.Debugf("[%s] [ActorEnter] response PlayerId = %d,ActorName = %s", p.TagName, p.PlayerId, p.PlayerName)
	return nil
}

func (p *Robot) RandSleep() {
	time.Sleep(time.Duration(rand.Int31n(300)) * time.Millisecond)
}

func (p *Robot) Debug(args ...interface{}) {
	if p.PrintLog {
		cherryLogger.Debug(args...)
	}

}

func (p *Robot) Debugf(template string, args ...interface{}) {
	if p.PrintLog {
		cherryLogger.Debugf(template, args...)
	}
}
