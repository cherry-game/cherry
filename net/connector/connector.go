package cherryConnector

import (
	"crypto/tls"
	"github.com/cherry-game/cherry/logger"
	"net"
	"net/http"
)

// GetNetListener 证书构造 net.Listener
//
// address 监听地址
//
// certFile 证书文件
//
// keyFile 公钥文件
func GetNetListener(address, certFile, keyFile string) (net.Listener, error) {
	if certFile == "" || keyFile == "" {
		return net.Listen("tcp", address)
	}

	crt, err := tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		cherryLogger.Fatalf("failed to listen: %s", err.Error())
	}
	tlsCfg := &tls.Config{Certificates: []tls.Certificate{crt}}

	return tls.Listen("tcp", address, tlsCfg)
}

//CheckOrigin
//
//请求检查函数 防止跨站请求伪造 true则不检查
func CheckOrigin(_ *http.Request) bool {
	return true
}
