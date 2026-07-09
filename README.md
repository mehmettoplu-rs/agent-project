# Go Multi-Agent System (v1.0)

This is an autonomous multi-agent system written in Go, featuring a Chief Orchestrator and specialized sub-agents. It leverages the Groq API (LLaMA 3.1) to execute tasks sequentially by delegating work and using dynamic tools.

## Architecture & Go Runtime Mechanics

* **Chief Orchestrator:** Acts as the Manager agent. It does not write code or run commands directly but delegates tasks using the `DelegateTool`.
* **Sub-Agents:** Specialized agents (e.g., Coder) that possess specific tools (Run Command, Read File, Write File) to perform actual system operations.
* **Memory Management:** Designed with Go's runtime in mind. Slice re-slicing (`O(1)` operations), `strings.Builder` for minimizing allocations, and streaming JSON decoders (`json.NewDecoder`) are utilized to keep Garbage Collector (GC) pressure to an absolute minimum.
* **Concurrency Safety:** Contexts with timeouts and `defer cancel()` mechanisms prevent goroutine and memory leaks during execution.
* **I/O Safety:** Output streams from terminal commands are strictly capped using `io.LimitReader` to prevent Out-Of-Memory (OOM) crashes.

## Getting Started

1. Clone the repository.
2. Set your Groq API key:
   `export GROQ_API_KEY="your-api-key"` (Linux/Mac)
   `set GROQ_API_KEY="your-api-key"` (Windows PowerShell: `$env:GROQ_API_KEY="your-api-key"`)
3. Run the system:
   `go run main.go`
