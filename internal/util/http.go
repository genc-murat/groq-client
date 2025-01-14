package util

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"math/rand"
	"mime/multipart"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
)

var (
	ErrRequestFailed     = errors.New("request failed")
	ErrResponseParsing   = errors.New("response parsing failed")
	ErrRateLimitExceeded = errors.New("rate limit exceeded")
	ErrTimeout           = errors.New("request timeout")
)

type HTTPClient struct {
	client      *fasthttp.Client
	rateLimit   *RateLimiter
	retryConfig *RetryConfig
	baseHeaders map[string]string
	mu          sync.RWMutex
}

type HTTPClientConfig struct {
	MaxRequestTimeout time.Duration
	RequestsPerSecond int
	MaxRetries        int
	RetryWaitTime     time.Duration
	BaseHeaders       map[string]string
}

// NewHTTPClient creates a new instance of HTTPClient with the provided configuration.
// It sets default values for MaxRequestTimeout, RequestsPerSecond, MaxRetries, and RetryWaitTime
// if they are not provided in the config. It also initializes base headers if provided.
//
// Parameters:
//   - config: HTTPClientConfig containing configuration options for the HTTP client.
//
// Returns:
//   - *HTTPClient: A pointer to the newly created HTTPClient instance.
func NewHTTPClient(config HTTPClientConfig) *HTTPClient {
	if config.MaxRequestTimeout == 0 {
		config.MaxRequestTimeout = 30 * time.Second
	}
	if config.RequestsPerSecond == 0 {
		config.RequestsPerSecond = 10
	}
	if config.MaxRetries == 0 {
		config.MaxRetries = 3
	}
	if config.RetryWaitTime == 0 {
		config.RetryWaitTime = time.Second
	}

	baseHeaders := make(map[string]string)
	if config.BaseHeaders != nil {
		for k, v := range config.BaseHeaders {
			baseHeaders[k] = v
		}
	}

	client := &HTTPClient{
		client: &fasthttp.Client{
			ReadTimeout:  config.MaxRequestTimeout,
			WriteTimeout: config.MaxRequestTimeout,
		},
		rateLimit: NewRateLimiter(config.RequestsPerSecond),
		retryConfig: &RetryConfig{
			MaxRetries:    config.MaxRetries,
			RetryWaitTime: config.RetryWaitTime,
		},
		baseHeaders: baseHeaders,
		mu:          sync.RWMutex{},
	}

	fmt.Printf("Base Headers initialized with: %v\n", baseHeaders)

	return client
}

// GetClient returns the underlying fasthttp.Client instance used by the HTTPClient.
// This allows for direct manipulation or configuration of the client if needed.
func (h *HTTPClient) GetClient() *fasthttp.Client {
	return h.client
}

// DoRequest sends an HTTP request with the specified method, URL, body, and headers,
// and returns the response body or an error if the request fails.
//
// Parameters:
//   - ctx: The context to control the request lifetime.
//   - method: The HTTP method to use (e.g., "GET", "POST").
//   - url: The URL to send the request to.
//   - body: The request body as a byte slice.
//   - headers: A map of additional headers to include in the request.
//
// Returns:
//   - A byte slice containing the response body.
//   - An error if the request fails or the response status code is 400 or higher.
//
// The function respects rate limiting and retries the request if necessary.
// It also sets base headers defined in the HTTPClient and additional headers provided in the headers parameter.
func (c *HTTPClient) DoRequest(ctx context.Context, method, url string, body []byte, headers map[string]string) ([]byte, error) {
	if err := c.rateLimit.Wait(ctx); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrRateLimitExceeded, err)
	}

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(url)
	req.Header.SetMethod(method)

	c.mu.RLock()
	fmt.Printf("Setting base headers: %v\n", c.baseHeaders)
	for k, v := range c.baseHeaders {
		req.Header.Set(k, v)
	}
	c.mu.RUnlock()

	if headers != nil {
		fmt.Printf("Setting request headers: %v\n", headers)
		for k, v := range headers {
			req.Header.Set(k, v)
		}
	}

	if len(body) > 0 {
		req.SetBody(body)
	}

	req.Header.VisitAll(func(key, value []byte) {
		fmt.Printf("Final Header - %s: %s\n", string(key), string(value))
	})

	err := c.doRequestWithRetry(ctx, req, resp)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode() >= 400 {
		return nil, fmt.Errorf("%w: status code %d", ErrRequestFailed, resp.StatusCode())
	}

	respBody := make([]byte, len(resp.Body()))
	copy(respBody, resp.Body())

	return respBody, nil
}

