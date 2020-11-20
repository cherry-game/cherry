package cherryInterfaces

//ISerializer 消息序列化
type ISerializer interface {
	Marshal(interface{}) ([]byte, error)

	Unmarshal([]byte, interface{}) error

	Name() string
}
