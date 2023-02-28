package cherryFacade

type (
	// INetParser 前端网络数据包解析器
	INetParser interface {
		Load(application IApplication)
		AddConnector(connector IConnector)
	}
)
