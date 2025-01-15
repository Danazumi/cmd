package main

// A simple program demonstrating the text area component from the Bubbles
// component library.

//What was rendered in HTMX was .ClientID and .Text

import (
	"cmd/server"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/gorilla/websocket"

	"github.com/charmbracelet/bubbles/cursor"
	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

// Transfer main.go here to the main.go
func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "Oof: %v\n", err)
	}
}

type model struct {
	viewport    viewport.Model
	messages    []string
	textarea    textarea.Model
	senderStyle lipgloss.Style
	err         error
	clt         server.Client
	mzg         server.Message
	connected   bool
	msgChan     chan string
}

// Build and return the initial model containing all the UI components and their initial states.
func initialModel() model {
	ta := textarea.New()
	ta.Placeholder = "Send a message..."
	ta.Focus()

	//set visual prompt at start of inbox
	ta.Prompt = "┃ "
	ta.CharLimit = 280

	ta.SetWidth(30)
	ta.SetHeight(3)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(30, 5)
	vp.SetContent(`Welcome to the chat room!
Type a message and press Enter to send.`)

	ta.KeyMap.InsertNewline.SetEnabled(false)

	//Establish a WS connection
	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:3000/ws", nil)
	connected := err == nil

	clientId := uuid.New().String()

	msgChan := make(chan string)

	m := model{
		connected:   connected,
		textarea:    ta,
		messages:    []string{},
		viewport:    vp,
		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
		err:         nil,
		clt: server.Client{
			Id:   clientId,
			Hub:  nil,
			Conn: conn,
			Send: make(chan []byte),
		},
		// mzg: server.Message{
		// 	ClientID: "",
		// 	Text:     "",
		// },
		//Initialize the lst 2 with default values
		msgChan: msgChan,
	}

	go func() {
		for {
			_, message, err := conn.ReadMessage()
			if err != nil {
				log.Printf("read error: %v", err)
				return
			}

			var msg server.Message
			if err := json.Unmarshal(message, &msg); err != nil {
				log.Printf("unmarshal error: %v", err)
				continue
			}

			// Add message to view

			displayId := "You"
			if msg.ClientID != m.clt.Id {
				displayId = msg.ClientID[:4]
			}
			m.messages = append(m.messages, fmt.Sprintf("%s: %s", displayId, msg.Text))
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.viewport.GotoBottom()
		}
	}()

	return m

}

func (m model) Init() tea.Cmd {
	return textarea.Blink
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.viewport.Width = msg.Width
		m.textarea.SetWidth(msg.Width)
		return m, nil

	case tea.KeyMsg:
		switch msg.String() {
		case "esc", "ctrl+c":
			// Quit.
			fmt.Println(m.textarea.Value())
			return m, tea.Quit

		case "enter":
			// Get the input from the textarea.
			v := m.textarea.Value()

			if v == "" {
				// Don't send empty messages.
				return m, nil
			}

			if !m.connected {
				m.messages = append(m.messages, "Not connected to server!")
				m.viewport.SetContent(strings.Join(m.messages, "\n"))
				m.textarea.Reset()
				m.viewport.GotoBottom()
				return m, nil
			}

			// Send the message over the WebSocket connection.
			if m.clt.Conn != nil {
				// Construct the message.
				message := server.Message{
					ClientID: m.clt.Id,
					Text:     v,
				}

				// Serialize the message into JSON.
				messageJSON, err := json.Marshal(message)
				if err != nil {
					m.err = fmt.Errorf("failed to serialize message: %w", err)
					return m, nil
				}

				// Send the serialized message.
				err = m.clt.Conn.WriteMessage(websocket.TextMessage, messageJSON)
				if err != nil {
					m.err = fmt.Errorf("failed to send message: %w", err)
					return m, nil
				}
			} else {
				m.err = fmt.Errorf("WebSocket connection not initialized")
			}

			// Update the local chat history.
			// m.messages = append(m.messages, m.senderStyle.Render("You: ")+v)
			shortId := m.clt.Id[:4] // Get first 4 characters of UUID
			m.messages = append(m.messages, fmt.Sprintf("%s: %s", shortId, v))
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.textarea.Reset()
			m.viewport.GotoBottom()
			return m, nil

		default:
			// Send all other keypresses to the textarea.
			var cmd tea.Cmd
			m.textarea, cmd = m.textarea.Update(msg)
			return m, cmd
		}

	case cursor.BlinkMsg:
		// Textarea should also process cursor blinks.
		var cmd tea.Cmd
		m.textarea, cmd = m.textarea.Update(msg)
		return m, cmd

	default:
		return m, nil
	}
}

func (m model) View() string {
	return fmt.Sprintf(
		"%s\n\n%s",
		m.viewport.View(),
		m.textarea.View(),
	) + "\n\n"
}

func (m *model) cleanup() {
	if m.clt.Conn != nil {
		m.clt.Conn.Close()
		m.connected = false
	}
}
