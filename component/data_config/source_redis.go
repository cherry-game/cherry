package cherryDataConfig

type SourceRedis struct {
}

func (r *SourceRedis) Name() string {
	return "redis"
}

func (r *SourceRedis) Init(_ IDataConfig) {

}

func (r *SourceRedis) Stop() {

}
