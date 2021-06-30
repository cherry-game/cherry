package cherryFacade

type (
	IPacketCodec interface {
		PacketDecode(data []byte) ([]IPacket, error)
		PacketEncode(typ byte, data []byte) ([]byte, error)
	}

	IPacket interface {
		Type() byte
		Len() int
		Data() []byte
	}
)
