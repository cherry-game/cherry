package cherryDataConfig

type RedisSource struct {
}

func (r *RedisSource) Name() string {
	return "redis"
}

func (r *RedisSource) Init(dataConfig IDataConfig) {

}

func (r *RedisSource) Stop() {

}
