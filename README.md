# scissors
Rock Paper Scissors Game

## Game State Machine

```mermaid
stateDiagram-v2
    [*] --> Lobby
    Lobby --> WaitingForPartner : find_partner
    WaitingForPartner --> GameReady : partner_found
    GameReady --> PlayerTurn : start_game
    PlayerTurn --> WaitingForOpponent : move_submitted
    WaitingForOpponent --> ShowResults : both_moves_received
    ShowResults --> GameReady : play_again
    ShowResults --> Lobby : leave_game
    GameReady --> Lobby : leave_game
```
