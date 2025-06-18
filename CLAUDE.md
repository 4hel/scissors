# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a real-time multiplayer Rock Paper Scissors game built in Go using WebSockets. The application consists of:

- **Server** (`cmd/server/main.go`): WebSocket server handling matchmaking, game sessions, and move evaluation
- **Client** (`cmd/client/main.go`): Interactive CLI client for playing the game
- **Events** (`internal/events/events.go`): Shared event system and state machines

## Architecture

The application uses an event-driven architecture with well-defined state machines for both client and server:

### Client State Machine
- `Idle` → `FindingPartner` → `InGame` → `MakingMove` → `WaitingForResult` → `ViewingResults`
- Players can loop back to `InGame` (play again) or return to `Idle` (leave game)

### Server State Machine  
- `MatchmakingPool` → `GameSession` → `CollectingMoves` → `EvaluatingRound`
- Supports returning to `CollectingMoves` (play again) or `MatchmakingPool` (leave game)

### Key Components

- **Server**: Handles concurrent players with goroutines, manages game sessions, implements matchmaking
- **Client**: Interactive CLI with real-time WebSocket communication
- **Events**: Type-safe event system with structured payloads for client/server communication

## Common Commands

```bash
# Build and run server
go run cmd/server/main.go

# Build and run client (in separate terminal)
go run cmd/client/main.go

# Build both binaries
go build -o bin/server cmd/server/main.go
go build -o bin/client cmd/client/main.go

# Run tests
go test ./...

# Get dependencies
go mod tidy
```

## Development Notes

- Server runs on port 8080 with WebSocket endpoint `/ws`
- Client connects to `ws://localhost:8080/ws` by default
- All communication uses JSON-encoded events with type safety
- Server uses mutex locks for thread-safe game state management
- Game sessions are identified by unique IDs generated from timestamps
- Players are identified by unique IDs generated from timestamps