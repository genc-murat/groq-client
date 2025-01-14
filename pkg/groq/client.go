package groq

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"
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

// GetCacheKey returns a string that can be used as a cache key for the message content.
// For string content, returns the string directly.
// For multimodal content (array of ContentType), concatenates all text contents with spaces.
// For other content types, returns the string representation of the content.
func (m ChatMessage) GetCacheKey() string {
	switch content := m.Content.(type) {
	case string:
		return content
	case []ContentType:
		// For vision/multimodal messages, concat text contents
		var texts []string
		for _, c := range content {
			if c.Type == "text" {
				texts = append(texts, c.Text)
			}
		}
		return strings.Join(texts, " ")
	default:
		return fmt.Sprintf("%v", content)
	}
}

// CreateChatCompletion sends a chat completion request to the Groq API.
// It takes a context and a ChatCompletionRequest as input.
// The function first validates the request, then checks if a cached response exists.
// If no cache hit occurs, it makes an HTTP POST request to the chat completions endpoint.
// The response is cached (if caching is enabled) before being returned.
//
// Parameters:
//   - ctx: Context for the request, used for timeouts and cancellation
//   - req: Pointer to ChatCompletionRequest containing the chat messages and parameters
//
// Returns:
//   - *ChatCompletionResponse: Contains the API's response including generated message
//   - error: Non-nil if request validation fails, API request fails, or other errors occur
func (c *Client) CreateChatCompletion(ctx context.Context, req *ChatCompletionRequest) (*ChatCompletionResponse, error) {
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrInvalidRequest, err)
	}

	lastMsg := req.Messages[len(req.Messages)-1]
	cacheKey := lastMsg.GetCacheKey()

	if c.cache != nil {
		if resp, found := c.cache.Get(ctx, cacheKey); found {
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
		_ = c.cache.Set(ctx, cacheKey, &result)
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

// CreateTranscription sends an audio file to be transcribed into text using the specified model.
// If no model is specified, it defaults to Whisper Large v3.
//
// The audio file must be in one of the supported formats: flac, mp3, mp4, mpeg, mpga, m4a, ogg, wav, or webm.
//
// Parameters:
//   - ctx: The context for the request
//   - req: TranscriptionRequest containing:
//   - File: The audio file to transcribe
//   - FileName: Name of the audio file with extension
//   - Model: (Optional) The model to use for transcription
//   - Language: (Optional) The language of the audio
//   - Prompt: (Optional) Text to guide the model's transcription
//   - ResponseFormat: (Optional) The format of the transcription response
//   - Temperature: (Optional) Sampling temperature for the model
//
// Returns:
//   - *TranscriptionResponse: Contains the transcribed text and other response data
//   - error: Any error that occurred during the request
func (c *Client) CreateTranscription(ctx context.Context, req *TranscriptionRequest) (*TranscriptionResponse, error) {
	if req.Model == "" {
		req.Model = ModelWhisperLargeV3
	}

	ext := filepath.Ext(req.FileName)
	if !isValidAudioFormat(ext) {
		return nil, fmt.Errorf("invalid audio format: %s. Supported formats: flac, mp3, mp4, mpeg, mpga, m4a, ogg, wav, webm", ext)
	}

	form := map[string]interface{}{
		"file":     req.File,
		"filename": req.FileName,
		"model":    string(req.Model),
	}

	if req.Language != "" {
		form["language"] = req.Language
	}
	if req.Prompt != "" {
		form["prompt"] = req.Prompt
	}
	if req.ResponseFormat != "" {
		form["response_format"] = req.ResponseFormat
	}
	if req.Temperature != 0 {
		form["temperature"] = fmt.Sprintf("%.2f", req.Temperature)
	}

	var result TranscriptionResponse
	err := c.httpClient.DoMultipartForm(
		ctx,
		"POST",
		fmt.Sprintf("%s/audio/transcriptions", c.baseURL),
		form,
		&result,
	)
	if err != nil {
		return nil, fmt.Errorf("transcription request failed: %w", err)
	}

	return &result, nil
}

// CreateTranslation sends an audio file to be translated into English.
// It accepts a TranslationRequest containing the audio file and optional parameters,
// and returns a TranslationResponse with the translated text.
//
// The audio file must be in one of the supported formats:
// flac, mp3, mp4, mpeg, mpga, m4a, ogg, wav, webm
//
// If no model is specified in the request, it defaults to ModelWhisperLargeV3.
//
// Parameters:
//   - ctx: Context for the request
//   - req: TranslationRequest containing:
//   - File: The audio file to translate
//   - FileName: Name of the audio file including extension
//   - Model: (Optional) The model to use for translation
//   - Prompt: (Optional) Text to guide the model's style or continue a previous audio segment
//   - ResponseFormat: (Optional) The format of the translation response
//   - Temperature: (Optional) Sampling temperature between 0 and 1
//
// Returns:
//   - *TranslationResponse: Contains the translated text and other response data
//   - error: Any error encountered during the translation request
func (c *Client) CreateTranslation(ctx context.Context, req *TranslationRequest) (*TranslationResponse, error) {
	if req.Model == "" {
		req.Model = ModelWhisperLargeV3
	}

	ext := filepath.Ext(req.FileName)
	if !isValidAudioFormat(ext) {
		return nil, fmt.Errorf("invalid audio format: %s. Supported formats: flac, mp3, mp4, mpeg, mpga, m4a, ogg, wav, webm", ext)
	}

	form := map[string]interface{}{
		"file":     req.File,
		"filename": req.FileName,
		"model":    string(req.Model),
	}

	if req.Prompt != "" {
		form["prompt"] = req.Prompt
	}
	if req.ResponseFormat != "" {
		form["response_format"] = req.ResponseFormat
	}
	if req.Temperature != 0 {
		form["temperature"] = fmt.Sprintf("%.2f", req.Temperature)
	}

	var result TranslationResponse
	err := c.httpClient.DoMultipartForm(
		ctx,
		"POST",
		fmt.Sprintf("%s/audio/translations", c.baseURL),
		form,
		&result,
	)
	if err != nil {
		return nil, fmt.Errorf("translation request failed: %w", err)
	}

	return &result, nil
}

// isValidAudioFormat checks if the provided file extension is a supported audio format.
// Returns true if the extension is one of: .flac, .mp3, .mp4, .mpeg, .mpga, .m4a, .ogg, .wav, .webm.
// The extension should include the dot prefix (e.g. ".mp3").
func isValidAudioFormat(ext string) bool {
	validFormats := map[string]bool{
		".flac": true,
		".mp3":  true,
		".mp4":  true,
		".mpeg": true,
		".mpga": true,
		".m4a":  true,
		".ogg":  true,
		".wav":  true,
		".webm": true,
	}
	return validFormats[ext]
}
