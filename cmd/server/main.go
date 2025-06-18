package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"scissors/internal/events"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow connections from any origin
	},
}

type Player struct {
	ID   string
	Conn *websocket.Conn
	Move events.Move
}

type GameSession struct {
	ID      string
	State   events.ServerState
	Player1 *Player
	Player2 *Player
	Moves   map[string]events.Move
}

type Server struct {
	mu            sync.RWMutex
	waitingPlayer *Player
	gameSessions  map[string]*GameSession
	players       map[string]*Player
}

func NewServer() *Server {
	return &Server{
		gameSessions: make(map[string]*GameSession),
		players:      make(map[string]*Player),
	}
}

func (s *Server) handleConnection(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade connection: %v", err)
		return
	}
	defer conn.Close()

	playerID := fmt.Sprintf("player_%d", time.Now().UnixNano())
	player := &Player{
		ID:   playerID,
		Conn: conn,
	}

	s.mu.Lock()
	s.players[playerID] = player
	s.mu.Unlock()

	log.Printf("Player %s connected", playerID)

	for {
		var event events.Event
		err := conn.ReadJSON(&event)
		if err != nil {
			log.Printf("Player %s disconnected: %v", playerID, err)
			break
		}

		s.handleEvent(player, event)
	}

	s.mu.Lock()
	delete(s.players, playerID)
	s.mu.Unlock()
}

func (s *Server) handleEvent(player *Player, event events.Event) {
	switch event.Type {
	case events.FindPartnerEvent:
		s.handleFindPartner(player)
	case events.MoveSubmittedEvent:
		s.handleMoveSubmitted(player, event)
	case events.PlayAgainEvent:
		s.handlePlayAgain(player)
	case events.LeaveGameEvent:
		s.handleLeaveGame(player)
	}
}

func (s *Server) handleFindPartner(player *Player) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.waitingPlayer == nil {
		// No one waiting, this player becomes the waiting player
		s.waitingPlayer = player
		log.Printf("Player %s is waiting for a partner", player.ID)
	} else {
		// Match with waiting player
		gameID := fmt.Sprintf("game_%d", time.Now().UnixNano())
		game := &GameSession{
			ID:      gameID,
			State:   events.ServerGameSession,
			Player1: s.waitingPlayer,
			Player2: player,
			Moves:   make(map[string]events.Move),
		}

		s.gameSessions[gameID] = game
		
		// Notify both players
		partnerData := events.PartnerFoundData{GameID: gameID}
		s.sendEvent(s.waitingPlayer, events.Event{
			Type: events.PartnerFoundEvent,
			Data: partnerData,
		})
		s.sendEvent(player, events.Event{
			Type: events.PartnerFoundEvent,
			Data: partnerData,
		})

		// Start the game
		s.sendEvent(s.waitingPlayer, events.Event{Type: events.StartGameEvent})
		s.sendEvent(player, events.Event{Type: events.StartGameEvent})

		game.State = events.ServerCollectingMoves
		s.waitingPlayer = nil

		log.Printf("Game %s started between %s and %s", gameID, s.waitingPlayer.ID, player.ID)
	}
}

func (s *Server) handleMoveSubmitted(player *Player, event events.Event) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find the game this player is in
	var game *GameSession
	for _, g := range s.gameSessions {
		if (g.Player1 != nil && g.Player1.ID == player.ID) || 
		   (g.Player2 != nil && g.Player2.ID == player.ID) {
			game = g
			break
		}
	}

	if game == nil {
		log.Printf("No game found for player %s", player.ID)
		return
	}

	// Parse move data
	moveDataBytes, _ := json.Marshal(event.Data)
	var moveData events.MoveData
	json.Unmarshal(moveDataBytes, &moveData)

	game.Moves[player.ID] = moveData.Move
	log.Printf("Player %s submitted move: %s", player.ID, moveData.Move)

	// Check if both players have submitted moves
	if len(game.Moves) == 2 {
		game.State = events.ServerEvaluatingRound
		s.evaluateRound(game)
	}
}

func (s *Server) evaluateRound(game *GameSession) {
	player1Move := game.Moves[game.Player1.ID]
	player2Move := game.Moves[game.Player2.ID]

	result1 := s.determineResult(player1Move, player2Move)
	result2 := s.determineResult(player2Move, player1Move)

	// Send results to both players
	s.sendEvent(game.Player1, events.Event{
		Type: events.GameResultsEvent,
		Data: events.GameResultsData{
			YourMove:     player1Move,
			OpponentMove: player2Move,
			Result:       result1,
		},
	})

	s.sendEvent(game.Player2, events.Event{
		Type: events.GameResultsEvent,
		Data: events.GameResultsData{
			YourMove:     player2Move,
			OpponentMove: player1Move,
			Result:       result2,
		},
	})

	// Clear moves for next round
	game.Moves = make(map[string]events.Move)
	game.State = events.ServerGameSession

	log.Printf("Game %s round completed: %s vs %s", game.ID, player1Move, player2Move)
}

func (s *Server) determineResult(playerMove, opponentMove events.Move) string {
	if playerMove == opponentMove {
		return "tie"
	}

	switch playerMove {
	case events.Rock:
		if opponentMove == events.Scissors {
			return "win"
		}
	case events.Paper:
		if opponentMove == events.Rock {
			return "win"
		}
	case events.Scissors:
		if opponentMove == events.Paper {
			return "win"
		}
	}
	return "lose"
}

func (s *Server) handlePlayAgain(player *Player) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find the game and reset it for another round
	for _, game := range s.gameSessions {
		if (game.Player1 != nil && game.Player1.ID == player.ID) || 
		   (game.Player2 != nil && game.Player2.ID == player.ID) {
			game.State = events.ServerCollectingMoves
			s.sendEvent(game.Player1, events.Event{Type: events.StartGameEvent})
			s.sendEvent(game.Player2, events.Event{Type: events.StartGameEvent})
			break
		}
	}
}

func (s *Server) handleLeaveGame(player *Player) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Find and remove the game
	for gameID, game := range s.gameSessions {
		if (game.Player1 != nil && game.Player1.ID == player.ID) || 
		   (game.Player2 != nil && game.Player2.ID == player.ID) {
			delete(s.gameSessions, gameID)
			break
		}
	}

	// If this player was waiting, clear them
	if s.waitingPlayer != nil && s.waitingPlayer.ID == player.ID {
		s.waitingPlayer = nil
	}
}

func (s *Server) sendEvent(player *Player, event events.Event) {
	if err := player.Conn.WriteJSON(event); err != nil {
		log.Printf("Failed to send event to player %s: %v", player.ID, err)
	}
}

func main() {
	server := NewServer()
	
	http.HandleFunc("/ws", server.handleConnection)
	
	log.Println("Rock Paper Scissors server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", nil))
}