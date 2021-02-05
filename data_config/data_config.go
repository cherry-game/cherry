package cherryDataConfig

import (
	jsoniter "github.com/json-iterator/go"
)

var (
	parserMap = make(map[string]Parser)      //文件格式解析器
	sourceMap = make(map[string]IDataSource) //配置文件数据源
)

func GetParser(name string) Parser {
	return parserMap[name]
}

func RegisterParser(name string, parse Parser) {
	parserMap[name] = parse
}

func GetDataSource(name string) IDataSource {
	return sourceMap[name]
}

func RegisterDataSource(source IDataSource) {
	sourceMap[source.Name()] = source
}

func init() {
	// register default json parser
	RegisterParser("json", jsoniter.Unmarshal)
	RegisterDataSource(new(FileSource))
}
