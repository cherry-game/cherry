package main

import (
	"github.com/nats-io/nats.go"
	"log"
	"os"
	"os/signal"
	"time"
)

func main() {
	urls := nats.DefaultURL
	var opts []nats.Option
	opts = setupConnOptions(opts)

	var subj = "cherry.nodes.game-1.10001"

	nc, err := nats.Connect(urls, opts...)
	if err != nil {
		log.Fatal(err)
	}

	var i = 0
	for {
		if i == 10 {
			break
		}

		nc.Publish(subj, []byte("aaa"))
		time.Sleep(1 * time.Second)
		i++
	}

	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)
	<-c
	log.Println()
	log.Printf("Draining...")
	nc.Drain()
	log.Fatalf("Exiting")
}

func setupConnOptions(opts []nats.Option) []nats.Option {
	totalWait := 10 * time.Minute
	reconnectDelay := time.Second

	opts = append(opts, nats.ReconnectWait(reconnectDelay))

	opts = append(opts, nats.MaxReconnects(int(totalWait/reconnectDelay)))

	opts = append(opts, nats.DisconnectErrHandler(func(conn *nats.Conn, err error) {
		log.Printf("Disconnected: will attempt reconnects for %.0fm", totalWait.Minutes())
	}))

	opts = append(opts, nats.ReconnectHandler(func(nc *nats.Conn) {
		log.Printf("Reconnected [%s]", nc.ConnectedUrl())
	}))
	opts = append(opts, nats.ClosedHandler(func(nc *nats.Conn) {
		log.Fatalf("Exiting: %v", nc.LastError())
	}))

	return opts
}
