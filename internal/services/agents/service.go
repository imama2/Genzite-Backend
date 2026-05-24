package agents

// AgentService defines the interface for interacting with the Python agent services.
type AgentService interface {
	// TODO: Define methods for interacting with agents, e.g., the Architect agent via gRPC.
}

// agentServiceClient is the concrete implementation of AgentService.
type agentServiceClient struct {
	// TODO: Add gRPC client connections.
}

// NewService creates a new AgentService client.
func NewService() AgentService {
	return &agentServiceClient{}
}
