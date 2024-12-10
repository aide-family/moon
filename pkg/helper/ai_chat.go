package helper

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	nhttp "net/http"
	"strings"
	"time"

	"github.com/aide-family/moon/pkg/util/types"

	"github.com/go-kratos/kratos/v2/transport/http"
)

// Request is a struct representing an Ollama request.
type Request struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
	Stream   bool      `json:"stream"`
}

// Message is a struct representing an Ollama message.
type Message struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// Response is a struct representing an Ollama response.
type Response struct {
	CreatedAt          time.Time `json:"created_at"`
	Done               bool      `json:"done"`
	DoneReason         string    `json:"done_reason"`
	EvalCount          int       `json:"eval_count"`
	EvalDuration       int64     `json:"eval_duration"`
	LoadDuration       int       `json:"load_duration"`
	Message            *Message  `json:"message"`
	Model              string    `json:"model"`
	PromptEvalCount    int       `json:"prompt_eval_count"`
	PromptEvalDuration int       `json:"prompt_eval_duration"`
	TotalDuration      int64     `json:"total_duration"`
}

// NewOllama creates a new instance of Ollama with the given URL and options.
func NewOllama(url string, opts ...OllamaOption) *Ollama {
	o := &Ollama{
		Model: "jimscard/devopd:7b",
		URL:   url,
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// Ollama represents an Ollama instance.
type Ollama struct {
	Model string `json:"model"`
	URL   string `json:"url"`
}

// OllamaOption is a function that configures an Ollama instance.
type OllamaOption func(o *Ollama)

// WithOllamaModel sets the model for the Ollama instance.
func WithOllamaModel(model string) OllamaOption {
	return func(o *Ollama) {
		o.Model = model
	}
}

// HandleChat returns an HTTP handler function that handles chat requests.
func (o *Ollama) HandleChat() http.HandlerFunc {
	return func(ctx http.Context) error {
		o.handleChat(ctx.Response(), ctx.Request())
		return nil
	}
}

func (o *Ollama) handleChat(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	flusher, ok := w.(nhttp.Flusher)
	if !ok {
		nhttp.Error(w, "Streaming unsupported!", nhttp.StatusInternalServerError)
		return
	}

	msg := Message{
		Role:    "user",
		Content: r.URL.Query().Get("message") + ", 用中文回复我！",
	}
	if msg.Content == "" {
		fmt.Fprintf(w, "data: [DONE]\n\n")
		flusher.Flush()
		return
	}

	req := Request{
		Model:    o.Model,
		Stream:   true,
		Messages: []Message{msg},
	}

	if err := streamFromOllama(context.Background(), o.URL, req, w, flusher); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}

func streamFromOllama(ctx context.Context, url string, ollamaReq Request, w http.ResponseWriter, flusher http.Flusher) error {
	js, err := types.Marshal(&ollamaReq)
	if err != nil {
		return fmt.Errorf("marshal error: %w", err)
	}

	client := nhttp.Client{}
	httpReq, err := nhttp.NewRequestWithContext(ctx, nhttp.MethodPost, url, bytes.NewReader(js))
	if err != nil {
		return fmt.Errorf("create request error: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	fmt.Printf("Sending request to Ollama: %s\n", string(js))
	httpResp, err := client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("do request error: %w", err)
	}
	defer httpResp.Body.Close()

	if httpResp.StatusCode != nhttp.StatusOK {
		return fmt.Errorf("unexpected status code: %d", httpResp.StatusCode)
	}

	reader := bufio.NewReader(httpResp.Body)
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Printf("Read error: %v\n", err)
			break
		}

		if len(strings.TrimSpace(line)) == 0 {
			continue
		}

		fmt.Printf("Received line: %s\n", line)
		var response Response
		if err := types.Unmarshal([]byte(line), &response); err != nil {
			fmt.Printf("Unmarshal error: %v\n", err)
			continue
		}

		safeMessage := strings.ReplaceAll(response.Message.Content, "\n", "\\n")
		if response.Message.Content != "" {
			fmt.Fprintf(w, "data: %s\n\n", safeMessage)
			flusher.Flush()
		}

		if response.Done {
			break
		}
	}
	return nil
}
