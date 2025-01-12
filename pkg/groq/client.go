package groq

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"time"

	"github.com/genc-murat/groq-client/internal/util"
)

const (
	DefaultBaseURL = "https://api.groq.com/openai/v1"
	defaultTimeout = 30 * time.Second
)

type Client struct {
	baseURL    string
	httpClient *util.HTTPClient
	config     *Config
	cache      Cache
}

// NewClient creates a new instance of Client with the provided API key and optional configurations.
// It sets up the HTTP client with default settings and base headers including the Authorization header.
// If the base headers are not set properly, it will panic.
//
// Parameters:
//   - apiKey: The API key used for authorization.
//   - opts: Optional configurations that can be applied to the Client.
//
// Returns:
//   - *Client: A pointer to the newly created Client instance.
func NewClient(apiKey string, opts ...Option) *Client {
	baseHeaders := map[string]string{
		"Authorization": fmt.Sprintf("Bearer %s", apiKey),
		"Content-Type":  "application/json",
	}

	httpConfig := util.HTTPClientConfig{
		MaxRequestTimeout: defaultTimeout,
		RequestsPerSecond: 10,
		MaxRetries:        3,
		RetryWaitTime:     time.Second,
		BaseHeaders:       baseHeaders,
	}

	httpClient := util.NewHTTPClient(httpConfig)

	currentHeaders := httpClient.GetBaseHeaders()
	if len(currentHeaders) == 0 || currentHeaders["Authorization"] == "" {
		panic(fmt.Sprintf("Base headers not set properly. Current headers: %v", currentHeaders))
	}

	c := &Client{
		baseURL:    DefaultBaseURL,
		httpClient: httpClient,
		config:     defaultConfig(),
	}

	for _, opt := range opts {
		opt(c)
	}

	return c
}

// CreateChatCompletion sends a request to create a chat completion based on the provided request.
// It validates the request, checks the cache for a previous response, and if not found, sends the request
// to the server. The response is then cached if caching is enabled.
//
// Parameters:
//   - ctx: The context for the request, used for cancellation and timeouts.
//   - req: The ChatCompletionRequest containing the details of the chat completion request.
//
// Returns:
//   - *ChatCompletionResponse: The response from the server containing the chat completion result.
//   - error: An error if the request validation fails, the request to the server fails, or any other issue occurs.
func (c *Client) CreateChatCompletion(ctx context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidRequest, err)
	}

	if c.cache != nil {
		if resp, found := c.cache.Get(ctx, req.Messages[len(req.Messages)-1].Content); found {
			return resp, nil
		}
	}

	headers := map[string]string{
		"Content-Type": "application/json",
	}

	var result ChatCompletionResponse
	err := c.httpClient.DoJSON(
		ctx,
		"POST",
		fmt.Sprintf("%s/chat/completions", c.baseURL),
		req,
		&result,
		headers,
	)
	if err != nil {
		return nil, fmt.Errorf("chat completion request failed: %w", err)
	}

	if c.cache != nil {
		_ = c.cache.Set(ctx, req.Messages[len(req.Messages)-1].Content, &result)
	}

	return &result, nil
}

// GetCache returns the current cache instance associated with the Client.
// This cache can be used to store and retrieve data to improve performance
// by avoiding redundant operations.
func (c *Client) GetCache() Cache {
	return c.cache
}

// ClearCache clears the client's cache if it is not nil.
// It takes a context.Context as an argument and returns an error if the cache
// clearing operation fails. If the cache is nil, it returns nil.
func (c *Client) ClearCache(ctx context.Context) error {
	if c.cache != nil {
		return c.cache.Clear(ctx)
	}
	return nil
}

// GetCacheStats retrieves the statistics of the cache if it exists.
// It returns a pointer to a CacheStats struct containing the cache statistics,
// or nil if the cache is not initialized.
func (c *Client) GetCacheStats() *CacheStats {
	if c.cache != nil {
		stats := c.cache.GetStats()
		return &stats
	}
	return nil
}

// CreateChatCompletionStream sends a chat completion request to the server and processes the response stream.
// It validates the request, marshals it to JSON, and sends it via an HTTP POST request.
// The response is expected to be a stream of events, which are read and processed line by line.
// Each line is expected to be a JSON-encoded ChatCompletionChunk, which is passed to the provided handler function.
//
// The function returns an error if the request validation fails, if there is an error during the HTTP request,
// if there is an error reading the stream, or if the handler function returns an error.
//
// Parameters:
// - ctx: The context for controlling the request lifetime.
// - req: The chat completion request to be sent.
// - handler: A function to handle each chunk of the chat completion response.
//
// Returns:
// - An error if any step of the process fails, or if the context is canceled.
func (c *Client) CreateChatCompletionStream(ctx context.Context, req *ChatCompletionRequest, handler StreamHandler) error {
	if err := req.Validate(); err != nil {
		return fmt.Errorf("%w: %v", ErrInvalidRequest, err)
	}

	req.Stream = true

	reqBody, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal request: %w", err)
	}

	headers := map[string]string{
		"Accept":       "text/event-stream",
		"Content-Type": "application/json",
	}

	respBody, err := c.httpClient.DoRequest(
		ctx,
		"POST",
		fmt.Sprintf("%s/chat/completions", c.baseURL),
		reqBody,
		headers,
	)
	if err != nil {
		return err
	}

	reader := bufio.NewReader(bytes.NewReader(respBody))

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		line, err := reader.ReadBytes('\n')
		if err != nil {
			if err == io.EOF {
				return nil
			}
			return fmt.Errorf("error reading stream: %v", err)
		}

		line = bytes.TrimSpace(line)
		if len(line) == 0 {
			continue
		}

		if !bytes.HasPrefix(line, []byte("data: ")) {
			continue
		}
		line = bytes.TrimPrefix(line, []byte("data: "))

		if string(line) == "[DONE]" {
			return nil
		}

		var chunk ChatCompletionChunk
		if err := json.Unmarshal(line, &chunk); err != nil {
			return fmt.Errorf("%w: %v", ErrJSONDecoding, err)
		}

		if err := handler(&chunk); err != nil {
			return fmt.Errorf("stream handler error: %v", err)
		}
	}
}
