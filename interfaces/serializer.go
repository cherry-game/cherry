package cherryInterfaces

//ISerializer 消息序列化
type ISerializer interface {
	Marshal(interface{}) ([]byte, error) // 编码
	Unmarshal([]byte, interface{}) error // 解码
	Name() string                        // 序列化类型的名称
}
