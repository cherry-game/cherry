package cherryDataConfig

type SourceRedis struct {
}

func (r *SourceRedis) Name() string {
	return "redis"
}

func (r *SourceRedis) Init(_ IDataConfig) {

}

func (r *SourceRedis) ReadBytes(configName string) (data []byte, error error) {
	return nil, nil
}

func (r *SourceRedis) OnChange(fn ConfigChangeFn) {

}

func (r *SourceRedis) Stop() {

}
