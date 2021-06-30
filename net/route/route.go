package cherryRoute

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

// New create a new route
func New(nodeType, handleName, method string) *Route {
	return &Route{nodeType, handleName, method}
}

func NewByName(routeName string) *Route {
	r, err := Decode(routeName)
	if err != nil {
		return nil
	}
	return r
}

// String transforms the route into a string
func (r *Route) String() string {
	return fmt.Sprintf("%s.%s.%s", r.nodeType, r.handleName, r.method)
}

// Decode decodes the route
func Decode(route string) (*Route, error) {
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

	return New(r[0], r[1], r[2]), nil
}
