package cherryProfile

//type NatsConfig struct {
//	Address        string
//	ReconnectDelay time.Duration
//	MaxReconnects  int
//	RequestTimeout time.Duration
//	User           string
//	Password       string
//}

//func NewNatsConfig() *NatsConfig {
//	cluster := Config().Get("cluster")
//	nats := cluster.Get("nats")
//	return NewNatsWithConfig(nats)
//}

//func NewNatsWithConfig(config jsoniter.Any) *NatsConfig {
//	address := config.Get("address").ToString()
//	if address == "" {
//		panic("address is empty!")
//	}
//
//	reconnectDelay := config.Get("reconnect_delay").ToInt64()
//	if reconnectDelay < 1 {
//		reconnectDelay = 1
//	}
//
//	maxReconnects := config.Get("max_reconnects").ToInt()
//	if maxReconnects < 1 {
//		maxReconnects = 10
//	}
//
//	requestTimeout := config.Get("request_timeout").ToInt64()
//	if requestTimeout < 1 {
//		requestTimeout = 1
//	}
//
//	user := config.Get("user").ToString()
//	password := config.Get("password").ToString()
//
//	cfg := &NatsConfig{
//		Address:        address,
//		ReconnectDelay: time.Duration(reconnectDelay) * time.Second,
//		MaxReconnects:  maxReconnects,
//		RequestTimeout: time.Duration(requestTimeout) * time.Second,
//		User:           user,
//		Password:       password,
//	}
//
//	return cfg
//}

//func (p *NatsConfig) GetOptions() []nats.Option {
//	var options []nats.Option
//
//	if p.ReconnectDelay > 0 {
//		options = append(options, nats.ReconnectWait(p.ReconnectDelay))
//	}
//
//	if p.MaxReconnects > 0 {
//		options = append(options, nats.MaxReconnects(p.MaxReconnects))
//	}
//
//	options = append(options, nats.DisconnectErrHandler(func(conn *nats.Conn, err error) {
//		if err != nil {
//			cherryLogger.Warnf("disconnect error. [error = %v]", err)
//		}
//	}))
//
//	options = append(options, nats.ReconnectHandler(func(nc *nats.Conn) {
//		cherryLogger.Warnf("reconnected [%s]", nc.ConnectedUrl())
//	}))
//
//	options = append(options, nats.ClosedHandler(func(nc *nats.Conn) {
//		cherryLogger.Infof("exiting... %s", p.Address)
//		if nc.LastError() != nil {
//			cherryLogger.Infof(".[error = %v]", nc.LastError())
//		}
//	}))
//
//	if p.User != "" {
//		options = append(options, nats.UserInfo(p.User, p.Password))
//	}
//
//	return options
//}