// DoJSON sends an HTTP request with a JSON body and decodes the JSON response.
//
// Parameters:
//   - ctx: The context for the request.
//   - method: The HTTP method (e.g., "GET", "POST").
//   - url: The URL to send the request to.
//   - reqBody: The request body to be marshaled to JSON. Can be nil.
//   - respBody: The response body to be unmarshaled from JSON. Can be nil.
//   - headers: Additional headers to include in the request. Can be nil.
//
// Returns:
//   - error: An error if the request fails or the response cannot be parsed.
func (c *HTTPClient) DoJSON(ctx context.Context, method, url string, reqBody interface{}, respBody interface{}, headers map[string]string) error {
	var bodyBytes []byte
	var err error

	if reqBody != nil {
		bodyBytes, err = json.Marshal(reqBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	if headers == nil {
		headers = make(map[string]string)
	}

	headers["Content-Type"] = "application/json"

	respBytes, err := c.DoRequest(ctx, method, url, bodyBytes, headers)
	if err != nil {
		return err
	}

	if respBody != nil {
		if err := json.Unmarshal(respBytes, respBody); err != nil {
			return fmt.Errorf("%w: %v", ErrResponseParsing, err)
		}
	}

	return nil
}

// SetBaseHeaders sets the base headers for the HTTP client.
// It takes a map of headers as input and updates the client's base headers
// with the provided key-value pairs. The method is thread-safe as it locks
// the mutex before updating the headers and unlocks it after the update.
//
// Parameters:
//   - headers: A map[string]string containing the headers to be set.
//
// Example:
//
//	headers := map[string]string{
//	    "Content-Type": "application/json",
//	    "Authorization": "Bearer token",
//	}
//	client.SetBaseHeaders(headers)
func (c *HTTPClient) SetBaseHeaders(headers map[string]string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.baseHeaders = make(map[string]string)

	for k, v := range headers {
		c.baseHeaders[k] = v
	}

	fmt.Printf("Base headers updated to: %v\n", c.baseHeaders)
}

// GetBaseHeaders returns a copy of the base headers of the HTTP client.
// It acquires a read lock to ensure thread-safe access to the baseHeaders map.
func (c *HTTPClient) GetBaseHeaders() map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	headers := make(map[string]string)
	for k, v := range c.baseHeaders {
		headers[k] = v
	}

	return headers
}

// doRequestWithRetry sends an HTTP request and retries it upon failure based on the retry configuration.
// It will retry the request up to MaxRetries times, waiting RetryWaitTime * attempt between each retry.
// If the context is done before the request succeeds, it returns the context's error.
// If the response status code is not retryable, it returns nil.
// If the maximum number of retries is exceeded, it returns an error indicating the last encountered error.
//
// Parameters:
//
//	ctx - the context to control cancellation and timeout
//	req - the HTTP request to be sent
//	resp - the HTTP response to be populated
//
// Returns:
//
//	error - an error if the request fails after the maximum number of retries or if the context is done
func (c *HTTPClient) doRequestWithRetry(ctx context.Context, req *fasthttp.Request, resp *fasthttp.Response) error {
	var lastErr error

	for attempt := 0; attempt <= c.retryConfig.MaxRetries; attempt++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		if attempt > 0 {
			time.Sleep(c.retryConfig.RetryWaitTime * time.Duration(attempt))
		}

		err := c.client.Do(req, resp)
		if err == nil {
			if !isRetryableStatusCode(resp.StatusCode()) {
				return nil
			}
			lastErr = fmt.Errorf("received status code %d", resp.StatusCode())
			continue
		}

		lastErr = err
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

type RateLimiter struct {
	ticker *time.Ticker
	tokens chan struct{}
}

// NewRateLimiter creates a new RateLimiter that allows a specified number of requests per second.
// It initializes a ticker that ticks at intervals based on the requestsPerSecond parameter,
// and a buffered channel to hold the tokens.
//
// Parameters:
//   - requestsPerSecond: The number of requests allowed per second.
//
// Returns:
//   - *RateLimiter: A pointer to the newly created RateLimiter instance.
func NewRateLimiter(requestsPerSecond int) *RateLimiter {
	rl := &RateLimiter{
		ticker: time.NewTicker(time.Second / time.Duration(requestsPerSecond)),
		tokens: make(chan struct{}, requestsPerSecond),
	}

	for i := 0; i < requestsPerSecond; i++ {
		rl.tokens <- struct{}{}
	}

	go rl.refillTokens()

	return rl
}

// Wait blocks until a token is available or the context is done.
// It returns nil if a token is acquired, or an error if the context is done.
//
// Parameters:
//
//	ctx - The context to use for cancellation.
//
// Returns:
//
//	error - nil if a token is acquired, or the context's error if it is done.
func (rl *RateLimiter) Wait(ctx context.Context) error {
	select {
	case <-rl.tokens:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

// refillTokens is a method of RateLimiter that continuously refills the token bucket.
// It listens to a ticker channel and attempts to add a token to the tokens channel
// whenever the ticker ticks. If the tokens channel is full, it discards the token.
func (rl *RateLimiter) refillTokens() {
	for range rl.ticker.C {
		select {
		case rl.tokens <- struct{}{}:
		default:
		}
	}
}

type RetryConfig struct {
	MaxRetries    int
	RetryWaitTime time.Duration
}

// isRetryableStatusCode checks if the given HTTP status code is considered retryable.
// Retryable status codes include:
// - 429 (Too Many Requests)
// - 500 (Internal Server Error)
// - 502 (Bad Gateway)
// - 503 (Service Unavailable)
// - 504 (Gateway Timeout)
//
// Returns true if the status code is retryable, otherwise false.
func isRetryableStatusCode(statusCode int) bool {
	switch statusCode {
	case
		fasthttp.StatusTooManyRequests,
		fasthttp.StatusInternalServerError,
		fasthttp.StatusBadGateway,
		fasthttp.StatusServiceUnavailable,
		fasthttp.StatusGatewayTimeout:
		return true
	default:
		return false
	}
}

func (c *HTTPClient) DoMultipartForm(ctx context.Context, method, url string, form map[string]interface{}, respBody interface{}) error {
	if err := c.rateLimit.Wait(ctx); err != nil {
		return fmt.Errorf("%w: %v", ErrRateLimitExceeded, err)
	}

	var buf bytes.Buffer
	writer := multipart.NewWriter(&buf)

	for key, value := range form {
		if key == "file" {
			if reader, ok := value.(io.Reader); ok {
				if fileName, ok := form["filename"].(string); ok {
					part, err := writer.CreateFormFile("file", fileName)
					if err != nil {
						return fmt.Errorf("error creating form file: %w", err)
					}
					if _, err := io.Copy(part, reader); err != nil {
						return fmt.Errorf("error copying file data: %w", err)
					}
				}
			}
		} else if key != "filename" {
			switch v := value.(type) {
			case []string:
				for _, item := range v {
					if err := writer.WriteField(key, item); err != nil {
						return fmt.Errorf("error writing array field: %w", err)
					}
				}
			default:
				if err := writer.WriteField(key, fmt.Sprintf("%v", v)); err != nil {
					return fmt.Errorf("error writing field: %w", err)
				}
			}
		}
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("error closing multipart writer: %w", err)
	}

	req := fasthttp.AcquireRequest()
	resp := fasthttp.AcquireResponse()
	defer fasthttp.ReleaseRequest(req)
	defer fasthttp.ReleaseResponse(resp)

	req.SetRequestURI(url)
	req.Header.SetMethod(method)
	req.SetBody(buf.Bytes())
	req.Header.SetContentType(writer.FormDataContentType())

	c.mu.RLock()
	for k, v := range c.baseHeaders {
		if k != "Content-Type" {
			req.Header.Set(k, v)
		}
	}
	c.mu.RUnlock()

	err := c.doRequestWithRetry(ctx, req, resp)
	if err != nil {
		return err
	}

	if resp.StatusCode() >= 400 {
		bodyStr := string(resp.Body())
		return fmt.Errorf("%w: status code %d, body: %s", ErrRequestFailed, resp.StatusCode(), bodyStr)
	}

	if respBody != nil {
		if err := json.Unmarshal(resp.Body(), respBody); err != nil {
			return fmt.Errorf("%w: %v", ErrResponseParsing, err)
		}
	}

	return nil
}

func generateBoundary() string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, 30)
	for i := range b {
		b[i] = charset[rand.Intn(len(charset))]
	}
	return string(b)
}
