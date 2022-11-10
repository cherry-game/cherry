package cherryMessage

import (
	cerr "github.com/cherry-game/cherry/error"
	"strings"
)

const (
	dot = "."
)

// Route struct
type Route struct {
	nodeType   string // node server type name
	handleName string // handle name
	method     string // method name
}

func (r *Route) NodeType() string {
	return r.nodeType
}

func (r *Route) HandleName() string {
	return r.handleName
}

func (r *Route) Method() string {
	return r.method
}

// NewRoute create a new route
func NewRoute(nodeType, handleName, method string) *Route {
	return &Route{nodeType, handleName, method}
}

// String transforms the route into a string
func (r *Route) String() string {
	return r.nodeType + dot + r.handleName + dot + r.method
}

// DecodeRoute decodes the route
func DecodeRoute(route string) (*Route, error) {
	if route == "" {
		return nil, cerr.RouteFieldCantEmpty
	}

	r := strings.Split(route, dot)
	for _, s := range r {
		if strings.TrimSpace(s) == "" {
			return nil, cerr.RouteFieldCantEmpty
		}
	}

	if len(r) != 3 {
		return nil, cerr.RouteInvalid
	}

	return NewRoute(r[0], r[1], r[2]), nil
}
