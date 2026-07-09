package agent

import (
	"agent-orchestrator/llm"
	"context"
	"fmt"
)

// Orchestrator manages the primary ManagerAgent and holds references to all available SubAgents.
type Orchestrator struct {
	ManagerAgent *Agent
	SubAgents    map[string]*Agent
}

// NewOrchestrator initializes the multi-agent system.
// It creates the ManagerAgent on the heap and initializes the SubAgents map.
func NewOrchestrator(llmClient llm.LLMProvider) *Orchestrator {
	manager := &Agent{
		ID:            "Manager-001",
		Role:          "You are the Chief Orchestrator. Your ONLY job is to listen to the user and delegate technical tasks to your sub-agents using the ask_* tools. YOU CANNOT WRITE CODE OR RUN COMMANDS YOURSELF. If the user asks for files, folders, or code, you MUST use the ask_coder tool.",
		LanguageModel: llmClient,
	}

	return &Orchestrator{
		ManagerAgent: manager,
		SubAgents:    make(map[string]*Agent),
	}
}

// RegisterSubAgent adds a new agent to the Orchestrator's internal map.
// Maps in Go are reference types; this operation modifies the underlying hash table.
func (o *Orchestrator) RegisterSubAgent(name string, agent *Agent) {
	o.SubAgents[name] = agent
	fmt.Printf("[SYSTEM] The new agent has been added to the system: %s (Rol: %s)\n", name, agent.Role)
}

// DelegateTool is a dynamic tool that allows the Manager to pass tasks to a specific SubAgent.
type DelegateTool struct {
	TargetAgentName string
	TargetAgent     *Agent // Pointer to avoid copying the entire struct (Escape Analysis applies here).
}

func (t *DelegateTool) Name() string {
	return "ask_" + t.TargetAgentName
}

func (t *DelegateTool) Description() string {
	return fmt.Sprintf(`TOOL: ask_%s (To assign a task to the %s)
    {
        "thinking": "I need to assign this specific task to the %s.",
        "action": "ask_%s",
        "data": "Detailed task description for the %s.",
        "content": ""
    }`, t.TargetAgentName, t.TargetAgentName, t.TargetAgentName, t.TargetAgentName, t.TargetAgentName)
}

// Execute implements the tool.Tool interface. It blocks and runs the target agent's task synchronously.
func (t *DelegateTool) Execute(data, content string) string {
	fmt.Printf("[ORCHESTRATOR] Delegating task to %s...\n", t.TargetAgentName)

	// Note: Using context.Background() creates a fresh, empty context without a timeout.
	// This means the sub-agent could potentially run forever if not handled carefully.
	ctx := context.Background()

	result, err := t.TargetAgent.RunTask(ctx, data)
	if err != nil {
		return fmt.Sprintf("SYSTEM ERROR (From %s): %v", t.TargetAgentName, err)
	}

	return fmt.Sprintf("SYSTEM DATA (Result from %s):\n%s", t.TargetAgentName, result)
}
