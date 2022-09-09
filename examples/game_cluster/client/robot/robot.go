package main

import (
	"fmt"
	cherryError "github.com/cherry-game/cherry/error"
	"github.com/cherry-game/cherry/examples/game_cluster/internal/code"
	"github.com/cherry-game/cherry/examples/game_cluster/internal/pb"
	cherryHttp "github.com/cherry-game/cherry/extend/http"
	cherryTime "github.com/cherry-game/cherry/extend/time"
	cherryLogger "github.com/cherry-game/cherry/logger"
	cherryClient "github.com/cherry-game/cherry/net/client"
	jsoniter "github.com/json-iterator/go"
	"math/rand"
	"time"
)

type (
	// Robot client robot
	Robot struct {
		*cherryClient.Client
		PrintLog  bool
		Token     string
		ServerId  int32
		PID       int32
		UID       int64
		OpenId    string
		ActorId   int64
		ActorName string
		StartTime cherryTime.CherryTime
	}
)

func New(client *cherryClient.Client) *Robot {
	return &Robot{
		Client: client,
	}
}

// GetToken http://172.16.124.137/login?pid=2126003&account=test1&password=test1
func (p *Robot) GetToken(url string, pid, userName, password string) error {
	requestURL := fmt.Sprintf("%s/login", url)
	jsonBytes, _, err := cherryHttp.GET(requestURL, map[string]string{
		"pid":      pid,
		"account":  userName,
		"password": password,
	})

	if err != nil {
		return err
	}

	rsp := code.Result{}
	if err = jsoniter.Unmarshal(jsonBytes, &rsp); err != nil {
		return err
	}

	if code.IsFail(rsp.Code) {
		return cherryError.Errorf("get Token fail. [message = %s]", rsp.Message)
	}

	p.Token = rsp.Data.(string)
	p.TagName = fmt.Sprintf("%s_%s", pid, userName)
	p.StartTime = cherryTime.Now()

	return nil
}

func (p *Robot) UserLogin(serverId int32) error {
	route := "gate.userHandler.login"

	p.Debugf("[%s] [UserLogin] request ServerId = %d", p.TagName, serverId)

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

func (p *Robot) ActorSelect() error {
	route := "game.actorHandler.select"

	msg, err := p.Request(route, &pb.None{})
	if err != nil {
		return err
	}

	rsp := &pb.ActorSelectResponse{}
	err = p.Serializer().Unmarshal(msg.Data, rsp)
	if err != nil {
		return err
	}

	if len(rsp.ActorList) < 1 {
		p.Debugf("[%s] not found actor list.", p.TagName)
		return nil
	}

	p.ActorId = rsp.ActorList[0].ActorId
	p.ActorName = rsp.ActorList[0].ActorName

	p.Debugf("[%s] [ActorSelect] response ActorId = %d,ActorName = %s", p.TagName, p.ActorId, p.ActorName)

	return nil
}

func (p *Robot) ActorCreate() error {
	if p.ActorId > 0 {
		p.Debugf("[%s] deny create actor", p.TagName)
		return nil
	}

	route := "game.actorHandler.create"
	gender := rand.Int31n(1)

	req := &pb.ActorCreateRequest{
		ActorName: "actor" + p.OpenId,
		Gender:    gender,
	}

	msg, err := p.Request(route, req)
	if err != nil {
		return err
	}

	rsp := &pb.ActorCreateResponse{}
	err = p.Serializer().Unmarshal(msg.Data, rsp)
	if err != nil {
		return err
	}

	p.ActorId = rsp.ActorInfo.ActorId
	p.ActorName = rsp.ActorInfo.ActorName

	p.Debugf("[%s] [ActorCreate] ActorId = %d,ActorName = %s", p.TagName, p.ActorId, p.ActorName)

	return nil
}

func (p *Robot) ActorEnter() error {
	route := "game.actorHandler.enter"
	req := &pb.Int64{
		Value: p.ActorId,
	}

	msg, err := p.Request(route, req)
	if err != nil {
		return err
	}

	rsp := &pb.ActorEnterResponse{}
	err = p.Serializer().Unmarshal(msg.Data, rsp)
	if err != nil {
		return err
	}

	p.Debugf("[%s] [ActorEnter] response ActorId = %d,ActorName = %s", p.TagName, p.ActorId, p.ActorName)
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
