package gate

import (
	"context"
	"github.com/cherry-game/cherry"
	"github.com/cherry-game/cherry/examples/game_cluster/internal/code"
	"github.com/cherry-game/cherry/examples/game_cluster/internal/data"
	"github.com/cherry-game/cherry/examples/game_cluster/internal/pb"
	rpcCenter "github.com/cherry-game/cherry/examples/game_cluster/internal/rpc/center"
	rpcGame "github.com/cherry-game/cherry/examples/game_cluster/internal/rpc/game"
	sessionKey "github.com/cherry-game/cherry/examples/game_cluster/internal/session_key"
	"github.com/cherry-game/cherry/examples/game_cluster/internal/token"
	cstring "github.com/cherry-game/cherry/extend/string"
	facade "github.com/cherry-game/cherry/facade"
	clog "github.com/cherry-game/cherry/logger"
	cdiscovery "github.com/cherry-game/cherry/net/cluster/discovery"
	ch "github.com/cherry-game/cherry/net/handler"
	cm "github.com/cherry-game/cherry/net/message"
	cproto "github.com/cherry-game/cherry/net/proto"
	cs "github.com/cherry-game/cherry/net/session"
)

type (
	UserHandler struct {
		ch.Handler
	}
)

const (
	// 客户端连接后，必需先执行第一条协议，进行token验证后，才能进行后续的逻辑
	firstRouteName = "gate.userHandler.login"
)

var (
	notLoginRsp = &pb.Int32{
		Value: code.ActorDenyLogin,
	}

	duplicateLoginRSp = &cproto.Response{
		Code: code.ActorDuplicateLogin,
	}
)

func (p *UserHandler) Name() string {
	return "userHandler"
}

func (p *UserHandler) OnInit() {
	p.AddLocal("login", p.login)

	// beforeFilter 过滤器，确认session为登录绑定状态，才可以进行其他的消息请求
	p.AddBeforeFilter(func(ctx context.Context, session *cs.Session, message *cm.Message) bool {
		if !session.IsBind() && message.Route != firstRouteName {
			session.Kick(notLoginRsp, true)
			return false
		}
		return true
	})

	cs.AddOnCloseListener(func(session *cs.Session) (next bool) {
		// 当客户端断开连接通，通知游戏节点关闭session
		rpcGame.SessionClose(session)
		return true
	})
}

func (p *UserHandler) login(session *cs.Session, req *pb.LoginRequest) (*pb.LoginResponse, int32) {
	// 验证token
	userToken, errCode := p.validateToken(req.Token)
	if code.IsFail(errCode) {
		return nil, errCode
	}

	// 验证pid是否配置
	sdkRow := data.SdkConfig.Get(userToken.PID)
	if sdkRow == nil {
		return nil, code.PIDError
	}

	// 根据token带来的sdk参数，从中心节点获取uid
	uid, errCode := rpcCenter.GetUID(sdkRow.SdkId, userToken.PID, userToken.OpenID)
	if uid == 0 || code.IsFail(errCode) {
		return nil, code.AccountBindFail
	}

	p.checkGateSession(session, uid)

	session.Set(sessionKey.ServerID, cstring.ToString(req.ServerId))
	session.Set(sessionKey.PID, cstring.ToString(userToken.PID))
	session.Set(sessionKey.OpenID, userToken.OpenID)

	response := &pb.LoginResponse{
		Uid:    uid,
		Pid:    userToken.PID,
		OpenId: userToken.OpenID,
	}

	// uid和session绑定
	err := cs.Bind(session.SID(), uid)
	if err != nil {
		clog.Warn(err)
		return nil, code.AccountBindFail
	}

	return response, code.OK
}

func (p *UserHandler) validateToken(base64Token string) (*token.Token, int32) {
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

func (p *UserHandler) checkGateSession(session *cs.Session, uid facade.UID) {
	oldSession, found := cs.GetByUID(uid)
	//如果当前uid已经登录，并且与当前session的sid不一样，则踢除旧的session
	if found && oldSession.SID() != session.SID() {
		oldSession.Kick(duplicateLoginRSp, true)
	}

	// 当前uid已登陆该网关，通知其他网关需做踢出操作
	members := cdiscovery.ListByType(p.NodeType(), p.NodeId())
	if len(members) > 0 {
		for _, member := range members {
			cherry.Kick(member.GetNodeId(), uid, duplicateLoginRSp, true)
		}
	}
}
