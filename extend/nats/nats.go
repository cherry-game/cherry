package cherryNATS

import (
	cherryError "github.com/cherry-game/cherry/error"
	cherryLogger "github.com/cherry-game/cherry/logger"
	cherryProfile "github.com/cherry-game/cherry/profile"
	"github.com/nats-io/nats.go"
)

func Connect(cfg *cherryProfile.NatsConfig) (*nats.Conn, error) {
	options := ParseOptions(cfg)
	nats, err := nats.Connect(cfg.Address, options...)
	if err != nil {
		return nil, cherryError.Errorf("nats address = %s, err = %s", cfg.Address, err)
	}
	return nats, nil
}

func ParseOptions(cfg *cherryProfile.NatsConfig) []nats.Option {
	var options []nats.Option

	if cfg.ReconnectDelay > 0 {
		options = append(options, nats.ReconnectWait(cfg.ReconnectDelay))
	}

	if cfg.MaxReconnects > 0 {
		options = append(options, nats.MaxReconnects(cfg.MaxReconnects))
	}

	options = append(options, nats.DisconnectErrHandler(func(conn *nats.Conn, err error) {
		if err != nil {
			cherryLogger.Warnf("disconnect error. [error = %v]", err)
		}
	}))

	options = append(options, nats.ReconnectHandler(func(nc *nats.Conn) {
		cherryLogger.Warnf("reconnected [%s]", nc.ConnectedUrl())
	}))

	options = append(options, nats.ClosedHandler(func(nc *nats.Conn) {
		cherryLogger.Infof("exiting... %s", cfg.Address)
		if nc.LastError() != nil {
			cherryLogger.Infof(".[error = %v]", nc.LastError())
		}
	}))

	if cfg.User != "" {
		options = append(options, nats.UserInfo(cfg.User, cfg.Password))
	}

	return options
}
