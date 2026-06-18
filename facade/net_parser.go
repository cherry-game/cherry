// Package cherryFacade defines the core interfaces for the Cherry framework.
//
// This file defines the frontend network parser abstraction:
//   - INetParser: protocol codec + connector assembly + agent Actor loading
package cherryFacade

type (
	// INetParser assembles the frontend access layer for the Application.
	// It provides protocol-specific decoding, creates agent Actors for each
	// connected client, and manages the set of network connectors.
	//
	// Typical implementations (pomelo, simple) load their protocol handlers
	// during Load(), attach to connector callbacks, and bridge raw connections
	// to the Actor system via sessions and agent Actors.
	//
	// If the Application is configured with isFrontend == true, a parser must
	// be set or startup panics.
	INetParser interface {
		Load(application IApplication)       // register protocol-specific agent Actors
		AddConnector(connector IConnector)   // attach a network connector to this parser
		Connectors() []IConnector            // return all attached connectors
	}
)
