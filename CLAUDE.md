# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build & Test

```bash
go build ./...           # full build
go test ./...            # full test suite
go test -run TestX ./net/actor/  # single test
go test -count=1 ./net/discovery/  # no cache
```

## Architecture

cherry is a Go game server framework built on the Actor Model. Understand the codebase through 4 execution chains:

| Chain | Key Paths | Role |
|-------|-----------|------|
| App Assembly | `./cherry.go` → `./application.go` | Component registration, lifecycle, startup |
| Actor Execution | `./net/actor/` | Per-Actor goroutine, serial mailbox, local/remote/event dispatch |
| Cluster Communication | `./net/cluster/` + `./net/nats/` + `./net/discovery/` | Cross-node RPC, NATS transport, member discovery |
| Frontend Access | `./net/parser/` + `./net/connector/` | Protocol decode, session, agent Actor, WebSocket/TCP |

## Startup Order

1. `Configure(...)` creates `Application`
2. `AppBuilder.Startup()` auto-injects `cluster_component` + `discovery_component` in Cluster mode
3. Register custom components
4. `Application.Startup()` auto-registers `actor_component`
5. If `netParser` is set, its connectors are registered as components
6. All components run `Set → Init → OnAfterInit` in registration order
7. If `isFrontend == true`, `netParser` must be set or `panic`
8. Block until signal or `Shutdown()`; components stop in reverse order: `OnBeforeStop → OnStop`

## Key Conventions

- Actor path format: `{nodeID}.{actorID}` or `{nodeID}.{actorID}.{childID}`
- Component `Name()` must be globally unique within an Application
- `Application.Register()` only works before startup
- `actor_component` is registered automatically — never register it manually
- `Cluster` mode auto-registers `cluster_component` and `discovery_component`
- `Application.Startup()` blocks until shutdown, not "init and return"
- `Message` uses an object pool; receiver recycles on success, caller must `Recycle()` on delivery failure
- Child Actors can only be created by parent Actors; children cannot create further children
- `sync.Map` zero value is usable — no explicit initialization needed (see `net/discovery/`)

## Module Reference

- **facade/** — framework contracts (`IApplication`, `IComponent`, `IActorSystem`, `IDiscovery`, `ICluster`, `INetParser`, `ISerializer`, `Message`, `ActorPath`). Define interfaces here first, then implement.
- **net/actor/** — execution kernel. `CreateActor()` starts a goroutine. `Call()` / `CallWait()` dispatch locally or via Cluster. Default timeouts: `callTimeout=3s`, `arrivalTimeOut=100ms`, `executionTimeout=100ms` (`./net/actor/system.go:30`).
- **net/cluster/** — wraps discovery + NATS for cross-node message routing (`PublishLocal`, `PublishRemote`, `RequestRemote`).
- **net/nats/** — NATS connection pool + per-connection state (`Connect.options`, `waiters`, `subs`). `RequestSync()` uses reply subject per connection, not global. Disconnect clears all waiters.
- **net/discovery/** — node discovery with 3 backends: `default` (profile file), `nats` (master/worker heartbeat), `etcd` (separate repo). `ComponentDefault` is a composable base class with `memberMap` + listeners.
- **net/parser/** — frontend access layer. Not just codec — it assembles connectors, creates sessions and agent Actors. `pomelo/` and `simple/` implementations.
- **profile/** — config loader with package-level global state. `GetDuration()` returns raw `time.Duration` (no unit — caller multiplies). Include files merged with main config override.

## Global State (package-level singletons)

- `profile/` — `profilePath`, `jsonConfig`, `env`, `debug`, `printLevel`
- `net/nats/` — `connectPool`, `roundIndex`
- `logger/` — `DefaultLogger`
- `net/discovery/` — `discoveryMap` (component registry)

Multiple Application instances in one process will conflict on these.

## Pitfalls

- `isFrontend == true` without a parser → `panic` at startup
- `INetParser`'s real job is assembling connectors and loading agent Actors, not just encode/decode
- Thread-safe assumptions: Actor handlers are serial; shared state across Actors needs explicit synchronization
- `net/nats/` reply subject is per-connection, not global. `RequestSync()` timeout leaves late-arriving replies discarded.
- `net/discovery/` nats mode relies on background goroutines for heartbeat; memory leaks if `ctx` is not cancelled on shutdown.
- Child Actor lifecycle is tied to parent — a stuck child blocks parent shutdown.

## Cross-Module Checks

- Change `application.go` → check `./cherry.go`, `./facade/application.go`
- Change `facade/*` → check corresponding implementation directories
- Change `net/actor/*` → check `./facade/actor.go`, `./net/cluster/`, `./code/`, `./error/`
- Change `net/discovery/*` → check `./profile/`, `./net/nats/`, `./facade/cluster.go`
- Change `net/cluster/*` → check `./net/nats/`, `./net/discovery/`
- Change `net/nats/*` → check `./net/cluster/`, `./net/discovery/`
- Change `net/parser/*` → check `./net/connector/`, `./facade/net_parser.go`
- Change `profile/*` → check `./application.go`, `./logger/`, `./net/discovery/`
