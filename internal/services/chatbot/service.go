package chatbot

import (
	"log/slog"
	"sync"

	"github.com/imama2/Genzite-Backend/internal/core/broker"
	"github.com/imama2/Genzite-Backend/internal/services/agents"
)

// ChatService manages all active chat sessions and orchestrates the agent-based workflows.
type ChatService struct {
	logger       *slog.Logger
	sessions     map[string]*Session // Map sessionID to Session
	mu           sync.RWMutex
	agentClient  agents.AgentService
	brokerClient broker.BrokerService
}

// NewService creates a new ChatService.
func NewService(logger *slog.Logger, agentClient agents.AgentService, brokerClient broker.BrokerService) *ChatService {
	return &ChatService{
		logger:       logger,
		sessions:     make(map[string]*Session),
		agentClient:  agentClient,
		brokerClient: brokerClient,
	}
}

// HandleNewConnection creates a new session for a new WebSocket connection.
func (s *ChatService) HandleNewConnection(userID string, sendChan chan<- []byte) *Session {
	s.mu.Lock()
	defer s.mu.Unlock()

	session := NewSession(userID, s.logger, sendChan, s.agentClient, s.brokerClient)
	s.sessions[session.ID] = session
	s.logger.Info("New chat session created", "sessionID", session.ID, "userID", userID)
	return session
}

// HandleDisconnect removes a session when a client disconnects.
func (s *ChatService) HandleDisconnect(sessionID string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.sessions, sessionID)
	s.logger.Info("Chat session closed", "sessionID", sessionID)
}
