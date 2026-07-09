package agent

import (
	"agent-orchestrator/llm"
	"agent-orchestrator/tools"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

const baseSystemPrompt = `You are an autonomous system execution agent. You interact with the underlying OS and file system through strictly defined
  tools.

    CRITICAL EXECUTION RULES:
    1. OUTPUT FORMAT: You MUST respond ONLY with ONE SINGLE raw, valid JSON object.
    2. STEP-BY-STEP EXECUTION: You can only use ONE tool at a time. If the user asks for multiple tasks, do them sequentially. Execute the first step,
  WAIT for the SYSTEM DATA result, and then execute the second step in your next response. DO NOT combine multiple actions into one message.
    3. JSON ESCAPING: You must properly escape all special characters. Backslashes (\) in file paths and inner double quotes (") MUST be escaped.
    4. SECURITY CLEARANCE: You are strictly FORBIDDEN from executing destructive OS commands.
    5. DETERMINISM: Base your actions strictly on the provided SYSTEM DATA.

    AVAILABLE TOOLS:`

type AgentResponse struct {
	Thinking string `json:"thinking"`
	Action   string `json:"action"`
	Data     string `json:"data"`
	Content  string `json:"content"`
}

type Agent struct {
	ID            string
	Role          string
	Status        string
	Memory        []string
	LanguageModel llm.LLMProvider
	Tools         map[string]tools.Tool
}

const maxMemorySize = 10

// AddTool registers a new tool to the agent's Tools map.
// If the map is uninitialized, it allocates it using make().
func (a *Agent) AddTool(t tools.Tool) {
	if a.Tools == nil {
		a.Tools = make(map[string]tools.Tool)
	}
	a.Tools[t.Name()] = t
}

// buildSystemPrompt dynamically constructs the system prompt based on the agent's role and registered tools.
func (a *Agent) buildSystemPrompt() string {
	// strings.Builder is used here to avoid multiple costly string concatenations and allocations.
	var sb strings.Builder
	sb.WriteString(baseSystemPrompt + "\n\n")
	sb.WriteString("YOUR SPECIFIC ROLE: ")
	sb.WriteString(a.Role)
	sb.WriteString("\n\n")

	sb.WriteString(`TOOL: reply_to_user (To answer directly)
    {
        "thinking": "I have the final answer.",
        "action": "reply_to_user",
        "data": "Your direct response to the user.",
        "content": ""
    }` + "\n\n")

	for _, t := range a.Tools {
		sb.WriteString(t.Description())
		sb.WriteString("\n\n")
	}
	return sb.String()
}

// Think sends the current prompt and conversation history to the LLM, then parses the JSON response.
func (a *Agent) Think(ctx context.Context, prompt string) (*AgentResponse, error) {
	// Append the new prompt to memory. This may cause an underlying array reallocation if capacity is reached.
	a.Memory = append(a.Memory, "USER: "+prompt)

	// Keep memory within maxMemorySize to avoid unbounded growth and context window limits.
	if len(a.Memory) > maxMemorySize {
		// Slice re-slicing: creates a new slice header pointing to the same backing array. Very cheap (O(1)).
		a.Memory = a.Memory[len(a.Memory)-maxMemorySize:]
	}

	history := strings.Join(a.Memory, "\n")

	fullContext := a.buildSystemPrompt() + "\n\nCONVERSATION HISTORY:\n" + history

	rawJSON, err := a.LanguageModel.Generate(ctx, fullContext)
	if err != nil {
		return nil, err
	}

	// Open log file in append mode. This avoids loading the whole file into memory.
	f, errLog := os.OpenFile("agent_debug.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if errLog == nil {
		f.WriteString("=== NEW LLM ANSWER ===\n")
		f.WriteString(rawJSON + "\n")
		f.WriteString("=======================\n\n")
		f.Close()
	}

	cleanJSON := strings.TrimSpace(rawJSON)
	cleanJSON = strings.TrimPrefix(cleanJSON, "```json")
	cleanJSON = strings.TrimPrefix(cleanJSON, "```")
	cleanJSON = strings.TrimSuffix(cleanJSON, "```")
	cleanJSON = strings.TrimSpace(cleanJSON)

	var response AgentResponse
	// json.NewDecoder uses a streaming approach via strings.NewReader, which is memory efficient.
	decoder := json.NewDecoder(strings.NewReader(cleanJSON))
	err = decoder.Decode(&response)
	if err != nil {
		return nil, err
	}

	a.Memory = append(a.Memory, "ASSISTANT: "+response.Data)
	return &response, nil
}

// RunTask is a synchronous loop that continuously asks the LLM to 'Think' and executes the requested actions.
func (a *Agent) RunTask(ctx context.Context, task string) (string, error) {
	currentPrompt := task

	// Infinite loop. Will only break when the LLM returns "reply_to_user" or an error occurs.
	for {
		response, err := a.Think(ctx, currentPrompt)
		if err != nil {
			return "", err
		}

		if response.Action == "reply_to_user" {
			return response.Data, nil
		}

		if tool, exists := a.Tools[response.Action]; exists {
			fmt.Printf("[%s] Executing tool: %s...\n", a.ID, response.Action)
			currentPrompt = tool.Execute(response.Data, response.Content)
		} else {
			fmt.Printf("[%s] ERROR: Unknown tool '%s'\n", a.ID, response.Action)
			currentPrompt = fmt.Sprintf("SYSTEM ERROR: Action '%s' does not exist.", response.Action)
		}
	}
}
