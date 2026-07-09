package llm

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
)

type OllamaClient struct {
	BaseURL string
	Model   string
}

type ollamaRequest struct {
	Model  string `json:"model"`
	Prompt string `json:"prompt"`
	Stream bool   `json:"stream"`
}

type ollamaResponse struct {
	Response string `json:"response"`
	Done     bool   `json:"done"`
}

func (o *OllamaClient) Generate(prompt string) (string, error) {
	req := ollamaRequest{
		Model:  o.Model,
		Prompt: prompt,
		Stream: false,
	}

	jsonData, err := json.Marshal(req)

	if err != nil {
		return "", err
	}

	fullURL := o.BaseURL + "/api/generate"

	resp, err := http.Post(fullURL, "application/json", bytes.NewBuffer(jsonData))

	if err != nil {
		return "", err
	}

	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)

	if err != nil {
		return "", err
	}

	var ollamaResp ollamaResponse

	err = json.Unmarshal(bodyBytes, &ollamaResp)

	if err != nil {
		return "", err
	}
	return ollamaResp.Response, nil
}
