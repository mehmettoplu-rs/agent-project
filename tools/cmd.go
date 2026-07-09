package tools

import (
	"context"
	"io"
	"os/exec"
	"runtime"
	"time"
)

// RunCommandTool implements the Tool interface to allow agents to execute safe OS commands.
type RunCommandTool struct{}

func (t *RunCommandTool) Name() string {
	return "run_command"
}

func (t *RunCommandTool) Description() string {
	return `TOOL: run_command (To execute safe terminal commands)
    {
        "thinking": "I need to run this command to verify my code or get environment info.",
        "action": "run_command",
        "data": "python test.py",
        "content": ""
    }`
}

func (t *RunCommandTool) Execute(data, content string) string {
	output := runCommand(data)
	return "SYSTEM DATA (Terminal Output):\n" + output + "\nAnalyze this output. If it's an error, fix your code/action. Otherwise, reply to the user."
}

func runCommand(command string) string {
	// Create a context with a 10-second timeout to prevent rogue commands from hanging indefinitely.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	// defer ensures cancel() is called when the function exits, preventing goroutine leaks.
	defer cancel()

	var cmd *exec.Cmd

	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(ctx, "cmd", "/c", command)
	} else {
		cmd = exec.CommandContext(ctx, "sh", "-c", command)
	}

	// Prevent unbounded memory allocation by capping the output read size (10 KB max).
	const maxOutputSize = 10 * 1024

	// Redirect stderr to stdout so we capture all error messages in the same stream.
	cmd.Stderr = cmd.Stdout

	stdoutPipe, err := cmd.StdoutPipe()
	if err != nil {
		return "ERROR: Unable to create a pipe - " + err.Error()
	}

	if err := cmd.Start(); err != nil {
		return "ERROR: Komut başlatilamadi - " + err.Error()
	}

	// io.LimitReader acts as a safety valve, stopping reads once maxOutputSize is reached.
	limitedReader := io.LimitReader(stdoutPipe, maxOutputSize)

	// io.ReadAll reads from the limited reader into a byte slice, allocating memory strictly up to maxOutputSize.
	outBytes, _ := io.ReadAll(limitedReader)

	err = cmd.Wait()

	result := string(outBytes)

	if err != nil {
		return "ERROR/TIMEOUT:\n" + result + "\n" + err.Error()
	}

	if len(result) == 0 {
		return "SUCCESS: Command executed with no output."
	}

	return "SUCCESS:\n" + result
}
