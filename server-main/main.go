package main

import (
	"cmd/server"
	"log"
	"net/http"
)

func serverIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.Error(w, "404 not found.", http.StatusNotFound)
		return
	}

	if r.Method != "GET" {
		http.Error(w, "not found.", http.StatusNotFound)
		return
	}

}

func main() {
	//create new instance of hub by calling NewHub fxtion
	hub := server.NewHub()

	//new go routine
	go hub.Run()

	//line reg the serverIndex function as the handler for the root path
	http.HandleFunc("/", serverIndex)

	//registers a nil handler for the /ws path
	//for the websocket connection
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		//serve the websocket and initialize the hub & client
		server.ServeWs(hub, w, r)
	})

	//line starts http server on port 3000
	log.Fatal(http.ListenAndServe(":3000", nil))

	// p := tea.NewProgram(initialModel())
	// if _, err := p.Run(); err != nil {
	// 	fmt.Fprintf(os.Stderr, "Oof: %v\n", err)
	// }

}

// package main

// // A simple program demonstrating the text area component from the Bubbles
// // component library.

// //What was rendered in HTMX was .ClientID and .Text

// import (
// 	"encoding/json"
// 	"fmt"
// 	"os"
// 	"strings"
// 	"github.com/google/uuid"

// 	"github.com/gorilla/websocket"

// 	"cmd/server"

// 	"github.com/charmbracelet/bubbles/cursor"
// 	"github.com/charmbracelet/bubbles/textarea"
// 	"github.com/charmbracelet/bubbles/viewport"
// 	tea "github.com/charmbracelet/bubbletea"
// 	"github.com/charmbracelet/lipgloss"
// )

// // Transfer main.go here to the main.go
// func main() {
// 	p := tea.NewProgram(initialModel())
// 	if _, err := p.Run(); err != nil {
// 		fmt.Fprintf(os.Stderr, "Oof: %v\n", err)
// 	}
// }

// type model struct {
// 	viewport    viewport.Model
// 	messages    []string
// 	textarea    textarea.Model
// 	senderStyle lipgloss.Style
// 	err         error
// 	clt         server.Client
// 	mzg         server.Message
// 	connected   bool
// 	msgChan		chan string
// }

// // Build and return the initial model containing all the UI components and their initial states.
// func initialModel() model {
// 	ta := textarea.New()
// 	ta.Placeholder = "Send a message..."
// 	ta.Focus()

// 	//set visual prompt at start of inbox
// 	ta.Prompt = "┃ "
// 	ta.CharLimit = 280

// 	ta.SetWidth(30)
// 	ta.SetHeight(3)

// 	// Remove cursor line styling
// 	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

// 	ta.ShowLineNumbers = false

// 	vp := viewport.New(30, 5)
// 	vp.SetContent(`Welcome to the chat room!
// Type a message and press Enter to send.`)

// 	ta.KeyMap.InsertNewline.SetEnabled(false)

// 	//Establish a WS connection
// 	conn, _, err := websocket.DefaultDialer.Dial("ws://localhost:3000/ws", nil)
// 	if err != nil {
//         log.Printf("WebSocket connection error: %v", err)
//     }

// 	clientId := uuid.New().String()

// 	msgChan := make(chan string)

// 	return model{
// 		connected:   connected,
// 		textarea:    ta,
// 		messages:    []string{},
// 		viewport:    vp,
// 		senderStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("5")),
// 		err:         nil,
// 		clt: server.Client{
// 			Id:   clientId,
// 			Hub:  nil,
// 			Conn: conn,
// 			Send: make(chan []byte),
// 		},
// 		mzg: server.Message{
// 			ClientID: "",
// 			Text:     "",
// 		},
// 		//Initialize the lst 2 with default values
// 		msgChan: msgChan,
// 	}
// }

// func (m model) Init() tea.Cmd {
// 	return textarea.Blink
// }

// // I will pass the message to the update function
// // func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
// // 	switch msg := msg.(type) {
// // 	case tea.WindowSizeMsg:
// // 		m.viewport.Width = msg.Width
// // 		m.textarea.SetWidth(msg.Width)
// // 		return m, nil
// // 	case tea.KeyMsg:
// // 		switch msg.String() {
// // 		case "esc", "ctrl+c":
// // 			// Quit.
// // 			fmt.Println(m.textarea.Value())
// // 			return m, tea.Quit
// // 		case "enter":
// // 			v := m.textarea.Value()

