package cherryNUID

import "github.com/nats-io/nuid"

var (
	id = nuid.New()
)

func Next() string {
	return id.Next()
}
