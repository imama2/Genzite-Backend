package chatbot

import (
	"log/slog"
	"sync"

	"github.com/google/uuid"
)

// Session represents a single user's conversational session.
// It manages the state, context, and interaction flow with the agent pipeline.
type Session struct {
	ID        string
	UserID    string
	logger    *slog.Logger
	mu        sync.Mutex
	state     State
	context   *ConversationContext
	lastError error
	// To send messages back to the user via websocket
	sendToClient chan<- []byte
}

// ConversationContext holds the history and artifacts of a build conversation.
type ConversationContext struct {
	BuildID            string
	UserPrompt         string
	ArchitectBlueprint string
	DesignTokens       string
	GeneratedArtifacts []string
}

// State represents the current state of the conversation.
type State int

const (
	Idle State = iota
	AwaitingUserInput
	ClarifyingWithArchitect
	RunningBuildPipeline
	Completed
	Failed
)

func NewSession(userID string, logger *slog.Logger, sendChan chan<- []byte) *Session {
	sessionID := uuid.New().String()
	return &Session{
		ID:     sessionID,
		UserID: userID,
		logger: logger.With("sessionID", sessionID, "userID", userID),
		state:  Idle,
		context: &ConversationContext{
			BuildID: uuid.New().String(),
		},
		sendToClient: sendChan,
	}
}

func (s *Session) ProcessMessage(message []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// TODO: Implement the state machine logic for processing user messages.
	// This will involve:
	// 1. Parsing the incoming message.
	// 2. Determining the current state (e.g., Idle, AwaitingUserInput).
	// 3. Calling the appropriate service (e.g., Architect agent via gRPC, enqueuing build job).
	// 4. Updating the session state and context.
	s.logger.Info("Processing message in session", "state", s.state, "message", string(message))

	// Placeholder: Echo the message back to the client.
	s.sendToClient <- []byte("Echo: " + string(message))

	return nil
}