// // 			if v == "" {
// // 				// Don't send empty messages.
// // 				return m, nil
// // 			}

// // 			// Simulate sending a message. In your application you'll want to
// // 			// also return a custom command to send the message off to
// // 			// a server.
// // 			m.messages = append(m.messages, m.senderStyle.Render("You: ")+v)
// // 			m.viewport.SetContent(strings.Join(m.messages, "\n"))
// // 			m.textarea.Reset()
// // 			m.viewport.GotoBottom()
// // 			return m, nil
// // 		default:
// // 			// Send all other keypresses to the textarea.
// // 			var cmd tea.Cmd
// // 			m.textarea, cmd = m.textarea.Update(msg)
// // 			return m, cmd
// // 		}

// // 	case cursor.BlinkMsg:
// // 		// Textarea should also process cursor blinks.
// // 		var cmd tea.Cmd
// // 		m.textarea, cmd = m.textarea.Update(msg)
// // 		return m, cmd

// // 	default:
// // 		return m, nil
// // 	}
// // }

// func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
// 	switch msg := msg.(type) {
// 	case tea.WindowSizeMsg:
// 		m.viewport.Width = msg.Width
// 		m.textarea.SetWidth(msg.Width)
// 		return m, nil

// 	case tea.KeyMsg:
// 		switch msg.String() {
// 		case "esc", "ctrl+c":
// 			// Quit.
// 			fmt.Println(m.textarea.Value())
// 			return m, tea.Quit

// 		case "enter":
// 			// Get the input from the textarea.
// 			v := m.textarea.Val ue()

// 			if v == "" {
// 				// Don't send empty messages.
// 				return m, nil
// 			}

// 			if !m.connected {
// 				m.messages = append(m.messages, "Not connected to server!")
// 				m.viewport.SetContent(strings.Join(m.messages, "\n"))
// 				m.textarea.Reset()
// 				m.viewport.GotoBottom()
// 				return m, nil
// 			}

// 			// Send the message over the WebSocket connection.
// 			if m.clt.Conn != nil {
// 				// Construct the message.
// 				message := server.Message{
// 					ClientID: m.clt.Id,
// 					Text:     v,
// 				}

// 				// Serialize the message into JSON.
// 				messageJSON, err := json.Marshal(message)
// 				if err != nil {
// 					m.err = fmt.Errorf("failed to serialize message: %w", err)
// 					return m, nil
// 				}

// 				// Send the serialized message.
// 				err = m.clt.Conn.WriteMessage(websocket.TextMessage, messageJSON)
// 				if err != nil {
// 					m.err = fmt.Errorf("failed to send message: %w", err)
// 					return m, nil
// 				}
// 			} else {
// 				m.err = fmt.Errorf("WebSocket connection not initialized")
// 			}

// 			// Update the local chat history.
// 			m.messages = append(m.messages, m.senderStyle.Render("You: ")+v)
// 			m.viewport.SetContent(strings.Join(m.messages, "\n"))
// 			m.textarea.Reset()
// 			m.viewport.GotoBottom()
// 			return m, nil

// 		default:
// 			// Send all other keypresses to the textarea.
// 			var cmd tea.Cmd
// 			m.textarea, cmd = m.textarea.Update(msg)
// 			return m, cmd
// 		}

// 	case cursor.BlinkMsg:
// 		// Textarea should also process cursor blinks.
// 		var cmd tea.Cmd
// 		m.textarea, cmd = m.textarea.Update(msg)
// 		return m, cmd

// 	default:
// 		return m, nil
// 	}
// }

// func (m model) View() string {
// 	return fmt.Sprintf(
// 		"%s\n\n%s",
// 		m.viewport.View(),
// 		m.textarea.View(),
// 	) + "\n\n"
// }

// func (m *model) cleanup() {
// 	if m.clt.Conn != nil {
// 		m.clt.Conn.Close()
// 		m.connected = false
// 	}
// }
