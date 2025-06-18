package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/gorilla/websocket"
	"scissors/internal/events"
)

type Client struct {
	conn  *websocket.Conn
	state events.ClientState
}

func NewClient() *Client {
	return &Client{
		state: events.ClientIdle,
	}
}

func (c *Client) connect(serverURL string) error {
	conn, _, err := websocket.DefaultDialer.Dial(serverURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to server: %v", err)
	}
	c.conn = conn
	return nil
}

func (c *Client) listenForEvents() {
	for {
		var event events.Event
		err := c.conn.ReadJSON(&event)
		if err != nil {
			log.Printf("Connection closed: %v", err)
			break
		}
		c.handleEvent(event)
	}
}

func (c *Client) handleEvent(event events.Event) {
	switch event.Type {
	case events.PartnerFoundEvent:
		c.state = events.ClientInGame
		fmt.Println("🎉 Partner found! Get ready to play!")

	case events.StartGameEvent:
		c.state = events.ClientMakingMove
		fmt.Println("\n🎮 New round starting!")
		c.promptForMove()

	case events.GameResultsEvent:
		c.state = events.ClientViewingResults
		c.handleGameResults(event)

	case events.BothMovesReceivedEvent:
		c.state = events.ClientWaitingForResult
		fmt.Println("⏳ Both moves received, waiting for results...")
	}
}

func (c *Client) handleGameResults(event events.Event) {
	resultsBytes, _ := json.Marshal(event.Data)
	var results events.GameResultsData
	json.Unmarshal(resultsBytes, &results)

	fmt.Printf("\n📊 ROUND RESULTS:\n")
	fmt.Printf("   Your move: %s\n", strings.ToUpper(string(results.YourMove)))
	fmt.Printf("   Opponent move: %s\n", strings.ToUpper(string(results.OpponentMove)))

	switch results.Result {
	case "win":
		fmt.Printf("   🏆 You WIN!\n")
	case "lose":
		fmt.Printf("   😞 You lose!\n")
	case "tie":
		fmt.Printf("   🤝 It's a TIE!\n")
	}

	fmt.Printf("\nWhat would you like to do?\n")
	fmt.Printf("1. Play again\n")
	fmt.Printf("2. Leave game\n")
	fmt.Printf("Enter your choice (1-2): ")

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		choice := strings.TrimSpace(scanner.Text())
		switch choice {
		case "1":
			c.sendEvent(events.Event{Type: events.PlayAgainEvent})
			fmt.Println("⏳ Waiting for next round...")
		case "2":
			c.sendEvent(events.Event{Type: events.LeaveGameEvent})
			c.state = events.ClientIdle
			fmt.Println("👋 Left the game. Thanks for playing!")
			c.showMainMenu()
		default:
			fmt.Println("Invalid choice, leaving game...")
			c.sendEvent(events.Event{Type: events.LeaveGameEvent})
			c.state = events.ClientIdle
			c.showMainMenu()
		}
	}
}

func (c *Client) promptForMove() {
	fmt.Printf("\nMake your move:\n")
	fmt.Printf("1. ROCK 🗿\n")
	fmt.Printf("2. PAPER 📄\n")
	fmt.Printf("3. SCISSORS ✂️\n")
	fmt.Printf("Enter your choice (1-3): ")

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		choice := strings.TrimSpace(scanner.Text())

		var move events.Move
		switch choice {
		case "1":
			move = events.Rock
		case "2":
			move = events.Paper
		case "3":
			move = events.Scissors
		default:
			fmt.Println("Invalid choice! Please enter 1, 2, or 3.")
			c.promptForMove()
			return
		}

		fmt.Printf("You chose: %s\n", strings.ToUpper(string(move)))
		c.sendEvent(events.Event{
			Type: events.MoveSubmittedEvent,
			Data: events.MoveData{Move: move},
		})

		c.state = events.ClientWaitingForResult
		fmt.Println("⏳ Move submitted! Waiting for opponent...")
	}
}

func (c *Client) sendEvent(event events.Event) {
	if err := c.conn.WriteJSON(event); err != nil {
		log.Printf("Failed to send event: %v", err)
	}
}

func (c *Client) showMainMenu() {
	if c.state != events.ClientIdle {
		return
	}

	fmt.Printf("\n🎮 Rock Paper Scissors Game\n")
	fmt.Printf("1. Find a partner to play\n")
	fmt.Printf("2. Quit\n")
	fmt.Printf("Enter your choice (1-2): ")

	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		choice := strings.TrimSpace(scanner.Text())
		switch choice {
		case "1":
			c.state = events.ClientFindingPartner
			c.sendEvent(events.Event{Type: events.FindPartnerEvent})
			fmt.Println("🔍 Looking for a partner...")
		case "2":
			fmt.Println("👋 Goodbye!")
			c.conn.Close()
			os.Exit(0)
		default:
			fmt.Println("Invalid choice! Please enter 1 or 2.")
			c.showMainMenu()
		}
	}
}

func main() {
	client := NewClient()

	// Connect to server
	serverURL := "wss://ws.dingodream.dev/ws"
	if err := client.connect(serverURL); err != nil {
		log.Fatal(err)
	}
	defer client.conn.Close()

	fmt.Println("🔗 Connected to Rock Paper Scissors server!")

	// Start listening for events in a goroutine
	go client.listenForEvents()

	// Show main menu
	client.showMainMenu()

	// Keep the main thread alive
	select {}
}
