package agents

import (
	"context"

	architect "github.com/imama2/Genzite-Backend/gen/go/agents"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// AgentService defines the interface for interacting with the Python agent services.
type AgentService interface {
	AnalyzePrompt(ctx context.Context, opts ...grpc.CallOption) (architect.Architect_AnalyzePromptClient, error)
	// TODO: Add FinalizeBlueprint method
}

// agentServiceClient is the concrete implementation of AgentService.
type agentServiceClient struct {
	architectConn *grpc.ClientConn
}

// NewService creates a new AgentService client.
// For now, it connects to a hardcoded address. This should be configurable.
func NewService() AgentService {
	// TODO: Make the agent address configurable.
	conn, err := grpc.NewClient("localhost:50051", grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		// In a real app, you'd want to handle this more gracefully.
		// For now, we panic because the orchestrator can't function without it.
		panic(err)
	}
	return &agentServiceClient{
		architectConn: conn,
	}
}

func (c *agentServiceClient) AnalyzePrompt(ctx context.Context, opts ...grpc.CallOption) (architect.Architect_AnalyzePromptClient, error) {
	client := architect.NewArchitectClient(c.architectConn)
	return client.AnalyzePrompt(ctx, opts...)
}
