# scissors
Rock Paper Scissors Game

## Game State Machine

```mermaid
stateDiagram-v2
    [*] --> Lobby
    Lobby --> WaitingForPartner : Find Partner
    WaitingForPartner --> GameReady : Partner Found
    GameReady --> PlayerTurn : Start Game
    PlayerTurn --> WaitingForOpponent : Move Submitted
    WaitingForOpponent --> ShowResults : Both Moves Received
    ShowResults --> GameReady : Play Again
    ShowResults --> Lobby : Leave Game
    GameReady --> Lobby : Leave Game
```
