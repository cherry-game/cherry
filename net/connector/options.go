package cherryConnector

import (
	clog "github.com/cherry-game/cherry/logger"
)

type (
	Options struct {
		address  string
		certFile string
		keyFile  string
		chanSize int
	}

	Option func(*Options)
)

func WithCert(certFile, keyFile string) Option {
	return func(o *Options) {
		if certFile != "" && keyFile != "" {
			o.certFile = certFile
			o.keyFile = keyFile
		} else {
			clog.Errorf("Cert config error.[cert = %s,key = %s]", certFile, keyFile)
		}
	}
}

func WithChanSize(size int) Option {
	return func(o *Options) {
		if size > 1 {
			o.chanSize = size
		}
	}
}
