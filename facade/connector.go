// Package cherryFacade defines the core interfaces for the Cherry framework.
//
// This file defines the network connector abstraction:
//   - IConnector: listen/accept connections and emit OnConnect callbacks
package cherryFacade

import "net"

// IConnector manages a network listener, accepting connections on a specific
// address and protocol. It emits each new connection through OnConnect callbacks,
// which downstream parsers use to create sessions and agent Actors.
//
// Typical implementations wrap a TCP or WebSocket listener and handle accept loops
// internally. Start/Stop are called by the framework during application startup/shutdown.
type IConnector interface {
	IComponent
	Start()                     // begin accepting connections
	Stop()                      // stop the listener and drain pending connections
	OnConnect(fn OnConnectFunc) // register a callback invoked for each new connection
}

// OnConnectFunc is called when a new connection is established.
type OnConnectFunc func(conn net.Conn)
