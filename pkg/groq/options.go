package groq

import (
	"time"

	"github.com/genc-murat/groq-client/internal/util"
)

type Option func(*Client)

// WithBaseURL sets the base URL for the client.
//
// Parameters:
//   - baseURL: The base URL to be set for the client.
//
// Returns:
//   - Option: A function that sets the base URL for the client.
func WithBaseURL(baseURL string) Option {
	return func(c *Client) {
		c.baseURL = baseURL
	}
}

// WithHTTPConfig returns an Option that configures the HTTP client of the Client
// with the provided HTTPClientConfig. It merges any existing base headers from
// the current HTTP client into the new configuration before creating a new
// HTTP client.
//
// Parameters:
//   - config: The HTTPClientConfig to use for configuring the HTTP client.
//
// Returns:
//   - Option: A function that applies the provided HTTPClientConfig to the Client.
func WithHTTPConfig(config util.HTTPClientConfig) Option {
	return func(c *Client) {
		currentHeaders := c.httpClient.GetBaseHeaders()
		if len(currentHeaders) > 0 {
			if config.BaseHeaders == nil {
				config.BaseHeaders = make(map[string]string)
			}
			for k, v := range currentHeaders {
				config.BaseHeaders[k] = v
			}
		}

		c.httpClient = util.NewHTTPClient(config)
	}
}

// WithTimeout returns an Option that sets the maximum request timeout for the HTTP client.
// The timeout parameter specifies the duration to wait before timing out a request.
// This function updates the client's HTTP client configuration with the provided timeout value.
//
// Parameters:
//   - timeout: The maximum duration to wait before timing out a request.
//
// Returns:
//   - Option: A function that modifies the client's HTTP client configuration.
func WithTimeout(timeout time.Duration) Option {
	return func(c *Client) {
		currentHeaders := c.httpClient.GetBaseHeaders()

		config := util.HTTPClientConfig{
			MaxRequestTimeout: timeout,
			RequestsPerSecond: c.config.RateLimit.RequestsPerMinute,
			MaxRetries:        c.config.RetryConfig.MaxRetries,
			RetryWaitTime:     c.config.RetryConfig.RetryDelay,
			BaseHeaders:       currentHeaders,
		}

		c.httpClient = util.NewHTTPClient(config)
	}
}

// WithRetryConfig sets the retry configuration for the client, including the maximum number of retries
// and the wait time between retries. It also updates the HTTP client configuration with the new retry settings.
//
// Parameters:
//   - maxRetries: The maximum number of retry attempts.
//   - retryWaitTime: The duration to wait between retry attempts.
//
// Returns:
//   - Option: A function that applies the retry configuration to the client.
func WithRetryConfig(maxRetries int, retryWaitTime time.Duration) Option {
	return func(c *Client) {
		c.config.RetryConfig.MaxRetries = maxRetries
		c.config.RetryConfig.RetryDelay = retryWaitTime

		currentHeaders := c.httpClient.GetBaseHeaders()
		config := util.HTTPClientConfig{
			MaxRequestTimeout: c.httpClient.GetClient().ReadTimeout,
			RequestsPerSecond: c.config.RateLimit.RequestsPerMinute,
			MaxRetries:        maxRetries,
			RetryWaitTime:     retryWaitTime,
			BaseHeaders:       currentHeaders,
		}

		c.httpClient = util.NewHTTPClient(config)
	}
}

// WithRateLimit sets the rate limit for the client in requests per minute.
// It updates the client's configuration to enable rate limiting and adjusts
// the HTTP client settings accordingly.
//
// Parameters:
//   - requestsPerMinute: The number of requests allowed per minute.
//
// Returns:
//   - Option: A function that modifies the client's configuration to apply the rate limit.
func WithRateLimit(requestsPerMinute int) Option {
	return func(c *Client) {
		c.config.RateLimit.RequestsPerMinute = requestsPerMinute
		c.config.RateLimit.Enabled = true

		currentHeaders := c.httpClient.GetBaseHeaders()
		config := util.HTTPClientConfig{
			MaxRequestTimeout: c.httpClient.GetClient().ReadTimeout,
			RequestsPerSecond: requestsPerMinute,
			MaxRetries:        c.config.RetryConfig.MaxRetries,
			RetryWaitTime:     c.config.RetryConfig.RetryDelay,
			BaseHeaders:       currentHeaders,
		}

		c.httpClient = util.NewHTTPClient(config)
	}
}

// WithBaseHeaders returns an Option that sets the base headers for the HTTP client.
// It takes a map of headers as input and merges them with the existing base headers
// of the client's HTTP client.
//
// headers: A map where the key is the header name and the value is the header value.
//
// Example usage:
//
//	client := NewClient()
//	client.ApplyOptions(WithBaseHeaders(map[string]string{"Authorization": "Bearer token"}))
func WithBaseHeaders(headers map[string]string) Option {
	return func(c *Client) {
		currentHeaders := c.httpClient.GetBaseHeaders()
		for k, v := range headers {
			currentHeaders[k] = v
		}

		c.httpClient.SetBaseHeaders(currentHeaders)
	}
}
