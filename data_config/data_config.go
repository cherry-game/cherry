package cherryDataConfig

var (
	parserMap     = make(map[string]IParser)     //文件格式解析器
	dataSourceMap = make(map[string]IDataSource) //数据配置数据源
)

func init() {
	RegisterParser(new(JsonParser))
	RegisterSource(new(FileSource))
}

func GetParser(name string) IParser {
	return parserMap[name]
}

func RegisterParser(parser IParser) {
	parserMap[parser.TypeName()] = parser
}

func GetDataSource(name string) IDataSource {
	return dataSourceMap[name]
}

func RegisterSource(dataSource IDataSource) {
	dataSourceMap[dataSource.Name()] = dataSource
}
