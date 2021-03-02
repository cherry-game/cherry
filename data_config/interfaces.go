package cherryDataConfig

type (
	// IDataConfig 数据配置接口
	IDataConfig interface {
		//Register(configName string, typ interface{}) // 注册映射文件
		//GetFiles() []IConfigFile                         // 获取注册的文件列表
		GetBytes(configName string) (data []byte, found bool) // 获取原始的数据
		GetParser() IParser                                   // 当前参数配置的数据格式解析器
		GetDataSource() IDataSource                           // 当前参数配置的获取数据源
	}

	// IDataSource 配置文件数据源
	IDataSource interface {
		Name() string                                          // 数据源名称
		Init(dataConfig IDataConfig)                           // 函数初始化时
		ReadData(configName string) (data []byte, error error) // 获取数据
		OnChange(changeFileFn ChangeFileFn)                    // 数据变更时
		Stop()                                                 // 停止
	}

	// ChangeFileFn 数据变更时触发该函数
	ChangeFileFn func(configName string, data []byte)

	// IParser 数据格式解析接口
	IParser interface {
		TypeName() string                           // 注册名称
		Unmarshal(text []byte, v interface{}) error // 文件格式解析器
	}

	// IConfigFile 配置文件接口
	IConfigFile interface {
		Name() string // 配置名称
		Init()        // 配置序列化后，执行该函数
		Reload()      // 配置重加载后，先执行Init(),再执行该函数
	}
)
