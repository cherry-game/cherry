package cherryDataConfig

type IDataConfig interface {
	GetFirst(index *IndexObject, params ...interface{}) interface{}

	GetList(tableName string) interface{}

	GetIndexList(index *IndexObject, params ...interface{}) interface{}

	Reload(fileName string, text []byte) error

	CheckFileName(fileName string, text []byte) error

	RegisterModel(models ...IConfigModel) error

	RegisterService(service IConfigService)
}

type IDataParse interface {
	Parse(text []byte, v interface{}) error
}

type IDataSource interface {
	Name() string

	Init()

	Destroy()

	SetParse(parse IDataParse)

	GetContent(fileName string) ([]byte, error)

	GetConfigNames() []string
}

type IConfigModel interface {
	FileName() string

	Init()
}

type IConfigService interface {
	Clean(typ interface{})

	Init(dataConfig IDataConfig)
}
