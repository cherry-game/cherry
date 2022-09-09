package token

import (
	"encoding/json"
	"fmt"
	"github.com/cherry-game/cherry/examples/game_cluster/internal/code"
	cherryCrypto "github.com/cherry-game/cherry/extend/crypto"
	cherryTime "github.com/cherry-game/cherry/extend/time"
	cherryLogger "github.com/cherry-game/cherry/logger"
)

const (
	hashFormat      = "pid:%d,openid:%s,timestamp:%d"
	tokenExpiredDay = 3
)

type Token struct {
	PID       int32  `json:"pid"`
	OpenID    string `json:"open_id"`
	Timestamp int64  `json:"tt"`
	Hash      string `json:"hash"`
}

func New(pid int32, openId string, appKey string) *Token {
	token := &Token{
		PID:       pid,
		OpenID:    openId,
		Timestamp: cherryTime.Now().ToMillisecond(),
	}

	token.Hash = BuildHash(token, appKey)
	return token
}

func (t *Token) ToBase64() string {
	bytes, _ := json.Marshal(t)
	return cherryCrypto.Base64Encode(string(bytes))
}

func DecodeToken(base64Token string) (*Token, bool) {
	if len(base64Token) < 1 {
		return nil, false
	}

	token := &Token{}
	bytes, err := cherryCrypto.Base64DecodeBytes(base64Token)
	if err != nil {
		cherryLogger.Warnf("base64Token = %s, validate error = %v", base64Token, err)
		return nil, false
	}

	err = json.Unmarshal(bytes, token)
	if err != nil {
		cherryLogger.Warnf("base64Token = %s, unmarshal error = %v", base64Token, err)
		return nil, false
	}

	return token, true
}

func Validate(token *Token, appKey string) (int32, bool) {
	now := cherryTime.Now()
	now.AddDays(tokenExpiredDay)

	if token.Timestamp > now.ToMillisecond() {
		cherryLogger.Warnf("token is expired, token = %s", token)
		return code.TokenValidateFail, false
	}

	newHash := BuildHash(token, appKey)
	if newHash != token.Hash {
		cherryLogger.Warnf("hash validate fail. newHash = %s, token = %s", token)
		return code.TokenValidateFail, false
	}

	return code.OK, true
}

func BuildHash(t *Token, appKey string) string {
	value := fmt.Sprintf(hashFormat, t.PID, t.OpenID, t.Timestamp)
	return cherryCrypto.MD5(value + appKey)
}
