package chatbot

import (
	"context"
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/google/uuid"
	architect "github.com/imama2/Genzite-Backend/gen/go/agents"
	"github.com/imama2/Genzite-Backend/internal/core/broker"
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
	brokerService   broker.BrokerService
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

func NewSession(userID string, logger *slog.Logger, sendChan chan<- []byte, agentService agents.AgentService, broker broker.BrokerService) *Session {
	sessionID := uuid.New().String()
	return &Session{
		ID:            sessionID,
		UserID:        userID,
		logger:        logger.With("sessionID", sessionID, "userID", userID),
		state:         StateIdle,
		sendToClient:  sendChan,
		agentService:  agentService,
		brokerService: broker,
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

// OutgoingMessage represents a message sent from the server to the client.
type OutgoingMessage struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
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
			s.logger.Info("User response received, passing to architect", "response", msg.Content)

			// Send the user's response to the architect agent
			req := &architect.AnalyzeRequest{
				Event: &architect.AnalyzeRequest_UserResponse_{
					UserResponse: &architect.AnalyzeRequest_UserResponse{
						QuestionId: s.context.LastQuestionID,
						Answer:     msg.Content,
					},
				},
			}
			if err := s.architectStream.Send(req); err != nil {
				s.logger.Error("Failed to send user response to architect", "error", err)
				// TODO: Handle error, maybe set state to Failed
			}
		} else if msg.Type == "finalize_blueprint" {
			s.state = StateFinalizingBlueprint
			go s.finalizeBlueprint()
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
	go s.handleArchitectStream()
}

// handleArchitectStream runs in a goroutine to listen for messages from the agent.
func (s *Session) handleArchitectStream() {
	for {
		resp, err := s.architectStream.Recv()
		if err != nil {
			// TODO: Handle different kinds of errors, e.g., io.EOF means stream closed cleanly.
			s.mu.Lock()
			s.state = StateFailed
			s.lastError = err
			s.mu.Unlock()
			s.logger.Error("Failed to receive from architect stream", "error", err)
			return
		}

		s.mu.Lock()
		switch event := resp.Event.(type) {
		case *architect.AnalyzeResponse_Question:
			s.state = StateAwaitingUserInput
			s.context.LastQuestionID = event.Question.QuestionId
			s.logger.Info("Received question from architect", "questionId", event.Question.QuestionId)
			s.sendMessageToClient("architect_question", event.Question)

		case *architect.AnalyzeResponse_Update:
			s.context.ArchitectBlueprint = event.Update.BlueprintChunkJson // Or append, depending on strategy
			s.logger.Info("Received blueprint update from architect")
			s.sendMessageToClient("blueprint_update", event.Update)

		case *architect.AnalyzeResponse_ConversationComplete_:
			s.logger.Info("Architect conversation complete. Ready to finalize.")
			s.sendMessageToClient("conversation_complete", nil)
			// The stream will be closed by the server, which we'll detect in Recv().
			// We can now allow the user to trigger finalization.
		}
		s.mu.Unlock()
	}
}

func (s *Session) finalizeBlueprint() {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.logger.Info("Finalizing blueprint", "buildId", s.context.BuildID)

	req := &architect.FinalizeBlueprintRequest{
		BuildId: s.context.BuildID,
	}

	resp, err := s.agentService.FinalizeBlueprint(context.Background(), req)
	if err != nil {
		s.state = StateFailed
		s.lastError = err
		s.logger.Error("Failed to finalize blueprint", "error", err)
		s.sendMessageToClient("finalize_failed", err.Error())
		return
	}

	s.state = StateCompleted
	s.context.ArchitectBlueprint = resp.BlueprintJson
	s.logger.Info("Blueprint finalized successfully")
	s.sendMessageToClient("finalize_success", resp)

	// Enqueue the build job
	s.enqueueBuildJob(resp.BlueprintJson)
}

func (s *Session) enqueueBuildJob(blueprintJSON string) {
	if s.brokerService == nil {
		s.logger.Error("Broker service is not initialized, cannot enqueue build job")
		return
	}

	jobPayload := broker.BuildJob{
		BuildID:       s.context.BuildID,
		UserID:        s.UserID,
		BlueprintJSON: blueprintJSON,
	}

	payloadBytes, err := json.Marshal(jobPayload)
	if err != nil {
		s.logger.Error("Failed to marshal build job payload", "error", err, "buildID", s.context.BuildID)
		// Optionally, notify the client of this internal error
		s.sendMessageToClient("enqueue_failed", "internal error: could not create build job")
		return
	}

	err = s.brokerService.Enqueue(context.Background(), broker.QueueBuild, payloadBytes)
	if err != nil {
		s.logger.Error("Failed to enqueue build job", "error", err, "buildID", s.context.BuildID)
		s.sendMessageToClient("enqueue_failed", "internal error: could not submit build job")
		return
	}

	s.logger.Info("Successfully enqueued build job", "buildID", s.context.BuildID, "queue", broker.QueueBuild)
	s.sendMessageToClient("build_enqueued", map[string]string{"buildId": s.context.BuildID})
}

// sendMessageToClient is a helper to marshal and send messages to the WebSocket client.
func (s *Session) sendMessageToClient(msgType string, payload interface{}) {
	msg := OutgoingMessage{
		Type:    msgType,
		Payload: payload,
	}
	jsonMsg, err := json.Marshal(msg)
	if err != nil {
		s.logger.Error("Failed to marshal outgoing message", "error", err)
		return
	}
	s.sendToClient <- jsonMsg
}
