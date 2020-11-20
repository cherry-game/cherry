package cherryDataConfig

type Storage struct {
	fileName string
	models   []IConfigModel
	index    map[string][]IConfigModel
	pkIndex  map[string]IConfigModel
}
