// Package cherryFacade defines the core interfaces for the Cherry framework.
//
// This file defines the Actor model abstractions:
//   - IActorSystem: Actor runtime — create, dispatch, call Actors
//   - IActor: single Actor instance — send messages, query state
//   - IActorHandler: Actor lifecycle and message routing callbacks
//   - IActorChild: parent Actor's child management
//   - IEventData: typed event payload
package cherryFacade

import (
	"time"

	creflect "github.com/cherry-game/cherry/extend/reflect"
)

type (
	// IActorSystem is the Actor runtime used by application-level code.
	// It provides Actor creation, fire-and-forget message posting
	// (local/remote/event), and RPC-style Call/CallWait/CallType invocations.
	//
	// The framework registers a single IActorSystem implementation automatically
	// via the actor_component. Business code accesses it through IApplication.ActorSystem().
	IActorSystem interface {
		GetIActor(id string) (IActor, bool)                                   // lookup an Actor by ID; nil, false if not found
		CreateActor(id string, handler IActorHandler) (IActor, error)          // create and start a new Actor
		PostRemote(m *Message) bool                                            // fire-and-forget to a remote Actor; returns false if delivery rejected
		PostLocal(m *Message) bool                                             // fire-and-forget to a local Actor; returns false if Actor not found
		PostEvent(data IEventData)                                             // broadcast an event to all subscribers matching data.Name()
		Call(source, target, funcName string, arg any) int32                   // async RPC to target actor, returns cherryCode status code
		CallWait(source, target, funcName string, arg, reply any) int32         // sync RPC to target actor with reply, returns cherryCode status code
		CallType(nodeType, actorID, funcName string, arg any) int32             // call a random Actor of the given node type, returns cherryCode status code
		SetLocalInvoke(invoke InvokeFunc)                                      // set the low-level dispatch hook for local messages
		SetRemoteInvoke(invoke InvokeFunc)                                     // set the low-level dispatch hook for remote messages
		SetCallTimeout(d time.Duration)                                        // set RPC call timeout (default 3s)
		SetArrivalTimeout(t int64)                                             // set message arrival timeout in ms (default 100ms)
		SetExecutionTimeout(t int64)                                           // set handler execution timeout in ms (default 100ms)
	}

	// InvokeFunc is the low-level dispatch hook called when a message arrives at an Actor.
	// It receives the owning Application, the resolved function metadata (FuncInfo),
	// and the message envelope.
	//
	// The framework sets this via SetLocalInvoke / SetRemoteInvoke during startup.
	// Custom implementations are rare — the default invoke resolves the handler,
	// deserializes args, and calls the registered function.
	InvokeFunc func(app IApplication, fi *creflect.FuncInfo, m *Message)

	// IActor is a single Actor instance running on its own goroutine.
	// All handlers registered in OnInit are invoked serially on that goroutine —
	// no locks are needed for Actor-local state.
	//
	// Actors are created via IActorSystem.CreateActor and live until Exit()
	// is called or the Application shuts down.
	IActor interface {
		App() IApplication                                                   // the owning Application
		ActorID() string                                                     // unique ID within the node
		Path() *ActorPath                                                    // parsed "nodeID.actorID" path
		Call(targetPath, funcName string, arg any) int32                     // async RPC to another Actor, returns cherryCode status code
		CallWait(targetPath, funcName string, arg, reply any) int32           // sync RPC with reply, returns cherryCode status code
		CallType(nodeType, actorID, funcName string, arg any) int32           // call a random Actor of the given node type, returns cherryCode status code
		PostRemote(m *Message)                                               // fire-and-forget to a remote Actor
		PostLocal(m *Message)                                                // fire-and-forget to a local Actor
		LastAt() int64                                                       // last activity timestamp in ms, updated on each message
		Exit()                                                               // stop this Actor; cannot be restarted
	}

	// IActorHandler defines the lifecycle and message routing callbacks that every
	// Actor must implement. Pass an IActorHandler to IActorSystem.CreateActor
	// to register a new Actor type.
	//
	// Lifecycle order:
	//   AliasID() → OnInit() → [message loop] → OnStop()
	IActorHandler interface {
		AliasID() string                          // the ID string this Actor is registered under
		OnInit()                                  // called once before the message loop starts; register handlers here
		OnStop()                                  // called once before the goroutine exits; release resources here
		OnLocalReceived(m *Message) (bool, bool)  // returns (handled, needRecycle) — handled=true means the message was processed; needRecycle=true means caller should Recycle the message
		OnRemoteReceived(m *Message) (bool, bool) // returns (handled, needRecycle) — handled=true means the message was processed; needRecycle=true means caller should Recycle the message
		OnFindChild(m *Message) (IActor, bool)    // resolve a child Actor from the message's Target path; returns the child and true if found
	}

	// IActorChild is implemented by parent Actors that can spawn and manage
	// children. Children share the parent's goroutine — all Call and CallWait
	// invocations are serial, so child handlers never run concurrently with
	// each other or with the parent.
	IActorChild interface {
		Create(id string, handler IActorHandler) (IActor, error)    // create a child Actor on the parent's goroutine
		Get(id string) (IActor, bool)                               // get a child by ID; nil, false if not found
		Remove(id string)                                           // remove and stop a child Actor
		Each(fn func(i IActor))                                     // iterate all children (the callback must not mutate the child set)
		Call(childID, funcName string, arg any)                     // call a handler on a child Actor, returns cherryCode status code
		CallWait(childID, funcName string, arg, reply any) int32     // call a handler on a child and wait for reply, returns cherryCode status code
	}
)

type (
	// IEventData is a typed event payload posted via IActorSystem.PostEvent.
	// Actors subscribe by event Name in OnInit via EventRegister.
	//
	// Multiple PostEvent calls for the same event may arrive at a subscriber;
	// UniqueID allows the subscriber to detect and skip duplicates.
	IEventData interface {
		Name() string    // event name used for subscription matching
		UniqueID() int64 // unique ID for deduplication — two events with the same ID are the same occurrence
	}
)
