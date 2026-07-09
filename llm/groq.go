package llm

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type GroqClient struct {
	APIKey string
	Model  string
}

type GroqRequest struct {
	Model    string        `json:"model"`
	Messages []GroqMessage `json:"messages"`
}

type GroqMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type GroqResponse struct {
	Choices []struct {
		Message GroqMessage `json:"message"`
	} `json:"choices"`
}

func (g *GroqClient) Generate(ctx context.Context, prompt string) (string, error) {
	req := GroqRequest{
		Model: g.Model,
		Messages: []GroqMessage{{
			Role:    "user",
			Content: prompt,
		}},
	}

	// json.Marshal converts the Go struct into a JSON byte slice, allocating new memory.
	jsonData, err := json.Marshal(req)

	if err != nil {
		return "", err
	}

	// bytes.NewBuffer wraps the byte slice to implement the io.Reader interface required by http.NewRequest.
	httpReq, err := http.NewRequestWithContext(ctx, "POST", "https://api.groq.com/openai/v1/chat/completions", bytes.NewBuffer(jsonData))

	if err != nil {
		return "", err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+g.APIKey)

	// Note: Creating a new http.Client for every request is inefficient.
	// In a high-throughput system, it's better to reuse a single *http.Client instance
	// to take advantage of HTTP keep-alive and connection pooling.
	client := &http.Client{}
	resp, err := client.Do(httpReq)

	if err != nil {
		return "", err
	}

	// Ensure the response body is always closed to prevent connection/file descriptor leaks.
	defer resp.Body.Close()

	// io.ReadAll reads the entire HTTP response body into memory at once.
	// We are doing this here instead of using json.NewDecoder directly on resp.Body.
	bodyBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	var groqResp GroqResponse

	err = json.Unmarshal(bodyBytes, &groqResp)

	if err != nil {
		return "", err
	}

	if len(groqResp.Choices) > 0 {
		return groqResp.Choices[0].Message.Content, nil
	}

	return "", fmt.Errorf("groq boş cevap döndürdü")
}
