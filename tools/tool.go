package tools

// Tool interface defines the contract that all agent capabilities must follow.
// In Go, interfaces are implemented implicitly (duck typing). Any struct that
// defines these exact three methods automatically satisfies the Tool interface.
type Tool interface {
	Name() string
	Description() string
	// Execute takes the parsed LLM action data and content, and returns the result as a string.
	Execute(data, content string) string
}
