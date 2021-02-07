package cherryDataConfig

type (
	IDataConfig interface {
		Register(file IConfigFile)         // 注册文件
		GetFiles() []IConfigFile           // 获取注册的文件列表
		Get(fileName string) interface{}   // 获取原始的数据
		Load(fileName string, data []byte) // 加载数据流
	}

	// IDataSource 配置文件数据源
	IDataSource interface {
		Name() string                // 数据源名称
		Init(dataConfig IDataConfig) // 函数初始化时
		Stop()                       // 停止
	}

	Parser func(text []byte, v interface{}) error // 文件格式解析器

	// IConfigFile 配置文件接口
	IConfigFile interface {
		FileName() string // 文件名
		Init()            // 文件序列化后，执行该函数
		Reload()          // 文件重加载后，先执行Init(),再执行该函数
	}
)
