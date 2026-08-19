package cherryNats

import (
	"time"

	cerror "github.com/cherry-game/cherry/error"
	cfacade "github.com/cherry-game/cherry/facade"
)

// NewConnectFromConfig builds a *Connect from the cluster.nats config section.
// name identifies the connect (used in logs and as the reply subject suffix);
// it must be unique per process among connects that subscribe the same reply
// subject base. The returned Connect is NOT connected and has no reply
// subscription; the caller decides when to call Connect() and, if the connect
// will issue RequestSync, whether to call SubscribeReply.
func NewConnectFromConfig(config cfacade.ProfileJSON, name string) (*Connect, error) {
	address := config.GetString("address")
	if address == "" {
		return nil, cerror.Error("nats address is empty")
	}

	conn := NewConnect(name,
		WithAddress(address),
		WithReconnectDelay(config.GetDuration("reconnect_delay", 1)*time.Second),
		WithRequestTimeout(config.GetDuration("request_timeout", 1)*time.Second),
		WithAuth(config.GetString("user"), config.GetString("password")),
		WithParams(config.GetInt("max_reconnects")),
		WithIsStats(config.GetBool("is_stats")),
		WithStatsInterval(config.GetInt("stats_interval", 30)),
	)
	return conn, nil
}
