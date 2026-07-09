package main

import (
	"agent-orchestrator/agent"
	"agent-orchestrator/llm"
	"agent-orchestrator/tools"
	"bufio"
	"context"
	"fmt"
	"os"
	"strings"
	"time"
)

func main() {
	// Read the API key from environment variables.
	apiKey := os.Getenv("GROQ_API_KEY")
	if apiKey == "" {
		fmt.Println("FATAL ERROR: GROQ_API_KEY environment variable not found.")
		return
	}

	// Initialize the LLM client (Groq). Allocates a new GroqClient struct on the heap.
	client := &llm.GroqClient{
		APIKey: apiKey,
		Model:  "openai/gpt-oss-120b",
	}

	// Create the orchestrator agent which will act as the Manager.
	orchestrator := agent.NewOrchestrator(client)

	// Set up the Coder sub-agent and attach its specific tools.
	coderAgent := &agent.Agent{
		ID:            "Coder-001",
		Role:          "Software engineer",
		LanguageModel: client,
	}

	researcherAgent := &agent.Agent{
		ID:            "Researcher-001",
		Role:          "Internet and Documentation Researcher. Your job is to fetch content from URLs using fetch_url tool, extract the core information, and summarize it accurately.",
		LanguageModel: client,
	}

	researcherAgent.AddTool(&tools.FetchURLTool{})

	coderAgent.AddTool(&tools.RunCommandTool{})
	coderAgent.AddTool(&tools.ReadFileTool{})
	coderAgent.AddTool(&tools.WriteFileTool{})

	// Register the Coder agent to the Orchestrator and give the Manager a tool to delegate tasks to it.
	orchestrator.RegisterSubAgent("coder", coderAgent)
	orchestrator.RegisterSubAgent("researcher", researcherAgent)

	orchestrator.ManagerAgent.AddTool(&agent.DelegateTool{
		TargetAgentName: "coder",
		TargetAgent:     coderAgent,
	})

	orchestrator.ManagerAgent.AddTool(&agent.DelegateTool{
		TargetAgentName: "researcher",
		TargetAgent:     researcherAgent,
	})

	fmt.Println("Multi-Agent System Initialized. Type 'exit' or 'quit' to shut down.")
	fmt.Println("---------------------------------------------------------")

	scanner := bufio.NewScanner(os.Stdin)

	for {
		fmt.Print("\nYou: ")

		if !scanner.Scan() {
			break
		}

		userInput := strings.TrimSpace(scanner.Text())

		if userInput == "" {
			continue
		}

		if strings.ToLower(userInput) == "exit" || strings.ToLower(userInput) == "quit" {
			fmt.Println("Shutting down the system. Goodbye!")
			break
		}

		// Create a context with a 3-minute timeout for the task to prevent indefinite hanging.
		ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)

		fmt.Printf("[SYSTEM] Orchestrator is analyzing your request...\n")

		// Run the task through the orchestrator.
		finalResponse, err := orchestrator.ManagerAgent.RunTask(ctx, userInput)
		if err != nil {
			fmt.Println("System Error:", err)
		} else {
			fmt.Printf("\nManager: %s\n", finalResponse)
		}
		cancel() // Release resources associated with the context.
	}

	if err := scanner.Err(); err != nil {
		fmt.Println("A fatal I/O error occurred:", err)
	}
}
