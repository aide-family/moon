package helper

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	nhttp "net/http"
	"strings"
	"time"

	strategyapi "github.com/aide-family/moon/api/admin/strategy"
	"github.com/aide-family/moon/pkg/util/safety"
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

// OllamaResponse is a struct representing an Ollama response.
type OllamaResponse struct {
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

// OpenAIAPIResponseByStream is a struct representing an OpenAI API response.
type OpenAIAPIResponseByStream struct {
	ID                string `json:"id"`
	Object            string `json:"object"`
	Created           int    `json:"created"`
	Model             string `json:"model"`
	SystemFingerprint string `json:"system_fingerprint"`
	Choices           []struct {
		Index int `json:"index"`
		Delta struct {
			Content string `json:"content"`
		} `json:"delta"`
		Logprobs     interface{} `json:"logprobs"`
		FinishReason interface{} `json:"finish_reason"`
	} `json:"choices"`
}

type OpenAIAPIResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int    `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index   int `json:"index"`
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
}

// String OpenAIAPIResponse to json
func (o *OpenAIAPIResponse) String() string {
	b, _ := types.Marshal(o)
	return string(b)
}

// NewOllama creates a new instance of Ollama with the given URL and options.
func NewOllama(url string, opts ...OllamaOption) *Ollama {
	o := &Ollama{
		Model: "gpt-4o-mini",
		URL:   url,
		Auth:  "",
		// 上下文容量
		contextSize: 10,
	}
	for _, opt := range opts {
		opt(o)
	}
	return o
}

// Ollama represents an Ollama instance.
type Ollama struct {
	Model   string `json:"model"`
	URL     string `json:"url"`
	Auth    string `json:"auth"`
	Type    string `json:"type"`
	context *safety.Map[string, []Message]
	// 上下文容量
	contextSize uint32
}

// OllamaOption is a function that configures an Ollama instance.
type OllamaOption func(o *Ollama)

// WithOllamaModel sets the model for the Ollama instance.
func WithOllamaModel(model string) OllamaOption {
	return func(o *Ollama) {
		o.Model = model
	}
}

// WithOllamaAuth sets the auth for the Ollama instance.
func WithOllamaAuth(auth string) OllamaOption {
	return func(o *Ollama) {
		o.Auth = auth
	}
}

// WithOllamaType sets the type for the Ollama instance.
func WithOllamaType(t string) OllamaOption {
	return func(o *Ollama) {
		o.Type = strings.ToLower(t)
	}
}

// WithOllamaContextSize sets the context size for the Ollama instance.
func WithOllamaContextSize(size uint32) OllamaOption {
	return func(o *Ollama) {
		o.contextSize = size
	}
}

// HandleChat returns an HTTP handler function that handles chat requests.
func (o *Ollama) HandleChat() http.HandlerFunc {
	o.context = safety.NewMap[string, []Message]()
	return func(ctx http.Context) error {
		o.handleChat(ctx.Response(), ctx.Request())
		return nil
	}
}

// HandlePushContext is a handler that pushes context to the Ollama instance.
func (o *Ollama) HandlePushContext() http.HandlerFunc {
	return func(ctx http.Context) error {
		return o.pushContext(ctx)
	}
}

func (o *Ollama) pushContext(ctx http.Context) error {
	token := ctx.Request().URL.Query().Get("token")
	if token == "" {
		return nil
	}

	body, err := io.ReadAll(ctx.Request().Body)
	if err != nil {
		return fmt.Errorf("read body error: %w", err)
	}

	var message Message
	if err := types.Unmarshal(body, &message); err != nil {
		return fmt.Errorf("unmarshal error: %w", err)
	}

	if ch, ok := o.context.Get(token); ok {
		// 判断上下文容量， 大于10之后，清空
		if len(ch) > int(o.contextSize) {
			ch = make([]Message, 0)
		}
		ch = append(ch, message)
		o.context.Set(token, ch)
	} else {
		ch = make([]Message, 0)
		ch = append(ch, message)
		o.context.Set(token, ch)
	}

	return nil
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

	token := r.URL.Query().Get("token")
	if token == "" {
		return
	}

	messages, ok := o.context.Get(token)
	if !ok {
		messages = make([]Message, 0)
		o.context.Set(token, messages)
	}

	ctx := r.Context()
	req := Request{
		Model:    o.Model,
		Stream:   true,
		Messages: messages,
	}

	if err := o.streamFromOllama(ctx, token, o.URL, req, w, flusher); err != nil {
		fmt.Printf("Error: %v\n", err)
		return
	}

	fmt.Fprintf(w, "data: [DONE]\n\n")
	flusher.Flush()
}

func (o *Ollama) GetAnnotation(ctx context.Context, strategyItem *strategyapi.CreateStrategyRequest) (*strategyapi.GetAnnotationReply, error) {
	return o.getAnnotation(ctx, strategyItem)
}

func (o *Ollama) getAnnotation(ctx context.Context, strategyItem *strategyapi.CreateStrategyRequest) (*strategyapi.GetAnnotationReply, error) {
	formatValue := `
	map[string]any{
		"value":  endPointValue.Value,
		"time":   endPointValue.Timestamp,
		"labels": labels.Map(),
		"ext":    extJSON,
	}`
	messages := []Message{
		{
			Role:    "user",
			Content: strategyItem.String(),
		},
		{
			Role:    "user",
			Content: fmt.Sprintf("模板按照：%s 格式， 通过template完成数据填充，我希望这个json格式是满足此数据的go template格式", formatValue),
		},
		{
			Role:    "user",
			Content: "请根据规则生成告警描述和告警总结的go template， 按照json格式返回，json格式如下：{\"description\": \"描述\", \"summary\": \"总结\"}, 回复不要有任何解释, 以及任何的markdown语法， 只需要纯粹的json字符串，我会根据这个回复通过Unmarshal解析出json数据",
		},
	}
	resp, err := o.getTemplate(ctx, messages)
	if err != nil {
		return nil, err
	}

	var reply strategyapi.GetAnnotationReply
	if len(resp.Choices) > 0 {
		if err := types.Unmarshal([]byte(resp.Choices[0].Message.Content), &reply); err != nil {
			return nil, err
		}
	}
	return &reply, nil
}

func (o *Ollama) getTemplate(ctx context.Context, msg []Message) (*OpenAIAPIResponse, error) {
	resp, err := o.responseFromOllama(ctx, msg)
	if err != nil {
		return nil, err
	}

	fmt.Printf("Response: %s\n", resp.String())

	return resp, nil
}

// responseFromOllama 从Ollama获取响应
func (o *Ollama) responseFromOllama(ctx context.Context, msg []Message) (*OpenAIAPIResponse, error) {
	req := Request{
		Model:    o.Model,
		Stream:   false,
		Messages: msg,
	}

	js, err := types.Marshal(&req)
	if err != nil {
		return nil, fmt.Errorf("marshal error: %w", err)
	}

	client := nhttp.Client{}
	httpReq, err := nhttp.NewRequestWithContext(ctx, nhttp.MethodPost, o.URL, bytes.NewReader(js))
	if err != nil {
		return nil, fmt.Errorf("create request error: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+o.Auth)

	httpResp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("do request error: %w", err)
	}
	defer httpResp.Body.Close()

	body, err := io.ReadAll(httpResp.Body)
	if err != nil {
		return nil, fmt.Errorf("read body error: %w", err)
	}

	var resp OpenAIAPIResponse
	if err := types.Unmarshal(body, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

func (o *Ollama) streamFromOllama(ctx context.Context, token string, url string, ollamaReq Request, w http.ResponseWriter, flusher http.Flusher) error {
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
	httpReq.Header.Set("Authorization", "Bearer "+o.Auth)

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
	resp := ""
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
		switch o.Type {
		case "openai":
			var response OpenAIAPIResponseByStream
			line = strings.TrimPrefix(line, "data:")
			if err := types.Unmarshal([]byte(line), &response); err != nil {
				fmt.Printf("Unmarshal error: %v, data: %v\n", err, line)
				continue
			}

			for _, choice := range response.Choices {
				if choice.Delta.Content != "" {
					resp += choice.Delta.Content
					fmt.Fprintf(w, "data: %s\n\n", choice.Delta.Content)
					flusher.Flush()
				}

				if choice.FinishReason == "stop" {
					break
				}
				continue
			}
		default:
			var response OllamaResponse
			if err := types.Unmarshal([]byte(line), &response); err != nil {
				fmt.Printf("Unmarshal error: %v\n", err)
				continue
			}

			safeMessage := strings.ReplaceAll(response.Message.Content, "\n", "\\n")
			if response.Message.Content != "" {
				resp += safeMessage
				fmt.Fprintf(w, "data: %s\n\n", safeMessage)
				flusher.Flush()
			}

			if response.Done {
				break
			}
		}
	}
	if token != "" {
		msg, ok := o.context.Get(token)
		if !ok {
			return nil
		}
		msg = append(msg, Message{
			Role:    "assistant",
			Content: resp,
		})
		o.context.Set(token, msg)
	}
	return nil
}
