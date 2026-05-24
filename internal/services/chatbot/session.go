package chatbot

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/google/uuid"
	architect "github.com/imama2/Genzite-Backend/gen/go/agents"
	"github.com/imama2/Genzite-Backend/internal/services/agents"
)

// Session represents a single user's conversational session.
type Session struct {
	ID           string
	UserID       string
	logger       *slog.Logger
	mu           sync.Mutex
	state        State
	context      *ConversationContext
	lastError    error
	sendToClient chan<- []byte // To send messages back to the user via websocket

	agentService    agents.AgentService
	architectStream architect.Architect_AnalyzePromptClient
}

// ConversationContext holds the history and artifacts of a build conversation.
type ConversationContext struct {
	BuildID            string
	UserPrompt         string
	ArchitectBlueprint string // Stores the evolving blueprint
	LastQuestionID     string // To track the last question asked by the architect
	GeneratedArtifacts []string
}

// State represents the current state of the conversation.
type State int

const (
	StateIdle State = iota
	StateArchitectConversation
	StateAwaitingUserInput
	StateFinalizingBlueprint
	StateCompleted
	StateFailed
)

func NewSession(userID string, logger *slog.Logger, sendChan chan<- []byte, agentService agents.AgentService) *Session {
	sessionID := uuid.New().String()
	return &Session{
		ID:           sessionID,
		UserID:       userID,
		logger:       logger.With("sessionID", sessionID, "userID", userID),
		state:        StateIdle,
		sendToClient: sendChan,
		agentService: agentService,
		context: &ConversationContext{
			BuildID: uuid.New().String(),
		},
	}
}

// IncomingMessage represents the structure of a message from the client.
type IncomingMessage struct {
	Type    string `json:"type"`
	Content string `json:"content"`
}

// ProcessMessage is the entry point for all incoming messages from the user.
// It acts as a state machine, routing the message based on the session's current state.
func (s *Session) ProcessMessage(rawMessage []byte) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var msg IncomingMessage
	if err := json.Unmarshal(rawMessage, &msg); err != nil {
		s.logger.Error("Failed to unmarshal incoming message", "error", err)
		// TODO: Send error back to client
		return err
	}

	s.logger.Info("Processing message in session", "state", s.state, "messageType", msg.Type)

	switch s.state {
	case StateIdle:
		if msg.Type == "start_prompt" {
			s.context.UserPrompt = msg.Content
			s.state = StateArchitectConversation
			go s.startArchitectConversation()
		} else {
			s.logger.Warn("Received unexpected message in Idle state", "messageType", msg.Type)
		}

	case StateAwaitingUserInput:
		if msg.Type == "user_response" {
			s.state = StateArchitectConversation
			// The response will be sent by the handleArchitectStream goroutine
			// For now, we just log it.
			s.logger.Info("User response received, passing to architect", "response", msg.Content)
			// This requires sending the response to the active stream, which will be handled
			// in the `handleArchitectStream` method.
		} else {
			s.logger.Warn("Received unexpected message in AwaitingUserInput state", "messageType", msg.Type)
		}

	default:
		s.logger.Warn("Message received in unhandled state", "state", s.state)
	}

	return nil
}

// startArchitectConversation kicks off the conversation with the Architect agent.
func (s *Session) startArchitectConversation() {
	s.logger.Info("Starting architect conversation")

	stream, err := s.agentService.AnalyzePrompt(context.Background())
	if err != nil {
		s.mu.Lock()
		s.state = StateFailed
		s.lastError = err
		s.mu.Unlock()
		s.logger.Error("Failed to start architect stream", "error", err)
		// TODO: Send error to client
		return
	}
	s.architectStream = stream

	// Send the initial prompt
	initialReq := &architect.AnalyzeRequest{
		Event: &architect.AnalyzeRequest_InitialPrompt_{
			InitialPrompt: &architect.AnalyzeRequest_InitialPrompt{
				BuildId:    s.context.BuildID,
				UserPrompt: s.context.UserPrompt,
			},
		},
	}
	if err := stream.Send(initialReq); err != nil {
		s.logger.Error("Failed to send initial prompt to architect", "error", err)
		// TODO: Handle error
		return
	}

	// Now, listen for responses from the architect in a separate goroutine
	// This will be implemented in the next step.
	// go s.handleArchitectStream()
}
