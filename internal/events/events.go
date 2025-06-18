package events

// Move represents a player's move in rock paper scissors
type Move string

const (
	Rock     Move = "rock"
	Paper    Move = "paper"
	Scissors Move = "scissors"
)

// ClientState represents the client's current state
type ClientState string

const (
	ClientIdle             ClientState = "idle"
	ClientFindingPartner   ClientState = "finding_partner"
	ClientInGame           ClientState = "in_game"
	ClientMakingMove       ClientState = "making_move"
	ClientWaitingForResult ClientState = "waiting_for_result"
	ClientViewingResults   ClientState = "viewing_results"
)

// ServerState represents the server's current state for a game session
type ServerState string

const (
	ServerMatchmakingPool ServerState = "matchmaking_pool"
	ServerGameSession     ServerState = "game_session"
	ServerCollectingMoves ServerState = "collecting_moves"
	ServerEvaluatingRound ServerState = "evaluating_round"
)

// EventType represents the type of event being sent
type EventType string

const (
	// Client to Server events
	FindPartnerEvent        EventType = "find_partner"
	MoveSubmittedEvent      EventType = "move_submitted"
	PlayAgainEvent          EventType = "play_again"
	LeaveGameEvent          EventType = "leave_game"

	// Server to Client events
	PartnerFoundEvent       EventType = "partner_found"
	StartGameEvent          EventType = "start_game"
	BothMovesReceivedEvent  EventType = "both_moves_received"
	GameResultsEvent        EventType = "game_results"
)

// Event represents a message sent between client and server
type Event struct {
	Type EventType   `json:"type"`
	Data interface{} `json:"data,omitempty"`
}

// MoveData represents the data for a move submission
type MoveData struct {
	Move Move `json:"move"`
}

// GameResultsData represents the results of a game round
type GameResultsData struct {
	YourMove     Move   `json:"your_move"`
	OpponentMove Move   `json:"opponent_move"`
	Result       string `json:"result"` // "win", "lose", "tie"
}

// PartnerFoundData represents data when a partner is found
type PartnerFoundData struct {
	GameID string `json:"game_id"`
}