package cherryMessage

import (
	"fmt"
	"github.com/cherry-game/cherry/error"
	"strings"
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

func NewRouteByName(routeName string) *Route {
	r, err := DecodeRoute(routeName)
	if err != nil {
		return nil
	}
	return r
}

// String transforms the route into a string
func (r *Route) String() string {
	return fmt.Sprintf("%s.%s.%s", r.nodeType, r.handleName, r.method)
}

// DecodeRoute decodes the route
func DecodeRoute(route string) (*Route, error) {
	if route == "" {
		return nil, cherryError.RouteFieldCantEmpty
	}

	r := strings.Split(route, ".")
	for _, s := range r {
		if strings.TrimSpace(s) == "" {
			return nil, cherryError.RouteFieldCantEmpty
		}
	}

	if len(r) != 3 {
		return nil, cherryError.RouteInvalid
	}

	return NewRoute(r[0], r[1], r[2]), nil
}
