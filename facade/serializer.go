// Package cherryFacade defines the core interfaces for the Cherry framework.
//
// This file defines the serialization abstraction:
//   - ISerializer: pluggable codec for message encoding/decoding
package cherryFacade

// ISerializer is a pluggable codec for message encoding and decoding.
// It is used by the cluster, Actor system, and connectors to serialize
// messages for cross-process transfer and deserialize incoming payloads.
//
// Multiple implementations may coexist (e.g. "json", "protobuf");
// the Name() method identifies which codec is in use.
type ISerializer interface {
	Marshal(v interface{}) ([]byte, error)   // encode a value into bytes
	Unmarshal(data []byte, v interface{}) error // decode bytes into a value
	Name() string                             // codec name (e.g. "json", "protobuf")
}
