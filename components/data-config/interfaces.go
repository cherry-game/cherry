package cherryDataConfig

type (
	// IDataConfig 数据配置接口
	IDataConfig interface {
		Register(configFile ...IConfig)                       // 注册映射文件
		GetBytes(configName string) (data []byte, found bool) // 获取原始的数据
		GetParser() IDataParser                               // 当前参数配置的数据格式解析器
		GetDataSource() IDataSource                           // 当前参数配置的获取数据源
	}

	// IDataParser 数据格式解析接口
	IDataParser interface {
		TypeName() string                           // 注册名称
		Unmarshal(text []byte, v interface{}) error // 文件格式解析器
	}

	// IDataSource 配置文件数据源
	IDataSource interface {
		Name() string                                           // 数据源名称
		Init(dataConfig IDataConfig)                            // 函数初始化时
		ReadBytes(configName string) (data []byte, error error) // 获取数据流
		OnChange(fn ConfigChangeFn)                             // 数据变更时
		Stop()                                                  // 停止
	}

	// ConfigChangeFn 数据变更时触发该函数
	ConfigChangeFn func(configName string, data []byte)

	// IConfig 配置接口
	IConfig interface {
		Name() string                                      // 配置名称
		Init()                                             // 结构体初始化
		OnLoad(maps interface{}, reload bool) (int, error) // 配置序列化后，执行该函数 (size,error)
		OnAfterLoad(reload bool)                           // 所有配置加载后再执行该函数
	}
)
