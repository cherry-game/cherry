package gate

import (
	"github.com/cherry-game/cherry/examples/demo_game_cluster/internal/code"
	"github.com/cherry-game/cherry/examples/demo_game_cluster/internal/data"
	"github.com/cherry-game/cherry/examples/demo_game_cluster/internal/pb"
	rpcCenter "github.com/cherry-game/cherry/examples/demo_game_cluster/internal/rpc/center"
	sessionKey "github.com/cherry-game/cherry/examples/demo_game_cluster/internal/session_key"
	"github.com/cherry-game/cherry/examples/demo_game_cluster/internal/token"
	cstring "github.com/cherry-game/cherry/extend/string"
	cfacade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cactor "github.com/cherry-game/cherry/net/actor"
	"github.com/cherry-game/cherry/net/parser/pomelo"
	cproto "github.com/cherry-game/cherry/net/proto"
)

var (
	duplicateLoginCode []byte
)

type (
	// ActorAgent 每个网络连接对应一个ActorAgent
	ActorAgent struct {
		cactor.Base
	}
)

func (p *ActorAgent) OnInit() {
	duplicateLoginCode, _ = p.App().Serializer().Marshal(&cproto.I32{
		Value: code.PlayerDuplicateLogin,
	})

	p.Local().Register("login", p.login)
	p.Remote().Register("setSession", p.setSession)
}

func (p *ActorAgent) setSession(req *pb.StringKeyValue) {
	if req.Key == "" {
		return
	}

	if agent, ok := pomelo.GetAgent(p.ActorID()); ok {
		agent.Session().Set(req.Key, req.Value)
	}
}

// login 用户登录，验证帐号 (*pb.LoginResponse, int32)
func (p *ActorAgent) login(session *cproto.Session, req *pb.LoginRequest) {
	agent, found := pomelo.GetAgent(p.ActorID())
	if !found {
		return
	}

	// 验证token
	userToken, errCode := p.validateToken(req.Token)
	if code.IsFail(errCode) {
		agent.Response(session, errCode)
		return
	}

	// 验证pid是否配置
	sdkRow := data.SdkConfig.Get(userToken.PID)
	if sdkRow == nil {
		agent.ResponseCode(session, code.PIDError, true)
		return
	}

	// 根据token带来的sdk参数，从中心节点获取uid
	uid, errCode := rpcCenter.GetUID(p.App(), sdkRow.SdkId, userToken.PID, userToken.OpenID)
	if uid == 0 || code.IsFail(errCode) {
		agent.ResponseCode(session, code.AccountBindFail, true)
		return
	}

	p.checkGateSession(uid)

	if err := agent.Bind(uid); err != nil {
		clog.Warn(err)
		agent.ResponseCode(session, code.AccountBindFail, true)
		return
	}

	agent.Session().Set(sessionKey.ServerID, cstring.ToString(req.ServerId))
	agent.Session().Set(sessionKey.PID, cstring.ToString(userToken.PID))
	agent.Session().Set(sessionKey.OpenID, userToken.OpenID)

	response := &pb.LoginResponse{
		Uid:    uid,
		Pid:    userToken.PID,
		OpenId: userToken.OpenID,
	}

	agent.Response(session, response)
}

func (p *ActorAgent) validateToken(base64Token string) (*token.Token, int32) {
	userToken, ok := token.DecodeToken(base64Token)
	if ok == false {
		return nil, code.AccountTokenValidateFail
	}

	platformRow := data.SdkConfig.Get(userToken.PID)
	if platformRow == nil {
		return nil, code.PIDError
	}

	statusCode, ok := token.Validate(userToken, platformRow.Salt)
	if ok == false {
		return nil, statusCode
	}

	return userToken, code.OK
}

func (p *ActorAgent) checkGateSession(uid cfacade.UID) {
	if agent, found := pomelo.GetAgentWithUID(uid); found {
		agent.Kick(duplicateLoginCode, true)
	}

	rsp := &cproto.PomeloKick{
		Uid:    uid,
		Reason: duplicateLoginCode,
	}

	// 遍历所有网关节点，踢除旧的session
	members := p.App().Discovery().ListByType(p.App().NodeType(), p.App().NodeId())
	for _, member := range members {
		// user是gate.go里自定义的agentActorID
		actorPath := cfacade.NewPath(member.GetNodeId(), "user")
		p.Call(actorPath, pomelo.KickFuncName, rsp)
	}
}

// onSessionClose  当agent断开时，关闭对应的ActorAgent
func (p *ActorAgent) onSessionClose(agent *pomelo.Agent) {
	session := agent.Session()
	serverId := session.GetString(sessionKey.ServerID)
	if serverId == "" {
		return
	}

	// 通知game节点关闭session
	childId := cstring.ToString(session.Uid)
	if childId != "" {
		targetPath := cfacade.NewChildPath(serverId, "player", childId)
		p.Call(targetPath, "sessionClose", nil)
	}

	// 自己退出
	p.Exit()
	clog.Infof("sessionClose path = %s", p.Path())
}
