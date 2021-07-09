package cherryMessage

// Message types
const (
	Request  Type = 0x00 // ----000-
	Notify   Type = 0x01 // ----001-
	Response Type = 0x02 // ----010-
	Push     Type = 0x03 // ----011-
)

// Type represents the type of message, which could be Request/Notify/Response/Push
type Type byte

var types = map[Type]string{
	Request:  "Request",
	Notify:   "Notify",
	Response: "Response",
	Push:     "Push",
}

func (t *Type) String() string {
	return types[*t]
}

func routable(t Type) bool {
	return t == Request || t == Notify || t == Push
}

func invalidType(t Type) bool {
	return t < Request || t > Push
}
