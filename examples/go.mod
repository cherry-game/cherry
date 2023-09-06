module github.com/cherry-game/cherry/examples

go 1.18

require (
	github.com/ahmetb/go-linq/v3 v3.2.0
	github.com/cherry-game/cherry v1.3.4
	github.com/cherry-game/cherry/components/cron v1.3.4
	github.com/cherry-game/cherry/components/data-config v1.3.4
	github.com/cherry-game/cherry/components/gin v1.3.4
	github.com/cherry-game/cherry/components/gops v1.3.4
	github.com/cherry-game/cherry/components/gorm v1.3.4
	github.com/gin-gonic/gin v1.8.2
	github.com/go-redis/redis/v8 v8.11.5
	github.com/goburrow/cache v0.1.4
	github.com/json-iterator/go v1.1.12
	github.com/nats-io/nats.go v1.28.0
	github.com/spf13/cast v1.5.1
	github.com/urfave/cli/v2 v2.25.7
	go.uber.org/zap v1.23.0
	google.golang.org/protobuf v1.31.0
	gorm.io/gorm v1.25.4
)

require (
	github.com/cespare/xxhash/v2 v2.1.2 // indirect
	github.com/cpuguy83/go-md2man/v2 v2.0.2 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-playground/locales v0.14.0 // indirect
	github.com/go-playground/universal-translator v0.18.0 // indirect
	github.com/go-playground/validator/v10 v10.11.1 // indirect
	github.com/go-sql-driver/mysql v1.6.0 // indirect
	github.com/goccy/go-json v0.9.11 // indirect
	github.com/google/gops v0.3.28 // indirect
	github.com/gorilla/websocket v1.5.0 // indirect
	github.com/jinzhu/inflection v1.0.0 // indirect
	github.com/jinzhu/now v1.1.5 // indirect
	github.com/klauspost/compress v1.16.7 // indirect
	github.com/leodido/go-urn v1.2.1 // indirect
	github.com/lestrrat-go/strftime v1.0.6 // indirect
	github.com/mattn/go-isatty v0.0.16 // indirect
	github.com/minio/highwayhash v1.0.2 // indirect
	github.com/modern-go/concurrent v0.0.0-20180228061459-e0a39a4cb421 // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/nats-io/nkeys v0.4.4 // indirect
	github.com/nats-io/nuid v1.0.1 // indirect
	github.com/pelletier/go-toml/v2 v2.0.6 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/radovskyb/watcher v1.0.7 // indirect
	github.com/robfig/cron/v3 v3.0.1 // indirect
	github.com/russross/blackfriday/v2 v2.1.0 // indirect
	github.com/ugorji/go/codec v1.2.7 // indirect
	github.com/xrash/smetrics v0.0.0-20201216005158-039620a65673 // indirect
	go.uber.org/atomic v1.7.0 // indirect
	go.uber.org/multierr v1.6.0 // indirect
	golang.org/x/crypto v0.11.0 // indirect
	golang.org/x/net v0.10.0 // indirect
	golang.org/x/sys v0.11.0 // indirect
	golang.org/x/text v0.11.0 // indirect
	golang.org/x/time v0.3.0 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gorm.io/driver/mysql v1.3.6 // indirect
)

replace (
	github.com/cherry-game/cherry => ../../cherry
	github.com/cherry-game/cherry/components/cron => ../../cherry/components/cron
	github.com/cherry-game/cherry/components/data-config => ../../cherry/components/data-config
	github.com/cherry-game/cherry/components/gin => ../../cherry/components/gin
	github.com/cherry-game/cherry/components/gops => ../../cherry/components/gops
	github.com/cherry-game/cherry/components/gorm => ../../cherry/components/gorm
)
