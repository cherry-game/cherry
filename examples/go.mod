module github.com/cherry-game/cherry/examples

go 1.17

replace (
	github.com/cherry-game/cherry => ./../
	github.com/cherry-game/cherry/components/cron => ../components/cron
	github.com/cherry-game/cherry/components/data-config => ../components/data-config
	github.com/cherry-game/cherry/components/gin => ../components/gin
)

require (
	github.com/ahmetb/go-linq/v3 v3.2.0
	github.com/cherry-game/cherry v1.2.2
	github.com/cherry-game/cherry/components/cron v0.0.0-00010101000000-000000000000
	github.com/cherry-game/cherry/components/data-config v0.0.0-00010101000000-000000000000
	github.com/cherry-game/cherry/components/gin v0.0.0-00010101000000-000000000000
	github.com/gin-gonic/gin v1.8.1
	github.com/go-redis/redis/v8 v8.11.5
	github.com/goburrow/cache v0.1.4
	github.com/gogo/protobuf v1.3.2
	github.com/json-iterator/go v1.1.12
	github.com/nats-io/nats.go v1.16.0
	github.com/spf13/cast v1.4.1
	github.com/urfave/cli/v2 v2.3.0
	go.uber.org/atomic v1.7.0
	go.uber.org/zap v1.22.0
)

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.0-20190314233015-f79a8a8ca69d // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-playground/locales v0.14.0 // indirect
	github.com/go-playground/universal-translator v0.18.0 // indirect
	github.com/go-playground/validator/v10 v10.10.0 // indirect
	github.com/goccy/go-json v0.9.7 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/lestrrat-go/strftime v1.0.6 // indirect
	github.com/mattn/go-isatty v0.0.14 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/nats-io/nkeys v0.3.0 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/pelletier/go-toml/v2 v2.0.1 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/radovskyb/watcher v1.0.7 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/russross/blackfriday/v2 v2.0.1 // indirect
	github.com/shurcooL/sanitized_anchor_name v1.0.0 // indirect
	github.com/ugorji/go/codec v1.2.7 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/crypto v0.0.0-20220315160706-3147a52a75dd // indirect
	golang.org/x/net v0.0.0-20211112202133-69e39bad7dc2 // indirect
	golang.org/x/sys v0.0.0-20220111092808-5a964db01320 // indirect
	golang.org/x/text v0.3.6 // indirect
	google.golang.org/protobuf v1.28.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
)
