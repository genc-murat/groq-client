package util

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestNewHTTPClient_DefaultConfig(t *testing.T) {
	config := HTTPClientConfig{}
	client := NewHTTPClient(config)

	assert.NotNil(t, client)
	assert.Equal(t, 30*time.Second, client.client.ReadTimeout)
	assert.Equal(t, 30*time.Second, client.client.WriteTimeout)
	assert.Equal(t, 10, cap(client.rateLimit.tokens))
	assert.Equal(t, 3, client.retryConfig.MaxRetries)
	assert.Equal(t, time.Second, client.retryConfig.RetryWaitTime)
	assert.Empty(t, client.baseHeaders)
}

func TestNewHTTPClient_CustomConfig(t *testing.T) {
	config := HTTPClientConfig{
		MaxRequestTimeout: 15 * time.Second,
		RequestsPerSecond: 20,
		MaxRetries:        5,
		RetryWaitTime:     2 * time.Second,
		BaseHeaders: map[string]string{
			"Authorization": "Bearer token",
		},
	}
	client := NewHTTPClient(config)

	assert.NotNil(t, client)
	assert.Equal(t, 15*time.Second, client.client.ReadTimeout)
	assert.Equal(t, 15*time.Second, client.client.WriteTimeout)
	assert.Equal(t, 20, cap(client.rateLimit.tokens))
	assert.Equal(t, 5, client.retryConfig.MaxRetries)
	assert.Equal(t, 2*time.Second, client.retryConfig.RetryWaitTime)
	assert.Equal(t, map[string]string{"Authorization": "Bearer token"}, client.baseHeaders)
}

func TestNewHTTPClient_BaseHeadersNil(t *testing.T) {
	config := HTTPClientConfig{
		BaseHeaders: nil,
	}
	client := NewHTTPClient(config)

	assert.NotNil(t, client)
	assert.Empty(t, client.baseHeaders)
}

func TestHTTPClient_GetClient(t *testing.T) {
	config := HTTPClientConfig{}
	client := NewHTTPClient(config)

	fastHTTPClient := client.GetClient()

	assert.NotNil(t, fastHTTPClient)
	assert.Equal(t, client.client, fastHTTPClient)
}
